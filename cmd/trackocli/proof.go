package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/light-client/commands/proofs"
	trcmd "github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

func init() {
	//Register the custom commands with the proof state command
	proofs.RegisterProofStateSubcommand(GetQueryProfileCmd())
	proofs.RegisterProofStateSubcommand(GetQueryProfilesCmd())
	proofs.RegisterProofStateSubcommand(GetQueryInvoiceCmd())
	proofs.RegisterProofStateSubcommand(GetQueryInvoicesCmd())
	proofs.RegisterProofStateSubcommand(GetQueryPaymentCmd())
	proofs.RegisterProofStateSubcommand(GetQueryPaymentsCmd())
}

//nolint - These functions represent what is being called by the proof state commands
// A funtion which returns a command is necessary because the ProofCommander
// must be passed into the query functions
func GetQueryInvoiceCmd() *cobra.Command {
	cmd := trcmd.QueryInvoiceCmd
	cmd.RunE = queryInvoiceCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryInvoicesCmd() *cobra.Command {
	cmd := trcmd.QueryInvoicesCmd
	cmd.RunE = queryInvoicesCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryProfileCmd() *cobra.Command {
	cmd := trcmd.QueryProfileCmd
	cmd.RunE = queryProfileCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryProfilesCmd() *cobra.Command {
	cmd := trcmd.QueryProfilesCmd
	cmd.RunE = queryProfilesCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryPaymentCmd() *cobra.Command {
	cmd := trcmd.QueryPaymentCmd
	cmd.RunE = queryPaymentCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryPaymentsCmd() *cobra.Command {
	cmd := trcmd.QueryPaymentsCmd
	cmd.RunE = queryPaymentsCmd
	cmd.SilenceUsage = true
	return cmd
}

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

func queryProfile(name string) (profile types.Profile, err error) {
	if len(name) == 0 {
		return profile, trcmd.ErrBadQuery("name")
	}
	key := invoicer.ProfileKey(name)
	proof, err := proofs.StateProverCommander.GetProof(key, 0) //0 height means latest block
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
	proof, err := proofs.StateProverCommander.GetProof(key, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetPaymentFromWire(proof.Data())
}

func queryInvoice(id []byte) (invoice types.Invoice, err error) {
	key := invoicer.InvoiceKey(id)
	proof, err := proofs.StateProverCommander.GetProof(key, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetInvoiceFromWire(proof.Data())
}

func queryListString(key []byte) (list []string, err error) {
	proof, err := proofs.StateProverCommander.GetProof(key, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetListStringFromWire(proof.Data())
}

func queryListBytes(key []byte) (list [][]byte, err error) {
	proof, err := proofs.StateProverCommander.GetProof(key, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetListBytesFromWire(proof.Data())
}
