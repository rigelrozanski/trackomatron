package invoicer

import (
	"github.com/pkg/errors"

	abci "github.com/tendermint/abci/types"
)

var (
	errStateNotFound = errors.New("State not found")

	abciErrNoSender           = abci.ErrUnknownRequest.AppendLog("Senders profile doesn't exist")
	abciErrNoReceiver         = abci.ErrUnknownRequest.AppendLog("Receiver profile doesn't exist")
	abciErrProfileNonExistent = abci.ErrUnknownRequest.AppendLog("Cannot modify a non-existent profile")
	abciErrProfileExists      = abci.ErrInternalError.AppendLog("Cannot create an already existing profile")
	abciErrDupInvoice         = abci.ErrInternalError.AppendLog("Duplicate invoice, edit the invoice notes to make them unique")
	abciErrGetProfiles        = abci.ErrUnknownRequest.AppendLog("error retrieving active profile list")
	abciErrGetInvoices        = abci.ErrUnknownRequest.AppendLog("error retrieving active invoice list")
	abciErrInvoiceMissing     = abci.ErrUnknownRequest.AppendLog("error retrieving invoice to modify")
)

func wrapErrDecodingState(err error) error {
	//note will return nil if err is nil
	return errors.Wrap(err, "Error decoding state")
}

func abciErrDecodingTX(err error) abci.Result {
	return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
}
