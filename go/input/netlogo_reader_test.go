package input

import (
	"reflect"
	"testing"
)

func TestReadFromString(t *testing.T) {
	cases := []struct {
		input string
		want  any
	}{
		{"10", 10.0},
		{"-0.005", -0.005},
		{".5", 0.5}, // as in the NetLogo weather list (finding 1, see trans/netlogo-findings.md#finding-1)
		{"1.0E-05", 1.0e-05},
		{`"Hedgerow"`, "Hedgerow"},
		{`"say \"hi\""`, `say "hi"`},
		{"TRUE", true},
		{"false", false},
		{"[]", []any{}},
		{`["Grassland*" "Garden"]`, []any{"Grassland*", "Garden"}},
		{`[["Average_Willow" 1.275] ["Margin"]]`, []any{[]any{"Average_Willow", 1.275}, []any{"Margin"}}},
	}
	for _, c := range cases {
		got, err := ReadFromString(c.input)
		if err != nil {
			t.Errorf("ReadFromString(%q): %v", c.input, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("ReadFromString(%q) = %#v, want %#v", c.input, got, c.want)
		}
	}
}

func TestReadFromStringErrors(t *testing.T) {
	for _, input := range []string{"", "abc", `"open`, "[1 2", "1 2"} {
		if _, err := ReadFromString(input); err == nil {
			t.Errorf("ReadFromString(%q): expected an error", input)
		}
	}
}

func TestReadValuePosition(t *testing.T) {
	// file-read leaves the position directly after the value; the rest of the line is read with file-read-line.
	s := `20  "no info" `
	v, end, err := ReadValue(s, 0)
	if err != nil || v != 20.0 || s[end:] != `  "no info" ` {
		t.Errorf("got %v, rest %q, err %v", v, s[end:], err)
	}
}
