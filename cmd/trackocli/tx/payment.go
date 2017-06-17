package tx

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	btypes "github.com/tendermint/basecoin/types"
	txcmd "github.com/tendermint/light-client/commands/txs"
	cmn "github.com/tendermint/tmlibs/common"

	trcmn "github.com/tendermint/trackomatron/cmd/trackocli/common"
	"github.com/tendermint/trackomatron/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var PaymentCmd = &cobra.Command{
	Use:   "payment [receiver]",
	Short: "Pay invoices and expenses with transaction infomation",
	RunE:  paymentCmd,
}

func init() {
	FSTxPayment := flag.NewFlagSet("", flag.ContinueOnError)
	FSTxPayment.String(trcmn.FlagIDs, "", "IDs to close during this transaction <id1>,<id2>,<id3>... ")
	FSTxPayment.String(trcmn.FlagTransactionID, "", "Completed transaction ID")
	FSTxPayment.String(trcmn.FlagPaid, "", "Payment amount in the format <decimal><currency> eg. 10.23usd")
	FSTxPayment.String(trcmn.FlagDate, "", "Date payment in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	FSTxPayment.String(trcmn.FlagDateRange, "",
		"Autoselect IDs within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")

	PaymentCmd.Flags().AddFlagSet(FSTxPayment)
}

func paymentCmd(cmd *cobra.Command, args []string) error {
	// Read the standard app-tx flags
	gas, fee, txInput, err := bcmd.ReadAppTxFlags()
	if err != nil {
		return err
	}

	// Retrieve the app-specific flags/args
	var receiver string
	if len(args) != 1 {
		return trcmn.ErrCmdReqArg("receiver")
	}
	receiver = args[0]

	data, err := paymentTx(txInput.Address, receiver)
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

// paymentTx Generates the tendermint TX used by the light and heavy client
func paymentTx(senderAddr []byte, receiver string) ([]byte, error) {

	flagIDs := viper.GetString(trcmn.FlagIDs)
	flagDateRange := viper.GetString(trcmn.FlagDateRange)

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
				return nil, trcmn.ErrBadHexID
			}
			id, err := hex.DecodeString(cmn.StripHex(idHex))
			if err != nil {
				return nil, err
			}
			ids = append([][]byte{id}, ids...)
		}
	}

	date, err := common.ParseDate(viper.GetString(trcmn.FlagDate))
	if err != nil {
		return nil, err
	}
	amt, err := types.ParseAmtCurTime(viper.GetString(trcmn.FlagPaid), date)
	if err != nil {
		return nil, err
	}

	tx := types.TxPayment{
		TransactionID: viper.GetString(trcmn.FlagTransactionID),
		SenderAddr:    senderAddr,
		IDs:           ids,
		Receiver:      receiver,
		Amt:           amt,
		StartDate:     startDate,
		EndDate:       endDate,
	}

	return invoicer.MarshalWithTB(tx, invoicer.TBTxPayment), nil
}
