package commands

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	errBadHexID = fmt.Errorf("HexID is not formatted correctly, must start with 0x")
)

func errCmdReqArg(arg string) error {
	return fmt.Errorf("command requires an argument ([%v])", arg) //never stack trace
}

func errBadQuery(arg string) error {
	return errors.Errorf("invalid query %v", arg)
}
