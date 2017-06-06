//nolint
package commands

const (

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
	FlagInvoiceAmount = "invoice-name"
	FlagReceiverName  = "payment-name"
)
