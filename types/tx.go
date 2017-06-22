package types

// TxProfile is the transaction struct sent through tendermint
type TxProfile struct {
	Address         []byte
	Name            string
	AcceptedCur     string
	DepositInfo     string
	DueDurationDays int
}

// TxInvoice is the transaction struct sent through tendermint
type TxInvoice struct {
	EditID      []byte
	Amount      string
	SenderAddr  []byte
	To          string
	DepositInfo string
	Notes       string
	Cur         string
	Date        string
	DueDate     string
	Receipt     string
	TaxesPaid   string
}

// TxPayment is the transaction struct sent through tendermint
type TxPayment struct {
	TransactionID string
	SenderAddr    []byte
	IDs           [][]byte
	Receiver      string
	Amt           *AmtCurTime
	DateRange     string
}
