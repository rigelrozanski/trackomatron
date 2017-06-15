package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	bcmd "github.com/tendermint/basecoin/cmd/commands"

	trcmd "github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
)

//nolint
var (
	//Commands

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

	PaymentCmd = &cobra.Command{
		Use:   "payment [receiver]",
		Short: "pay invoices and expenses with transaction infomation",
		RunE:  paymentCmd,
	}
)

func init() {

	ProfileOpenCmd.Flags().AddFlagSet(trcmd.FSTxProfile)
	ProfileEditCmd.Flags().AddFlagSet(trcmd.FSTxProfile)

	ContractOpenCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ContractEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoiceEdit)

	ExpenseOpenCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ExpenseOpenCmd.Flags().AddFlagSet(trcmd.FSTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoice)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSTxExpense)
	ExpenseEditCmd.Flags().AddFlagSet(trcmd.FSTxInvoiceEdit)

	PaymentCmd.Flags().AddFlagSet(trcmd.FSTxPayment)

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
			return trcmd.ErrCmdReqArg("name")
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
		return trcmd.ErrCmdReqArg("amount<amt><cur>")
	}
	amountStr := args[0]

	address, err := getAddress()
	if err != nil {
		return errors.Wrap(err, "Error loading address")
	}

	txBytes, err := trcmd.InvoiceTx(TBTx, address, amountStr)
	if err != nil {
		return err
	}
	return bcmd.AppTx(invoicer.Name, txBytes)
}

func paymentCmd(cmd *cobra.Command, args []string) error {
	var receiver string
	if len(args) != 1 {
		return trcmd.ErrCmdReqArg("receiver")
	}
	receiver = args[0]

	address, err := getAddress()
	if err != nil {
		return errors.Wrap(err, "Error loading address")
	}

	txBytes, err := trcmd.PaymentTx(address, receiver)
	if err != nil {
		return err
	}
	return bcmd.AppTx(invoicer.Name, txBytes)
}
