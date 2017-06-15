//nolint
package adapters

import (
	"encoding/hex"
	"errors"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"

	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/trackocli/common"
	"github.com/tendermint/trackomatron/types"
)

var (
	ContractOpenCmd = &cobra.Command{
		Use:   "contract-open [amount]",
		Short: "Send a contract invoice of amount <value><currency>",
		RunE:  contractOpenCmd,
	}

	ContractEditCmd = &cobra.Command{
		Use:   "contract-edit [amount]",
		Short: "Edit an open contract invoice to amount <value><currency>",
		RunE:  contractEditCmd,
	}

	ExpenseOpenCmd = &cobra.Command{
		Use:   "expense-open [amount]",
		Short: "Send an expense invoice of amount <value><currency>",
		RunE:  expenseOpenCmd,
	}

	ExpenseEditCmd = &cobra.Command{
		Use:   "expense-edit [amount]",
		Short: "Edit an open expense invoice to amount <value><currency>",
		RunE:  expenseEditCmd,
	}

	_ lightclient.TxReader = InvoiceTxReader{}
)

func init() {

	ProfileOpenCmd.Flags().AddFlagSet(trcmd.FSTxProfile)
	ProfileEditCmd.Flags().AddFlagSet(trcmd.FSTxProfile)

	ContractOpenCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoiceEdit)

	ExpenseOpenCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ExpenseOpenCmd.Flags().AddFlagSet(trcmd.FSTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoiceEdit)

	PaymentCmd.Flags().AddFlagSet(trcmd.FSTxPayment)

	//register commands
	InvoicerCmd.AddCommand(
		ProfileOpenCmd,
		ProfileEditCmd,
		ProfileDeactivateCmd,
		ContractOpenCmd,
		ContractEditCmd,
		ExpenseOpenCmd,
		ExpenseEditCmd,
		PaymentCmd,
	)
	bcmd.RegisterTxSubcommand(InvoicerCmd)
}

type InvoiceTxMaker struct {
	TBTx byte
}

func (m InvoiceTxMaker) MakeReader() (lightclient.TxReader, error) {
	chainID := viper.GetString(commands.ChainFlag)
	return InvoiceTxReader{
		App:  bcmd.AppTxReader{ChainID: chainID},
		TBTx: m.TBTx,
	}, nil
}

// define flags

type InvoiceFlags struct {
	bcmd.AppFlags `mapstructure:",squash"`
}

func (m InvoiceTxMaker) Flags() (*flag.FlagSet, interface{}) {
	fs, app := bcmd.AppFlagSet()
	fs.AddFlagSet(common.FSTxInvoice)
	fs.String(common.FlagInvoiceAmount, "", "Name of the new invoice to open")

	//add additional flags, as necessary
	switch m.TBTx {
	case invoicer.TBTxExpenseOpen:
		fs.AddFlagSet(common.FSTxExpense)
	case invoicer.TBTxExpenseEdit:
		fs.AddFlagSet(common.FSTxExpense)
		fs.AddFlagSet(common.FSTxInvoiceEdit)
	case invoicer.TBTxContractEdit:
		fs.AddFlagSet(common.FSTxInvoiceEdit)
	}

	return fs, &InvoiceFlags{AppFlags: app}
}

// parse flags

type InvoiceTxReader struct {
	App  bcmd.AppTxReader
	TBTx byte
}

func (t InvoiceTxReader) ReadTxJSON(data []byte, pk crypto.PubKey) (interface{}, error) {
	return t.App.ReadTxJSON(data, pk)
}

func (t InvoiceTxReader) ReadTxFlags(flags interface{}, pk crypto.PubKey) (interface{}, error) {
	data := flags.(*InvoiceFlags)
	amount := viper.GetString(common.FlagInvoiceAmount)
	senderAddr := pk.Address()
	txBytes, err := InvoiceTx(t.TBTx, senderAddr, amount)
	if err != nil {
		return nil, err
	}
	return t.App.ReadTxFlags(&data.AppFlags, invoicer.Name, txBytes, pk)
}

// InvoiceTx Generates the Tendermint tx
func InvoiceTx(TBTx byte, senderAddr []byte, amountStr string) ([]byte, error) {

	var id []byte

	//if editing
	var err error
	if TBTx == invoicer.TBTxContractEdit || //require this flag if
		TBTx == invoicer.TBTxExpenseEdit { //require this flag if

		//get the old id to remove if editing
		idRaw := viper.GetString(FlagID)
		if len(idRaw) == 0 {
			return nil, errors.New("Need the id to edit, please specify through the flag --id")
		}
		if !cmn.IsHex(idRaw) {
			return nil, ErrBadHexID
		}
		id, err = hex.DecodeString(cmn.StripHex(idRaw))
		if err != nil {
			return nil, err
		}
	}

	//check for expenses flags required
	if TBTx == invoicer.TBTxExpenseOpen ||
		TBTx == invoicer.TBTxExpenseEdit {

		if len(viper.GetString(FlagTaxesPaid)) == 0 {
			return nil, errors.New("Need --taxes flag")
		}
	}

	tx := types.TxInvoice{
		EditID:      id,
		Amount:      amountStr,
		SenderAddr:  senderAddr,
		To:          viper.GetString(FlagTo),
		DepositInfo: viper.GetString(FlagDepositInfo),
		Notes:       viper.GetString(FlagNotes),
		Cur:         viper.GetString(FlagCur),
		Date:        viper.GetString(FlagDate),
		DueDate:     viper.GetString(FlagDueDate),
		Receipt:     viper.GetString(FlagReceipt),
		TaxesPaid:   viper.GetString(FlagTaxesPaid),
	}

	return invoicer.MarshalWithTB(tx, TBTx), nil
}
