package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	bcmd "github.com/tendermint/basecoin/cmd/commands"

	"github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
func init() {

	//Set the commands run function
	commands.QueryInvoiceCmd.RunE = queryInvoiceCmd
	commands.QueryInvoicesCmd.RunE = queryInvoicesCmd
	commands.QueryProfileCmd.RunE = queryProfileCmd
	commands.QueryProfilesCmd.RunE = queryProfilesCmd
	commands.QueryPaymentCmd.RunE = queryPaymentCmd
	commands.QueryPaymentsCmd.RunE = queryPaymentsCmd

	//register commands
	bcmd.RegisterQuerySubcommand(commands.QueryInvoicesCmd)
	bcmd.RegisterQuerySubcommand(commands.QueryInvoiceCmd)
	bcmd.RegisterQuerySubcommand(commands.QueryProfileCmd)
	bcmd.RegisterQuerySubcommand(commands.QueryProfilesCmd)
	bcmd.RegisterQuerySubcommand(commands.QueryPaymentCmd)
	bcmd.RegisterQuerySubcommand(commands.QueryPaymentsCmd)
}

type cmdQuery struct {
	tmAddr string
}

func newCmdQuery(cmd *cobra.Command) cmdQuery {
	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()
	return cmdQuery{tmAddr}
}

func queryInvoiceCmd(cmd *cobra.Command, args []string) error {
	c := newCmdQuery(cmd)
	return commands.DoQueryInvoiceCmd(cmd, args, c.queryInvoice)
}

func queryInvoicesCmd(cmd *cobra.Command, args []string) error {
	c := newCmdQuery(cmd)
	return commands.DoQueryInvoicesCmd(cmd, args, c.queryListBytes, c.queryInvoice)
}

func queryProfileCmd(cmd *cobra.Command, args []string) error {
	c := newCmdQuery(cmd)
	return commands.DoQueryProfileCmd(cmd, args, c.queryProfile)
}

func queryProfilesCmd(cmd *cobra.Command, args []string) error {
	c := newCmdQuery(cmd)
	return commands.DoQueryProfilesCmd(cmd, args, c.queryListString)
}

func queryPaymentCmd(cmd *cobra.Command, args []string) error {
	c := newCmdQuery(cmd)
	return commands.DoQueryPaymentCmd(cmd, args, c.queryPayment)
}

func queryPaymentsCmd(cmd *cobra.Command, args []string) error {
	c := newCmdQuery(cmd)
	return commands.DoQueryPaymentsCmd(cmd, args, c.queryListString, c.queryPayment)
}

///////////////////////////////////////////////////////////////////

func (c cmdQuery) queryProfile(name string) (profile types.Profile, err error) {

	if len(name) == 0 {
		return profile, commands.ErrBadQuery("name")
	}
	key := invoicer.ProfileKey(name)

	res, err := c.query(key)
	if err != nil {
		return profile, err
	}

	return invoicer.GetProfileFromWire(res)
}

func (c cmdQuery) queryInvoice(id []byte) (invoice types.Invoice, err error) {

	if len(id) == 0 {
		return invoice, commands.ErrBadQuery("id")
	}

	key := invoicer.InvoiceKey(id)
	res, err := c.query(key)
	if err != nil {
		return invoice, err
	}

	return invoicer.GetInvoiceFromWire(res)
}

func (c cmdQuery) queryPayment(transactionID string) (payment types.Payment, err error) {

	if len(transactionID) == 0 {
		return payment, commands.ErrBadQuery("transactionID")
	}

	key := invoicer.PaymentKey(transactionID)
	res, err := c.query(key)
	if err != nil {
		return payment, err
	}

	return invoicer.GetPaymentFromWire(res)
}

func (c cmdQuery) queryListString(key []byte) (list []string, err error) {
	res, err := c.query(key)
	if err != nil {
		return
	}
	return invoicer.GetListStringFromWire(res)
}

func (c cmdQuery) queryListBytes(key []byte) (list [][]byte, err error) {
	res, err := c.query(key)
	if err != nil {
		return
	}
	return invoicer.GetListBytesFromWire(res)
}

//Wrap the basecoin query function with a response code check
func (c cmdQuery) query(key []byte) ([]byte, error) {
	resp, err := bcmd.Query(c.tmAddr, key)
	if err != nil {
		return nil, err
	}
	if !resp.Code.IsOK() {
		return nil, errors.Errorf("Query for key (%v) returned non-zero code (%v): %v",
			string(key), resp.Code, resp.Log)
	}
	return resp.Value, nil
}
