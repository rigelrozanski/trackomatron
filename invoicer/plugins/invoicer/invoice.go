package invoicer

import (
	"bytes"
	"time"

	abci "github.com/tendermint/abci/types"
	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin-examples/invoicer/types"
)

func validateInvoiceCtx(ctx *types.Context) abci.Result {
	//Validate Tx
	switch {
	case len(ctx.Sender) == 0:
		return abci.ErrInternalError.AppendLog("invoice must have a sender")
	case len(ctx.Receiver) == 0:
		return abci.ErrInternalError.AppendLog("invoice must have a receiver")
	case len(ctx.AcceptedCur) == 0:
		return abci.ErrInternalError.AppendLog("invoice must have an accepted currency")
	case ctx.Payable == nil:
		return abci.ErrInternalError.AppendLog("invoice amount is nil")
	case ctx.Due.Before(time.Now()):
		return abci.ErrInternalError.AppendLog("cannot issue overdue invoice")
	default:
		return abci.OK
	}
}

func runTxInvoice(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte, shouldExist bool) (res abci.Result) {

	// Decode the new invoice from cli
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

	invoices, err := getListBytes(store, ListInvoiceKey())
	if err != nil {
		return abciErrGetInvoices
	}

	//Remove before editing, invoice.ID will be empty if not editing
	if len(invoice.GetID()) > 0 {
		found := false

		for i, v := range invoices {
			if bytes.Compare(v, invoice.GetID()) == 0 {

				//Can only edit if the current invoice is still open
				storeInvoice, err := getInvoice(store, v)
				if err != nil {
					return abciErrInvoiceClosed
				}
				if !storeInvoice.GetCtx().Open {
					return abciErrInvoiceClosed
				}

				invoices = append(invoices[:i], invoices[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return abciErrInvoiceMissing
		}

		store.Set(ListInvoiceKey(), wire.BinaryBytes(invoices))
	}

	//Set the id if it doesn't yet exist
	if len(invoice.GetID()) == 0 {
		invoice.SetID()
	}

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
