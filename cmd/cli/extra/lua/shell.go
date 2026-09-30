package luaextra

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cjoudrey/gluahttp"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	"golang.org/x/term"
	luaJson "layeh.com/gopher-json"
)

var defaultHttpClient = &http.Client{}

// NewShellState creates a Lua state initialized with standard libraries and extras (json, phttp).
func NewShellState() *lua.LState {
	L := lua.NewState()

	// Preload extra modules
	L.PreloadModule("json", luaJson.Loader)
	L.PreloadModule("phttp", gluahttp.NewHttpModule(defaultHttpClient).Loader)

	// Set helpful globals
	L.SetGlobal("help", L.NewFunction(func(ls *lua.LState) int {
		helpText := `Potatoverse Lua Shell
Available modules:
  - Standard: string, table, math, io, os, debug, coroutine, package
  - Preloaded: json, phttp (use require("json"), require("phttp"))

Commands:
  - Type expressions (e.g. 1 + 2, math.sqrt(16)) to evaluate directly
  - Type statements (e.g. x = 10, function foo() ... end)
  - Type "exit" or press Ctrl+D to quit
`
		ls.Push(lua.LString(helpText))
		return 1
	}))

	return L
}

func isIncomplete(err error) bool {
	if lerr, ok := err.(*lua.ApiError); ok {
		if perr, ok := lerr.Cause.(*parse.Error); ok {
			return perr.Pos.Line == parse.EOF
		}
	}
	return false
}

func formatLuaValue(val lua.LValue) string {
	if val == nil || val == lua.LNil {
		return "nil"
	}
	switch v := val.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return v.String()
	case lua.LBool:
		return fmt.Sprintf("%t", bool(v))
	case *lua.LTable:
		return fmt.Sprintf("table: %p", v)
	default:
		return val.String()
	}
}

// RunShell starts an interactive Lua REPL session.
func RunShell() error {
	L := NewShellState()
	defer L.Close()

	isTerm := term.IsTerminal(int(os.Stdin.Fd()))

	if isTerm {
		return runInteractiveShell(L)
	}
	return runPipedShell(L)
}

func runInteractiveShell(L *lua.LState) error {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	terminal := term.NewTerminal(struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}, "> ")

	// Redirect Lua print output to write with CRLF
	L.SetGlobal("print", L.NewFunction(func(ls *lua.LState) int {
		top := ls.GetTop()
		var parts []string
		for i := 1; i <= top; i++ {
			parts = append(parts, ls.ToStringMeta(ls.Get(i)).String())
		}
		_, _ = terminal.Write([]byte(strings.Join(parts, "\t") + "\r\n"))
		return 0
	}))

	banner := "Potatoverse Lua Shell (" + lua.PackageName + " " + lua.PackageVersion + ", " + lua.LuaVersion + ")\r\n" +
		"Type \"exit\" or press Ctrl+D to quit, \"help()\" for information.\r\n"
	_, _ = terminal.Write([]byte(banner))

	for {
		terminal.SetPrompt("> ")
		line, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				_, _ = terminal.Write([]byte("\r\nGoodbye!\r\n"))
				return nil
			}
			return err
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed == "exit" || trimmed == "quit" || trimmed == "exit()" || trimmed == "quit()" {
			_, _ = terminal.Write([]byte("Goodbye!\r\n"))
			return nil
		}

		executeLine(L, line, func(prompt string) (string, error) {
			terminal.SetPrompt(prompt)
			return terminal.ReadLine()
		}, func(out string) {
			_, _ = terminal.Write([]byte(out + "\r\n"))
		})
	}
}

func runPipedShell(L *lua.LState) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed == "exit" || trimmed == "quit" || trimmed == "exit()" || trimmed == "quit()" {
			break
		}

		executeLine(L, line, func(prompt string) (string, error) {
			if scanner.Scan() {
				return scanner.Text(), nil
			}
			return "", io.EOF
		}, func(out string) {
			fmt.Println(out)
		})
	}

	return scanner.Err()
}

func executeLine(L *lua.LState, initialLine string, readNext func(string) (string, error), output func(string)) {
	code := initialLine

	// Check for = expression shortcut
	if strings.HasPrefix(code, "=") {
		expr := strings.TrimSpace(code[1:])
		evalExpression(L, expr, output)
		return
	}

	// Try evaluating as expression first (e.g. 1+1, x, math.sin(0))
	if _, err := L.LoadString("return " + code); err == nil {
		evalExpression(L, code, output)
		return
	}

	// Try loading as statement/chunk
	for {
		_, err := L.LoadString(code)
		if err == nil {
			// Chunk is complete and valid, execute it
			runChunk(L, code, output)
			return
		}

		if !isIncomplete(err) {
			// Syntax error, not an incomplete block
			output(err.Error())
			return
		}

		// Incomplete block, ask for continuation
		nextLine, readErr := readNext(">> ")
		if readErr != nil {
			output(err.Error())
			return
		}
		code = code + "\n" + nextLine
	}
}

func evalExpression(L *lua.LState, expr string, output func(string)) {
	oldTop := L.GetTop()
	err := L.DoString("return " + expr)
	if err != nil {
		output(err.Error())
		return
	}

	newTop := L.GetTop()
	if newTop > oldTop {
		var results []string
		for i := oldTop + 1; i <= newTop; i++ {
			results = append(results, formatLuaValue(L.Get(i)))
		}
		output(strings.Join(results, "\t"))
	}
	L.SetTop(oldTop)
}

func runChunk(L *lua.LState, code string, output func(string)) {
	oldTop := L.GetTop()
	err := L.DoString(code)
	if err != nil {
		output(err.Error())
		return
	}

	newTop := L.GetTop()
	if newTop > oldTop {
		var results []string
		for i := oldTop + 1; i <= newTop; i++ {
			results = append(results, formatLuaValue(L.Get(i)))
		}
		output(strings.Join(results, "\t"))
	}
	L.SetTop(oldTop)
}
