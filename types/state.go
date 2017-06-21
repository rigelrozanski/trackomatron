package types

import (
	"time"

	"github.com/tendermint/tmlibs/merkle"
)

// Profile is the state used to store an invoicer profile
type Profile struct {
	Address         []byte //identifier for querying
	Name            string //identifier for querying
	AcceptedCur     string //currency you will accept payment in
	DepositInfo     string //default deposit information (mostly for fiat)
	DueDurationDays int    //default duration until a sent invoice due date
	Active          bool   //default duration until a sent invoice due date
}

// NewProfile create a new active profile
func NewProfile(Address []byte, Name, AcceptedCur, DepositInfo string,
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

//nolint - Autogenerator code for the invoicer types
// used to hold contracts and expense invoices
// +gen holder:"Invoice,Impl[*Contract,*Expense]"
type InvoiceInner interface {
	SetID()
	GetID() []byte
	GetCtx() *Context
}

//for checking errors at compile time
var _ InvoiceInner = new(Contract)
var _ InvoiceInner = new(Expense)

// Contract state struct of type Invoice
type Contract struct {
	ID  []byte
	Ctx *Context
}

// Context struct used for hash to determine ID for invoices
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

// Unpaid calculates the total remaining unpaid portion of an invoice
func (c *Context) Unpaid() (*AmtCurTime, error) {
	return c.Payable.Minus(c.Paid)
}

// Pay makes the maximum payment to the invoice from the fund
//   are then reduced and returned through the variable leftover
func (c *Context) Pay(fund *AmtCurTime) (leftover *AmtCurTime, err error) {
	unpaid, err := c.Unpaid()
	if err != nil {
		return fund, err
	}
	gte, err := fund.GTE(unpaid)
	if err != nil {
		return fund, err
	}
	if gte {
		c.Paid = c.Payable
		c.Open = false
		fund, err = fund.Minus(c.Payable)
		if err != nil {
			return fund, err
		}
	} else {
		//TODO better way of duplicating value of fund here
		c.Paid = &AmtCurTime{CurrencyTime{fund.CurTime.Cur, fund.CurTime.Date}, fund.Amount}
		fund.Amount = "0" //empty the fund
	}
	return fund, nil
}

// NewContract creates a new open Contract invoice
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

// SetID generates the Contract ID from the context
func (w *Contract) SetID() {
	w.ID = merkle.SimpleHashFromBinary(w.Ctx)
}

// GetID get the Contract ID
func (w *Contract) GetID() []byte {
	return w.ID
}

// GetCtx return the context
func (w *Contract) GetCtx() *Context {
	return w.Ctx
}

// Expense state struct of type Invoice
type Expense struct {
	ID           []byte
	Ctx          *Context
	Document     []byte
	DocFileName  string
	ExpenseTaxes *AmtCurTime
}

// NewExpense creates a new open Expense invoice
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

// SetID generates the Expense ID from the context
func (e *Expense) SetID() {
	e.ID = merkle.SimpleHashFromBinary(e.Ctx)
}

// GetID get the Expense ID
func (e *Expense) GetID() []byte {
	return e.ID
}

// GetCtx return the context
func (e *Expense) GetCtx() *Context {
	return e.Ctx
}

/////////////////////////////////////////////////////////////////////////

// Payment state struct for paying invoices
type Payment struct {
	TransactionID  string
	InvoiceIDs     [][]byte //List of ID's to close with transaction
	Sender         string   //Intended sender profile name of the payment
	Receiver       string   //Intended receiver profile name of the payment
	PaymentCurTime *AmtCurTime
	StartDate      time.Time //Optional start date of payments to query for
	EndDate        time.Time //Optional end date of payments to query
}

// NewPayment creates a new payment state
func NewPayment(InvoiceIDs [][]byte, TransactionID, Sender, Receiver string,
	PaymentCurTime *AmtCurTime, StartDate, EndDate time.Time) *Payment {

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
