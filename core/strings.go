package core

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

func Contains(keys []string, candidate string) bool {
	for _, key := range keys {
		if key == candidate {
			return true
		}
	}
	return false
}

func FlattenMap(m map[string][]string) (arr []string) {
	for _, vs := range m {
		for _, v := range vs {
			if Contains(arr, v) {
				continue
			}

			arr = append(arr, v)
		}
	}

	return
}

func IsSimilar(s string, t string) bool {
	if s == t {
		return true
	}

	return s == t || strings.ToLower(CamelCase(s)) == strings.ToLower(CamelCase(t))
}

func CamelCase(s string) string {
	if s == "" {
		return s
	}

	t := strings.Trim(s, "_- ")
	r := regexp.MustCompile("[\\s_-]+(.)")
	o := r.ReplaceAllFunc([]byte(t), func(b []byte) []byte {
		return bytes.ToUpper([]byte{b[len(b)-1]})
	})

	return LowerFirst(string(o))
}

func Classify(s string) string {
	if s == "" {
		return ""
	}

	t := CamelCase(s)
	a := []rune(t)
	a[0] = unicode.ToUpper(a[0])

	return string(a)
}

func SnakeCase(s string) string {
	if s == "" {
		return s
	}

	t := CamelCase(s)
	r := regexp.MustCompile("[A-Z][a-z_-]")
	o := r.ReplaceAllFunc([]byte(t), func(b []byte) []byte {
		return append([]byte("_"), bytes.ToLower(b)...)
	})

	return strings.Trim(strings.ToLower(string(o)), "_")
}

func EncapCase(op string, s string) string {
	return Classify(fmt.Sprintf("%v %v", op, s))
}

func LowerFirst(s string) string {
	if s == "" {
		return s
	}

	a := []rune(s)
	a[0] = unicode.ToLower(a[0])

	return string(a)
}

func UpperFirst(s string) string {
	if s == "" {
		return s
	}

	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])

	return string(a)
}
