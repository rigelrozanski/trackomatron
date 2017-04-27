package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin-examples/invoicer/plugins/invoicer"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

var (
	//flags
	num         int
	short       bool
	typeflag    string
	from        string //list of addresses by commas
	to          string //list of addresses by commas
	date        string // with colon start:end, :start, or end:
	downloadExp string //path to download expenses

	//commands
	QueryInvoiceCmd = &cobra.Command{
		Use:   "invoice [hexID]",
		Short: "Query an invoice by invoice ID",
		RunE:  queryInvoiceCmd,
	}

	QueryInvoicesCmd = &cobra.Command{
		Use:   "invoices",
		Short: "Query all invoice",
		RunE:  queryInvoicesCmd,
	}
)

func init() {
	//register flags

	downloadExpFlag := bcmd.Flag2Register{&download, "download-expenses", false, "download expenses pdfs to the relative path specified"}

	invoiceFlags := []bcmd.Flag2Register{
		downloadExpFlag,
	}

	invoicesFlags := []bcmd.Flag2Register{
		{&num, "n", 0, "number of results to display, use 0 for no limit"},
		{&short, "short", false, "output fields: paid, amount, date, sender, receiver"},
		{&typeflg, "type", "",
			"limit the scope by using any of the following modifiers with commas: invoice,expense,paid,unpaid"},
		{&date, "date", "",
			"Query within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:"},
		{&from, "from", "", "only query for invoices from these addresses in the format <ADDR1>,<ADDR2>, etc."},
		{&to, "to", "", "only query for invoices to these addresses in the format <ADDR1>,<ADDR2>, etc."},
		downloadExpFlag,
	}

	bcmd.RegisterFlags(QueryInvoiceCmd, invoiceFlags)
	bcmd.RegisterFlags(QueryInvoicesCmd, invoicesFlags)

	//register commands
	bcmd.RegisterQuerySubcommand(QueryInvoicesCmd)
	bcmd.RegisterQuerySubcommand(QueryInvoiceCmd)
}

func queryInvoiceCmd(cmd *cobra.Command, args []string) error {

	//get the parent context
	parentContext := cmd.Parent()

	//get the issue, generate issue key
	if len(args) != 1 {
		return fmt.Errorf("query command requires an argument ([hexID])") //never stack trace
	}
	hexID := args[0]
	issueKey := invoicer.InvoiceKey(issue)

	//perform the query, get response
	resp, err := bcmd.Query(parentContext.Flag("node").Value.String(), issueKey)
	if err != nil {
		return err
	}
	if !resp.Code.IsOK() {
		return errors.Errorf("Query for issueKey (%v) returned non-zero code (%v): %v",
			string(issueKey), resp.Code, resp.Log)
	}

	//get the invoicer issue object and print it
	p2vIssue, err := invoicer.GetIssueFromWire(resp.Value)
	if err != nil {
		return err
	}
	fmt.Println(string(wire.JSONBytes(p2vIssue)))
	return nil
}

func queryInvoicesCmd(cmd *cobra.Command, args []string) error {
}
