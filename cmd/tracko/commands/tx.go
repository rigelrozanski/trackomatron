package commands

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	bcmd "github.com/tendermint/basecoin/cmd/commands"

	trcmd "github.com/tendermint/trackomatron/commands"
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
)

func init() {

	ProfileOpenCmd.Flags().AddFlagSet(trcmd.FSProfile)
	ProfileEditCmd.Flags().AddFlagSet(trcmd.FSProfile)

	ContractOpenCmd.Flags().AddFlagSet(trcmd.FSInvoice)
	ContractEditCmd.Flags().AddFlagSet(trcmd.FSInvoice)
	ContractEditCmd.Flags().AddFlagSet(trcmd.FSEdit)

	ExpenseOpenCmd.Flags().AddFlagSet(trcmd.FSInvoice)
	ExpenseOpenCmd.Flags().AddFlagSet(trcmd.FSExpense)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSInvoice)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSExpense)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSEdit)

	PaymentCmd.Flags().AddFlagSet(trcmd.FSPayment)

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

func getAddress() (addr []byte, err error) {
	keyPath := viper.GetString("from") //TODO update to proper basecoin key once integrated
	key, err := bcmd.LoadKey(keyPath)
	if key == nil {
		return
	}
	return key.Address[:], err
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

	txBytes := trcmd.ProfileTx(TBTx, address, name)
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

	txBytes, err := trcmd.InvoiceTx(TBTx, tmAddr, amountStr)
	if err != nil {
		return err
	}
	return bcmd.AppTx(invoicer.Name, txBytes)
}

func paymentCmd(cmd *cobra.Command, args []string) error {
	var receiver string
	if len(args) != 1 {
		return ErrCmdReqArg("receiver")
	}
	receiver = args[0]

	tmAddr := cmd.Parent().Flag("node").Value.String()

	txBytes, err := trcmd.PaymentTx(tmAddr, receiver)
	if err != nil {
		return err
	}
	return bcmd.AppTx(invoicer.Name, txBytes)
}
