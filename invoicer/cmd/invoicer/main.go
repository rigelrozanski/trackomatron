package main

import (
	"github.com/spf13/cobra"
	"os"
	"path"

	_ "github.com/tendermint/basecoin-examples/invoicer/cmd/invoicer/commands"
	"github.com/tendermint/basecoin-examples/invoicer/plugins/invoicer"
	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/tmlibs/cli"
)

func main() {

	var RootCmd = &cobra.Command{
		Use: invoicer.Name,
	}

	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.KeyCmd,
		commands.VerifyCmd,
		commands.BlockCmd,
		commands.AccountCmd,
		commands.UnsafeResetAllCmd,
		commands.QuickVersionCmd("0.1.0"),
	)

	cmd := cli.PrepareMainCmd(
		RootCmd,
		"INV",
		os.ExpandEnv(path.Join("$HOME", "."+invoicer.Name)),
	)
	cmd.Execute()
}
