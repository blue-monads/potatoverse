package libx

import (
	"errors"
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

			stackStr := strings.Split(string(stack), "\n")
			causeStr := fmt.Sprintf("%v", cause)

			// Print the error and the stack trace
			err = &PanicError{
				Err:   errors.New(causeStr),
				Stack: stackStr,
			}

			pp.Println(causeStr)

			pp.Println("@stack", stackStr)

		}
	}()
	wrapped()
	return err
}
