package types

import (
	"time"

	bcmd "github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/merkle"
)

const (
	TBIDExpense = iota
	TBIDContract
	TBIDPayment
)

func TxBytes(object interface{}, tb byte) []byte {
	data := wire.BinaryBytes(object)
	return append([]byte{tb}, data...)
}

type Profile struct {
	Address         bcmd.Address //identifier for querying
	Name            string       //identifier for querying
	AcceptedCur     string       //currency you will accept payment in
	DepositInfo     string       //default deposit information (mostly for fiat)
	DueDurationDays int          //default duration until a sent invoice due date
	Active          bool         //default duration until a sent invoice due date
}

func NewProfile(Address bcmd.Address, Name, AcceptedCur, DepositInfo string,
	DueDurationDays int) *Profile {
	return &Profile{
		Address:         Address,
		Name:            Name,
		AcceptedCur:     AcceptedCur,
		DepositInfo:     DepositInfo,
		DueDurationDays: DueDurationDays,
		Active:          true,
	}
}

//////////////////////////////////////////////////////////////////////

// +gen holder:"Invoice,Impl[*Contract,*Expense]"
type InvoiceInner interface {
	SetID()
	GetID() []byte
	GetCtx() *Context
}

//for checking errors at compile time
var _ InvoiceInner = new(Contract)
var _ InvoiceInner = new(Expense)

type Contract struct {
	ID  []byte
	Ctx *Context
}

//struct used for hash to determine ID
type Context struct {
	Sender      string
	Receiver    string
	DepositInfo string
	Notes       string
	AcceptedCur string
	Due         time.Time

	Open     bool        //Is this invoice open
	Invoiced *AmtCurTime //Amount Invoiced (likely fiat)
	Payable  *AmtCurTime //Payable Amount (likely crypto)
	Paid     *AmtCurTime //Amount Paid towards this invoice
}

func (c *Context) Unpaid() (*AmtCurTime, error) {
	return c.Payable.Minus(c.Paid)
}

//This function will make the maximum payment to the invoice from the fund
//funds should be reduced from the the fund and returned throught the pointer
func (c *Context) Pay(fund *AmtCurTime) error {
	unpaid, err := c.Unpaid()
	if err != nil {
		return err
	}
	gte, err := fund.GTE(unpaid)
	if err != nil {
		return err
	}
	if gte {
		c.Paid = c.Payable
		c.Open = false
		fund, err = fund.Minus(c.Payable)
		if err != nil {
			return err
		}
	} else {
		//TODO better way of duplicating value of fund here
		c.Paid = &AmtCurTime{CurrencyTime{fund.CurTime.Cur, fund.CurTime.Date}, fund.Amount}
		fund.Amount = "0" //empty the fund
	}
	return nil
}

func NewContract(ID []byte, Sender, Receiver, DepositInfo, Notes string,
	AcceptedCur string, Due time.Time, Amount, Payable *AmtCurTime) *Contract {

	return &Contract{
		ID: ID,
		Ctx: &Context{
			Sender:      Sender,
			Receiver:    Receiver,
			DepositInfo: DepositInfo,
			Notes:       Notes,
			AcceptedCur: AcceptedCur,
			Due:         Due,

			Open:     true,
			Invoiced: Amount,
			Payable:  Payable,
			Paid:     nil,
		},
	}
}

func (w *Contract) SetID() {
	hashBytes := merkle.SimpleHashFromBinary(w.Ctx)
	w.ID = append([]byte{TBIDContract}, hashBytes...)
}

func (w *Contract) GetID() []byte {
	return w.ID
}

func (w *Contract) GetCtx() *Context {
	return w.Ctx
}

type Expense struct {
	ID           []byte
	Ctx          *Context
	Document     []byte
	DocFileName  string
	ExpenseTaxes *AmtCurTime
}

func NewExpense(ID []byte, Sender, Receiver, DepositInfo, Notes string,
	AcceptedCur string, Due time.Time, Amount, Payable *AmtCurTime,
	Document []byte, DocFileName string, ExpenseTaxes *AmtCurTime) *Expense {

	return &Expense{
		ID: ID,
		Ctx: &Context{
			Sender:      Sender,
			Receiver:    Receiver,
			DepositInfo: DepositInfo,
			Notes:       Notes,
			AcceptedCur: AcceptedCur,
			Due:         Due,

			Open:     true,
			Invoiced: Amount,
			Payable:  Payable,
			Paid:     nil,
		},
		Document:     Document,
		DocFileName:  DocFileName,
		ExpenseTaxes: ExpenseTaxes,
	}
}

func (e *Expense) SetID() {
	hashBytes := merkle.SimpleHashFromBinary(e.Ctx)
	e.ID = append([]byte{TBIDExpense}, hashBytes...)
}

func (e *Expense) GetID() []byte {
	return e.ID
}

func (e *Expense) GetCtx() *Context {
	return e.Ctx
}

/////////////////////////////////////////////////////////////////////////

type Payment struct {
	TransactionID  string
	InvoiceIDs     [][]byte //List of ID's to close with transaction
	Sender         string   //Intended sender profile name of the payment
	Receiver       string   //Intended receiver profile name of the payment
	PaymentCurTime *AmtCurTime
	StartDate      *time.Time //Optional start date of payments to query for
	EndDate        *time.Time //Optional end date of payments to query
}

func NewPayment(InvoiceIDs [][]byte, TransactionID, Sender, Receiver string,
	PaymentCurTime *AmtCurTime, StartDate, EndDate *time.Time) *Payment {

	return &Payment{
		TransactionID:  TransactionID,
		InvoiceIDs:     InvoiceIDs,
		Sender:         Sender,
		Receiver:       Receiver,
		PaymentCurTime: PaymentCurTime,
		StartDate:      StartDate,
		EndDate:        EndDate,
	}
}
