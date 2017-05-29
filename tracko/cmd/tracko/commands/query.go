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

	bcmd "github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin-examples/tracko/common"
	"github.com/tendermint/basecoin-examples/tracko/plugins/invoicer"
	"github.com/tendermint/basecoin-examples/tracko/types"
)

var (
	//commands
	QueryInvoiceCmd = &cobra.Command{
		Use:   "invoice [id]",
		Short: "Query an invoice by ID",
		RunE:  queryInvoiceCmd,
	}

	QueryInvoicesCmd = &cobra.Command{
		Use:   "invoices",
		Short: "Query all invoice",
		RunE:  queryInvoicesCmd,
	}

	QueryProfileCmd = &cobra.Command{
		Use:   "profile [name]",
		Short: "Query a profile",
		RunE:  queryProfileCmd,
	}

	QueryProfilesCmd = &cobra.Command{
		Use:   "profiles",
		Short: "List all open profiles",
		RunE:  queryProfilesCmd,
	}

	QueryPaymentCmd = &cobra.Command{
		Use:   "payment [id]",
		Short: "List historical payment",
		RunE:  queryPaymentCmd,
	}

	QueryPaymentsCmd = &cobra.Command{
		Use:   "payments",
		Short: "List historical payments",
		RunE:  queryPaymentsCmd,
	}
)

func init() {
	//register flags
	fsDownload := flag.NewFlagSet("", flag.ContinueOnError)
	fsDownload.String(FlagDownloadExp, "", "download expenses pdfs to the relative path specified")

	fsInvoices := flag.NewFlagSet("", flag.ContinueOnError)
	fsInvoices.Int(FlagNum, 0, "number of results to display, use 0 for no limit")
	fsInvoices.String(FlagType, "",
		"limit the scope by using any of the following modifiers with commas: invoice,expense,paid,unpaid")
	fsInvoices.String(FlagDateRange, "",
		"Query within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")
	fsInvoices.String(FlagFrom, "", "Only query for invoices from these addresses in the format <ADDR1>,<ADDR2>, etc.")
	fsInvoices.String(FlagTo, "", "Only query for invoices to these addresses in the format <ADDR1>,<ADDR2>, etc.")
	fsInvoices.Bool(FlagSum, false, "Sum invoice values by sender")

	fsPayments := flag.NewFlagSet("", flag.ContinueOnError)
	fsPayments.Int(FlagNum, 0, "number of results to display, use 0 for no limit")
	fsPayments.String(FlagDateRange, "",
		"Query within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")
	fsPayments.String(FlagFrom, "", "Only query for invoices from these addresses in the format <ADDR1>,<ADDR2>, etc.")
	fsPayments.String(FlagTo, "", "Only query for payments to these addresses in the format <ADDR1>,<ADDR2>, etc.")

	QueryInvoiceCmd.Flags().AddFlagSet(fsDownload)
	QueryInvoicesCmd.Flags().AddFlagSet(fsDownload)
	QueryInvoicesCmd.Flags().AddFlagSet(fsInvoices)
	QueryProfilesCmd.Flags().Bool(FlagInactive, false, "list inactive profiles")
	QueryPaymentsCmd.Flags().AddFlagSet(fsPayments)

	//register commands
	bcmd.RegisterQuerySubcommand(QueryInvoicesCmd)
	bcmd.RegisterQuerySubcommand(QueryInvoiceCmd)
	bcmd.RegisterQuerySubcommand(QueryProfileCmd)
	bcmd.RegisterQuerySubcommand(QueryProfilesCmd)
	bcmd.RegisterQuerySubcommand(QueryPaymentCmd)
	bcmd.RegisterQuerySubcommand(QueryPaymentsCmd)
}

func queryInvoiceCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return errCmdReqArg("id")
	}
	if !isHex(args[0]) {
		return errBadHexID
	}
	id, err := hex.DecodeString(StripHex(args[0]))
	//return errors.Errorf("%x\n%v\n", id, args[0])
	if err != nil {
		return err
	}

	//get the invoicer object and print it
	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()
	invoice, err := queryInvoice(tmAddr, id)
	if err != nil {
		return err
	}

	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(invoice))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(invoice)))
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

func queryInvoicesCmd(cmd *cobra.Command, args []string) error {
	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()
	listInvoices, err := queryListBytes(tmAddr, invoicer.ListInvoiceKey())
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
	contractFilt, expenseFilt, paidFilt, unpaidFilt := true, true, true, true
	if len(ty) > 0 {
		contractFilt, expenseFilt, paidFilt, unpaidFilt = false, false, false, false
		if strings.Contains(ty, "contract") {
			contractFilt = true
		}
		if strings.Contains(ty, "expense") {
			expenseFilt = true
		}
		if strings.Contains(ty, "paid") {
			paidFilt = true
		}
		if strings.Contains(ty, "unpaid") {
			unpaidFilt = true
		}

		//if a whole catagory is missing, turn it on
		if !contractFilt && !expenseFilt {
			contractFilt, expenseFilt = true, true
		}
		if !paidFilt && !unpaidFilt {
			paidFilt, unpaidFilt = true, true
		}
	}

	//get the date range to query
	startDate, endDate, err := processFlagDateRange()
	if err != nil {
		return err
	}

	//Loop through the invoices and query out the valid ones
	var invoices []types.Invoice
	for _, id := range listInvoices {

		invoice, err := queryInvoice(tmAddr, id)
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
		if !((contractFilt && isContract) ||
			(expenseFilt && isExpense)) ||
			!((paidFilt && ctx.Open) ||
				(unpaidFilt && !ctx.Open)) {

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

func queryProfileCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errCmdReqArg("name")
	}

	name := args[0]

	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()
	profile, err := queryProfile(tmAddr, name)
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

func queryProfilesCmd(cmd *cobra.Command, args []string) error {

	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()

	var key []byte
	if viper.GetBool(FlagInactive) {
		key = invoicer.ListProfileInactiveKey()
	} else {
		key = invoicer.ListProfileActiveKey()
	}

	listProfiles, err := queryListString(tmAddr, key)
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

func queryPaymentCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return errCmdReqArg("id")
	}
	transactionID := args[0]

	//get the invoicer object and print it
	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()
	invoice, err := queryPayment(tmAddr, transactionID)
	if err != nil {
		return err
	}

	switch viper.GetString("output") {
	case "text":
		fmt.Println(string(wire.JSONBytes(invoice))) //TODO Actually make text
	case "json":
		fmt.Println(string(wire.JSONBytes(invoice)))
	}
	return nil
}

func queryPaymentsCmd(cmd *cobra.Command, args []string) error {

	//TODO Upgrade to viper once basecoin viper upgrade complete
	tmAddr := cmd.Parent().Flag("node").Value.String()

	listPayments, err := queryListString(tmAddr, invoicer.ListPaymentKey())
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

		payment, err := queryPayment(tmAddr, transactionID)
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

///////////////////////////////////////////////////////////////////

func queryProfile(tmAddr, name string) (profile types.Profile, err error) {

	if len(name) == 0 {
		return profile, errBadQuery("name")
	}
	key := invoicer.ProfileKey(name)

	res, err := query(tmAddr, key)
	if err != nil {
		return profile, err
	}

	return invoicer.GetProfileFromWire(res)
}

func queryInvoice(tmAddr string, id []byte) (invoice types.Invoice, err error) {

	if len(id) == 0 {
		return invoice, errBadQuery("id")
	}

	key := invoicer.InvoiceKey(id)
	res, err := query(tmAddr, key)
	if err != nil {
		return invoice, err
	}

	return invoicer.GetInvoiceFromWire(res)
}

func queryPayment(tmAddr string, transactionID string) (payment types.Payment, err error) {

	if len(transactionID) == 0 {
		return payment, errBadQuery("transactionID")
	}

	key := invoicer.PaymentKey(transactionID)
	res, err := query(tmAddr, key)
	if err != nil {
		return payment, err
	}

	return invoicer.GetPaymentFromWire(res)
}

func queryListString(tmAddr string, key []byte) (profile []string, err error) {
	res, err := query(tmAddr, key)
	if err != nil {
		return profile, err
	}
	return invoicer.GetListStringFromWire(res)
}

func queryListBytes(tmAddr string, key []byte) (invoice [][]byte, err error) {
	res, err := query(tmAddr, key)
	if err != nil {
		return invoice, err
	}
	return invoicer.GetListBytesFromWire(res)
}

//Wrap the basecoin query function with a response code check
func query(tmAddr string, key []byte) ([]byte, error) {
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
