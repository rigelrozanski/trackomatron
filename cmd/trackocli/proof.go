package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/light-client/commands/proofs"
	trcmd "github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

//nolint
type ProofCommander struct {
	proofs.ProofCommander
}

//nolint
func GetQueryProfileCmd(lp proofs.ProofCommander) *cobra.Command {
	p := ProofCommander{lp}
	cmd := &cobra.Command{
		Use:          "profile [name]",
		Short:        "Query a profile",
		RunE:         p.queryProfileCmd,
		SilenceUsage: true,
	}
	return cmd
}

func (p ProofCommander) queryProfileCmd(cmd *cobra.Command, args []string) error {
	return trcmd.DoQueryProfileCmd(cmd, args, p.queryProfile)
}

func (p ProofCommander) queryProfile(name string) (profile types.Profile, err error) {
	if len(name) == 0 {
		return profile, trcmd.ErrBadQuery("name")
	}
	proof, err := p.GetProof(trcmd.AppAdapterProfile, name, 0) //0 height means latest block
	if err != nil {
		return
	}
	return invoicer.GetProfileFromWire(proof.Data())
}

func init() {
	proofs.RegisterProofStateSubcommand(GetQueryProfileCmd)
}
