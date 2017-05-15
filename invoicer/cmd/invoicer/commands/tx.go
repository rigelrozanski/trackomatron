package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin-examples/invoicer/plugins/invoicer"
	"github.com/tendermint/basecoin-examples/invoicer/types"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
)

var (
	//commands
	InvoicerCmd = &cobra.Command{
		Use:   invoicer.Name,
		Short: "commands relating to invoicer system",
	}

	ProfileOpenCmd = &cobra.Command{
		Use:   "profile-open [name]",
		Short: "open a profile for sending/receiving invoices and expense claims",
		RunE:  profileOpenCmd,
	}

	ProfileEditCmd = &cobra.Command{
		Use:   "profile-edit [name]",
		Short: "open a profile for sending/receiving invoices and expense claims",
		RunE:  profileEditCmd,
	}

	ProfileCloseCmd = &cobra.Command{
		Use:   "profile-close [name]",
		Short: "open a profile for sending/receiving invoices and expense claims",
		RunE:  profileCloseCmd,
	}

	WageOpenCmd = &cobra.Command{
		Use:   "wage-open [sender][amount]",
		Short: "send an invoice",
		RunE:  wageOpenCmd,
	}

	WageEditCmd = &cobra.Command{
		Use:   "wage-edit [sender][amount]",
		Short: "send an invoice",
		RunE:  wageEditCmd,
	}

	ExpenseOpenCmd = &cobra.Command{
		Use:   "expense-open [sender][amount]",
		Short: "send an expense",
		RunE:  expenseOpenCmd,
	}

	ExpenseEditCmd = &cobra.Command{
		Use:   "expense-edit [sender][amount]",
		Short: "send an expense",
		RunE:  expenseEditCmd,
	}

	CloseInvoiceCmd = &cobra.Command{
		Use:   "close-invoice [ID]",
		Short: "close an invoice or expense",
		RunE:  closeInvoiceCmd,
	}
)

func init() {

	//register flags
	fsProfile := flag.NewFlagSet("", flag.ContinueOnError)
	fsProfile.String(FlagTo, "", "Destination address for the bits")
	fsProfile.String(FlagCur, "btc", "currencies accepted for invoice payments")
	fsProfile.String(FlagDepositInfo, "", "default deposit information to be provided")
	fsProfile.Int(FlagDueDurationDays, 14, "default number of days until invoice is due from invoice submission")
	fsProfile.String(FlagTimezone, "UTC", "timezone for invoice calculations")

	fsInvoice := flag.NewFlagSet("", flag.ContinueOnError)
	fsInvoice.String(FlagTo, "allinbits", "name of the invoice/expense receiver")
	fsInvoice.String(FlagDepositInfo, "", "deposit information for invoice payment (default: profile)")
	fsInvoice.String(FlagNotes, "", "notes regarding the expense")
	fsInvoice.String(FlagTimezone, "", "invoice/expense timezone (default: profile)")
	fsInvoice.String(FlagCur, "btc", "currency which invoice/expense should be paid in")
	fsInvoice.String(FlagDueDate, "", "invoice/expense due date in the format YYYY-MM-DD eg. 2016-12-31 (default: profile)")

	fsExpense := flag.NewFlagSet("", flag.ContinueOnError)
	fsExpense.String(FlagReceipt, "", "directory to receipt document file")
	fsExpense.String(FlagTaxesPaid, "", "taxes amount in the format <decimal><currency> eg. 10.23usd")

	fsClose := flag.NewFlagSet("", flag.ContinueOnError)
	fsClose.String(FlagTransactionID, "", "completed transaction ID")
	fsClose.String(FlagCur, "", "payment amount in the format <decimal><currency> eg. 10.23usd")
	fsClose.String(FlagDate, "", "date payment in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")

	fsEdit := flag.NewFlagSet("", flag.ContinueOnError)
	fsEdit.String(FlagTransactionID, "", "Hex ID of the invoice to modify")

	ProfileOpenCmd.Flags().AddFlagSet(fsProfile)
	ProfileEditCmd.Flags().AddFlagSet(fsProfile)

	WageOpenCmd.Flags().AddFlagSet(fsInvoice)
	WageEditCmd.Flags().AddFlagSet(fsInvoice)
	WageEditCmd.Flags().AddFlagSet(fsEdit)

	ExpenseOpenCmd.Flags().AddFlagSet(fsInvoice)
	ExpenseOpenCmd.Flags().AddFlagSet(fsExpense)
	ExpenseEditCmd.Flags().AddFlagSet(fsInvoice)
	ExpenseEditCmd.Flags().AddFlagSet(fsExpense)
	ExpenseEditCmd.Flags().AddFlagSet(fsEdit)

	CloseInvoiceCmd.Flags().AddFlagSet(fsClose)

	//register commands
	InvoicerCmd.AddCommand(
		ProfileOpenCmd,
		ProfileEditCmd,
		ProfileCloseCmd,
		WageOpenCmd,
		WageEditCmd,
		ExpenseOpenCmd,
		ExpenseEditCmd,
		CloseInvoiceCmd,
	)
	bcmd.RegisterTxSubcommand(InvoicerCmd)
}

func profileOpenCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(args, types.TBTxProfileOpen)
}

func profileEditCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(args, types.TBTxProfileEdit)
}

func profileCloseCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(args, types.TBTxProfileClose)
}

func profileCmd(args []string, TxTB byte) error {
	if len(args) != 1 {
		return errCmdReqArg("name")
	}
	name := args[0]

	timezone, err := time.LoadLocation(viper.GetString(FlagTimezone))
	if err != nil {
		return errors.Wrap(err, "error loading timezone")
	}

	profile := types.NewProfile(
		name,
		viper.GetString(FlagCur),
		viper.GetString(FlagDepositInfo),
		viper.GetInt(FlagDueDurationDays),
		*timezone,
	)

	txBytes := types.TxBytes(*profile, TxTB)
	return bcmd.AppTx(invoicer.Name, txBytes)
}

func wageOpenCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, types.TBTxWageOpen)
}

func wageEditCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, types.TBTxWageEdit)
}

func expenseOpenCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, types.TBTxExpenseOpen)
}

func expenseEditCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(cmd, args, types.TBTxExpenseEdit)
}

func invoiceCmd(cmd *cobra.Command, args []string, txTB byte) error {
	if len(args) != 2 {
		return fmt.Errorf("Command requires two arguments ([sender][<amt><cur>])") //never stack trace
	}
	sender := args[0]
	amountStr := args[1]

	profile, err := queryProfile(cmd.Parent().Flag("node").Value.String(), sender)
	if err != nil {
		return err
	}

	date, err := types.ParseDate(viper.GetString(FlagDate), viper.GetString(FlagTimezone))
	if err != nil {
		return err
	}
	amt, err := types.ParseAmtCurTime(amountStr, date)
	if err != nil {
		return err
	}

	//retrieve flags, or if they aren't used, use the senders profile's default

	var dueDate time.Time
	if len(viper.GetString(FlagDueDate)) > 0 {
		dueDate, err = types.ParseDate(viper.GetString(FlagDueDate), viper.GetString(FlagTimezone))
		if err != nil {
			return err
		}
	} else {
		dueDate = time.Now().AddDate(0, 0, profile.DueDurationDays)
	}

	var depositInfo string
	if len(viper.GetString(FlagDepositInfo)) > 0 {
		depositInfo = viper.GetString(FlagDepositInfo)
	} else {
		depositInfo = profile.DepositInfo
	}

	var accCur string
	if len(viper.GetString(FlagCur)) > 0 {
		accCur = viper.GetString(FlagCur)
	} else {
		accCur = profile.AcceptedCur
	}

	//get the old id to remove if editing
	var id []byte = nil
	idRaw := viper.GetString(FlagTransactionID)
	if len(idRaw) > 0 {
		if !isHex(idRaw) {
			return errBadHexID
		}
		id, err = hex.DecodeString(StripHex(idRaw))
		if err != nil {
			return err
		}
	} else if txTB == types.TBTxWageEdit || //require this flag if
		txTB == types.TBTxExpenseEdit { //require this flag if
		errors.New("need the id to edit, please specify through the flag --id")
	}

	var invoice types.Invoice

	switch txTB {
	//if not an expense then we're almost done!
	case types.TBTxWageOpen, types.TBTxWageEdit:
		invoice = types.NewWage(
			id,
			sender,
			viper.GetString(FlagTo),
			depositInfo,
			viper.GetString(FlagNotes),
			amt,
			accCur,
			dueDate,
		)
	case types.TBTxExpenseOpen, types.TBTxExpenseEdit:
		if len(viper.GetString(FlagTaxesPaid)) == 0 {
			return errors.New("need --taxes flag")
		}

		taxes, err := types.ParseAmtCurTime(viper.GetString(FlagTaxesPaid), date)
		if err != nil {
			return err
		}
		docBytes, err := ioutil.ReadFile(viper.GetString(FlagReceipt))
		if err != nil {
			return errors.Wrap(err, "Problem reading receipt file")
		}

		_, filename := path.Split(viper.GetString(FlagReceipt))
		invoice = types.NewExpense(
			id,
			sender,
			viper.GetString(FlagTo),
			depositInfo,
			viper.GetString(FlagNotes),
			amt,
			accCur,
			dueDate,
			docBytes,
			filename,
			taxes,
		)
	default:
		return errors.New("Unrecognized TypeBytes")
	}

	//txBytes := invoice.TxBytesOpen()
	txBytes := types.TxBytes(struct{ types.Invoice }{invoice}, txTB)
	return bcmd.AppTx(invoicer.Name, txBytes)
}

func closeInvoiceCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errCmdReqArg("HexID")
	}
	if !isHex(args[0]) {
		return errBadHexID
	}
	id, err := hex.DecodeString(StripHex(args[0]))
	if err != nil {
		return err
	}

	date, err := types.ParseDate(viper.GetString(FlagDate), viper.GetString(FlagTimezone))
	if err != nil {
		return err
	}
	act, err := types.ParseAmtCurTime(viper.GetString(FlagCur), date)
	if err != nil {
		return err
	}

	closeInvoice := types.NewCloseInvoice(
		id,
		viper.GetString(FlagTransactionID),
		act,
	)
	//txBytes := closeInvoice.TxBytes()
	txBytes := types.TxBytes(*closeInvoice, types.TBTxCloseInvoice)
	return bcmd.AppTx(invoicer.Name, txBytes)
}
