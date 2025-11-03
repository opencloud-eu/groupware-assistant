package tools

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"
)

const (
	ProductName = "GroupwareAssistant"
)

func ToHtml(text string) string {
	return strings.Join(HtmlJoin(SplitParas(text)), "\n")
}

func SplitParas(text string) []string {
	return paraSplitter.Split(text, -1)
}

func HtmlJoin(parts []string) []string {
	var result []string
	for i := range parts {
		result = append(result, fmt.Sprintf("<p>%v</p>", parts[i]))
	}
	return result
}

var paraSplitter = regexp.MustCompile("[\r\n]+")

func ToBoolMap(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, e := range s {
		m[e] = true
	}
	return m
}

func ToBoolMapS(s ...string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, e := range s {
		m[e] = true
	}
	return m
}

func PickRandom[T any](s ...T) T {
	return s[rand.IntN(len(s))]
}
