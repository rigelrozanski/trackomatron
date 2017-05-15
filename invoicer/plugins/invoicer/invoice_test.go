package invoicer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/basecoin-examples/invoicer/types"
	wire "github.com/tendermint/go-wire"
)

func TestRunInvoice(t *testing.T) {

	amt, err := types.ParseAmtCurTime("100BTC", time.Now())
	require.Nil(t, err)

	var invoice types.Invoice

	invoice = types.NewWage(
		nil,
		"foo", //from
		"bar", //to
		"deposit info",
		"notes",
		amt,
		"btc",
		time.Now().Add(time.Hour*100),
	)

	//txBytes := types.TxBytes(invoice, 0x01)
	txBytes := types.TxBytes(struct{ types.Invoice }{invoice}, 0x01)
	//txBytes := wire.BinaryBytes(struct{ types.Invoice }{invoice})

	var invoiceRead = new(types.Invoice)

	//err = wire.ReadBinaryBytes(txBytes, invoiceRead)
	err = wire.ReadBinaryBytes(txBytes[1:], invoiceRead)
	require.Nil(t, err)
	require.NotNil(t, *invoiceRead)
	_, ok := (*invoiceRead).(*types.Wage)
	require.True(t, ok)
}
