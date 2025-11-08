package libx

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/blue-monads/turnix/backend/utils/qq"
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

			qq.Println(causeStr)

			qq.Println("@stack", stackStr)

		}
	}()
	wrapped()
	return err
}
