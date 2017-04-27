package invoicer

import (
	"github.com/shopspring/decimal"

	"github.com/tendermint/basecoin-examples/invoicer/types"

	abci "github.com/tendermint/abci/btypes"
	"github.com/tendermint/basecoin/state"
	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
)

const invoicerName = "invoicer"

type Invoicer struct {
	name string
}

func New() *Invoicer {
	return &Invoicer{
		name: invoicerName,
	}
}

///////////////////////////////////////////////////

func newP2VIssue(issue string, feePerVote btypes.Coins) P2VIssue {
	return P2VIssue{
		Issue:        issue,
		FeePerVote:   feePerVote,
		VotesFor:     0,
		VotesAgainst: 0,
	}
}

func ProfileKey(name string) []byte {
	return []byte(cmn.Fmt("%v,Profile=%v", invoicerName, name))
}

func InvoicerKey(ID int) []byte {
	return []byte(cmn.Fmt("%v,ID=%v", invoicerName, ID))
}

//get an Invoice from store bytes
func GetProfileFromWire(profileBytes []byte) (profile Profile, err error) {
	out, err := getFromWire(profileBytes, profile)
	return out.(Profile), err
}

func GetInvoiceFromWire(invoiceBytes []byte) (invoice Invoice, err error) {
	out, err := getFromWire(invoiceBytes, invoice)
	return out.(Invoice), err
}

func GetExpenseFromWire(expenseBytes []byte) (expense Expense, err error) {
	out, err := getFromWire(expenseBytes, expense)
	return out.(Expense), err
}

func getFromWire(bytes []byte, destination interface{}) (interface{}, error) {
	var err error

	//Determine if the issue already exists and load
	if len(profileBytes) > 0 { //is there a record of the issue existing?
		err = wire.ReadBinaryBytes(profileBytes, &destination)
		if err != nil {
			err = abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	} else {
		err = abci.ErrInternalError.AppendLog("state not found")
	}
	return in, err
}

func getProfile(store btypes.KVStore, name string) (profile Profile, err error) {
	bytes := store.Get(ProfileKey(name))
	return GetProfileFromWire(bytes)
}

func getInvoice(store btypes.KVStore, ID int) (invoice Invoice, err error) {
	bytes := store.Get(InvoicerKey(address))
	return GetInvoiceFromWire(bytes)
}

func getExpense(store btypes.KVStore, ID int) (expense Expense, err error) {
	bytes := store.Get(ExpenseKey(address))
	return GetExpenseFromWire(bytes)
}

///////////////////////////////////////////////////

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
	case types.TBTxNewProfile:
		return inv.runTxNewProfile(store, ctx, txBytes[1:])
	case types.TBTxOpenInvoice:
		return inv.runTxOpenInvoice(store, ctx, txBytes[1:])
	case types.TBTxOpenExpense:
		return inv.runTxOpenExpense(store, ctx, txBytes[1:])
	case types.TBTxClose:
		return inv.runTxClose(store, ctx, txBytes[1:])
	default:
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: bad prepended bytes")
	}
}

func chargeFee(store btypes.KVStore, ctx btypes.CallContext, fee btypes.Coins) {

	//Charge the Fee from the context coins
	leftoverCoins := ctx.Coins.Minus(fee)
	if !leftoverCoins.IsZero() {
		acc := ctx.CallerAccount
		//return leftover coins
		acc.Balance = acc.Balance.Plus(leftoverCoins)   // subtract fees
		state.SetAccount(store, ctx.CallerAddress, acc) // save the new balance
	}
}

func (inv *Invoicer) runTxNewProfile(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte) (res abci.Result) {

	// Decode tx
	var profile Profile
	err := wire.ReadBinaryBytes(txBytes, &profile)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	fee := btypes.Coins{{"ProfileToken", 1}}

	//Validate Tx
	switch {
	case len(tx.Issue) == 0:
		return abci.ErrInternalError.AppendLog("P2VTx.Issue must have a length greater than 0")
	case len(profile.Nickname) == 0:
		return abci.ErrInternalError.AppendLog("new profile must have nickname")
	case len(profile.AcceptedCur) == 0:
		return abci.ErrInternalError.AppendLog("new profile must have at least one accepted currency")
	case DueDurationDays < 0:
		return abci.ErrInternalError.AppendLog("new profile due duration must be non-negative")
	case !ctx.Coins.IsGTE(fee): // Did the caller provide enough coins?
		return abci.ErrInsufficientFunds.AppendLog("Tx Funds insufficient for creating a new profile")
	}

	//Return if the issue already exists, aka no error was thrown
	if _, err := getProfile(store, profile.Name); err == nil {
		return abci.ErrInternalError.AppendLog("Cannot create an already existing Profile")
	}

	//Store profile and charge fee
	store.Set(ProfileKey(profile.Name), wire.BinaryBytes(profile))
	chargeFee(store, ctx, fee)
	return abci.OK
}

func (inv *Invoicer) runTxVote(store btypes.KVStore, ctx btypes.CallContext, txBytes []byte) (res abci.Result) {

	//Decode tx
	var tx voteTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	//Validate Tx
	if len(tx.Issue) == 0 {
		return abci.ErrInternalError.AppendLog("transaction issue must have a length greater than 0")
	}

	//Load P2VIssue
	p2vIssue, err := getIssue(store, tx.Issue)
	if err != nil {
		return abci.ErrInternalError.AppendLog("error loading issue: " + err.Error())
	}

	//Did the caller provide enough coins?
	if !ctx.Coins.IsGTE(p2vIssue.FeePerVote) {
		return abci.ErrInsufficientFunds.AppendLog("Tx Funds insufficient for voting")
	}

	//Transaction Logic
	switch tx.VoteTypeByte {
	case TypeByteVoteFor:
		p2vIssue.VotesFor += 1
	case TypeByteVoteAgainst:
		p2vIssue.VotesAgainst += 1
	default:
		return abci.ErrInternalError.AppendLog("P2VTx.VoteTypeByte was not recognized")
	}

	//Save P2VIssue, charge fee, return
	store.Set(IssueKey(tx.Issue), wire.BinaryBytes(p2vIssue))
	chargeFee(store, ctx, p2vIssue.FeePerVote)
	return abci.OK
}

func (inv *Invoicer) InitChain(store btypes.KVStore, vals []*abci.Validator) {
}

func (inv *Invoicer) BeginBlock(store btypes.KVStore, hash []byte, header *abci.Header) {
}

func (inv *Invoicer) EndBlock(store btypes.KVStore, height uint64) (res abci.ResponseEndBlock) {
	return
}
