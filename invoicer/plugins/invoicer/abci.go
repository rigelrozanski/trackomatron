package invoicer

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/state"
	btypes "github.com/tendermint/basecoin/types"
)

const Name = "invoicer"

type Invoicer struct {
	name string
}

func New() *Invoicer {
	return &Invoicer{
		name: Name,
	}
}

func (inv *Invoicer) Name() string {
	return inv.name
}

func (inv *Invoicer) SetOption(store btypes.KVStore, key string, value string) (log string) {
	return ""
}

func (inv *Invoicer) RunTx(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte) (res abci.Result) {

	defer func() {
		//Return the ctx coins to the wallet if there is an error
		if res.IsErr() {
			acc := ctx.CallerAccount
			acc.Balance = acc.Balance.Plus(ctx.Coins)       // add the context transaction coins
			state.SetAccount(store, ctx.CallerAddress, acc) // save the new balance
		}
	}()

	//Determine the transaction type and then send to the appropriate transaction function
	if len(txBytes) < 1 {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: no tx bytes")
	}

	//Note that the zero position of txBytes contains the type-byte for the tx type
	switch txBytes[0] {
	case TBTxProfileOpen:
		return runTxProfile(store, ctx, txBytes[1:], false, writeProfile)
	case TBTxProfileEdit:
		return runTxProfile(store, ctx, txBytes[1:], true, writeProfile)
	case TBTxProfileDeactivate:
		return runTxProfile(store, ctx, txBytes[1:], true, deactivateProfile)
	case TBTxContractOpen:
		return runTxInvoice(store, ctx, txBytes[1:], false)
	case TBTxContractEdit:
		return runTxInvoice(store, ctx, txBytes[1:], true)
	case TBTxExpenseOpen:
		return runTxInvoice(store, ctx, txBytes[1:], false)
	case TBTxExpenseEdit:
		return runTxInvoice(store, ctx, txBytes[1:], true)
	case TBTxPayment:
		return runTxPayment(store, ctx, txBytes[1:])
	default:
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: bad prepended bytes")
	}
}

func (inv *Invoicer) InitChain(store btypes.KVStore, vals []*abci.Validator) {
}

func (inv *Invoicer) BeginBlock(store btypes.KVStore, hash []byte, header *abci.Header) {
}

func (inv *Invoicer) EndBlock(store btypes.KVStore, height uint64) (res abci.ResponseEndBlock) {
	return
}
