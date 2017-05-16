package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	date  = time.Date(2015, time.Month(12), 31, 0, 0, 0, 0, time.UTC)
	date2 = time.Date(2016, time.Month(12), 31, 0, 0, 0, 0, time.UTC)
)

func testParse(t *testing.T) {
	assert := assert.New(t)

	parse, err := ParseAmtCurTime("100BTC", date)
	assert.Nil(err)
	assert.True(parse.EQ(&AmtCurTime{CurrencyTime{"BTC", date}, "100"}))

	_, err = ParseAmtCurTime("BTC100", date)
	assert.NotNil(err)
	_, err = ParseAmtCurTime("100", date)
	assert.NotNil(err)
	_, err = ParseAmtCurTime("BTC", date)
	assert.NotNil(err)
	_, err = ParseAmtCurTime("", date)
	assert.NotNil(err)
}

func testEqualities(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	base, err := ParseAmtCurTime("100BTC", date)
	require.Nil(err)
	baseCopy, err := ParseAmtCurTime("100BTC", date)
	require.Nil(err)
	small, err := ParseAmtCurTime("1BTC", date)
	require.Nil(err)
	diffCur, err := ParseAmtCurTime("100ETH", date)
	require.Nil(err)
	diffDate, err := ParseAmtCurTime("100BTC", date2)
	require.Nil(err)

	//First test to make sure the equality returns errors
	// when comparing different currencies dates
	_, err = base.EQ(diffCur)
	assert.NotNil(err)
	_, err = base.EQ(diffDate)
	assert.NotNil(err)

	//Test ==
	res, err := base.EQ(baseCopy)
	assert.Nil(err)
	assert.True(res)
	_, err = base.EQ(small)
	assert.NotNil(err)

	//Test >=
	res, err = base.GTE(small)
	assert.Nil(err)
	assert.True(res)
	res, err = base.GTE(baseCopy)
	assert.Nil(err)
	assert.True(res)
	_, err = small.GTE(base)
	assert.NotNil(err)

	//Test >
	res, err = base.GT(small)
	assert.Nil(err)
	assert.True(res)
	_, err = base.GT(baseCopy)
	assert.NotNil(err)
	_, err = small.GT(base)
	assert.NotNil(err)

	//Test <=
	res, err = small.LTE(base)
	assert.Nil(err)
	assert.True(res)
	res, err = base.LTE(baseCopy)
	assert.Nil(err)
	assert.True(res)
	_, err = base.LTE(small)
	assert.NotNil(err)

	//Test <
	res, err = small.LT(base)
	assert.Nil(err)
	assert.True(res)
	_, err = base.LT(baseCopy)
	assert.NotNil(err)
	_, err = base.LT(small)
	assert.NotNil(err)

}

func testAddMinus(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	btc0, err := ParseAmtCurTime("0BTC", date)
	require.Nil(err)
	btc100, err := ParseAmtCurTime("100BTC", date)
	require.Nil(err)
	btc200, err := ParseAmtCurTime("200BTC", date)
	require.Nil(err)

	diffCur, err := ParseAmtCurTime("100ETH", date)
	require.Nil(err)
	diffDate, err := ParseAmtCurTime("100BTC", date2)
	require.Nil(err)

	//First test to make sure the equality returns errors
	// when comparing different currencies dates
	_, err = btc100.Add(diffCur)
	assert.NotNil(err)
	_, err = btc100.Add(diffDate)
	assert.NotNil(err)
	_, err = btc100.Minus(diffCur)
	assert.NotNil(err)
	_, err = btc100.Minus(diffDate)
	assert.NotNil(err)

	//Test Add
	res, err := btc100.Add(btc100)
	assert.Nil(err)
	eq, err := res.EQ(btc200)
	assert.Nil(err)
	assert.True(eq)

	//Test Minus
	res, err = btc100.Minus(btc100)
	assert.Nil(err)
	eq, err = res.EQ(btc0)
	assert.Nil(err)
	assert.True(eq)

}
