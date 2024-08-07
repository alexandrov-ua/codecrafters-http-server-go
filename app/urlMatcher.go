package main

import (
	"bufio"
	"strings"
)

type Matcher func(s string) (params map[string]string, ok bool)

func CreateUrlMatcher(template string) (Matcher, error) {
	return func(s string) (params map[string]string, ok bool) {
		temp := bufio.NewReader(strings.NewReader(template))
		input := bufio.NewReader(strings.NewReader(s))
		return Match(temp, input)
	}, nil
}

func Match(template *bufio.Reader, input *bufio.Reader) (map[string]string, bool) {
	params := make(map[string]string, 0)
	for {
		next, _ := template.ReadByte()
		switch {
		case next == '{':
			if t, e := template.Peek(1); e == nil && t[0] == '*' {
				template.ReadByte()
				paramName := readUntil(template, '}')
				template.ReadByte()
				paramVal := readUntil(input, 0)
				params[paramName] = paramVal
			} else {
				paramName := readUntil(template, '}')
				template.ReadByte()
				paramVal := readUntil(input, '/')
				params[paramName] = paramVal
			}
		default:
			tmp, _ := input.ReadByte()
			if tmp != next {
				return params, false
			}
			if next == 0 && tmp == 0 {
				return params, true
			}
		}
	}
}

func readUntil(r *bufio.Reader, c byte) string {
	sb := strings.Builder{}
	for {
		if a, e := r.Peek(1); e == nil && a[0] == c {
			break
		}
		if b, e := r.ReadByte(); e == nil {
			sb.WriteByte(b)
		} else {
			break
		}

	}
	return sb.String()
}
