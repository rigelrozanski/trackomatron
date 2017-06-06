//nolint
package adapters

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"
	"github.com/tendermint/light-client/commands/txs"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"

	trcmd "github.com/tendermint/trackomatron/cmd/tracko/commands"
	"github.com/tendermint/trackomatron/plugins/invoicer"
	trtypes "github.com/tendermint/trackomatron/types"
)

type PaymentPresenter struct{}

func (_ PaymentPresenter) MakeKey(str string) ([]byte, error) {
	key := invoicer.PaymentKey(str)
	return key, nil
}

func (_ PaymentPresenter) ParseData(raw []byte) (interface{}, error) {
	var payment trtypes.Payment
	err := wire.ReadBinaryBytes(raw, &payment)
	return payment, err
}

/**** build out the tx ****/

var (
	_ txs.ReaderMaker      = PaymentTxMaker{}
	_ lightclient.TxReader = PaymentTxReader{}
)

type PaymentTxMaker struct{}

func (m PaymentTxMaker) MakeReader() (lightclient.TxReader, error) {
	chainID := viper.GetString(commands.ChainFlag)
	return PaymentTxReader{
		App: bcmd.AppTxReader{ChainID: chainID},
	}, nil
}

// define flags

type PaymentFlags struct {
	bcmd.AppFlags `mapstructure:",squash"`
}

func (m PaymentTxMaker) Flags() (*flag.FlagSet, interface{}) {
	fs, app := bcmd.AppFlagSet()
	fs.AddFlagSet(trcmd.FSPayment)
	fs.String(trcmd.FlagReceiverName, "", "Name of the receiver of the payment")
	return fs, &PaymentFlags{AppFlags: app}
}

// parse flags

type PaymentTxReader struct {
	App bcmd.AppTxReader
}

func (t PaymentTxReader) ReadTxJSON(data []byte, pk crypto.PubKey) (interface{}, error) {
	return t.App.ReadTxJSON(data, pk)
}

func (t PaymentTxReader) ReadTxFlags(flags interface{}, pk crypto.PubKey) (interface{}, error) {
	data := flags.(*PaymentFlags)

	receiver := viper.GetString(trcmd.FlagReceiverName)
	tmAddr := viper.GetString(commands.NodeFlag)

	txBytes, err := trcmd.PaymentTx(tmAddr, receiver)
	if err != nil {
		return nil, err
	}
	return t.App.ReadTxFlags(&data.AppFlags, invoicer.Name, txBytes, pk)
}
