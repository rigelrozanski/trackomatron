package main

import (
	"os"

	"github.com/spf13/cobra"

	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/light-client/commands"
	"github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/commands/proxy"
	"github.com/tendermint/light-client/commands/seeds"
	"github.com/tendermint/light-client/commands/txs"
	"github.com/tendermint/tmlibs/cli"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	adapters "github.com/tendermint/trackomatron/cmd/trackocli/adapters"
	invplug "github.com/tendermint/trackomatron/plugins/invoicer"
)

// TrackoCli represents the base command when called without any subcommands
var TrackoCli = &cobra.Command{
	Use:   "trackocli",
	Short: "Light client for trackomatron",
	Long:  `trackocli is an version of basecli`,
}

func main() {
	commands.AddBasicFlags(TrackoCli)

	//initialize proofs and txs
	proofs.StatePresenters.Register("account", bcmd.AccountPresenter{})
	proofs.TxPresenters.Register("base", bcmd.BaseTxPresenter{})
	proofs.StatePresenters.Register("profile", adapters.ProfilePresenter{})

	txs.Register("send", bcmd.SendTxMaker{})
	txs.Register("profile-open", adapters.ProfileTxMaker{invplug.TBTxProfileOpen})
	txs.Register("profile-edit", adapters.ProfileTxMaker{invplug.TBTxProfileEdit})
	txs.Register("profile-deactivate", adapters.ProfileTxMaker{invplug.TBTxProfileDeactivate})

	// set up the various commands to use
	TrackoCli.AddCommand(
		keycmd.RootCmd,
		commands.InitCmd,
		seeds.RootCmd,
		proofs.RootCmd,
		txs.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(TrackoCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	cmd.Execute()
}
