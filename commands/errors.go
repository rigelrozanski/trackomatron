//nolint
package commands

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	// ErrBadHexID is an error for improperly formatted hex strings
	ErrBadHexID = fmt.Errorf("HexID is not formatted correctly, must start with 0x")
)

func ErrCmdReqArg(arg string) error {
	return fmt.Errorf("command requires an argument ([%v])", arg) //never stack trace
}

func ErrBadQuery(arg string) error {
	return errors.Errorf("invalid query %v", arg)
}
