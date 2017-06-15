package query

import (
	"github.com/spf13/cobra"

	cmdproofs "github.com/tendermint/light-client/commands/proofs"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Handle custom app proofs for state of abci app",
}

func init() {
	//Register the app commands with the proof state command
	cmdproofs.RootCmd.AddCommand(appCmd)
}
