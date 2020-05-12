package fintypes

import (
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"strings"
	"unicode"
)

// Fix error coin names like "Ethereum (ETH)" or "FileCoin [Future]", they are data source error.
func FixNameSymbol(s string) string {
	acceptSpecialChars := "$-+"

	tmp, err := gstring.ReplaceWithTags(s, "(", ")", "", 10)
	if err == nil {
		s = tmp
	}
	tmp, err = gstring.ReplaceWithTags(s, "[", "]", "", 10)
	if err == nil {
		s = tmp
	}

	// letter/digit/space/acceptSpecialChars accepted
	var rName []rune
	for _, v := range s {
		if unicode.IsLetter(v) || unicode.IsDigit(v) || unicode.IsSpace(v) || strings.Contains(acceptSpecialChars, string(v)) {
			rName = append(rName, v)
		} else {
			rName = append(rName, '-')
		}
	}
	s = string(rName)

	// use '-' instead of space
	for {
		if strings.Contains(s, " ") {
			s = strings.Replace(s, " ", "-", -1)
		} else {
			break
		}
	}
	for {
		if strings.Contains(s, "--") {
			s = strings.Replace(s, "--", "-", -1)
		} else {
			break
		}
	}
	s = gstring.TrimLeftAll(s, " ")
	s = gstring.TrimRightAll(s, " ")
	s = gstring.TrimLeftAll(s, "-")
	s = gstring.TrimRightAll(s, "-")
	return s
}

// Fix coin name like "bytom"
func FixCoinName(name string) string {
	return strings.ToLower(FixNameSymbol(name))
}

// Fix coin symbols like "BTM*"
func FixCoinSymbol(symbol string) string {
	return strings.ToUpper(FixNameSymbol(symbol))
}
