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
	trquery "github.com/tendermint/trackomatron/cmd/trackocli/query"
	trtx "github.com/tendermint/trackomatron/cmd/trackocli/tx"
)

// TrackoCli represents the base command when called without any subcommands
var TrackoCli = &cobra.Command{
	Use:   "trackocli",
	Short: "Light client for trackomatron",
}

func main() {
	//Add the basic flags
	commands.AddBasicFlags(TrackoCli)

	// Prepare queries
	proofs.RootCmd.AddCommand(
		//basecoin commands
		bcmd.AccountQueryCmd,
		//custom commands
		trquery.QueryInvoiceCmd,
		trquery.QueryInvoicesCmd,
		trquery.QueryProfileCmd,
		trquery.QueryProfilesCmd,
		trquery.QueryPaymentCmd,
		trquery.QueryPaymentsCmd,
	)

	//Initialize proofs and txs default basecoin behaviour
	//proofs.StateGetPresenters.Register("account", bcmd.AccountPresenter{})
	proofs.TxPresenters.Register("base", bcmd.BaseTxPresenter{})
	txs.RootCmd.AddCommand(
		//basecoin commands
		bcmd.SendTxCmd,
		//custom commands
		trtx.ProfileOpenCmd,
		trtx.ProfileEditCmd,
		trtx.ProfileDeactivateCmd,
		trtx.ContractOpenCmd,
		trtx.ContractEditCmd,
		trtx.ExpenseOpenCmd,
		trtx.ExpenseEditCmd,
		trtx.PaymentCmd,
	)

	// set up the various commands to use
	TrackoCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		keycmd.RootCmd,
		seeds.RootCmd,
		proofs.RootCmd,
		txs.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(TrackoCli, "TRC", os.ExpandEnv("$HOME/.trackocli"))
	cmd.Execute()
}
