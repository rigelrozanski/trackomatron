package commands

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/trackomatron/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var (
//Exposed flagsets
)

func init() {

	//register flags
	FSTxProfile.String(FlagTo, "", "Who you're invoicing")
	FSTxProfile.String(FlagCur, "BTC", "Payment curreny accepted")
	FSTxProfile.String(FlagDepositInfo, "", "Default deposit information to be provided")
	FSTxProfile.Int(FlagDueDurationDays, 14, "Default number of days until invoice is due from invoice submission")

	FSTxPayment.String(FlagIDs, "", "IDs to close during this transaction <id1>,<id2>,<id3>... ")
	FSTxPayment.String(FlagTransactionID, "", "Completed transaction ID")
	FSTxPayment.String(FlagPaid, "", "Payment amount in the format <decimal><currency> eg. 10.23usd")
	FSTxPayment.String(FlagDate, "", "Date payment in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	FSTxPayment.String(FlagDateRange, "",
		"Autoselect IDs within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")

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
