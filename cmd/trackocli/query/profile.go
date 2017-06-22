package query

import (
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"

	trcmn "github.com/tendermint/trackomatron/cmd/trackocli/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
)

//nolint
var (
	QueryProfileCmd = &cobra.Command{
		Use:          "profile [name]",
		Short:        "Query a profile",
		SilenceUsage: true,
		RunE:         queryProfileCmd,
	}

	QueryProfilesCmd = &cobra.Command{
		Use:          "profiles",
		Short:        "List all open profiles",
		SilenceUsage: true,
		RunE:         queryProfilesCmd,
	}
)

func init() {
	FSQueryProfiles := flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryProfiles.Bool(trcmn.FlagInactive, false, "List inactive profiles")
	QueryProfilesCmd.Flags().AddFlagSet(FSQueryProfiles)
}

// DoQueryProfileCmd is the workhorse of the heavy and light cli query profile commands
func queryProfileCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return trcmn.ErrCmdReqArg("name")
	}

	name := args[0]
	if len(name) == 0 {
		return trcmn.ErrBadQuery("name")
	}
	key := invoicer.ProfileKey(name)
	proof, err := getProof(key)
	if err != nil {
		return err
	}
	profile, err := invoicer.GetProfileFromWire(proof.Data())

	if err != nil {
		return err
	}
	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(profile))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(profile)))
	}
	return nil
}

// DoQueryProfilesCmd is the workhorse of the heavy and light cli query profiles commands
func queryProfilesCmd(cmd *cobra.Command, args []string) error {

	var key []byte
	if viper.GetBool(trcmn.FlagInactive) {
		key = invoicer.ListProfileInactiveKey()
	} else {
		key = invoicer.ListProfileActiveKey()
	}

	proof, err := getProof(key)
	if err != nil {
		return err
	}
	listProfiles, err := invoicer.GetListStringFromWire(proof.Data())
	if err != nil {
		return err
	}
	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(listProfiles))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(listProfiles)))
	}
	return nil
}
