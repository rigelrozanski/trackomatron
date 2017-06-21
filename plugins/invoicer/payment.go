package invoicer

import (
	"fmt"
	"time"

	abci "github.com/tendermint/abci/types"
	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"

	types "github.com/tendermint/trackomatron/types"
)

func validatePayment(ctx types.Context) abci.Result {
	//Validate Tx
	switch {
	case len(ctx.Sender) == 0:
		return abci.ErrInternalError.AppendLog("Invoice must have a sender")
	case len(ctx.Receiver) == 0:
		return abci.ErrInternalError.AppendLog("Invoice must have a receiver")
	case len(ctx.AcceptedCur) == 0:
		return abci.ErrInternalError.AppendLog("Invoice must have an accepted currency")
	case ctx.Payable == nil:
		return abci.ErrInternalError.AppendLog("Invoice amount is nil")
	case ctx.Due.Before(time.Now()):
		return abci.ErrInternalError.AppendLog("Cannot issue overdue invoice")
	default:
		return abci.OK
	}
}

func runTxPayment(store btypes.KVStore, txBytes []byte) (res abci.Result) {

	// Decode tx
	var tx = new(types.TxPayment)
	err := wire.ReadBinaryBytes(txBytes[1:], tx)
	if err != nil {
		return abciErrDecodingTX(err)
	}

	//get the sender's address
	profile, err := getProfileFromAddress(store, tx.SenderAddr)
	if err != nil {
		abciErrInternal(err)
	}
	sender := profile.Name

	payment := types.NewPayment(
		tx.IDs,
		tx.TransactionID,
		sender,
		tx.Receiver,
		tx.Amt,
		tx.StartDate,
		tx.EndDate,
	)

	//If there are no IDs provided in payment tx
	// then populate them based on date
	if len(payment.InvoiceIDs) == 0 {
		listInvoices, err := getListBytes(store, ListInvoiceKey())
		if err != nil {
			return abci.ErrInternalError.AppendLog(err.Error())
		}
		for _, id := range listInvoices {

			invoice, err := getInvoice(store, id)
			if err != nil {
				return abci.ErrInternalError.AppendLog(
					fmt.Sprintf("Bad invoice in active invoice list %x \n%x \n%v", id, listInvoices, err))
			}
			ctx := invoice.GetCtx()

			//skip record if out of the date range
			d := ctx.Invoiced.CurTime.Date
			if (!payment.StartDate.IsZero() && d.Before(payment.StartDate)) ||
				(!payment.EndDate.IsZero() && d.After(payment.EndDate)) {
				continue
			}

			payment.InvoiceIDs = append(payment.InvoiceIDs, invoice.GetID())
		}
	}

	//Validate Tx
	switch {
	case len(payment.InvoiceIDs) == 0:
		return abci.ErrInternalError.AppendLog("Payment doesn't contain any IDs to close!")
	case len(payment.TransactionID) == 0:
		return abci.ErrInternalError.AppendLog("Payment must include a transaction ID")
	}

	//Get all invoices, verify the ID
	var invoices []*types.Invoice
	for _, invoiceID := range payment.InvoiceIDs {
		invoice, err := getInvoice(store, invoiceID)
		if err != nil {
			return abciErrInvoiceMissing
		}
		invoices = append([]*types.Invoice{&invoice}, invoices...)
		if invoice.GetCtx().Sender != payment.Receiver {
			return abci.ErrInternalError.AppendLog(
				fmt.Sprintf("Invoice ID %x has receiver %v but the payment is to receiver %v!",
					invoice.GetID(),
					invoice.GetCtx().Receiver,
					payment.Receiver))
		}
	}

	//Make sure that the invoice is not paying too much!
	var totalCost *types.AmtCurTime
	for _, invoice := range invoices {
		unpaid, err := invoice.GetCtx().Unpaid()
		if err != nil {
			return abciErrDecimal(err)
		}
		totalCost, err = totalCost.Add(unpaid)
		if err != nil {
			return abciErrDecimal(err)
		}
	}
	gt, err := payment.PaymentCurTime.GT(totalCost)
	if err != nil {
		return abciErrDecimal(err)
	}
	if gt {
		return abciErrOverPayment
	}

	//calculate and write changes to the set of all invoices
	bal := payment.PaymentCurTime
	for _, invoice := range invoices {
		//pay the funds to the invoice, reduce funds from bal
		bal, err = invoice.GetCtx().Pay(bal)
		if err != nil {
			return abci.ErrUnauthorized.AppendLog("Error paying invoice: " + err.Error())
		}
		store.Set(InvoiceKey(invoice.GetID()), wire.BinaryBytes(*invoice))
	}

	//add the payment object to the store
	store.Set(PaymentKey(payment.TransactionID), wire.BinaryBytes(*payment))
	payments, err := getListString(store, ListPaymentKey())
	if err != nil {
		return abciErrGetPayments
	}
	payments = append(payments, payment.TransactionID)
	store.Set(ListPaymentKey(), wire.BinaryBytes(payments))

	return abci.OK
}
