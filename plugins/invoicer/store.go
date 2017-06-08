package invoicer

import (
	"bytes"
	"errors"

	btypes "github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/trackomatron/types"
)

//nolint Transaction Type-Bytes
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

// MarshalWithTB marshals the object and then prepends a typebyte
func MarshalWithTB(object interface{}, tb byte) []byte {
	data := wire.BinaryBytes(object)
	return append([]byte{tb}, data...)
}

// ProfileKey generates a store key based on profile name
func ProfileKey(name string) []byte {
	return []byte(cmn.Fmt("%v,Profile=%v", Name, name))
}

// InvoiceKey generates a store key based on invoice id bytes
func InvoiceKey(id []byte) []byte {
	return []byte(cmn.Fmt("%v,ID=%x", Name, id))
}

// PaymentKey generates a store key based on transaction id string
func PaymentKey(transactionID string) []byte {
	return []byte(cmn.Fmt("%v,Payment=%v", Name, transactionID))
}

// ListProfileActiveKey generates the store key for the list of active profiles
func ListProfileActiveKey() []byte {
	return []byte(cmn.Fmt("%v,Profiles", Name))
}

// ListProfileInactiveKey generates the store key for the list of inactive profiles
func ListProfileInactiveKey() []byte {
	return []byte(cmn.Fmt("%v,ProfilesAll", Name))
}

// ListInvoiceKey generates the store key for the list of invoices
func ListInvoiceKey() []byte {
	return []byte(cmn.Fmt("%v,Invoices", Name))
}

// ListPaymentKey generates the store key for the list of invoice payments
func ListPaymentKey() []byte {
	return []byte(cmn.Fmt("%v,Payments", Name))
}

// GetProfileFromWire profile from marshalled bytes
func GetProfileFromWire(bytes []byte) (profile types.Profile, err error) {
	if len(bytes) == 0 {
		return profile, errStateNotFound
	}

	err = wire.ReadBinaryBytes(bytes, &profile)
	return profile, wrapErrDecodingState(err)
}

// GetInvoiceFromWire invoice from marshalled bytes
func GetInvoiceFromWire(bytes []byte) (invoice types.Invoice, err error) {
	inv := struct{ types.Invoice }{}
	if len(bytes) == 0 {
		return invoice, errStateNotFound
	}
	err = wire.ReadBinaryBytes(bytes, &inv)
	return inv.Invoice, wrapErrDecodingState(err)
}

// GetPaymentFromWire payment from marshalled bytes
func GetPaymentFromWire(bytes []byte) (payment types.Payment, err error) {
	if len(bytes) == 0 {
		return payment, errStateNotFound
	}

	err = wire.ReadBinaryBytes(bytes, &payment)
	return payment, wrapErrDecodingState(err)
}

// GetListStringFromWire string array from marshalled bytes,
//   currently used from profile and payment lists
func GetListStringFromWire(bytes []byte) (out []string, err error) {

	//if list uninitilialized return new
	if len(bytes) == 0 {
		return out, nil
	}
	err = wire.ReadBinaryBytes(bytes, &out)
	return out, wrapErrDecodingState(err)
}

// GetListBytesFromWire string array from marshalled bytes,
//   currently used from invoice-id lists
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

////////////////////////////////////////////////////////////////////////////////

func getProfileFromAddress(store btypes.KVStore, address []byte) (profile *types.Profile, err error) {

	profiles, err := getListString(store, ListProfileActiveKey())
	if err != nil {
		return profile, err
	}
	found := false
	for _, name := range profiles {
		p, err := getProfile(store, name)
		if err != nil {
			return profile, err
		}
		if bytes.Compare(p.Address[:], address[:]) == 0 {
			profile = &p
			found = true
			break
		}
	}
	if !found {
		return profile, errors.New("Could not retreive profile from address")
	}
	return profile, nil
}
