package invoicer

import (
	"fmt"
	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/basecoin-examples/invoicer/types"
)

func ProfileKey(name string) []byte {
	return []byte(cmn.Fmt("%v,Profile=%v", Name, name))
}

func InvoiceKey(ID []byte) []byte {
	return []byte(cmn.Fmt("%v,ID=%x", Name, ID))
}

func ListProfileKey() []byte {
	return []byte(cmn.Fmt("%v,Profiles", Name))
}

func ListInvoiceKey() []byte {
	return []byte(cmn.Fmt("%v,Invoices", Name))
}

//Get objects from query bytes

func GetProfileFromWire(bytes []byte) (profile types.Profile, err error) {
	if len(bytes) == 0 {
		return profile, errStateNotFound
	}

	err = wire.ReadBinaryBytes(bytes, &profile)
	fmt.Printf("%+v\n", err)
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

func GetListProfileFromWire(bytes []byte) (profiles []string, err error) {

	//if list uninitilialized return new
	if len(bytes) == 0 {
		return profiles, nil
	}
	err = wire.ReadBinaryBytes(bytes, &profiles)
	return profiles, wrapErrDecodingState(err)
}

func GetListInvoiceFromWire(bytes []byte) (invoices [][]byte, err error) {

	//if list uninitilialized return new
	if len(bytes) == 0 {
		return invoices, nil
	}
	err = wire.ReadBinaryBytes(bytes, &invoices)
	return invoices, wrapErrDecodingState(err)
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

func getListProfile(store btypes.KVStore) ([]string, error) {
	bytes := store.Get(ListProfileKey())
	return GetListProfileFromWire(bytes)
}

func getListInvoice(store btypes.KVStore) ([][]byte, error) {
	bytes := store.Get(ListInvoiceKey())
	return GetListInvoiceFromWire(bytes)
}
