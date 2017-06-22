package query

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"

	trcmn "github.com/tendermint/trackomatron/cmd/trackocli/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var (
	QueryPaymentCmd = &cobra.Command{
		Use:          "payment [id]",
		Short:        "List historical payment",
		SilenceUsage: true,
		RunE:         queryPaymentCmd,
	}

	QueryPaymentsCmd = &cobra.Command{
		Use:          "payments",
		Short:        "List historical payments",
		SilenceUsage: true,
		RunE:         queryPaymentsCmd,
	}
)

func init() {

	FSQueryPayments := flag.NewFlagSet("", flag.ContinueOnError)
	FSQueryPayments.Int(trcmn.FlagNum, 0, "Number of results to display, use 0 for no limit")
	FSQueryPayments.String(trcmn.FlagDateRange, "",
		"Query within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")
	FSQueryPayments.String(trcmn.FlagFrom, "", "Only query for invoices from these addresses in the format <ADDR1>,<ADDR2>, etc.")
	FSQueryPayments.String(trcmn.FlagTo, "", "Only query for payments to these addresses in the format <ADDR1>,<ADDR2>, etc.")

	QueryPaymentsCmd.Flags().AddFlagSet(FSQueryPayments)
}

func queryPayment(transactionID string) (payment types.Payment, err error) {
	key := invoicer.PaymentKey(transactionID)
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetPaymentFromWire(proof.Data())
}

// DoQueryPaymentCmd is the workhorse of the heavy and light cli query profile commands
func queryPaymentCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return trcmn.ErrCmdReqArg("transactionID")
	}
	transactionID := args[0]

	//get the invoicer object and print it
	key := invoicer.PaymentKey(transactionID)
	proof, err := getProof(key)
	if err != nil {
		return err
	}
	payment, err := invoicer.GetPaymentFromWire(proof.Data())
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
func queryPaymentsCmd(cmd *cobra.Command, args []string) error {

	key := invoicer.ListPaymentKey()
	proof, err := getProof(key)
	if err != nil {
		return err
	}
	listPayments, err := invoicer.GetListStringFromWire(proof.Data())
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

		key := invoicer.PaymentKey(transactionID)
		proof, err := getProof(key)
		if err != nil {
			return err
		}
		payment, err := invoicer.GetPaymentFromWire(proof.Data())
		if err != nil {
			return errors.Errorf("Bad invoice in active invoice list %v \n%v \n%v", transactionID, listPayments, err)
		}

		//skip record if out of the date range
		d := payment.PaymentCurTime.CurTime.Date
		if (!startDate.IsZero() && d.Before(startDate)) ||
			(!endDate.IsZero() && d.After(endDate)) {
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
		maxInv := viper.GetInt(trcmn.FlagNum)
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
