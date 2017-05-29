package invoicer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/basecoin-examples/tracko/common"
	"github.com/tendermint/basecoin-examples/tracko/types"
	wire "github.com/tendermint/go-wire"
)

func TestRunTxInvoice(t *testing.T) {
	require := require.New(t)

	date := time.Date(2015, time.Month(12), 31, 0, 0, 0, 0, time.UTC)

	accCur := "BTC" //accepted currency for payment
	amt, err := types.ParseAmtCurTime("1000USD", date)
	require.Nil(err)

	taxes, err := types.ParseAmtCurTime("10USD", date)
	require.Nil(err)

	//calculate payable amount based on invoiced and accepted cur
	payable, err := common.ConvertAmtCurTime(accCur, amt)
	require.Nil(err)

	var invoices [2]types.Invoice

	invoices[0] = types.NewContract(
		nil,
		"foo", //from
		"bar", //to
		"deposit info",
		"notes",
		accCur,
		time.Now().Add(time.Hour*24*14),
		amt,
		payable,
	).Wrap()

	invoices[1] = types.NewExpense(
		nil,
		"foo", //from
		"bar", //to
		"deposit info",
		"notes: expense",
		accCur,
		time.Now().Add(time.Hour*24*14),
		amt,
		payable,
		[]byte("docbytes"),
		"dummy.txt",
		taxes,
	).Wrap()

	for _, invoice := range invoices {
		txBytes := types.TxBytes(invoice, 0x01)
		var invoiceRead = new(types.Invoice)
		err = wire.ReadBinaryBytes(txBytes[1:], invoiceRead)
		require.Nil(err)
		require.False(invoiceRead.Empty())
	}
}
