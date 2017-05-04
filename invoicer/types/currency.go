package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type CurTime struct {
	Cur  string
	Date time.Time
}

type AmtCurTime struct {
	Cur    CurTime
	Amount string //Decimal Number
}

func ParseAmtCurTime(amtCur string, date time.Time) (*AmtCurTime, error) {

	if len(amtCur) == 0 {
		return nil, errors.New("not enought information to parse AmtCurTime")
	}

	var reAmt = regexp.MustCompile("([\\d\\.]+)")
	var reCur = regexp.MustCompile("([^\\d\\W]+)")
	amt := reAmt.FindString(amtCur)
	cur := reCur.FindString(amtCur)

	return &AmtCurTime{CurTime{cur, date}, amt}, nil
}

func ParseDate(date string, timezone string) (t time.Time, err error) {

	//get the time of invoice
	t = time.Now()
	if len(timezone) > 0 {

		tz := time.UTC
		if len(timezone) > 0 {
			tz, err = time.LoadLocation(timezone)
			if err != nil {
				return t, fmt.Errorf("error loading timezone, error: ", err) //never stack trace
			}
		}

		str := strings.Split(date, "-")
		var ymd = []int{}
		for _, i := range str {
			j, err := strconv.Atoi(i)
			if err != nil {
				return t, err
			}
			ymd = append(ymd, j)
		}
		if len(ymd) != 3 {
			return t, fmt.Errorf("bad date parsing, not 3 segments") //never stack trace
		}

		t = time.Date(ymd[0], time.Month(ymd[1]), ymd[2], 0, 0, 0, 0, tz)

	}

	return t, nil
}
