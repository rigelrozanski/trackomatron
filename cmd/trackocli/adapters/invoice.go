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

type ProfilePresenter struct{}

func (_ ProfilePresenter) MakeKey(str string) ([]byte, error) {
	key := invoicer.ProfileKey(str)
	return key, nil
}

func (_ ProfilePresenter) ParseData(raw []byte) (interface{}, error) {
	var profile trtypes.Profile
	err := wire.ReadBinaryBytes(raw, &profile)
	return profile, err
}

/**** build out the tx ****/

var (
	_ txs.ReaderMaker      = ProfileTxMaker{}
	_ lightclient.TxReader = ProfileTxReader{}
)

type ProfileTxMaker struct {
	TBTx byte
}

func (m ProfileTxMaker) MakeReader() (lightclient.TxReader, error) {
	chainID := viper.GetString(commands.ChainFlag)
	return ProfileTxReader{
		App:  bcmd.AppTxReader{ChainID: chainID},
		TBTx: m.TBTx,
	}, nil
}

// define flags

type ProfileFlags struct {
	bcmd.AppFlags `mapstructure:",squash"`
}

func (m ProfileTxMaker) Flags() (*flag.FlagSet, interface{}) {
	fs, app := bcmd.AppFlagSet()
	fs.AddFlagSet(trcmd.FSProfile)

	if m.TBTx == invoicer.TBTxProfileOpen {
		// need the name here because no args in light-cli
		fs.String("profile-name", "", "Name of the new profile to open")
	}
	return fs, &ProfileFlags{AppFlags: app}
}

// parse flags

type ProfileTxReader struct {
	App  bcmd.AppTxReader
	TBTx byte
}

func (t ProfileTxReader) ReadTxJSON(data []byte, pk crypto.PubKey) (interface{}, error) {
	return t.App.ReadTxJSON(data, pk)
}

func (t ProfileTxReader) ReadTxFlags(flags interface{}, pk crypto.PubKey) (interface{}, error) {
	data := flags.(*ProfileFlags)

	var name string
	if t.TBTx == invoicer.TBTxProfileOpen {
		name = viper.GetString("profile-name")
	}

	address := pk.Address()
	txBytes := trcmd.ProfileTx(t.TBTx, address, name)
	return t.App.ReadTxFlags(&data.AppFlags, invoicer.Name, txBytes, pk)
}
