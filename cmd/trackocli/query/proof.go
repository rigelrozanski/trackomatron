package query

import (
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"
	cmdproofs "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	"github.com/tendermint/trackomatron/types"
)

func getProof(key []byte) (lc.Proof, error) {
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
	height := cmdproofs.GetHeight()
	return cmdproofs.GetProof(node, prover, key, height)
}

func queryInvoice(id []byte) (invoice types.Invoice, err error) {
	key := invoicer.InvoiceKey(id)
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetInvoiceFromWire(proof.Data())
}

func queryListString(key []byte) (list []string, err error) {
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetListStringFromWire(proof.Data())
}

func queryListBytes(key []byte) (list [][]byte, err error) {
	proof, err := getProof(key)
	if err != nil {
		return
	}
	return invoicer.GetListBytesFromWire(proof.Data())
}
