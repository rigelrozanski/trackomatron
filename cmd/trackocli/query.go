package main

import (
	"github.com/spf13/cobra"

	lc "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"
	cmdproofs "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"
	trcmd "github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Handle custom app proofs for state of abci app",
}

func init() {
	//Set the commands run function
	trcmd.QueryInvoiceCmd.RunE = queryInvoiceCmd
	trcmd.QueryInvoicesCmd.RunE = queryInvoicesCmd
	trcmd.QueryProfileCmd.RunE = queryProfileCmd
	trcmd.QueryProfilesCmd.RunE = queryProfilesCmd
	trcmd.QueryPaymentCmd.RunE = queryPaymentCmd
	trcmd.QueryPaymentsCmd.RunE = queryPaymentsCmd

	//Register the custom commands with the proof state command
	appCmd.AddCommand(trcmd.QueryProfileCmd)
	appCmd.AddCommand(trcmd.QueryProfilesCmd)
	appCmd.AddCommand(trcmd.QueryInvoiceCmd)
	appCmd.AddCommand(trcmd.QueryInvoicesCmd)
	appCmd.AddCommand(trcmd.QueryPaymentCmd)
	appCmd.AddCommand(trcmd.QueryPaymentsCmd)
	cmdproofs.RootCmd.AddCommand(appCmd)
}

//These functions represent what is being called by the query app commands
func queryInvoiceCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryInvoiceCmd(cmd, args, queryInvoice)
}

func queryInvoicesCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryInvoicesCmd(cmd, args, queryListBytes, queryInvoice)
}

func queryProfileCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryProfileCmd(cmd, args, queryProfile)
}

func queryProfilesCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryProfilesCmd(cmd, args, queryListString)
}

func queryPaymentCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryPaymentCmd(cmd, args, queryPayment)
}

func queryPaymentsCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryPaymentsCmd(cmd, args, queryListString, queryPayment)
}

///////////////////////////////////////////////

func getProof(key []byte) (lc.Proof, error) {
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
	height := cmdproofs.GetHeight()
	return cmdproofs.GetProof(node, prover, key, height)
}

func queryProfile(name string) (profile types.Profile, err error) {
	if len(name) == 0 {
		return profile, trcmd.ErrBadQuery("name")
	}
	key := invoicer.ProfileKey(name)
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetProfileFromWire(proof.Data())
}

func queryPayment(transactionID string) (payment types.Payment, err error) {
	if len(transactionID) == 0 {
		return payment, trcmd.ErrBadQuery("transactionID")
	}
	key := invoicer.PaymentKey(transactionID)
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetPaymentFromWire(proof.Data())
}

func queryInvoice(id []byte) (invoice types.Invoice, err error) {
	key := invoicer.InvoiceKey(id)
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetInvoiceFromWire(proof.Data())
}

func queryListString(key []byte) (list []string, err error) {
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetListStringFromWire(proof.Data())
}

func queryListBytes(key []byte) (list [][]byte, err error) {
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetListBytesFromWire(proof.Data())
}
