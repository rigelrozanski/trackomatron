//nolint
package common

const (

	///////////////////////////////////
	// Flags

	//Common
	FlagTo          string = "to"
	FlagCur         string = "cur"
	FlagDate        string = "date"
	FlagDateRange   string = "date-range"
	FlagDepositInfo string = "info"
	FlagNotes       string = "notes"
	FlagID          string = "id"
	FlagIDs         string = "ids"

	//Query
	FlagNum         string = "num"
	FlagSum         string = "sum"
	FlagLedger      string = "ledger"
	FlagType        string = "type"
	FlagFrom        string = "from"
	FlagDownloadExp string = "download-expense"
	FlagInactive    string = "inactive"

	//Transaction
	//Profile flags
	FlagDueDurationDays string = "due-days"

	//Invoice flags
	FlagDueDate string = "due-date"

	//Expense flags
	FlagReceipt   string = "receipt"
	FlagTaxesPaid string = "taxes"

	//Payment flags
	FlagTransactionID string = "tx-id"
	FlagPaid          string = "paid"

	//Light-client flags
	//The flags replace what are arguments in the full node
	FlagProfileName   = "profile-name"
	FlagInvoiceAmount = "invoice-amount"
	FlagReceiverName  = "receiver-name"

	///////////////////////////////////
	// Tx Names
	TxNameProfileOpen       = "profile-open"
	TxNameProfileEdit       = "profile-edit"
	TxNameProfileDeactivate = "profile-deactivate"
	TxNameContractOpen      = "contract-open"
	TxNameContractEdit      = "contract-edit"
	TxNameExpenseOpen       = "expense-open"
	TxNameExpenseEdit       = "expense-edit"
	TxNamePayment           = "payment"

	///////////////////////////////////
	// light-client presenter apps
	AppAdapterProfile             = "profile"
	AppAdapterInvoice             = "invoice"
	AppAdapterPayment             = "payment"
	AppAdapterListProfileActive   = "profiles"
	AppAdapterListProfileInactive = "profiles-inactive"
	AppAdapterListPayment         = "payments"
	AppAdapterListInvoice         = "invoices"
)
