package commands

var (
	//common
	FlagTo          string = "to"
	FlagCur         string = "cur"
	FlagDate        string = "date"
	FlagDateRange   string = "date-range"
	FlagDepositInfo string = "info"
	FlagNotes       string = "notes"
	FlagID          string = "id"
	FlagIDs         string = "ids"

	//query
	FlagNum         string = "num"
	FlagSum         string = "sum"
	FlagType        string = "type"
	FlagFrom        string = "from"
	FlagDownloadExp string = "download-expense"
	FlagInactive    string = "inactive"

	//transaction
	//profile flags
	FlagDueDurationDays string = "due-days"

	//invoice flags
	FlagDueDate string = "due-date"

	//expense flags
	FlagReceipt   string = "receipt"
	FlagTaxesPaid string = "taxes"

	//payment flags
	FlagTransactionID string = "tx-id"
	FlagPaid          string = "paid"
)
