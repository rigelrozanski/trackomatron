package types

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/shopspring/decimal"
)

const (
	TBTxNewProfile  byte = 0x01
	TBTxOpenInvoice byte = 0x02
	TBTxOpenExpense byte = 0x03
	TBTxClose       byte = 0x04
)

//////////////////////////////

type Currency string

type CurTime struct {
	cur  currency
	date time.Time
}

type AmtCurTime struct {
	cur    curTime
	amount decimal
}

///////////////////////////////

type Profile struct {
	Name               string        //identifier for querying
	AcceptedCur        currency      //currency you will accept payment in
	DefaultDepositInfo string        //default deposit information (mostly for fiat)
	DueDurationDays    int           //default duration until a sent invoice due date
	Timezone           time.Location //default duration until a sent invoice due date
}

func NewProfile(Name string, AcceptedCur currency,
	DefaultDepositInfo string, DueDurationDays int) Profile {
	return Proflie{
		Name:               Name,
		AcceptedCur:        AcceptedCur,
		DefaultDepositInfo: DefaultDepositInfo,
		DueDurationDays:    DueDurationDays,
	}
}

func NewTxBytesNewProfile(Name string, AcceptedCur currency,
	DefaultDepositInfo string, DueDurationDays int) []byte {

	data := wire.BinaryBytes(NewProfile(Address, Nickname, LegalName,
		AcceptedCur, DefaultDepositInfo, DueDurationDays))
	data = append([]byte{TBTxNewProfile}, data...)
	return data
}

type Invoice struct {
	ID             int
	Sender         string
	Receiver       string
	DepositInfo    string
	Amount         *AmtCurTime
	AcceptedCur    Currency
	TransactionID  string      //empty when unpaid
	PaymentCurTime *AmtCurTime //currency used to pay invoice, empty when unpaid
}

func NewInvoice(ID int, Sender []byte, Receiver []byte, DepositInfo string,
	Amount AmtCurTime, AcceptedCur Currency) Invoice {
	return Invoice{
		ID:             ID,
		Sender:         Sender,
		Receiver:       Receiver,
		DepositInfo:    DepositInfo,
		Amount:         Amount,
		AcceptedCur:    AcceptedCur,
		TransactionID:  "",
		PaymentCurTime: nil,
	}
}

func NewTxBytesOpenInvoice(ID int, AccSender []byte, AccReceiver []byte, DepositInfo string,
	Amount *AmtCurTime, AcceptedCur Currency, TransactionID string, PaymentCurTime *AmtCurTime) []byte {

	data := wire.BinaryBytes(NewInvoice(ID, AccSender, AccReceiver, DepositInfo,
		Amount, AcceptedCur, TransactionID, PaymentCurTime))
	data = append([]byte{TBTxOpenInvoice}, data...)
	return data
}

type Expense struct {
	Invoice
	PDFReceipt  []byte
	PDFFileName string
	Notes       string
	TaxesPaid   AmtCurTime
}

func NewExpense(ID int, AccSender []byte, AccReceiver []byte, DepositInfo string,
	Amount AmtCurTime, AcceptedCur []Currency, TransactionID string, PaymentCurTime *AmtCurTime,
	pdfReceipt []byte, notes string, taxesPaid *AmtCurTime) Expense {

	return Expense{
		ID:             ID,
		AccSender:      AccSender,
		AccReceiver:    AccReceiver,
		DepositInfo:    DepositInfo,
		Amount:         Amount,
		AcceptedCur:    AcceptedCur,
		TransactionID:  TransactionID,
		PaymentCurTime: PaymentCurTime,
		pdfReceipt:     pdfReceipt,
		notes:          notes,
		taxesPaid:      taxesPaid,
	}
}

func NewTxBytesOpenExpense(ID int, AccSender []byte, AccReceiver []byte, DepositInfo string,
	Amount AmtCurTime, AcceptedCur []Currency, TransactionID string, PaymentCurTime *AmtCurTime,
	pdfReceipt []byte, notes string, taxesPaid *AmtCurTime) []byte {

	data := wire.BinaryBytes(NewExpense(ID, AccSender, AccReceiver, DepositInfo,
		Amount, AcceptedCur, TransactionID, PaymentCurTime,
		pdfReceipt, notes, taxesPaid))
	data = append([]byte{TBTxOpenExpense}, data...)
	return data
}

type Close struct {
	ID             int
	TransactionID  string      //empty when unpaid
	PaymentCurTime *AmtCurTime //currency used to pay invoice, empty when unpaid
}

func NewClose(ID int, TransactionID string, PaymentCurTime *AmtCurTime) Close {
	return Close{
		ID:             ID,
		TransactionID:  TransactionID,
		PaymentCurTime: PaymentCurTime,
	}

}

func NewTxBytesClose(ID int) []byte {
	data := wire.BinaryBytes(NewClose(ID))
	data = append([]byte{TBTxClose}, data...)
	return data
}
