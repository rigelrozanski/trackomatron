package commands

//TODO
// edit open profile
// edit an unpaid invoice,
// bulk import from csv,
// JSON imports
// interoperability with ebuchman rates tool

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin-examples/invoicer/plugins/invoicer"
	"github.com/tendermint/basecoin-examples/invoicer/types"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/go-wire"
)

const invoicerName = "invoicer"

var (
	//profile flags
	flagCur                string
	flagDefaultDepositInfo string
	flagDueDurationDays    int
	flagTimezone           string

	//invoice flags
	flagSender      string //hex
	flagReceiver    string //hex
	flagDepositInfo string
	flagAmount      string //AmtCurTime
	flagDate        string
	flagCur         string

	//expense flags
	flagPdfReceipt string //hex
	flagNotes      string
	flagTaxesPaid  string //AmtCurTime

	//close flags
	flagID             string //hex
	flagTransactionID  string //empty when unpaid
	flagPaymentCurTime string //AmtCurTime

	//commands
	InvoicerCmd = &cobra.Command{
		Use:   "invoicer",
		Short: "commands relating to invoicer system",
	}

	NewProfileCmd = &cobra.Command{
		Use:   "new-profile [name]",
		Short: "open a profile for sending/receiving invoices and expense claims",
		RunE:  newProfileCmd,
	}

	OpenInvoiceCmd = &cobra.Command{
		Use:   "invoice",
		Short: "send an invoice",
		RunE:  openInvoiceCmd,
	}

	OpenExpenseCmd = &cobra.Command{
		Use:   "expense",
		Short: "send an expense",
		RunE:  openExpenseCmd,
	}

	CloseCmd = &cobra.Command{
		Use:   "close",
		Short: "close an invoice or expense",
		RunE:  openExpenseCmd,
	}
)

func init() {

	//register flags
	//issueFlag2Reg := bcmd.Flag2Register{&issueFlag, "issue", "default issue", "name of the issue to generate or vote for"}

	profileFlags := []bcmd.Flag2Register{
		{&flagAcceptedCur, "cur", "btc", "currencies accepted for invoice payments"},
		{&flagDefaultDepositInfo, "deposit-info", "", "default deposit information to be provided"},
		{&flagDueDurationDays, "due-days", 14, "default number of days until invoice is due from invoice submission"},
		{&flagTimezone, "timezone", "UTC", "timezone for invoice calculations"},
	}

	invoiceFlags := []bcmd.Flag2Register{
		{&flagSender, "sender", "", "name of invoice/expense sender"},
		{&lagReceiver, "receiver", "allinbits", "name of the invoice/expense receiver"},
		{&flagDepositInfo, "info", "", "deposit information for invoice payment"},
		{&flagAmount, "amount", "", "invoice/expense amount in the format <currency><decimal> eg. usd100.23"},
		{&flagInvoiceDate, "date", "", "invoice/expense date in the format YYYY-MM-DD eg. 2016-12-31 (default: today)"},
		{&flagTimezone, "timezone", "", "invoice/expense timezone (default: profile timezone)"},
		{&flagCur, "cur", "btc", "currency which invoice/expense should be paid in"},
	}

	expenseFlags := []bcmd.Flag2Register{
		{&flagPdfReceipt, "pdf", "", "directory to pdf document of receipt"},
		{&flagNotes, "notes", "", "notes regarding the expense"},
		{&flagTaxesPaid, "taxes", "", "taxes amount in the format <currency><decimal> eg. usd10.23"},
	}

	closeFlags := []bcmd.Flag2Register{
		{&flagID, "id", "", "Invoice ID"},
		{&flagTransactionID, "transaction", "", "completed transaction ID"},
		{&flagPaymentCurTime, "cur", "", "payment amount in the format <currency><decimal> eg. usd10.23"},
		{&flagPaymentDate, "date", "", "date payment in the format YYYY-MM-DD eg. 2016-12-31 (default: today)"},
	}

	bcmd.RegisterFlags(NewProfileCmd, profileFlags)
	bcmd.RegisterFlags(OpenInvoiceCmd, invoiceFlags)
	bcmd.RegisterFlags(OpenExpenseCmd, invoiceFlags)
	bcmd.RegisterFlags(OpenExpenseCmd, expenseFlags)
	bcmd.RegisterFlags(CloseCmd, closeFlags)

	//register commands
	InvoicerCmd.AddCommand(
		NewProfileCmd,
		OpenInvoiceCmd,
		OpenExpenseCmd,
		CloseCmd,
	)
	bcmd.RegisterTxSubcommand(InvoicerCmd)
}

func newProfileCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("new-profile command requires an argument ([name])") //never stack trace
	}
	name := StripHex(args[0])

	timezone, err := time.LoadLocation(flagTimezone)
	if err != nil {
		return fmt.Errorf("error loading timezone, error: ", err) //never stack trace
	}

	txBytes := types.NewTxBytesNewProfile(
		name,
		flagAcceptedCur.(types.Currency),
		flagDefaultDepositInfo,
		flagDueDurationDays,
		timezone,
	)
	return bcmd.AppTx(InvoicerName, txBytes)
}

////invoice flags
//flagSender      string //hex
//flagReceiver    string //hex
//flagDepositInfo string
//flagAmount      string //AmtCurTime
//flagDate        string
//flagCur         string

func openInvoiceCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invoice command requires an argument ([sender])") //never stack trace
	}

	sender := StripHex(args[0])

	t := time.Now()
	if len(flagTimezone) > 0 {

		tz := time.UTC
		if len(flagTimezone) > 0 {
			tz, err := time.LoadLocation(flagTimezone)
			if err != nil {
				return fmt.Errorf("error loading timezone, error: ", err) //never stack trace
			}
		}

		ymd := strings.Split(flagDate, "-")
		if len(ymd) != 3 {
			return fmt.Errorf("bad date parsing, not 3 segments") //never stack trace
		}

		t = time.Date(ymd[0], time.Month(ymd[1]), ymd[2], 0, 0, 0, 0, tz)

	}

	amt := types.AmtCurTime{
		flagAmount.(types.Currency),
		t,
	}

	//txBytes := NewTxBytesOpenInvoice(
	//ID int,
	//sender,
	//Receiver,
	//DepositInfo,
	//Amount *AmtCurTime,
	//AcceptedCur Currency,
	//TransactionID string,
	//PaymentCurTime *AmtCurTime,
	//)
	return bcmd.AppTx(InvoicerName, txBytes)
}
