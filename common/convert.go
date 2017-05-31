package common

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sethgrid/pester"
	"github.com/shopspring/decimal"

	"github.com/tendermint/basecoin-examples/tracko/types"
)

func ConvertAmtCurTime(denomOut string, in *types.AmtCurTime) (out *types.AmtCurTime, err error) {

	inDec, err := decimal.NewFromString(in.Amount)
	if err != nil {
		return out, err
	}

	outDec, err := convert(in.CurTime.Cur, denomOut, inDec, in.CurTime.Date)
	if err != nil {
		return out, err
	}

	return &types.AmtCurTime{
		types.CurrencyTime{
			denomOut,
			in.CurTime.Date,
		},
		outDec.String(),
	}, nil
}

//XXX NON-DETERMINISTIC
func convert(denomIn, denomOut string, amt decimal.Decimal, date time.Time) (out decimal.Decimal, err error) {
	dateStr := date.Format("2006-01-02")
	urlFiat2USD := fmt.Sprintf("http://api.fixer.io/%v?base=USD", dateStr)
	urlUSD2BTC := fmt.Sprintf("http://api.coindesk.com/v1/bpi/historical/close.json?start=%v&end=%v", dateStr, dateStr)

	//calculate the conversion factor
	conv := decimal.New(1, 0)
	if denomIn != "USD" {
		multiplier, err := getRate(urlFiat2USD, "rates", denomIn)
		if err != nil {
			return out, errors.WithMessage(err,
				fmt.Sprintf("denomIn %v, denomOut %v", denomIn, denomOut))
		}
		conv = conv.Mul(multiplier)
	}
	if denomOut == "BTC" {
		multiplier, err := getRate(urlUSD2BTC, "bpi", dateStr)
		if err != nil {
			return out, errors.WithMessage(err,
				fmt.Sprintf("denomIn %v, denomOut %v", denomIn, denomOut))
		}
		conv = conv.Mul(multiplier)
	} else if denomOut != "USD" {
		divisor, err := getRate(urlFiat2USD, "rates", denomOut)
		if err != nil {
			return out, errors.WithMessage(err,
				fmt.Sprintf("denomIn %v, denomOut %v", denomIn, denomOut))
		}
		conv = conv.Div(divisor)
	}

	return amt.Div(conv), nil
}

//XXX NON-DETERMINISTIC
//Get the conversion decimal from an http call
func getRate(url, index1, index2 string) (out decimal.Decimal, err error) {
	var temp map[string]interface{}

	resp, err := pester.Get(url)
	if err != nil {
		return out, err
	}
	err = json.NewDecoder(resp.Body).Decode(&temp)
	resp.Body.Close()
	if err != nil {
		return out, errors.WithMessage(err,
			fmt.Sprintf("Bad json decode for url %v, index1 %v, index2 %v", url, index1, index2))
	}
	temp2 := temp[index1].(map[string]interface{})

	if temp2[index2] == nil {
		return out, errors.Errorf("Rate is nil using: url %v, index1 %v, index2 %v", url, index1, index2)
	}

	return decimal.NewFromFloat(temp2[index2].(float64)), nil
}
