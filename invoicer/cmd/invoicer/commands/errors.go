package commands

import (
	"fmt"
)

var (
	errBadHexID = fmt.Errorf("HexID is not formatted correctly")
)

func errCmdReqArg(arg string) error {
	return fmt.Errorf("command requires an argument ([%v])", arg) //never stack trace
}
