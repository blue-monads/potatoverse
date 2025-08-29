package libx

import (
	"fmt"
)

func PanicWrapper(wrapped func()) (err error) {

	defer func() {

		if cause := recover(); cause != nil {
			err = fmt.Errorf("%v", cause)
		}
	}()
	wrapped()
	return err
}
