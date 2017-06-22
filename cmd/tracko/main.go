package main

import (
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/tmlibs/cli"
	"github.com/tendermint/trackomatron/plugins/invoicer"
)

func main() {

	var RootCmd = &cobra.Command{
		Use: "tracko",
	}

	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.RelayCmd,
		commands.UnsafeResetAllCmd,
		commands.QuickVersionCmd("0.1.0"),
	)

	commands.RegisterStartPlugin(invoicer.Name, func() types.Plugin { return invoicer.New() })
	cmd := cli.PrepareMainCmd(
		RootCmd,
		"TRK",
		os.ExpandEnv(path.Join("$HOME", ".tracko")),
	)
	cmd.Execute()
}
