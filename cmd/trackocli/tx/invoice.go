//nolint
package adapters

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	lightclient "github.com/tendermint/light-client"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"

	"github.com/tendermint/light-client/commands/txs"
	"github.com/tendermint/trackomatron/plugins/invoicer"
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

	FSTxInvoice = flag.NewFlagSet("", flag.ContinueOnError)
	FSTxExpense = flag.NewFlagSet("", flag.ContinueOnError)
	FSTxInvoiceEdit = flag.NewFlagSet("", flag.ContinueOnError)
	//only need to add common flags to this flagset as it's included in all invoice commands
	bcmd.AddAppTxFlags(FSTxInvoice)

	FSTxInvoice.String(FlagTo, "allinbits", "Name of the invoice receiver")
	FSTxInvoice.String(FlagDepositInfo, "", "Deposit information for invoice payment (default: profile)")
	FSTxInvoice.String(FlagNotes, "", "Notes regarding the expense")
	FSTxInvoice.String(FlagCur, "", "Currency which invoice should be paid in")
	FSTxInvoice.String(FlagDate, "", "Invoice demon date in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	FSTxInvoice.String(FlagDueDate, "", "Invoice due date in the format YYYY-MM-DD eg. 2016-12-31 (default: profile)")
	FSTxExpense.String(FlagReceipt, "", "Directory to receipt document file")
	FSTxExpense.String(FlagTaxesPaid, "", "Taxes amount in the format <decimal><currency> eg. 10.23usd")
	FSTxInvoiceEdit.String(FlagID, "", "ID (hex) of the invoice to modify")

	ContractOpenCmd.Flags().AddFlagSet(FSTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(FSTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(FSTxInvoiceEdit)
	ExpenseOpenCmd.Flags().AddFlagSet(FSTxInvoice)
	ExpenseOpenCmd.Flags().AddFlagSet(FSTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(FSTxInvoice)
	ExpenseEditCmd.Flags().AddFlagSet(FSTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(FSTxInvoiceEdit)

	txs.RootCmd.AddCommand(
		ContractOpenCmd,
		ContractEditCmd,
		ExpenseOpenCmd,
		ExpenseEditCmd,
	)
}

func contractOpenCmd(cmd *cobra.Command, args []string, TBTx byte) error {
	return invoiceCmd(cmd, args, invoicer.TBTxContractOpen)
}
func contractEditCmd(cmd *cobra.Command, args []string, TBTx byte) error {
	return invoiceCmd(cmd, args, invoicer.TBTxContractEdit)
}
func expenseOpenCmd(cmd *cobra.Command, args []string, TBTx byte) error {
	return invoiceCmd(cmd, args, invoicer.TBTxExpenseOpen)
}
func expenseEditCmd(cmd *cobra.Command, args []string, TBTx byte) error {
	return invoiceCmd(cmd, args, invoicer.TBTxExpenseEdit)
}

func invoiceCmd(cmd *cobra.Command, args []string, TBTx byte) error {
	// Note: we don't support loading apptx from json currently, so skip that

	// Read the standard app-tx flags
	gas, fee, txInput, err := bcmd.ReadAppTxFlags()
	if err != nil {
		return err
	}

	// Retrieve the app-specific flags
	if len(args) != 1 {
		return trcmd.ErrCmdReqArg("amount<amt><cur>")
	}
	amountStr := args[0]

	data, err := InvoiceTx(TBTx, txInput.Address, amountStr)
	if err != nil {
		return err
	}

	//create the name from the command
	name = strings.Split(cmd.Use, " ")[0]

	// Create AppTx and broadcast
	tx := &btypes.AppTx{
		Gas:   gas,
		Fee:   fee,
		Name:  name,
		Input: txInput,
		Data:  data,
	}
	res, err := bcmd.BroadcastAppTx(tx)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(res)
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
