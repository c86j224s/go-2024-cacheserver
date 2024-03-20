package awesomeerror

import (
	"errors"
	"fmt"
)

func New(args ...any) error {
	return errors.New(fmt.Sprint(args...))
}
