package types

import (
	"time"

	"github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/merkle"
)

const (
	TBIDExpense = iota
	TBIDWage

	TBTxProfileOpen
	TBTxProfileEdit
	TBTxProfileClose

	TBTxWageOpen
	TBTxWageEdit

	TBTxExpenseOpen
	TBTxExpenseEdit

	TBTxCloseInvoice
	TBTxBulkImport
)

func TxBytes(object interface{}, tb byte) []byte {
	data := wire.BinaryBytes(object)
	return append([]byte{tb}, data...)
}

type Profile struct {
	Name            string        //identifier for querying
	AcceptedCur     string        //currency you will accept payment in
	DepositInfo     string        //default deposit information (mostly for fiat)
	DueDurationDays int           //default duration until a sent invoice due date
	Timezone        time.Location //default duration until a sent invoice due date
}

func NewProfile(Name string, AcceptedCur string, DepositInfo string,
	DueDurationDays int, Timezone time.Location) *Profile {
	return &Profile{
		Name:            Name,
		AcceptedCur:     AcceptedCur,
		DepositInfo:     DepositInfo,
		DueDurationDays: DueDurationDays,
		Timezone:        Timezone,
	}
}

//////////////////////////////////////////////////////////////////////

// +gen holder:"Invoice,Impl[*Wage,*Expense]"
type InvoiceInner interface {
	SetID()
	GetID() []byte
	GetCtx() Context
	Close(close *CloseInvoice)
}

//var invoiceMapper = data.NewMapper(struct{ Invoice }{}).
//RegisterImplementation(&Wage{}, "wage", 0x01).
//RegisterImplementation(&Expense{}, "expense", 0x02)

//for checking errors at compile time
//var _ Invoice = new(Wage)
//var _ Invoice = new(Expense)

type Wage struct {
	Ctx            Context
	ID             []byte
	TransactionID  string      //empty when unpaid
	PaymentCurTime *AmtCurTime //currency used to pay invoice, empty when unpaid
}

//struct used for hash to determine ID
type Context struct {
	Sender      string
	Receiver    string
	DepositInfo string
	Notes       string
	Amount      *AmtCurTime
	AcceptedCur string
	Due         time.Time
}

func NewWage(ID []byte, Sender, Receiver, DepositInfo, Notes string,
	Amount *AmtCurTime, AcceptedCur string, Due time.Time) *Wage {

	return &Wage{
		Ctx: Context{
			Sender:      Sender,
			Receiver:    Receiver,
			DepositInfo: DepositInfo,
			Notes:       Notes,
			Amount:      Amount,
			AcceptedCur: AcceptedCur,
			Due:         Due,
		},
		ID:             ID,
		TransactionID:  "",
		PaymentCurTime: nil,
	}
}

func (w *Wage) SetID() {
	hashBytes := merkle.SimpleHashFromBinary(w.Ctx)
	w.ID = append([]byte{TBIDWage}, hashBytes...)
}

func (w *Wage) GetID() []byte {
	return w.ID
}

func (w *Wage) GetCtx() Context {
	return w.Ctx
}

func (w *Wage) Close(close *CloseInvoice) {
	w.TransactionID = close.TransactionID
	w.PaymentCurTime = close.PaymentCurTime
}

type Expense struct {
	Ctx            Context
	ID             []byte
	Document       []byte
	DocFileName    string
	ExpenseTaxes   *AmtCurTime
	TransactionID  string      //empty when unpaid
	PaymentCurTime *AmtCurTime //currency used to pay invoice, empty when unpaid
}

func NewExpense(ID []byte, Sender, Receiver, DepositInfo, Notes string,
	Amount *AmtCurTime, AcceptedCur string, Due time.Time,
	Document []byte, DocFileName string, ExpenseTaxes *AmtCurTime) *Expense {

	return &Expense{
		Ctx: Context{
			Sender:      Sender,
			Receiver:    Receiver,
			DepositInfo: DepositInfo,
			Notes:       Notes,
			Amount:      Amount,
			AcceptedCur: AcceptedCur,
			Due:         Due,
		},
		ID:             ID,
		Document:       Document,
		DocFileName:    DocFileName,
		ExpenseTaxes:   ExpenseTaxes,
		TransactionID:  "",
		PaymentCurTime: nil,
	}
}

func (e *Expense) SetID() {
	hashBytes := merkle.SimpleHashFromBinary(e.Ctx)
	e.ID = append([]byte{TBIDExpense}, hashBytes...)
}

func (e *Expense) GetID() []byte {
	return e.ID
}

func (e *Expense) GetCtx() Context {
	return e.Ctx
}

func (e *Expense) Close(close *CloseInvoice) {
	e.TransactionID = close.TransactionID
	e.PaymentCurTime = close.PaymentCurTime
}

/////////////////////////////////////////////////////////////////////////

type CloseInvoice struct {
	ID             []byte
	TransactionID  string      //empty when unpaid
	PaymentCurTime *AmtCurTime //currency used to pay invoice, empty when unpaid
}

func NewCloseInvoice(ID []byte, TransactionID string, PaymentCurTime *AmtCurTime) *CloseInvoice {
	return &CloseInvoice{
		ID:             ID,
		TransactionID:  TransactionID,
		PaymentCurTime: PaymentCurTime,
	}
}
