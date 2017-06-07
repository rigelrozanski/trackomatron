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
	trcmd "github.com/tendermint/trackomatron/commands"
	invplug "github.com/tendermint/trackomatron/plugins/invoicer"
)

// TrackoCli represents the base command when called without any subcommands
var TrackoCli = &cobra.Command{
	Use:   "trackocli",
	Short: "Light client for trackomatron",
}

func main() {
	commands.AddBasicFlags(TrackoCli)

	//initialize proofs and txs default basecoin behaviour
	proofs.StatePresenters.Register("account", bcmd.AccountPresenter{})
	proofs.TxPresenters.Register("base", bcmd.BaseTxPresenter{})
	txs.Register("send", bcmd.SendTxMaker{})

	//register invoicer plugin flags
	proofs.StatePresenters.Register(trcmd.AppAdapterProfile, adapters.ProfilePresenter{})
	proofs.StatePresenters.Register(trcmd.AppAdapterInvoice, adapters.InvoicePresenter{})
	proofs.StatePresenters.Register(trcmd.AppAdapterPayment, adapters.PaymentPresenter{})
	proofs.StatePresenters.Register(trcmd.AppAdapterListString, adapters.ListStringPresenter{})
	proofs.StatePresenters.Register(trcmd.AppAdapterListBytes, adapters.ListBytesPresenter{})

	txs.Register(trcmd.TxNameProfileOpen, adapters.ProfileTxMaker{TBTx: invplug.TBTxProfileOpen})
	txs.Register(trcmd.TxNameProfileEdit, adapters.ProfileTxMaker{TBTx: invplug.TBTxProfileEdit})
	txs.Register(trcmd.TxNameProfileDeactivate, adapters.ProfileTxMaker{TBTx: invplug.TBTxProfileDeactivate})
	txs.Register(trcmd.TxNameContractOpen, adapters.InvoiceTxMaker{TBTx: invplug.TBTxContractOpen})
	txs.Register(trcmd.TxNameContractEdit, adapters.InvoiceTxMaker{TBTx: invplug.TBTxContractEdit})
	txs.Register(trcmd.TxNameExpenseOpen, adapters.InvoiceTxMaker{TBTx: invplug.TBTxExpenseOpen})
	txs.Register(trcmd.TxNameExpenseEdit, adapters.InvoiceTxMaker{TBTx: invplug.TBTxExpenseEdit})
	txs.Register(trcmd.TxNamePayment, adapters.PaymentTxMaker{})

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
