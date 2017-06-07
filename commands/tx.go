package commands

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	bcmd "github.com/tendermint/basecoin/cmd/commands"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/trackomatron/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var (
	//Commands
	InvoicerCmd = &cobra.Command{
		Use:   invoicer.Name,
		Short: "Commands relating to invoicer system",
	}

	ProfileOpenCmd = &cobra.Command{
		Use:   "profile-open [name]",
		Short: "Open a profile for sending/receiving invoices",
		RunE:  profileOpenCmd,
	}

	ProfileEditCmd = &cobra.Command{
		Use:   "profile-edit",
		Short: "Edit an existing profile",
		RunE:  profileEditCmd,
	}

	ProfileDeactivateCmd = &cobra.Command{
		Use:   "profile-deactivate",
		Short: "Deactivate and existing profile",
		RunE:  profileDeactivateCmd,
	}

	ContractOpenCmd = &cobra.Command{
		Use:   "contract-open [amount]",
		Short: "Send a contract invoice of amount <value><currency>",
		RunE:  contractOpenCmd,
	}

	ContractEditCmd = &cobra.Command{
		Use:   "contract-edit [amount]",
		Short: "Edit an open contract invoice to amount <value><currency>",
		RunE:  contractEditCmd,
	}

	ExpenseOpenCmd = &cobra.Command{
		Use:   "expense-open [amount]",
		Short: "Send an expense invoice of amount <value><currency>",
		RunE:  expenseOpenCmd,
	}

	ExpenseEditCmd = &cobra.Command{
		Use:   "expense-edit [amount]",
		Short: "Edit an open expense invoice to amount <value><currency>",
		RunE:  expenseEditCmd,
	}

	PaymentCmd = &cobra.Command{
		Use:   "payment [receiver]",
		Short: "pay invoices and expenses with transaction infomation",
		RunE:  paymentCmd,
	}

	//Exposed flagsets
	FSProfile = flag.NewFlagSet("", flag.ContinueOnError)
	FSInvoice = flag.NewFlagSet("", flag.ContinueOnError)
	FSExpense = flag.NewFlagSet("", flag.ContinueOnError)
	FSPayment = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {

	//register flags
	FSProfile.String(FlagTo, "", "Who you're invoicing")
	FSProfile.String(FlagCur, "BTC", "Payment curreny accepted")
	FSProfile.String(FlagDepositInfo, "", "Default deposit information to be provided")
	FSProfile.Int(FlagDueDurationDays, 14, "Default number of days until invoice is due from invoice submission")

	FSInvoice.String(FlagTo, "allinbits", "Name of the invoice receiver")
	FSInvoice.String(FlagDepositInfo, "", "Deposit information for invoice payment (default: profile)")
	FSInvoice.String(FlagNotes, "", "Notes regarding the expense")
	FSInvoice.String(FlagCur, "", "Currency which invoice should be paid in")
	FSInvoice.String(FlagDate, "", "Invoice demon date in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	FSInvoice.String(FlagDueDate, "", "Invoice due date in the format YYYY-MM-DD eg. 2016-12-31 (default: profile)")

	FSExpense.String(FlagReceipt, "", "Directory to receipt document file")
	FSExpense.String(FlagTaxesPaid, "", "Taxes amount in the format <decimal><currency> eg. 10.23usd")

	FSPayment.String(FlagIDs, "", "IDs to close during this transaction <id1>,<id2>,<id3>... ")
	FSPayment.String(FlagTransactionID, "", "Completed transaction ID")
	FSPayment.String(FlagPaid, "", "Payment amount in the format <decimal><currency> eg. 10.23usd")
	FSPayment.String(FlagDate, "", "Date payment in the format YYYY-MM-DD eg. 2016-12-31 (default: today)")
	FSPayment.String(FlagDateRange, "",
		"Autoselect IDs within the date range start:end, where start/end are in the format YYYY-MM-DD, or empty. ex. --date 1991-10-21:")

	FSEdit := flag.NewFlagSet("", flag.ContinueOnError)
	FSEdit.String(FlagID, "", "ID (hex) of the invoice to modify")

	ProfileOpenCmd.Flags().AddFlagSet(FSProfile)
	ProfileEditCmd.Flags().AddFlagSet(FSProfile)

	ContractOpenCmd.Flags().AddFlagSet(FSInvoice)
	ContractEditCmd.Flags().AddFlagSet(FSInvoice)
	ContractEditCmd.Flags().AddFlagSet(FSEdit)

	ExpenseOpenCmd.Flags().AddFlagSet(FSInvoice)
	ExpenseOpenCmd.Flags().AddFlagSet(FSExpense)
	ExpenseEditCmd.Flags().AddFlagSet(FSInvoice)
	ExpenseEditCmd.Flags().AddFlagSet(FSExpense)
	ExpenseEditCmd.Flags().AddFlagSet(FSEdit)

	PaymentCmd.Flags().AddFlagSet(FSPayment)

	//register commands
	InvoicerCmd.AddCommand(
		ProfileOpenCmd,
		ProfileEditCmd,
		ProfileDeactivateCmd,
		ContractOpenCmd,
		ContractEditCmd,
		ExpenseOpenCmd,
		ExpenseEditCmd,
		PaymentCmd,
	)
	bcmd.RegisterTxSubcommand(InvoicerCmd)
}

func profileOpenCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(args, invoicer.TBTxProfileOpen)
}

func profileEditCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(args, invoicer.TBTxProfileEdit)
}

func profileDeactivateCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(args, invoicer.TBTxProfileDeactivate)
}

func profileCmd(args []string, TBTx byte) error {

	var name string
	if TBTx == invoicer.TBTxProfileOpen {
		if len(args) != 1 {
			return ErrCmdReqArg("name")
		}
		name = args[0]
	}

	address, err := getAddress()
	if err != nil {
		return errors.Wrap(err, "Error loading address")
	}

	txBytes := ProfileTx(TBTx, address, name)
	return bcmd.AppTx(invoicer.Name, txBytes)
}

func getAddress() (addr []byte, err error) {
	keyPath := viper.GetString("from") //TODO update to proper basecoin key once integrated
	key, err := bcmd.LoadKey(keyPath)
	if key == nil {
		return
	}
	return key.Address[:], err
}

// ProfileTx Generates the tendermint TX used by the light and heavy client
func ProfileTx(TBTx byte, address []byte, name string) []byte {

	profile := types.NewProfile(
		address,
		name,
		viper.GetString(FlagCur),
		viper.GetString(FlagDepositInfo),
		viper.GetInt(FlagDueDurationDays),
	)

	return invoicer.MarshalWithTB(*profile, TBTx)
}

//TODO optimize, move to the ABCI app
func getProfile(tmAddr string) (profile *types.Profile, err error) {

	//get the sender's address
	address, err := getAddress()
	if err != nil {
		return profile, errors.Wrap(err, "Error loading address")
	}

	profiles, err := queryListString(tmAddr, invoicer.ListProfileActiveKey())
	if err != nil {
		return profile, err
	}
	found := false
	for _, name := range profiles {
		p, err := queryProfile(tmAddr, name)
		if err != nil {
			return profile, err
		}
		if bytes.Compare(p.Address[:], address[:]) == 0 {
			profile = &p
			found = true
			break
		}
	}
	if !found {
		return profile, errors.New("Could not retreive profile from address")
	}
	return profile, nil
}

func contractOpenCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(invoicer.TBTxContractOpen, cmd, args)
}

func contractEditCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(invoicer.TBTxContractEdit, cmd, args)
}

func expenseOpenCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(invoicer.TBTxExpenseOpen, cmd, args)
}

func expenseEditCmd(cmd *cobra.Command, args []string) error {
	return invoiceCmd(invoicer.TBTxExpenseEdit, cmd, args)
}

func invoiceCmd(TBTx byte, cmd *cobra.Command, args []string) (err error) {
	if len(args) != 1 {
		return ErrCmdReqArg("amount<amt><cur>")
	}
	amountStr := args[0]

	tmAddr := cmd.Parent().Flag("node").Value.String()

	txBytes, err := InvoiceTx(TBTx, tmAddr, amountStr)
	if err != nil {
		return err
	}
	return bcmd.AppTx(invoicer.Name, txBytes)
}

// InvoiceTx Generates the tendermint TX used by the light and heavy client
func InvoiceTx(TBTx byte, tmAddr, amountStr string) ([]byte, error) {

	var id []byte

	//if editing
	var err error
	if TBTx == invoicer.TBTxContractEdit || //require this flag if
		TBTx == invoicer.TBTxExpenseEdit { //require this flag if

		//get the old id to remove if editing
		idRaw := viper.GetString(FlagID)
		if len(idRaw) == 0 {
			return nil, errors.New("Need the id to edit, please specify through the flag --id")
		}
		if !cmn.IsHex(idRaw) {
			return nil, ErrBadHexID
		}
		id, err = hex.DecodeString(cmn.StripHex(idRaw))
		if err != nil {
			return nil, err
		}
	}

	//get the sender's address
	profile, err := getProfile(tmAddr)
	if err != nil {
		return nil, err
	}
	sender := profile.Name

	var accCur string
	if len(viper.GetString(FlagCur)) > 0 {
		accCur = viper.GetString(FlagCur)
	} else {
		accCur = profile.AcceptedCur
	}

	dateStr := viper.GetString(FlagDate)
	date := time.Now()
	if len(dateStr) > 0 {
		date, err = common.ParseDate(dateStr)
		if err != nil {
			return nil, err
		}
	}
	amt, err := types.ParseAmtCurTime(amountStr, date)
	if err != nil {
		return nil, err
	}

	//calculate payable amount based on invoiced and accepted cur
	payable, err := common.ConvertAmtCurTime(accCur, amt)
	if err != nil {
		return nil, err
	}

	//retrieve flags, or if they aren't used, use the senders profile's default

	var dueDate time.Time
	if len(viper.GetString(FlagDueDate)) > 0 {
		dueDate, err = common.ParseDate(viper.GetString(FlagDueDate))
		if err != nil {
			return nil, err
		}
	} else {
		dueDate = time.Now().AddDate(0, 0, profile.DueDurationDays)
	}

	var depositInfo string
	if len(viper.GetString(FlagDepositInfo)) > 0 {
		depositInfo = viper.GetString(FlagDepositInfo)
	} else {
		depositInfo = profile.DepositInfo
	}

	var invoice types.Invoice

	switch TBTx {
	//if not an expense then we're almost done!
	case invoicer.TBTxContractOpen, invoicer.TBTxContractEdit:
		invoice = types.NewContract(
			id,
			sender,
			viper.GetString(FlagTo),
			depositInfo,
			viper.GetString(FlagNotes),
			accCur,
			dueDate,
			amt,
			payable,
		).Wrap()
	case invoicer.TBTxExpenseOpen, invoicer.TBTxExpenseEdit:
		if len(viper.GetString(FlagTaxesPaid)) == 0 {
			return nil, errors.New("Need --taxes flag")
		}

		taxes, err := types.ParseAmtCurTime(viper.GetString(FlagTaxesPaid), date)
		if err != nil {
			return nil, err
		}
		docBytes, err := ioutil.ReadFile(viper.GetString(FlagReceipt))
		if err != nil {
			return nil, errors.Wrap(err, "Problem reading receipt file")
		}
		_, filename := path.Split(viper.GetString(FlagReceipt))

		invoice = types.NewExpense(
			id,
			sender,
			viper.GetString(FlagTo),
			depositInfo,
			viper.GetString(FlagNotes),
			accCur,
			dueDate,
			amt,
			payable,
			docBytes,
			filename,
			taxes,
		).Wrap()
	default:
		return nil, errors.New("Unrecognized type-bytes")
	}

	return invoicer.MarshalWithTB(invoice, TBTx), nil
}

func paymentCmd(cmd *cobra.Command, args []string) error {
	var receiver string
	if len(args) != 1 {
		return ErrCmdReqArg("receiver")
	}
	receiver = args[0]

	tmAddr := cmd.Parent().Flag("node").Value.String()

	txBytes, err := PaymentTx(tmAddr, receiver)
	if err != nil {
		return err
	}
	return bcmd.AppTx(invoicer.Name, txBytes)
}

// PaymentTx Generates the tendermint TX used by the light and heavy client
func PaymentTx(tmAddr, receiver string) ([]byte, error) {

	//get the sender's address
	profile, err := getProfile(tmAddr)
	if err != nil {
		return nil, err
	}
	sender := profile.Name

	flagIDs := viper.GetString(FlagIDs)
	flagDateRange := viper.GetString(FlagDateRange)

	if len(flagIDs) > 0 && len(flagDateRange) > 0 {
		return nil, errors.New("Cannot use both the IDs flag and date-range flag")
	}
	if len(flagIDs) == 0 && len(flagDateRange) == 0 {
		return nil, errors.New("Must include an IDs flag or date-range flag")
	}

	//Get the date range or list of IDs
	var ids [][]byte
	var startDate, endDate *time.Time = nil, nil
	if len(flagDateRange) > 0 {
		var err error
		startDate, endDate, err = common.ParseDateRange(flagDateRange)
		if err != nil {
			return nil, err
		}
	} else {
		idsStr := strings.Split(flagIDs, ",")
		for _, idHex := range idsStr {
			if !cmn.IsHex(idHex) {
				return nil, ErrBadHexID
			}
			id, err := hex.DecodeString(cmn.StripHex(idHex))
			if err != nil {
				return nil, err
			}
			ids = append([][]byte{id}, ids...)
		}
	}

	date, err := common.ParseDate(viper.GetString(FlagDate))
	if err != nil {
		return nil, err
	}
	amt, err := types.ParseAmtCurTime(viper.GetString(FlagPaid), date)
	if err != nil {
		return nil, err
	}

	payment := types.NewPayment(
		ids,
		viper.GetString(FlagTransactionID),
		sender,
		receiver,
		amt,
		startDate,
		endDate,
	)
	return invoicer.MarshalWithTB(*payment, invoicer.TBTxPayment), nil
}
