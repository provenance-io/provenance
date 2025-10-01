package errors

import "fmt"

// ExitCodeError contains the exit code for cmd exit, and satisfies the error interface.
type ExitCodeError int

var _ error = (*ExitCodeError)(nil)

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit code: %d", e)
}
