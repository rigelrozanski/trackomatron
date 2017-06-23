package tx

import (
	"encoding/hex"
	"errors"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	btypes "github.com/tendermint/basecoin/types"
	txcmd "github.com/tendermint/light-client/commands/txs"
	cmn "github.com/tendermint/tmlibs/common"

	trcmn "github.com/tendermint/trackomatron/cmd/trackocli/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
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
)

func init() {

	fsTxInvoice := flag.NewFlagSet("", flag.ContinueOnError)
	fsTxExpense := flag.NewFlagSet("", flag.ContinueOnError)
	fsTxInvoiceEdit := flag.NewFlagSet("", flag.ContinueOnError)

	//only need to add apptx flags to this flagset as it's included in all invoice commands
	bcmd.AddAppTxFlags(fsTxInvoice)

	fsTxInvoice.String(trcmn.FlagTo, "allinbits", "Name of the invoice receiver")
	fsTxInvoice.String(trcmn.FlagDepositInfo, "", "Deposit information for invoice payment (default: profile)")
	fsTxInvoice.String(trcmn.FlagNotes, "", "Notes regarding the expense")
	fsTxInvoice.String(trcmn.FlagCur, "", "Currency which invoice should be paid in")
	fsTxInvoice.String(trcmn.FlagDate, "", "Invoice demon date in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	fsTxInvoice.String(trcmn.FlagDueDate, "", "Invoice due date in the format YYYY-MM-DD eg. 2016-12-31 (default: profile)")
	fsTxExpense.String(trcmn.FlagReceipt, "", "Directory to receipt document file")
	fsTxExpense.String(trcmn.FlagTaxesPaid, "", "Taxes amount in the format <decimal><currency> eg. 10.23usd")
	fsTxInvoiceEdit.String(trcmn.FlagID, "", "ID (hex) of the invoice to modify")

	ContractOpenCmd.Flags().AddFlagSet(fsTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(fsTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(fsTxInvoiceEdit)
	ExpenseOpenCmd.Flags().AddFlagSet(fsTxInvoice)
	ExpenseOpenCmd.Flags().AddFlagSet(fsTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(fsTxInvoice)
	ExpenseEditCmd.Flags().AddFlagSet(fsTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(fsTxInvoiceEdit)
}

func contractOpenCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, invoicer.TBTxContractOpen)
}
func contractEditCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, invoicer.TBTxContractEdit)
}
func expenseOpenCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, invoicer.TBTxExpenseOpen)
}
func expenseEditCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, invoicer.TBTxExpenseEdit)
}

func invoiceCmd(cmd *cobra.Command, args []string, TBTx byte) error {
	// Note: we don't support loading apptx from json currently, so skip that

	// Read the standard app-tx flags
	gas, fee, txInput, err := bcmd.ReadAppTxFlags()
	if err != nil {
		return err
	}

	// Retrieve the app-specific flags/args
	if len(args) != 1 {
		return trcmn.ErrCmdReqArg("amount<amt><cur>")
	}
	amountStr := args[0]

	data, err := invoiceTx(TBTx, txInput.Address, amountStr)
	if err != nil {
		return err
	}

	// Create AppTx and broadcast
	tx := &btypes.AppTx{
		Gas:   gas,
		Fee:   fee,
		Name:  invoicer.Name,
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

// invoiceTx Generates the Tendermint tx
func invoiceTx(TBTx byte, senderAddr []byte, amountStr string) ([]byte, error) {

	var id []byte

	//if editing
	var err error
	if TBTx == invoicer.TBTxContractEdit || //require this flag if
		TBTx == invoicer.TBTxExpenseEdit { //require this flag if

		//get the old id to remove if editing
		idRaw := viper.GetString(trcmn.FlagID)
		if len(idRaw) == 0 {
			return nil, errors.New("Need the id to edit, please specify through the flag --id")
		}
		if !cmn.IsHex(idRaw) {
			return nil, trcmn.ErrBadHexID
		}
		id, err = hex.DecodeString(cmn.StripHex(idRaw))
		if err != nil {
			return nil, err
		}
	}

	//check for expenses flags required
	if TBTx == invoicer.TBTxExpenseOpen ||
		TBTx == invoicer.TBTxExpenseEdit {

		if len(viper.GetString(trcmn.FlagTaxesPaid)) == 0 {
			return nil, errors.New("Need --taxes flag")
		}
	}

	tx := types.TxInvoice{
		EditID:      id,
		Amount:      amountStr,
		SenderAddr:  senderAddr,
		To:          viper.GetString(trcmn.FlagTo),
		DepositInfo: viper.GetString(trcmn.FlagDepositInfo),
		Notes:       viper.GetString(trcmn.FlagNotes),
		Cur:         viper.GetString(trcmn.FlagCur),
		Date:        viper.GetString(trcmn.FlagDate),
		DueDate:     viper.GetString(trcmn.FlagDueDate),
		Receipt:     viper.GetString(trcmn.FlagReceipt),
		TaxesPaid:   viper.GetString(trcmn.FlagTaxesPaid),
	}

	return invoicer.MarshalWithTB(tx, TBTx), nil
}
