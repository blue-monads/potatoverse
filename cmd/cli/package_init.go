package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/term"
)

type TemplateInfo struct {
	Name        string
	Description string
	RepoUrl     string
	Slug        string
}

var TEMPLATES = []TemplateInfo{
	{
		Name:        "React Lua",
		Description: "React frontend with Lua backend handlers",
		RepoUrl:     "https://github.com/blue-monads/potato-app-react-lua-template",
		Slug:        "potato-app-react-lua-template",
	},
	{
		Name:        "Vanilla Lua",
		Description: "Vanilla JavaScript with Lua backend handlers",
		RepoUrl:     "https://github.com/blue-monads/potato-app-vannilla-lua-template",
		Slug:        "potato-app-vannilla-lua-template",
	},
}

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

type PackageInitCmd struct {
	Template  string `name:"template" help:"Template slug or name to use. If omitted, choose interactively."`
	Slug      string `name:"slug" help:"Project slug (letters, digits and hyphens only). If omitted, prompt interactively."`
	Directory string `name:"directory" short:"d" help:"Destination directory for new project." type:"path" default:"."`
}

func (c *PackageInitCmd) Run(_ *kong.Context) error {
	templateInfo, err := c.chooseTemplate()
	if err != nil {
		return err
	}

	slug, err := c.chooseSlug()
	if err != nil {
		return err
	}

	destination := filepath.Join(c.Directory, slug)
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	}

	cleanDest := filepath.Clean(destination)
	fmt.Println()
	fmt.Printf("📦 Initializing project %s...\n", cyan(slug))
	fmt.Printf("  • Cloning template %s...\n", templateInfo.Name)

	if err := cloneTemplateRepo(templateInfo.RepoUrl, destination); err != nil {
		return err
	}

	fmt.Println("  • Configuring project files...")
	if err := replaceTemplateSlug(destination, templateInfo.Slug, slug); err != nil {
		return err
	}

	if err := resetGitState(destination); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("✨ Project created successfully at %s\n\n", bold(cleanDest))
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n\n", cleanDest)

	return nil
}

func (c *PackageInitCmd) chooseTemplate() (TemplateInfo, error) {
	if c.Template != "" {
		for _, t := range TEMPLATES {
			if strings.EqualFold(t.Slug, c.Template) || strings.EqualFold(t.Name, c.Template) {
				fmt.Printf("%s %s %s\n", symbolSuccess(), bold("Select a template:"), cyan(t.Name))
				return t, nil
			}
		}
		return TemplateInfo{}, fmt.Errorf("template %q not found", c.Template)
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return TemplateInfo{}, errors.New("interactive template selection requires a TTY or use --template")
	}

	return interactiveTemplateSelect()
}

func (c *PackageInitCmd) chooseSlug() (string, error) {
	if c.Slug != "" {
		if !slugPattern.MatchString(c.Slug) {
			return "", fmt.Errorf("invalid slug %q: only letters, digits, and hyphens are allowed", c.Slug)
		}
		fmt.Printf("%s %s %s\n", symbolSuccess(), bold("Project slug:"), cyan(c.Slug))
		return c.Slug, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New("interactive project slug prompt requires a TTY or use --slug")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s %s ", symbolQuestion(), bold("Project slug:"))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		slug := strings.TrimSpace(input)
		if slug == "" {
			fmt.Printf("  %s Slug is required. Use letters, digits and hyphens only.\n", symbolError())
			continue
		}
		if !slugPattern.MatchString(slug) {
			fmt.Printf("  %s Invalid slug %q. Only letters, digits and hyphens are allowed.\n", symbolError(), slug)
			continue
		}
		fmt.Printf("\x1b[1A\x1b[2K\r%s %s %s\n", symbolSuccess(), bold("Project slug:"), cyan(slug))
		return slug, nil
	}
}

func interactiveTemplateSelect() (TemplateInfo, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return TemplateInfo{}, err
	}
	defer func() {
		_ = term.Restore(fd, oldState)
		fmt.Print("\x1b[?25h\r")
	}()

	fmt.Print("\x1b[?25l") // hide cursor

	selection := 0
	numLines := len(TEMPLATES) + 3 // Header + Templates + Empty + Hint
	renderTemplateMenu(selection, true, numLines)

	for {
		var buf [1]byte
		if _, err := os.Stdin.Read(buf[:]); err != nil {
			return TemplateInfo{}, err
		}
		b := buf[0]

		if b == '\r' || b == '\n' {
			clearAndConfirmTemplateMenu(selection, numLines)
			return TEMPLATES[selection], nil
		}

		if b == '\x03' || b == 'q' || b == 'Q' {
			clearMenu(numLines)
			return TemplateInfo{}, errors.New("aborted")
		}

		if b == 'k' || b == 'K' {
			selection--
			if selection < 0 {
				selection = len(TEMPLATES) - 1
			}
			renderTemplateMenu(selection, false, numLines)
			continue
		}
		if b == 'j' || b == 'J' {
			selection++
			if selection >= len(TEMPLATES) {
				selection = 0
			}
			renderTemplateMenu(selection, false, numLines)
			continue
		}

		if b != '\x1b' {
			continue
		}

		if _, err := os.Stdin.Read(buf[:]); err != nil {
			return TemplateInfo{}, err
		}
		second := buf[0]
		if second != '[' && second != 'O' {
			continue
		}

		if _, err := os.Stdin.Read(buf[:]); err != nil {
			return TemplateInfo{}, err
		}
		third := buf[0]

		if third == 'A' {
			selection--
			if selection < 0 {
				selection = len(TEMPLATES) - 1
			}
			renderTemplateMenu(selection, false, numLines)
			continue
		}
		if third == 'B' {
			selection++
			if selection >= len(TEMPLATES) {
				selection = 0
			}
			renderTemplateMenu(selection, false, numLines)
			continue
		}
	}
}

func renderTemplateMenu(selection int, isFirst bool, numLines int) {
	if !isFirst {
		fmt.Printf("\x1b[%dA\r", numLines)
	}

	fmt.Printf("\x1b[2K\r%s %s\r\n", symbolQuestion(), bold("Select a template:"))
	for idx, template := range TEMPLATES {
		if idx == selection {
			name := boldCyan(fmt.Sprintf("%-13s", template.Name))
			desc := dim(template.Description)
			fmt.Printf("\x1b[2K\r  %s %s  %s\r\n", symbolPointer(), name, desc)
		} else {
			name := fmt.Sprintf("%-13s", template.Name)
			desc := dim(template.Description)
			fmt.Printf("\x1b[2K\r    %s  %s\r\n", name, desc)
		}
	}
	fmt.Print("\x1b[2K\r\r\n")
	fmt.Printf("\x1b[2K\r  %s\r\n", dim("(Use ↑/↓ or j/k to navigate, Enter to select)"))
}

func clearAndConfirmTemplateMenu(selection int, numLines int) {
	fmt.Printf("\x1b[%dA\r", numLines)
	fmt.Printf("\x1b[2K\r%s %s %s\r\n", symbolSuccess(), bold("Select a template:"), cyan(TEMPLATES[selection].Name))
	for i := 1; i < numLines; i++ {
		fmt.Print("\x1b[2K\r\n")
	}
	fmt.Printf("\x1b[%dA\r", numLines-1)
}

func clearMenu(numLines int) {
	fmt.Printf("\x1b[%dA\r", numLines)
	for i := 0; i < numLines; i++ {
		fmt.Print("\x1b[2K\r\n")
	}
	fmt.Printf("\x1b[%dA\r", numLines)
}

func cloneTemplateRepo(repoUrl, destination string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", repoUrl, destination)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone template: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func replaceTemplateSlug(root, oldSlug, newSlug string) error {
	oldBytes := []byte(oldSlug)
	newBytes := []byte(newSlug)
	pathsToRename := make([]string, 0)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}

		if !d.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if bytes.Contains(data, oldBytes) {
				info, err := os.Stat(path)
				if err != nil {
					return err
				}
				newData := bytes.ReplaceAll(data, oldBytes, newBytes)
				if err := os.WriteFile(path, newData, info.Mode()); err != nil {
					return err
				}
			}
		}

		if strings.Contains(d.Name(), oldSlug) {
			pathsToRename = append(pathsToRename, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(pathsToRename, func(i, j int) bool {
		return len(pathsToRename[i]) > len(pathsToRename[j])
	})

	for _, oldPath := range pathsToRename {
		base := filepath.Base(oldPath)
		newBase := strings.ReplaceAll(base, oldSlug, newSlug)
		newPath := filepath.Join(filepath.Dir(oldPath), newBase)
		if oldPath == newPath {
			continue
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return nil
}

func isParentGitRepo(targetDir string) bool {
	parentDir := filepath.Dir(targetDir)
	absParent, err := filepath.Abs(parentDir)
	if err != nil {
		absParent = parentDir
	}

	// 1. Check if git considers absParent to be inside a git work tree
	cmd := exec.Command("git", "-C", absParent, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "true" {
		return true
	}

	// 2. Also check if parent itself is a git dir (e.g. bare repo)
	cmd = exec.Command("git", "-C", absParent, "rev-parse", "--git-dir")
	if err := cmd.Run(); err == nil {
		return true
	}

	// 3. Fallback: walk up from absParent checking for .git directory or file
	curr := absParent
	for {
		gitPath := filepath.Join(curr, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return true
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return false
}

func resetGitState(root string) error {
	gitDir := filepath.Join(root, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("failed to remove template .git directory: %w", err)
	}

	if isParentGitRepo(root) {
		fmt.Println("  • Skipping git init (parent directory is already a git repository)")
		return nil
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize git repository: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	fmt.Println("  • Initialized empty git repository")

	return nil
}

// Styling and terminal UI helpers

func isColorEnabled() bool {
	return os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
}

func colorWrap(code, s string) string {
	if !isColorEnabled() {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func cyan(s string) string     { return colorWrap("36", s) }
func green(s string) string    { return colorWrap("32", s) }
func red(s string) string      { return colorWrap("31", s) }
func dim(s string) string      { return colorWrap("2", s) }
func bold(s string) string     { return colorWrap("1", s) }
func boldCyan(s string) string { return colorWrap("1;36", s) }

func symbolQuestion() string {
	if !isColorEnabled() {
		return "?"
	}
	return cyan("?")
}

func symbolSuccess() string {
	if !isColorEnabled() {
		return "✔"
	}
	return green("✔")
}

func symbolError() string {
	if !isColorEnabled() {
		return "✖"
	}
	return red("✖")
}

func symbolPointer() string {
	if !isColorEnabled() {
		return ">"
	}
	return cyan("❯")
}
