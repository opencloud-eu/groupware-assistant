package tools

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"slices"
	"strings"
)

const (
	ProductName = "GroupwareAssistant"
)

func ToHtml(text string) string {
	return "<!DOCTYPE html><html><body>" + strings.Join(HtmlJoin(SplitParas(text)), "\n") + "</body></html>"
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

func PickRandoms[T any](s ...T) []T {
	n := rand.IntN(len(s))
	if n == 0 {
		return []T{}
	}
	result := make([]T, n)
	o := make([]T, len(s))
	copy(o, s)
	for i := range n {
		p := rand.IntN(len(o))
		result[i] = slices.Delete(o, p, p)[0]
	}
	return result
}

func PickRandoms1[T any](s ...T) []T {
	n := 1 + rand.IntN(len(s)-1)
	result := make([]T, n)
	o := make([]T, len(s))
	copy(o, s)
	for i := range n {
		p := rand.IntN(len(o))
		result[i] = slices.Delete(o, p, p)[0]
	}
	return result
}

func PickLanguage() string {
	return PickRandom("en-US", "en-GB", "en-AU")
}
