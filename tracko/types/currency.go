package types

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type CurrencyTime struct {
	Cur  string
	Date time.Time
}

type AmtCurTime struct {
	CurTime CurrencyTime
	Amount  string //Decimal Number
}

func ParseAmtCurTime(amtCur string, date time.Time) (*AmtCurTime, error) {

	if len(amtCur) == 0 {
		return nil, errors.New("not enought information to parse AmtCurTime")
	}

	var reAmt = regexp.MustCompile("([\\d\\.]+)")
	var reCur = regexp.MustCompile("([^\\d\\W]+)")
	amt := reAmt.FindString(amtCur)
	cur := reCur.FindString(amtCur)

	return &AmtCurTime{CurrencyTime{cur, date}, amt}, nil
}

func (a *AmtCurTime) Add(a2 *AmtCurTime) (*AmtCurTime, error) {
	switch {
	case a == nil && a2 != nil:
		return a2, nil
	case a != nil && a2 == nil:
		return a, nil
	case a != nil && a2 != nil:
		amt1, amt2, err := getDecimals(a, a2)
		if err != nil {
			return nil, err
		}
		return &AmtCurTime{CurrencyTime{a.CurTime.Cur, a.CurTime.Date}, amt1.Add(amt2).String()}, nil
	case a == nil && a2 == nil:
		return nil, nil
	}
	return nil, nil //never called
}

func (a *AmtCurTime) Minus(a2 *AmtCurTime) (*AmtCurTime, error) {
	switch {
	case a == nil && a2 != nil:
		return nil, errors.New("a is nil")
	case a != nil && a2 == nil:
		return a, nil
	case a != nil && a2 != nil:
		amt1, amt2, err := getDecimals(a, a2)
		if err != nil {
			return nil, err
		}
		return &AmtCurTime{CurrencyTime{a.CurTime.Cur, a.CurTime.Date}, amt1.Sub(amt2).String()}, nil
	case a == nil && a2 == nil:
		return nil, errors.New("a is nil")
	}
	return nil, nil //never called
}

func (a *AmtCurTime) EQ(a2 *AmtCurTime) (bool, error) {
	amt1, amt2, err := getDecimals(a, a2)
	if err != nil {
		return false, err
	}
	return amt1.Equal(amt2), nil
}

func (a *AmtCurTime) GT(a2 *AmtCurTime) (bool, error) {
	amt1, amt2, err := getDecimals(a, a2)
	if err != nil {
		return false, err
	}
	return amt1.GreaterThan(amt2), nil
}

func (a *AmtCurTime) GTE(a2 *AmtCurTime) (bool, error) {
	amt1, amt2, err := getDecimals(a, a2)
	if err != nil {
		return false, err
	}
	return amt1.GreaterThanOrEqual(amt2), nil
}

func (a *AmtCurTime) LT(a2 *AmtCurTime) (bool, error) {
	amt1, amt2, err := getDecimals(a, a2)
	if err != nil {
		return false, err
	}
	return amt1.LessThan(amt2), nil
}

func (a *AmtCurTime) LTE(a2 *AmtCurTime) (bool, error) {
	amt1, amt2, err := getDecimals(a, a2)
	if err != nil {
		return false, err
	}
	return amt1.LessThanOrEqual(amt2), nil
}

func getDecimals(a1 *AmtCurTime, a2 *AmtCurTime) (amt1 decimal.Decimal, amt2 decimal.Decimal, err error) {

	if a1 == nil {
		return amt1, amt2, errors.New("input a1 is nil")
	}
	if a2 == nil {
		return amt1, amt2, errors.New("input a2 is nil")
	}

	amt1, err = decimal.NewFromString(a1.Amount)
	if err != nil {
		return
	}
	amt2, err = decimal.NewFromString(a2.Amount)
	if err != nil {
		return
	}
	err = a1.validateOperation(a2)
	return
}

func (a *AmtCurTime) validateOperation(a2 *AmtCurTime) error {
	if a.CurTime.Cur != a2.CurTime.Cur {
		return errors.New("Can't operate on two different currencies")
	}
	return nil
}
