package query

import (
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"
	trcmd "github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
)

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

	FSQueryProfiles = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	FSQueryProfiles.Bool(FlagInactive, false, "list inactive profiles")
	QueryProfilesCmd.Flags().AddFlagSet(FSQueryProfiles)
	appCmd.AddCommand(trcmd.QueryProfileCmd)
	appCmd.AddCommand(trcmd.QueryProfilesCmd)
}

// DoQueryProfileCmd is the workhorse of the heavy and light cli query profile commands
func queryProfileCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return ErrCmdReqArg("name")
	}

	name := args[0]

	if len(name) == 0 {
		return profile, trcmd.ErrBadQuery("name")
	}
	key := invoicer.ProfileKey(name)
	proof, err := getProof(key)
	if err != nil {
		return
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
	if viper.GetBool(FlagInactive) {
		key = invoicer.ListProfileInactiveKey()
	} else {
		key = invoicer.ListProfileActiveKey()
	}

	proof, err := getProof(key)
	if err != nil {
		return
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
