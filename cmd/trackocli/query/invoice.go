package query

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

	trcmn "github.com/tendermint/trackomatron/cmd/trackocli/common"
	"github.com/tendermint/trackomatron/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var (
	QueryInvoiceCmd = &cobra.Command{
		Use:          "invoice [id]",
		Short:        "Query an invoice by ID",
		SilenceUsage: true,
		RunE:         queryInvoiceCmd,
	}

	QueryInvoicesCmd = &cobra.Command{
		Use:          "invoices",
		Short:        "Query all invoice",
		SilenceUsage: true,
		RunE:         queryInvoicesCmd,
	}
)

func init() {
	//register flags
	FSQueryDownload := flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryInvoice := flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryInvoices := flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryDownload.String(trcmn.FlagDownloadExp, "", "Download expenses pdfs to the relative path specified")

	FSQueryInvoices.Int(trcmn.FlagNum, 0, "Number of results to display, use 0 for no limit")
	FSQueryInvoices.String(trcmn.FlagType, "",
		"Limit the scope by using any of the following modifiers with commas: invoice,expense,open,closed")
	FSQueryInvoices.String(trcmn.FlagDateRange, "",
		"Query within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")
	FSQueryInvoices.String(trcmn.FlagFrom, "", "Only query for invoices from these addresses in the format <ADDR1>,<ADDR2>, etc.")
	FSQueryInvoices.String(trcmn.FlagTo, "", "Only query for invoices to these addresses in the format <ADDR1>,<ADDR2>, etc.")
	FSQueryInvoices.Bool(trcmn.FlagSum, false, "Sum invoice values by sender")
	FSQueryInvoices.Bool(trcmn.FlagLedger, false, "open a Ledger Wallet Bitcoin with Sum Amount details filled in")

	FSQueryInvoice.Bool(trcmn.FlagLedger, false, "open a Ledger Wallet Bitcoin with transaction details filled in")

	QueryInvoiceCmd.Flags().AddFlagSet(FSQueryDownload)
	QueryInvoiceCmd.Flags().AddFlagSet(FSQueryInvoice)
	QueryInvoicesCmd.Flags().AddFlagSet(FSQueryDownload)
	QueryInvoicesCmd.Flags().AddFlagSet(FSQueryInvoices)
}

// DoQueryInvoiceCmd is the workhorse of the heavy and light cli query profile commands
func queryInvoiceCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return trcmn.ErrCmdReqArg("id")
	}
	if !cmn.IsHex(args[0]) {
		return trcmn.ErrBadHexID
	}
	id, err := hex.DecodeString(cmn.StripHex(args[0]))
	if err != nil {
		return err
	}

	key := invoicer.InvoiceKey(id)
	proof, err := getProof(key)
	if err != nil {
		return err
	}

	invoice, err := invoicer.GetInvoiceFromWire(proof.Data())
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

// DoQueryInvoicesCmd is the workhorse of the heavy and light cli query profiles commands
func queryInvoicesCmd(cmd *cobra.Command, args []string) error {

	key := invoicer.ListInvoiceKey()
	proof, err := getProof(key)
	if err != nil {
		return err
	}
	listInvoices, err := invoicer.GetListBytesFromWire(proof.Data())
	if err != nil {
		return err
	}

	//return fmt.Errorf("invoicexz %x\n", listInvoices)
	if len(listInvoices) == 0 {
		return fmt.Errorf("No save invoices to return") //never stack trace
	}

	//init flag variables
	froms, toes := processFlagFromTo()

	ty := viper.GetString(trcmn.FlagType)
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

		key := invoicer.InvoiceKey(id)
		proof, err := getProof(key)
		if err != nil {
			return err
		}

		invoice, err := invoicer.GetInvoiceFromWire(proof.Data())
		if err != nil {
			return errors.Errorf("Bad invoice in active invoice list %x \n%v \n%v", id, listInvoices, err)
		}

		ctx := invoice.GetCtx()

		//skip record if out of the date range
		d := ctx.Invoiced.CurTime.Date
		if (!startDate.IsZero() && d.Before(startDate)) ||
			(!endDate.IsZero() && d.After(endDate)) {
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
		maxInv := viper.GetInt(trcmn.FlagNum)
		if len(invoices) > maxInv && maxInv > 0 {
			break
		}
	}

	//compute the sum if flag is set
	if viper.GetBool(trcmn.FlagSum) {
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

func processFlagFromTo() (froms, toes []string) {
	from := viper.GetString(trcmn.FlagFrom)
	to := viper.GetString(trcmn.FlagTo)
	if len(froms) > 0 {
		froms = strings.Split(from, ",")
	}
	if len(toes) > 0 {
		toes = strings.Split(to, ",")
	}
	return
}

func processFlagDateRange() (startDate, endDate time.Time, err error) {
	flagDateRange := viper.GetString(trcmn.FlagDateRange)
	if len(flagDateRange) > 0 {
		startDate, endDate, err = common.ParseDateRange(flagDateRange)
		if err != nil {
			return
		}
	}
	return
}

func downloadExp(expense *types.Expense) error {
	savePath := viper.GetString(trcmn.FlagDownloadExp)
	if len(savePath) > 0 {
		savePath = path.Join(savePath, expense.DocFileName)
		err := ioutil.WriteFile(savePath, expense.Document, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
