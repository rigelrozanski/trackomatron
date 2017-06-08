package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/trackomatron/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var (
	//commands
	QueryInvoiceCmd = &cobra.Command{
		Use:   "invoice [id]",
		Short: "Query an invoice by ID",
	}

	QueryInvoicesCmd = &cobra.Command{
		Use:   "invoices",
		Short: "Query all invoice",
	}

	QueryProfileCmd = &cobra.Command{
		Use:   "profile [name]",
		Short: "Query a profile",
	}

	QueryProfilesCmd = &cobra.Command{
		Use:   "profiles",
		Short: "List all open profiles",
	}

	QueryPaymentCmd = &cobra.Command{
		Use:   "payment [id]",
		Short: "List historical payment",
	}

	QueryPaymentsCmd = &cobra.Command{
		Use:   "payments",
		Short: "List historical payments",
	}

	//exposed flagsets
	FSQueryDownload = flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryProfiles = flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryInvoices = flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryPayments = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	//register flags
	FSQueryDownload.String(FlagDownloadExp, "", "download expenses pdfs to the relative path specified")

	FSQueryProfiles.Bool(FlagInactive, false, "list inactive profiles")

	FSQueryInvoices.Int(FlagNum, 0, "number of results to display, use 0 for no limit")
	FSQueryInvoices.String(FlagType, "",
		"limit the scope by using any of the following modifiers with commas: invoice,expense,open,closed")
	FSQueryInvoices.String(FlagDateRange, "",
		"Query within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")
	FSQueryInvoices.String(FlagFrom, "", "Only query for invoices from these addresses in the format <ADDR1>,<ADDR2>, etc.")
	FSQueryInvoices.String(FlagTo, "", "Only query for invoices to these addresses in the format <ADDR1>,<ADDR2>, etc.")
	FSQueryInvoices.Bool(FlagSum, false, "Sum invoice values by sender")

	FSQueryPayments.Int(FlagNum, 0, "number of results to display, use 0 for no limit")
	FSQueryPayments.String(FlagDateRange, "",
		"Query within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")
	FSQueryPayments.String(FlagFrom, "", "Only query for invoices from these addresses in the format <ADDR1>,<ADDR2>, etc.")
	FSQueryPayments.String(FlagTo, "", "Only query for payments to these addresses in the format <ADDR1>,<ADDR2>, etc.")

	QueryInvoiceCmd.Flags().AddFlagSet(FSQueryDownload)
	QueryInvoicesCmd.Flags().AddFlagSet(FSQueryDownload)
	QueryInvoicesCmd.Flags().AddFlagSet(FSQueryInvoices)
	QueryProfilesCmd.Flags().AddFlagSet(FSQueryProfiles)
	QueryPaymentsCmd.Flags().AddFlagSet(FSQueryPayments)
}

// DoQueryInvoiceCmd is the workhorse of the heavy and light cli query profile commands
func DoQueryInvoiceCmd(cmd *cobra.Command, args []string,
	queryInvoice func(id []byte) (types.Invoice, error)) error {

	if len(args) != 1 {
		return ErrCmdReqArg("id")
	}
	if !cmn.IsHex(args[0]) {
		return ErrBadHexID
	}
	id, err := hex.DecodeString(cmn.StripHex(args[0]))
	if err != nil {
		return err
	}

	//get the invoicer object and print it
	invoice, err := queryInvoice(id)
	if err != nil {
		return err
	}

	jsonBytes, err := invoice.MarshalJSON()
	if err != nil {
		return err
	}

	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(jsonBytes)) //TODO Actually make text
	case "json":
		fmt.Println(string(jsonBytes)) //TODO Actually make text
	}

	expense, isExpense := invoice.Unwrap().(*types.Expense)
	if isExpense {
		err = downloadExp(expense)
		if err != nil {
			return errors.Errorf("Problem writing receipt file %v", err)
		}
	}

	return nil
}

func processFlagFromTo() (froms, toes []string) {
	from := viper.GetString(FlagFrom)
	to := viper.GetString(FlagTo)
	if len(froms) > 0 {
		froms = strings.Split(from, ",")
	}
	if len(toes) > 0 {
		toes = strings.Split(to, ",")
	}
	return
}

func processFlagDateRange() (startDate, endDate *time.Time, err error) {
	flagDateRange := viper.GetString(FlagDateRange)
	if len(flagDateRange) > 0 {
		startDate, endDate, err = common.ParseDateRange(flagDateRange)
		if err != nil {
			return
		}
	}
	return
}

// DoQueryInvoicesCmd is the workhorse of the heavy and light cli query profiles commands
func DoQueryInvoicesCmd(cmd *cobra.Command, args []string,
	queryListBytes func(key []byte) ([][]byte, error),
	queryInvoice func(id []byte) (types.Invoice, error)) error {

	listInvoices, err := queryListBytes(invoicer.ListInvoiceKey())
	if err != nil {
		return err
	}

	//return fmt.Errorf("invoicexz %x\n", listInvoices)
	if len(listInvoices) == 0 {
		return fmt.Errorf("No save invoices to return") //never stack trace
	}

	//init flag variables
	froms, toes := processFlagFromTo()

	ty := viper.GetString(FlagType)
	contractFilt, expenseFilt, openFilt, closedFilt := true, true, true, true

	if viper.GetBool("debug") {
		fmt.Printf("debug %v %v %v %v\n", len(ty), ty,
			strings.Contains(ty, "open"), strings.Contains(ty, "closed"))
	}
	if len(ty) > 0 {
		contractFilt, expenseFilt, openFilt, closedFilt = false, false, false, false
		if strings.Contains(ty, "contract") {
			contractFilt = true
		}
		if strings.Contains(ty, "expense") {
			expenseFilt = true
		}
		if strings.Contains(ty, "open") {
			openFilt = true
		}
		if strings.Contains(ty, "closed") {
			closedFilt = true
		}

		//if a whole catagory is missing, turn it on
		if !contractFilt && !expenseFilt {
			contractFilt, expenseFilt = true, true
		}
		if !openFilt && !closedFilt {
			openFilt, closedFilt = true, true
		}
	}
	if viper.GetBool("debug") {
		fmt.Printf("debug filts %v %v %v %v\n", contractFilt,
			expenseFilt, openFilt, closedFilt)
	}

	//get the date range to query
	startDate, endDate, err := processFlagDateRange()
	if err != nil {
		return err
	}

	//Loop through the invoices and query out the valid ones
	var invoices []types.Invoice
	for _, id := range listInvoices {

		invoice, err := queryInvoice(id)
		if err != nil {
			return errors.Errorf("Bad invoice in active invoice list %x \n%v \n%v", id, listInvoices, err)
		}
		ctx := invoice.GetCtx()

		//skip record if out of the date range
		d := ctx.Invoiced.CurTime.Date
		if (startDate != nil && d.Before(*startDate)) ||
			(endDate != nil && d.After(*endDate)) {
			continue
		}

		//continue if doesn't have the sender specified in the from or to flag
		cont := false
		for _, from := range froms {
			if from != ctx.Sender {
				cont = true
				break
			}
		}
		for _, to := range toes {
			if to != ctx.Receiver {
				cont = true
				break
			}
		}
		if cont {
			continue
		}

		//check the type filter flags
		expense, isExpense := invoice.Unwrap().(*types.Expense)
		_, isContract := invoice.Unwrap().(*types.Contract)

		if viper.GetBool("debug") {
			fmt.Printf("debug %v %v %v %v %v\n", isContract, isExpense, ctx.Open, openFilt, closedFilt)
		}
		switch {
		case isContract && !contractFilt && expenseFilt:
			continue
		case isExpense && contractFilt && !expenseFilt:
			continue
		case ctx.Open && !openFilt && closedFilt:
			continue
		case !ctx.Open && openFilt && !closedFilt:
			continue
		}

		if isExpense {
			err = downloadExp(expense)
			if err != nil {
				return errors.Errorf("problem writing receipt file %v", err)
			}
		}

		//all tests have passed so add to the invoices list
		invoices = append(invoices, invoice)

		//Limit the number of invoices retrieved
		maxInv := viper.GetInt(FlagNum)
		if len(invoices) > maxInv && maxInv > 0 {
			break
		}
	}

	//compute the sum if flag is set
	if viper.GetBool(FlagSum) {
		var sum *types.AmtCurTime
		for _, invoice := range invoices {
			unpaid, err := invoice.GetCtx().Unpaid()
			if err != nil {
				return err
			}
			sum, err = sum.Add(unpaid)
			if err != nil {
				return err
			}
		}
		out := struct {
			FinalInvoice types.Invoice
			SumDue       *types.AmtCurTime
		}{
			invoices[len(invoices)-1],
			sum,
		}

		switch viper.GetString("output") {
		case "text":
			fmt.Println(string(wire.JSONBytes(out))) //TODO Actually make text
		case "json":
			fmt.Println(string(wire.JSONBytes(out)))
		}
		return nil
	}

	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(invoices))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(invoices)))
	}
	return nil
}

func downloadExp(expense *types.Expense) error {
	savePath := viper.GetString(FlagDownloadExp)
	if len(savePath) > 0 {
		savePath = path.Join(savePath, expense.DocFileName)
		err := ioutil.WriteFile(savePath, expense.Document, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// DoQueryProfileCmd is the workhorse of the heavy and light cli query profile commands
func DoQueryProfileCmd(cmd *cobra.Command, args []string,
	queryProfile func(name string) (types.Profile, error)) error {
	if len(args) != 1 {
		return ErrCmdReqArg("name")
	}

	name := args[0]

	profile, err := queryProfile(name)
	if err != nil {
		return err
	}
	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(profile))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(profile)))
	}
	return nil
}

// DoQueryProfilesCmd is the workhorse of the heavy and light cli query profiles commands
func DoQueryProfilesCmd(cmd *cobra.Command, args []string,
	queryListString func(key []byte) ([]string, error)) error {

	var key []byte
	if viper.GetBool(FlagInactive) {
		key = invoicer.ListProfileInactiveKey()
	} else {
		key = invoicer.ListProfileActiveKey()
	}

	listProfiles, err := queryListString(key)
	if err != nil {
		return err
	}
	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(listProfiles))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(listProfiles)))
	}
	return nil
}

// DoQueryPaymentCmd is the workhorse of the heavy and light cli query profile commands
func DoQueryPaymentCmd(cmd *cobra.Command, args []string,
	queryPayment func(transactionID string) (types.Payment, error)) error {

	if len(args) != 1 {
		return ErrCmdReqArg("id")
	}
	transactionID := args[0]

	//get the invoicer object and print it
	payment, err := queryPayment(transactionID)
	if err != nil {
		return err
	}

	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(payment))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(payment)))
	}
	return nil
}

// DoQueryPaymentsCmd is the workhorse of the heavy and light cli query profiles commands
func DoQueryPaymentsCmd(cmd *cobra.Command, args []string,
	queryListString func(key []byte) ([]string, error),
	queryPayment func(transactionID string) (types.Payment, error)) error {

	listPayments, err := queryListString(invoicer.ListPaymentKey())
	if err != nil {
		return err
	}

	//return fmt.Errorf("invoicexz %x\n", listInvoices)
	if len(listPayments) == 0 {
		return fmt.Errorf("No save payments to return") //never stack trace
	}

	//init flag variables
	froms, toes := processFlagFromTo()

	//get the date range to query
	startDate, endDate, err := processFlagDateRange()
	if err != nil {
		return err
	}

	//Loop through the invoices and query out the valid ones
	var payments []types.Payment
	for _, transactionID := range listPayments {

		payment, err := queryPayment(transactionID)
		if err != nil {
			return errors.Errorf("Bad invoice in active invoice list %v \n%v \n%v", transactionID, listPayments, err)
		}

		//skip record if out of the date range
		d := payment.PaymentCurTime.CurTime.Date
		if (startDate != nil && d.Before(*startDate)) ||
			(endDate != nil && d.After(*endDate)) {
			continue
		}

		//continue if doesn't have the sender specified in the from or to flag
		cont := false
		for _, from := range froms {
			if from != payment.Sender {
				cont = true
				break
			}
		}
		for _, to := range toes {
			if to != payment.Sender {
				cont = true
				break
			}
		}
		if cont {
			continue
		}

		//all tests have passed so add to the invoices list
		payments = append(payments, payment)

		//Limit the number of invoices retrieved
		maxInv := viper.GetInt(FlagNum)
		if len(payments) > maxInv && maxInv > 0 {
			break
		}
	}

	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(payments))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(payments)))
	}
	return nil
}
