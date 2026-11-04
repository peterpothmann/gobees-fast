package input

import (
	"fmt"
	"strconv"
	"strings"
)

// ReadValue reads one NetLogo constant from s, starting at pos. Leading white space is skipped.
// It returns the value and the position directly after it.
//
// Supported, as needed by the input files: numbers (float64, also ".5"), strings in double quotes
// (string, with the escapes \" \\ \n \t \r), true/false (bool) and lists in brackets ([]any, nested).
//
// NetLogo: file-read, read-from-string.
func ReadValue(s string, pos int) (any, int, error) {
	pos = skipSpace(s, pos)
	if pos >= len(s) {
		return nil, pos, fmt.Errorf("end of input, expected a value")
	}
	switch s[pos] {
	case '"':
		return readString(s, pos)
	case '[':
		return readList(s, pos)
	}
	end := pos
	for end < len(s) && !isSpace(s[end]) && s[end] != '[' && s[end] != ']' && s[end] != '"' {
		end++
	}
	token := s[pos:end]
	switch strings.ToLower(token) {
	case "true":
		return true, end, nil
	case "false":
		return false, end, nil
	}
	x, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return nil, pos, fmt.Errorf("expected a constant at %q", token)
	}
	return x, end, nil
}

// ReadFromString reads exactly one NetLogo constant from s. Anything after it except white space is an error.
//
// NetLogo: read-from-string.
func ReadFromString(s string) (any, error) {
	v, end, err := ReadValue(s, 0)
	if err != nil {
		return nil, err
	}
	if rest := strings.TrimSpace(s[end:]); rest != "" {
		return nil, fmt.Errorf("unexpected %q after the value", rest)
	}
	return v, nil
}

func readString(s string, pos int) (any, int, error) {
	var b strings.Builder
	for i := pos + 1; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			return b.String(), i + 1, nil
		case c == '\\' && i+1 < len(s):
			i++
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			default:
				b.WriteByte(s[i])
			}
		default:
			b.WriteByte(c)
		}
	}
	return nil, pos, fmt.Errorf("string not closed")
}

func readList(s string, pos int) (any, int, error) {
	list := []any{}
	pos++ // '['
	for {
		pos = skipSpace(s, pos)
		if pos >= len(s) {
			return nil, pos, fmt.Errorf("list not closed")
		}
		if s[pos] == ']' {
			return list, pos + 1, nil
		}
		v, next, err := ReadValue(s, pos)
		if err != nil {
			return nil, pos, err
		}
		list = append(list, v)
		pos = next
	}
}

func skipSpace(s string, pos int) int {
	for pos < len(s) && isSpace(s[pos]) {
		pos++
	}
	return pos
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
