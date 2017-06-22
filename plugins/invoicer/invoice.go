package invoicer

import (
	"bytes"
	"io/ioutil"
	"path"
	"time"

	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"

	"github.com/tendermint/trackomatron/common"
	"github.com/tendermint/trackomatron/types"
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

func runTxInvoice(store btypes.KVStore, txBytes []byte) (res abci.Result) {

	tb := txBytes[0]

	// Decode tx
	var tx = new(types.TxInvoice)
	err := wire.ReadBinaryBytes(txBytes[1:], tx)
	if err != nil {
		return abciErrDecodingTX(err)
	}

	//get the sender's address
	profile, err := getProfileFromAddress(store, tx.SenderAddr)
	if err != nil {
		return abciErrInternal(err)
	}
	sender := profile.Name

	var accCur string
	if len(tx.Cur) > 0 {
		accCur = tx.Cur
	} else {
		accCur = profile.AcceptedCur
	}

	date := time.Now()
	if len(tx.Date) > 0 {
		date, err = time.Parse(common.TimeLayout, tx.Date)
		if err != nil {
			return abciErrInternal(err)
		}
	}
	amt, err := types.ParseAmtCurTime(tx.Amount, date)
	if err != nil {
		return abciErrInternal(err)
	}

	//calculate payable amount based on invoiced and accepted cur
	payable, err := common.ConvertAmtCurTime(accCur, amt)
	if err != nil {
		return abciErrInternal(err)
	}

	//retrieve flags, or if they aren't used, use the senders profile's default

	var dueDate time.Time
	if len(tx.DueDate) > 0 {
		date, err = time.Parse(common.TimeLayout, tx.DueDate)
		if err != nil {
			return abciErrInternal(err)
		}
	} else {
		dueDate = time.Now().AddDate(0, 0, profile.DueDurationDays)
	}

	var depositInfo string
	if len(tx.DepositInfo) > 0 {
		depositInfo = tx.DepositInfo
	} else {
		depositInfo = profile.DepositInfo
	}

	var invoice types.Invoice

	switch tb {
	//if not an expense then we're almost done!
	case TBTxContractOpen, TBTxContractEdit:
		invoice = types.NewContract(
			tx.EditID,
			sender,
			tx.To,
			depositInfo,
			tx.Notes,
			accCur,
			dueDate,
			amt,
			payable,
		).Wrap()
	case TBTxExpenseOpen, TBTxExpenseEdit:

		taxes, err := types.ParseAmtCurTime(tx.TaxesPaid, date)
		if err != nil {
			return abciErrInternal(err)
		}
		docBytes, err := ioutil.ReadFile(tx.Receipt)
		if err != nil {
			return abciErrInternal(errors.Wrap(err, "Problem reading receipt file"))
		}
		_, filename := path.Split(tx.Receipt)

		invoice = types.NewExpense(
			tx.EditID,
			sender,
			tx.To,
			depositInfo,
			tx.Notes,
			accCur,
			dueDate,
			amt,
			payable,
			docBytes,
			filename,
			taxes,
		).Wrap()
	default:
		return abciErrBadTypeByte
	}

	switch tb {
	case TBTxContractOpen, TBTxExpenseOpen:
		return runActionInvoice(store, invoice, false)
	case TBTxContractEdit, TBTxExpenseEdit:
		return runActionInvoice(store, invoice, true)
	}
	return abciErrBadTypeByte
}

func runActionInvoice(store btypes.KVStore, invoice types.Invoice, shouldExist bool) (res abci.Result) {

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
