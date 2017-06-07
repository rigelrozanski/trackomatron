package main

import (
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/tmlibs/cli"
	_ "github.com/tendermint/trackomatron/commands"
)

func main() {

	var RootCmd = &cobra.Command{
		Use: "tracko",
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
		"TRK",
		os.ExpandEnv(path.Join("$HOME", ".tracko")),
	)
	cmd.Execute()
}
