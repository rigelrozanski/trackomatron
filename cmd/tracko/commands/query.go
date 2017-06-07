package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	bcmd "github.com/tendermint/basecoin/cmd/commands"

	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
func init() {

	//Set the commands run function
	QueryInvoiceCmd.RunE = queryInvoiceCmd
	QueryInvoicesCmd.RunE = queryInvoicesCmd
	QueryProfileCmd.RunE = queryProfileCmd
	QueryProfilesCmd.RunE = queryProfilesCmd
	QueryPaymentCmd.RunE = queryPaymentCmd
	QueryPaymentsCmd.RunE = queryPaymentsCmd

	//register commands
	bcmd.RegisterQuerySubcommand(QueryInvoicesCmd)
	bcmd.RegisterQuerySubcommand(QueryInvoiceCmd)
	bcmd.RegisterQuerySubcommand(QueryProfileCmd)
	bcmd.RegisterQuerySubcommand(QueryProfilesCmd)
	bcmd.RegisterQuerySubcommand(QueryPaymentCmd)
	bcmd.RegisterQuerySubcommand(QueryPaymentsCmd)
}

func queryInvoiceCmd(cmd *cobra.Command, args []string) error {
	return DoQueryInvoiceCmd(cmd, args, queryInvoice)
}

func queryInvoicesCmd(cmd *cobra.Command, args []string) error {
	return DoQueryInvoicesCmd(cmd, args, queryListBytes, queryInvoice)
}

func queryProfileCmd(cmd *cobra.Command, args []string) error {
	return DoQueryProfileCmd(cmd, args, queryProfile)
}

func queryProfilesCmd(cmd *cobra.Command, args []string) error {
	return DoQueryProfilesCmd(cmd, args, queryListString)
}

func queryPaymentCmd(cmd *cobra.Command, args []string) error {
	return DoQueryPaymentCmd(cmd, args, queryPayment)
}

func queryPaymentsCmd(cmd *cobra.Command, args []string) error {
	return DoQueryPaymentsCmd(cmd, args, queryListString, queryPayment)
}

///////////////////////////////////////////////////////////////////

func queryProfile(name string) (profile types.Profile, err error) {

	if len(name) == 0 {
		return profile, ErrBadQuery("name")
	}
	key := invoicer.ProfileKey(name)

	res, err := query(key)
	if err != nil {
		return profile, err
	}

	return invoicer.GetProfileFromWire(res)
}

func queryInvoice(id []byte) (invoice types.Invoice, err error) {

	if len(id) == 0 {
		return invoice, ErrBadQuery("id")
	}

	key := invoicer.InvoiceKey(id)
	res, err := query(key)
	if err != nil {
		return invoice, err
	}

	return invoicer.GetInvoiceFromWire(res)
}

func queryPayment(transactionID string) (payment types.Payment, err error) {

	if len(transactionID) == 0 {
		return payment, ErrBadQuery("transactionID")
	}

	key := invoicer.PaymentKey(transactionID)
	res, err := query(key)
	if err != nil {
		return payment, err
	}

	return invoicer.GetPaymentFromWire(res)
}

func queryListString(key []byte) (list []string, err error) {
	res, err := query(key)
	if err != nil {
		return
	}
	return invoicer.GetListStringFromWire(res)
}

func queryListBytes(key []byte) (list [][]byte, err error) {
	res, err := query(key)
	if err != nil {
		return
	}
	return invoicer.GetListBytesFromWire(res)
}

//Wrap the basecoin query function with a response code check
func query(key []byte) ([]byte, error) {
	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()
	resp, err := bcmd.Query(tmAddr, key)
	if err != nil {
		return nil, err
	}
	if !resp.Code.IsOK() {
		return nil, errors.Errorf("Query for key (%v) returned non-zero code (%v): %v",
			string(key), resp.Code, resp.Log)
	}
	return resp.Value, nil
}
