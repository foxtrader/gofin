package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"strings"
)

/**
子账户

SubAcc的形式：
::spot::
::spot::cross
customSubAcc::spot::cross
*/
type (
	SubAcc string

	SubAccParsed struct {
		Name   string
		Market Market
		Margin Margin
	}
)

const (
	subAccDelimiter = "::"
)

func NewSubAccAddr(subAccName string, market Market, margin Margin) SubAcc {
	return SubAcc(fmt.Sprintf("%s%s%s%s%s", subAccName, subAccDelimiter, market, subAccDelimiter, margin))
}

func ParseSubAcc(s string) (sap SubAccParsed, err error) {
	defErr := gerror.Errorf("invalid SubAcc(%s)", s)

	ss := strings.Split(s, subAccDelimiter)
	if len(ss) != 3 {
		return SubAccParsed{}, defErr
	}

	market, err := ParseMarket(ss[1])
	if err != nil {
		return SubAccParsed{}, err
	}

	margin, err := ParseMargin(ss[2])
	if err != nil {
		return SubAccParsed{}, err
	}

	return SubAccParsed{
		Name:   ss[0],
		Market: market,
		Margin: margin,
	}, nil
}

func (sa SubAcc) Parse() (sap SubAccParsed, err error) {
	return ParseSubAcc(string(sa))
}

func (sa SubAcc) Verify() error {
	_, err := sa.Parse()
	return err
}

func (sa SubAcc) String() string {
	return string(sa)
}
