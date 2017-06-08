package commands

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/trackomatron/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var (
	//Exposed flagsets
	FSTxProfile     = flag.NewFlagSet("", flag.ContinueOnError)
	FSTxInvoice     = flag.NewFlagSet("", flag.ContinueOnError)
	FSTxExpense     = flag.NewFlagSet("", flag.ContinueOnError)
	FSTxPayment     = flag.NewFlagSet("", flag.ContinueOnError)
	FSTxInvoiceEdit = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {

	//register flags
	FSTxProfile.String(FlagTo, "", "Who you're invoicing")
	FSTxProfile.String(FlagCur, "BTC", "Payment curreny accepted")
	FSTxProfile.String(FlagDepositInfo, "", "Default deposit information to be provided")
	FSTxProfile.Int(FlagDueDurationDays, 14, "Default number of days until invoice is due from invoice submission")

	FSTxInvoice.String(FlagTo, "allinbits", "Name of the invoice receiver")
	FSTxInvoice.String(FlagDepositInfo, "", "Deposit information for invoice payment (default: profile)")
	FSTxInvoice.String(FlagNotes, "", "Notes regarding the expense")
	FSTxInvoice.String(FlagCur, "", "Currency which invoice should be paid in")
	FSTxInvoice.String(FlagDate, "", "Invoice demon date in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	FSTxInvoice.String(FlagDueDate, "", "Invoice due date in the format YYYY-MM-DD eg. 2016-12-31 (default: profile)")

	FSTxExpense.String(FlagReceipt, "", "Directory to receipt document file")
	FSTxExpense.String(FlagTaxesPaid, "", "Taxes amount in the format <decimal><currency> eg. 10.23usd")

	FSTxPayment.String(FlagIDs, "", "IDs to close during this transaction <id1>,<id2>,<id3>... ")
	FSTxPayment.String(FlagTransactionID, "", "Completed transaction ID")
	FSTxPayment.String(FlagPaid, "", "Payment amount in the format <decimal><currency> eg. 10.23usd")
	FSTxPayment.String(FlagDate, "", "Date payment in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	FSTxPayment.String(FlagDateRange, "",
		"Autoselect IDs within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")

	FSTxInvoiceEdit.String(FlagID, "", "ID (hex) of the invoice to modify")
}

// ProfileTx Generates the tendermint TX used by the light and heavy client
func ProfileTx(TBTx byte, address []byte, name string) []byte {
	tx := types.TxProfile{
		Address:         address,
		Name:            name,
		AcceptedCur:     viper.GetString(FlagCur),
		DepositInfo:     viper.GetString(FlagDepositInfo),
		DueDurationDays: viper.GetInt(FlagDueDurationDays),
	}
	return invoicer.MarshalWithTB(tx, TBTx)
}

//////////////////////////////////////////////////////////////////////////

// InvoiceTx Generates the tendermint TX used by the light and heavy client
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

//////////////////////////////////////////////////////////////////////////

// PaymentTx Generates the tendermint TX used by the light and heavy client
func PaymentTx(senderAddr []byte, receiver string) ([]byte, error) {

	flagIDs := viper.GetString(FlagIDs)
	flagDateRange := viper.GetString(FlagDateRange)

	if len(flagIDs) > 0 && len(flagDateRange) > 0 {
		return nil, errors.New("Cannot use both the IDs flag and date-range flag")
	}
	if len(flagIDs) == 0 && len(flagDateRange) == 0 {
		return nil, errors.New("Must include an IDs flag or date-range flag")
	}

	//Get the date range or list of IDs
	var ids [][]byte
	var startDate, endDate *time.Time = nil, nil
	if len(flagDateRange) > 0 {
		var err error
		startDate, endDate, err = common.ParseDateRange(flagDateRange)
		if err != nil {
			return nil, err
		}
	} else {
		idsStr := strings.Split(flagIDs, ",")
		for _, idHex := range idsStr {
			if !cmn.IsHex(idHex) {
				return nil, ErrBadHexID
			}
			id, err := hex.DecodeString(cmn.StripHex(idHex))
			if err != nil {
				return nil, err
			}
			ids = append([][]byte{id}, ids...)
		}
	}

	date, err := common.ParseDate(viper.GetString(FlagDate))
	if err != nil {
		return nil, err
	}
	amt, err := types.ParseAmtCurTime(viper.GetString(FlagPaid), date)
	if err != nil {
		return nil, err
	}

	tx := types.TxPayment{
		TransactionID: viper.GetString(FlagTransactionID),
		SenderAddr:    senderAddr,
		IDs:           ids,
		Receiver:      receiver,
		Amt:           amt,
		StartDate:     startDate,
		EndDate:       endDate,
	}

	return invoicer.MarshalWithTB(tx, invoicer.TBTxPayment), nil
}
