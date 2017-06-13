//nolint
package adapters

import (
	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/trackomatron/plugins/invoicer"
)

func parseString(raw []byte) (interface{}, error) {
	var list []string
	err := wire.ReadBinaryBytes(raw, list)
	return list, err
}

type ListProfileActivePresenter struct{}

func (_ ListProfileActivePresenter) MakeKey(str string) ([]byte, error) {
	return invoicer.ListProfileActiveKey(), nil
}
func (_ ListProfileActivePresenter) ParseData(raw []byte) (interface{}, error) {
	return parseString(raw)
}

type ListProfileInactivePresenter struct{}

func (_ ListProfileInactivePresenter) MakeKey(str string) ([]byte, error) {
	return invoicer.ListProfileInactiveKey(), nil
}
func (_ ListProfileInactivePresenter) ParseData(raw []byte) (interface{}, error) {
	return parseString(raw)
}

type ListPaymentPresenter struct{}

func (_ ListPaymentPresenter) MakeKey(str string) ([]byte, error) {
	return invoicer.ListPaymentKey(), nil
}
func (_ ListPaymentPresenter) ParseData(raw []byte) (interface{}, error) {
	return parseString(raw)
}

type ListInvoicePresenter struct{}

func (_ ListInvoicePresenter) MakeKey(str string) ([]byte, error) {
	return invoicer.ListPaymentKey(), nil
}
func (_ ListInvoicePresenter) ParseData(raw []byte) (interface{}, error) {
	var list [][]byte
	err := wire.ReadBinaryBytes(raw, list)
	return list, err
}
