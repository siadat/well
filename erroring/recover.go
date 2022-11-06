package erroring

import (
	"fmt"
)

func CallAndRecover[E error, T any](f func() T) (result T, retErr error) {
	defer func() {
		var err = recover()
		switch err := err.(type) {
		case nil:
			return
		case E:
			retErr = err
		default:
			retErr = fmt.Errorf("unexpected error of type %T: %s", err, err)
			PrintTrace()
		}
	}()
	result = f()
	return
}
