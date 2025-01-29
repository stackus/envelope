package envelope

import (
	"fmt"
)

type (
	ErrUnregisteredKey             string
	ErrReregisteredKey             string
	ErrFactoryReturnsNil           string
	ErrFactoryDoesNotReturnPointer string
)

func (e ErrUnregisteredKey) Error() string {
	return fmt.Sprintf("nothing has been registered for %q", string(e))
}

func (e ErrReregisteredKey) Error() string {
	return fmt.Sprintf("something has already been registered for %q", string(e))
}

func (e ErrFactoryReturnsNil) Error() string {
	return fmt.Sprintf("factory for %q returned nil", string(e))
}

func (e ErrFactoryDoesNotReturnPointer) Error() string {
	return fmt.Sprintf("factory for %q did not return a pointer", string(e))
}
