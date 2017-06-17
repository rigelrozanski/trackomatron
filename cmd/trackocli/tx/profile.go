package tx

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"

	btypes "github.com/tendermint/basecoin/types"
	trcmn "github.com/tendermint/trackomatron/cmd/trackocli/common"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
var (
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
)

func init() {
	FSTxProfile := flag.NewFlagSet("", flag.ContinueOnError)
	FSTxProfile.String(trcmn.FlagTo, "", "Who you're invoicing")
	FSTxProfile.String(trcmn.FlagCur, "BTC", "Payment curreny accepted")
	FSTxProfile.String(trcmn.FlagDepositInfo, "", "Default deposit information to be provided")
	FSTxProfile.Int(trcmn.FlagDueDurationDays, 14, "Default number of days until invoice is due from invoice submission")

	ProfileOpenCmd.Flags().AddFlagSet(FSTxProfile)
	ProfileEditCmd.Flags().AddFlagSet(FSTxProfile)
}

func profileOpenCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(cmd, args, invoicer.TBTxProfileOpen)
}

func profileEditCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(cmd, args, invoicer.TBTxProfileEdit)
}

func profileDeactivateCmd(cmd *cobra.Command, args []string) error {
	return profileCmd(cmd, args, invoicer.TBTxProfileDeactivate)
}

func profileCmd(cmd *cobra.Command, args []string, TBTx byte) error {

	// Read the standard app-tx flags
	gas, fee, txInput, err := bcmd.ReadAppTxFlags()
	if err != nil {
		return err
	}

	// Retrieve the app-specific flags/args
	var name string
	if TBTx == invoicer.TBTxProfileOpen {
		if len(args) != 1 {
			return trcmn.ErrCmdReqArg("name")
		}
		name = args[0]
	}

	data := profileTx(TBTx, txInput.Address, name)

	// Create AppTx and broadcast
	tx := &btypes.AppTx{
		Gas:   gas,
		Fee:   fee,
		Name:  invoicer.Name,
		Input: txInput,
		Data:  data,
	}
	res, err := bcmd.BroadcastAppTx(tx)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(res)
}

// profileTx Generates the tendermint TX used by the light and heavy client
func profileTx(TBTx byte, address []byte, name string) []byte {
	tx := types.TxProfile{
		Address:         address,
		Name:            name,
		AcceptedCur:     viper.GetString(trcmn.FlagCur),
		DepositInfo:     viper.GetString(trcmn.FlagDepositInfo),
		DueDurationDays: viper.GetInt(trcmn.FlagDueDurationDays),
	}
	return invoicer.MarshalWithTB(tx, TBTx)
}
