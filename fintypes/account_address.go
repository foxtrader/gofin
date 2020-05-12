package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"strings"
)

type (
	// example: buffett@gmail.com@binance
	AccountAddress string
)

const (
	AccountAddressNull AccountAddress = ""
)

func NewAccountAddress(email string, platform Platform) AccountAddress {
	return AccountAddress(fmt.Sprintf("%s@%s", email, platform.String()))
}

func (tai AccountAddress) String() string {
	return string(tai)
}

func (tai AccountAddress) Email() string {
	ss := strings.Split(string(tai), "@")
	if len(ss) != 3 {
		return ""
	}
	return ss[0] + "@" + ss[1]
}

func (tai AccountAddress) Platform() Platform {
	ss := strings.Split(string(tai), "@")
	if len(ss) != 3 {
		return PlatformUnknown
	}
	return Platform(strings.ToLower(ss[2]))
}

func ParseAccountAddress(s string) (AccountAddress, error) {
	if s == "" {
		return AccountAddressNull, nil
	}
	ss := strings.Split(s, "@")
	if len(ss) != 3 {
		return "", gerror.Errorf("invalid AccountAddress(%s)", s)
	}
	_, err := ParsePlatform(ss[2])
	if err != nil {
		return "", err
	}
	return AccountAddress(s), nil
}

func (tai AccountAddress) MarshalJSON() ([]byte, error) {
	return []byte(`"` + tai.String() + `"`), nil
}

func (tai *AccountAddress) UnmarshalJSON(b []byte) error {
	str := string(b)
	str = gstring.RemoveHead(str, 1)
	str = gstring.RemoveTail(str, 1)
	decTAI, err := ParseAccountAddress(str)
	if err != nil {
		return err
	}
	*tai = decTAI
	return nil
}
