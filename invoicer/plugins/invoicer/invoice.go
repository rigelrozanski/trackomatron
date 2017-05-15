package invoicer

import (
	"bytes"
	"time"

	abci "github.com/tendermint/abci/types"
	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin-examples/invoicer/types"
)

func validateInvoiceCtx(ctx types.Context) abci.Result {
	//Validate Tx
	switch {
	case len(ctx.Sender) == 0:
		return abci.ErrInternalError.AppendLog("invoice must have a sender")
	case len(ctx.Receiver) == 0:
		return abci.ErrInternalError.AppendLog("invoice must have a receiver")
	case len(ctx.AcceptedCur) == 0:
		return abci.ErrInternalError.AppendLog("invoice must have an accepted currency")
	case ctx.Amount == nil:
		return abci.ErrInternalError.AppendLog("invoice amount is nil")
	case ctx.Due.Before(time.Now()):
		return abci.ErrInternalError.AppendLog("cannot issue overdue invoice")
	default:
		return abci.OK
	}
}

func runTxInvoice(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte, shouldExist bool) (res abci.Result) {

	// Decode tx
	var reader = new(types.Invoice)
	err := wire.ReadBinaryBytes(txBytes, reader)
	if err != nil {
		return abciErrDecodingTX(err)
	}

	invoice := *reader

	//Validate
	res = validateInvoiceCtx(invoice.GetCtx())
	if res.IsErr() {
		return res
	}

	invoices, err := getListInvoice(store)
	if err != nil {
		return abciErrGetInvoices
	}

	//remove before editing, invoice.ID will be empty if not editing
	if len(invoice.GetID()) > 0 {
		found := false
		for i, v := range invoices {
			if bytes.Compare(v, invoice.GetID()) == 0 {
				invoices = append(invoices[:i], invoices[i+1:]...)
				found = true
				break
			}
		}
		if found {
			store.Set(ListInvoiceKey(), wire.BinaryBytes(invoices))
		} else {
			return abciErrInvoiceMissing
		}
	}

	//Set the id, then validate a bit more
	invoice.SetID()

	if _, err := getProfile(store, invoice.GetCtx().Sender); err != nil {
		return abciErrNoSender
	}
	if _, err := getProfile(store, invoice.GetCtx().Receiver); err != nil {
		return abciErrNoReceiver
	}

	//Return if the invoice already exists, aka no error was thrown
	_, err = getInvoice(store, invoice.GetID())
	if shouldExist && err != nil {
		return abciErrInvoiceMissing
	}
	if !shouldExist && err == nil {
		return abciErrDupInvoice
	}

	//Store invoice
	store.Set(InvoiceKey(invoice.GetID()), wire.BinaryBytes(invoice))

	invoices = append(invoices, invoice.GetID())
	store.Set(ListInvoiceKey(), wire.BinaryBytes(invoices))
	return abci.OK
}

func runTxCloseInvoice(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte) (res abci.Result) {

	// Decode tx
	var close = new(types.CloseInvoice)
	err := wire.ReadBinaryBytes(txBytes, close)
	if err != nil {
		return abciErrDecodingTX(err)
	}

	//Validate Tx
	switch {
	case len(close.ID) == 0:
		return abci.ErrInternalError.AppendLog("closer doesn't have an ID")
	case len(close.TransactionID) == 0:
		return abci.ErrInternalError.AppendLog("closer must include a transaction ID")
	}

	//actually write the changes
	invoice, err := getInvoice(store, close.ID)
	if err != nil {
		return abciErrInvoiceMissing
	}
	invoice.Close(close)

	store.Set(InvoiceKey(invoice.GetID()), wire.BinaryBytes(invoice))

	return abci.OK
}

//TODO add JSON imports
func runTxBulkImport(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte) (res abci.Result) {
	return abci.OK //TODO add functionality
}
