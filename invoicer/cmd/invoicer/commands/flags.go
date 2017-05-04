package commands

var (
	//common
	FlagTo          string = "to"
	FlagCur         string = "cur"
	FlagDate        string = "date"
	FlagDepositInfo string = "info"
	FlagNotes       string = "notes"

	//query
	FlagNum         string = "num"
	FlagShort       string = "short"
	FlagType        string = "type"
	FlagFrom        string = "from"
	FlagDownloadExp string = "download-expense"

	//transaction
	//profile flags
	FlagDueDurationDays string = "due-days"
	FlagTimezone        string = "timezone"

	//invoice flags
	FlagDueDate string = "due-date"

	//expense flags
	FlagReceipt   string = "receipt"
	FlagTaxesPaid string = "taxes"

	//close/edit flags
	FlagTransactionID string = "id"
)
