package query

import (
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"
	cmdproofs "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"
)

func getProof(key []byte) (lc.Proof, error) {
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
	height := cmdproofs.GetHeight()
	return cmdproofs.GetProof(node, prover, key, height)
}
