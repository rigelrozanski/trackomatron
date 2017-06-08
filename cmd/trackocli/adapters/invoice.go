//nolint
package adapters

import (
	"encoding/hex"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"
	"github.com/tendermint/light-client/commands/txs"
	cmn "github.com/tendermint/tmlibs/common"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"

	trcmd "github.com/tendermint/trackomatron/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	trtypes "github.com/tendermint/trackomatron/types"
)

type InvoicePresenter struct{}

func (_ InvoicePresenter) MakeKey(str string) ([]byte, error) {
	if !cmn.IsHex(str) {
		return nil, trcmd.ErrBadHexID
	}
	id, err := hex.DecodeString(cmn.StripHex(str))
	if err != nil {
		return nil, err
	}
	key := invoicer.InvoiceKey(id)
	return key, nil
}

func (_ InvoicePresenter) ParseData(raw []byte) (interface{}, error) {
	var invoice trtypes.Invoice
	err := wire.ReadBinaryBytes(raw, &invoice)
	return invoice, err
}

/**** build out the tx ****/

var (
	_ txs.ReaderMaker      = InvoiceTxMaker{}
	_ lightclient.TxReader = InvoiceTxReader{}
)

type InvoiceTxMaker struct {
	TBTx byte
}

func (m InvoiceTxMaker) MakeReader() (lightclient.TxReader, error) {
	chainID := viper.GetString(commands.ChainFlag)
	return InvoiceTxReader{
		App:  bcmd.AppTxReader{ChainID: chainID},
		TBTx: m.TBTx,
	}, nil
}

// define flags

type InvoiceFlags struct {
	bcmd.AppFlags `mapstructure:",squash"`
}

func (m InvoiceTxMaker) Flags() (*flag.FlagSet, interface{}) {
	fs, app := bcmd.AppFlagSet()
	fs.AddFlagSet(trcmd.FSTxInvoice)
	fs.String(trcmd.FlagInvoiceAmount, "", "Name of the new invoice to open")

	//add additional flags, as necessary
	switch m.TBTx {
	case invoicer.TBTxExpenseOpen:
		fs.AddFlagSet(trcmd.FSTxExpense)
	case invoicer.TBTxExpenseEdit:
		fs.AddFlagSet(trcmd.FSTxExpense)
		fs.AddFlagSet(trcmd.FSTxInvoiceEdit)
	case invoicer.TBTxContractEdit:
		fs.AddFlagSet(trcmd.FSTxInvoiceEdit)
	}

	return fs, &InvoiceFlags{AppFlags: app}
}

// parse flags

type InvoiceTxReader struct {
	App  bcmd.AppTxReader
	TBTx byte
}

func (t InvoiceTxReader) ReadTxJSON(data []byte, pk crypto.PubKey) (interface{}, error) {
	return t.App.ReadTxJSON(data, pk)
}

func (t InvoiceTxReader) ReadTxFlags(flags interface{}, pk crypto.PubKey) (interface{}, error) {
	data := flags.(*InvoiceFlags)
	amount := viper.GetString(trcmd.FlagInvoiceAmount)
	senderAddr := pk.Address()
	txBytes, err := trcmd.InvoiceTx(t.TBTx, senderAddr, amount)
	if err != nil {
		return nil, err
	}
	return t.App.ReadTxFlags(&data.AppFlags, invoicer.Name, txBytes, pk)
}