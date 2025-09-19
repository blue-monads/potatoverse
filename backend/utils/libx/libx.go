package libx

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/k0kubun/pp"
)

type PanicError struct {
	Err   error
	Stack []string
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("panic: %v\n\nstack trace:\n%s", p.Err, strings.Join(p.Stack, "\n"))
}

func (p *PanicError) Unwrap() error {
	return p.Err
}

func PanicWrapper(wrapped func()) (err error) {
	defer func() {
		if cause := recover(); cause != nil {
			// Get the stack trace
			stack := debug.Stack()

			// Print the error and the stack trace
			err = &PanicError{
				Err:   fmt.Errorf("panic: %v", cause),
				Stack: strings.Split(string(stack), "\n"),
			}

			pp.Println("@stack", stack)

		}
	}()
	wrapped()
	return err
}
