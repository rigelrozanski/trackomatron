//nolint
package adapters

import (
	wire "github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	trcmd "github.com/tendermint/trackomatron/commands"
)

type ListStringPresenter struct{}

func (_ ListStringPresenter) MakeKey(str string) ([]byte, error) {
	if !cmn.IsHex(str) {
		return nil, trcmd.ErrBadHexID
	}
	cmn.StripHex(str)
	return []byte(str), nil
}

func (_ ListStringPresenter) ParseData(raw []byte) (interface{}, error) {
	var list []string
	err := wire.ReadBinaryBytes(raw, &list)
	return list, err
}

type ListBytesPresenter struct{}

func (_ ListBytesPresenter) MakeKey(str string) ([]byte, error) {
	if !cmn.IsHex(str) {
		return nil, trcmd.ErrBadHexID
	}
	cmn.StripHex(str)
	return []byte(str), nil
}

func (_ ListBytesPresenter) ParseData(raw []byte) (interface{}, error) {
	var list [][]byte
	err := wire.ReadBinaryBytes(raw, &list)
	return list, err
}
