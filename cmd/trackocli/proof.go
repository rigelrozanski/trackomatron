package main

import (
	"encoding/hex"

	"github.com/spf13/cobra"

	"github.com/tendermint/light-client/commands/proofs"
	trcmd "github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
type ProofCommander struct {
	proofs.ProofCommander
}

func init() {
	//Set the commands run function
	proofs.RegisterProofStateSubcommand(GetQueryProfileCmd)
	proofs.RegisterProofStateSubcommand(GetQueryProfilesCmd)
	proofs.RegisterProofStateSubcommand(GetQueryInvoiceCmd)
	proofs.RegisterProofStateSubcommand(GetQueryInvoicesCmd)
	proofs.RegisterProofStateSubcommand(GetQueryPaymentCmd)
	proofs.RegisterProofStateSubcommand(GetQueryPaymentsCmd)
}

//nolint
func GetQueryInvoiceCmd(lp proofs.ProofCommander) *cobra.Command {
	p := ProofCommander{lp}
	cmd := trcmd.QueryInvoiceCmd
	cmd.RunE = p.queryInvoiceCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryInvoicesCmd(lp proofs.ProofCommander) *cobra.Command {
	p := ProofCommander{lp}
	cmd := trcmd.QueryInvoicesCmd
	cmd.RunE = p.queryInvoicesCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryProfileCmd(lp proofs.ProofCommander) *cobra.Command {
	p := ProofCommander{lp}
	cmd := trcmd.QueryProfileCmd
	cmd.RunE = p.queryProfileCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryProfilesCmd(lp proofs.ProofCommander) *cobra.Command {
	p := ProofCommander{lp}
	cmd := trcmd.QueryProfilesCmd
	cmd.RunE = p.queryProfilesCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryPaymentCmd(lp proofs.ProofCommander) *cobra.Command {
	p := ProofCommander{lp}
	cmd := trcmd.QueryPaymentCmd
	cmd.RunE = p.queryPaymentCmd
	cmd.SilenceUsage = true
	return cmd
}
func GetQueryPaymentsCmd(lp proofs.ProofCommander) *cobra.Command {
	p := ProofCommander{lp}
	cmd := trcmd.QueryPaymentsCmd
	cmd.RunE = p.queryPaymentsCmd
	cmd.SilenceUsage = true
	return cmd
}

func (p ProofCommander) queryInvoiceCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryInvoiceCmd(cmd, args, p.queryInvoice)
}

func (p ProofCommander) queryInvoicesCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryInvoicesCmd(cmd, args, p.queryListBytes, queryInvoice)
}

func (p ProofCommander) queryProfileCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryProfileCmd(cmd, args, p.queryProfile)
}

func (p ProofCommander) queryProfilesCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryProfilesCmd(cmd, args, p.queryListString)
}

func (p ProofCommander) queryPaymentCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryPaymentCmd(cmd, args, p.queryPayment)
}

func (p ProofCommander) queryPaymentsCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryPaymentsCmd(cmd, args, p.queryListString, queryPayment)
}

///////////////////////////////////////////////

func (p ProofCommander) queryProfile(name string) (profile types.Profile, err error) {
	if len(name) == 0 {
		return profile, trcmd.ErrBadQuery("name")
	}
	proof, err := p.GetProof(trcmd.AppAdapterProfile, name, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetProfileFromWire(proof.Data())
}

func (p ProofCommander) queryPayment(transactionID string) (payment types.Payment, err error) {
	if len(transactionID) == 0 {
		return payment, trcmd.ErrBadQuery("transactionID")
	}
	proof, err := p.GetProof(trcmd.AppAdapterPayment, transactionID, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetPaymentFromWire(proof.Data())
}

func (p ProofCommander) queryInvoice(id []byte) (invoice types.Invoice, err error) {
	idHexStr := "0x" + hex.EncodeToString(id)
	proof, err := p.GetProof(trcmd.AppAdapterInvoice, idHexStr, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetInvoiceFromWire(proof.Data())
}

func (p ProofCommander) queryListString(key []byte) (list []string, err error) {
	keyHexStr := "0x" + hex.EncodeToString(key)
	proof, err := p.GetProof(trcmd.AppAdapterListString, keyHexStr, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetListStringFromWire(proof.Data())
}

func (p ProofCommander) queryListBytes(key []byte) (list [][]byte, err error) {
	keyHexStr := "0x" + hex.EncodeToString(key)
	proof, err := p.GetProof(trcmd.AppAdapterListBytes, keyHexStr, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetListBytesFromWire(proof.Data())
}
