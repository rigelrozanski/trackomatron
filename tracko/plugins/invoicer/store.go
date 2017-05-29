package invoicer

import (
	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/basecoin-examples/tracko/types"
)

const (
	TBTxProfileOpen = iota
	TBTxProfileEdit
	TBTxProfileDeactivate

	TBTxContractOpen
	TBTxContractEdit

	TBTxExpenseOpen
	TBTxExpenseEdit

	TBTxPayment
)

func ProfileKey(name string) []byte {
	return []byte(cmn.Fmt("%v,Profile=%v", Name, name))
}

func InvoiceKey(ID []byte) []byte {
	return []byte(cmn.Fmt("%v,ID=%x", Name, ID))
}

func PaymentKey(transactionID string) []byte {
	return []byte(cmn.Fmt("%v,Payment=%v", Name, transactionID))
}

func ListProfileActiveKey() []byte {
	return []byte(cmn.Fmt("%v,Profiles", Name))
}

//Both active and inactive profiles
func ListProfileInactiveKey() []byte {
	return []byte(cmn.Fmt("%v,ProfilesAll", Name))
}

func ListInvoiceKey() []byte {
	return []byte(cmn.Fmt("%v,Invoices", Name))
}

func ListPaymentKey() []byte {
	return []byte(cmn.Fmt("%v,Payments", Name))
}

//Get objects from query bytes

func GetProfileFromWire(bytes []byte) (profile types.Profile, err error) {
	if len(bytes) == 0 {
		return profile, errStateNotFound
	}

	err = wire.ReadBinaryBytes(bytes, &profile)
	return profile, wrapErrDecodingState(err)
}

func GetInvoiceFromWire(bytes []byte) (invoice types.Invoice, err error) {
	inv := struct{ types.Invoice }{}
	if len(bytes) == 0 {
		return invoice, errStateNotFound
	}
	err = wire.ReadBinaryBytes(bytes, &inv)
	return inv.Invoice, wrapErrDecodingState(err)
}

func GetPaymentFromWire(bytes []byte) (payment types.Payment, err error) {
	if len(bytes) == 0 {
		return payment, errStateNotFound
	}

	err = wire.ReadBinaryBytes(bytes, &payment)
	return payment, wrapErrDecodingState(err)
}

func GetListStringFromWire(bytes []byte) (out []string, err error) {

	//if list uninitilialized return new
	if len(bytes) == 0 {
		return out, nil
	}
	err = wire.ReadBinaryBytes(bytes, &out)
	return out, wrapErrDecodingState(err)
}

func GetListBytesFromWire(bytes []byte) (out [][]byte, err error) {

	//if list uninitilialized return new
	if len(bytes) == 0 {
		return out, nil
	}
	err = wire.ReadBinaryBytes(bytes, &out)
	return out, wrapErrDecodingState(err)
}

//Get objects directly from the store

func getProfile(store btypes.KVStore, name string) (types.Profile, error) {
	bytes := store.Get(ProfileKey(name))
	return GetProfileFromWire(bytes)
}

func getInvoice(store btypes.KVStore, ID []byte) (types.Invoice, error) {
	bytes := store.Get(InvoiceKey(ID))
	return GetInvoiceFromWire(bytes)
}

func getPayment(store btypes.KVStore, transactionID []byte) (types.Payment, error) {
	bytes := store.Get(InvoiceKey(transactionID))
	return GetPaymentFromWire(bytes)
}

func getListString(store btypes.KVStore, key []byte) ([]string, error) {
	bytes := store.Get(key)
	return GetListStringFromWire(bytes)
}

func getListBytes(store btypes.KVStore, key []byte) ([][]byte, error) {
	bytes := store.Get(key)
	return GetListBytesFromWire(bytes)
}
