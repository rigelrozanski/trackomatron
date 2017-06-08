//nolint
package adapters

import wire "github.com/tendermint/go-wire"

type ListStringPresenter struct{}

func (_ ListStringPresenter) MakeKey(str string) ([]byte, error) {
	return []byte(str), nil
}

func (_ ListStringPresenter) ParseData(raw []byte) (interface{}, error) {
	var list []string
	err := wire.ReadBinaryBytes(raw, &list)
	return list, err
}

type ListBytesPresenter struct{}

func (_ ListBytesPresenter) MakeKey(str string) ([]byte, error) {
	return []byte(str), nil
}

func (_ ListBytesPresenter) ParseData(raw []byte) (interface{}, error) {
	var list [][]byte
	err := wire.ReadBinaryBytes(raw, &list)
	return list, err
}
