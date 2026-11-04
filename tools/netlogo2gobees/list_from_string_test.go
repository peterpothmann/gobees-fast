package main

import (
	"reflect"
	"testing"
)

func TestListFromStringREP(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"", []string{""}},               // NetLogo edge case, strings.Fields gives []
		{"a b ", []string{"a", "b", ""}}, // NetLogo edge case: trailing blank gives an empty last item
		{" a", []string{"a"}},            // leading blank is skipped
		{"a  b", []string{"a", "b"}},     // multiple blanks are skipped
		{"B_terrestris 500", []string{"B_terrestris", "500"}},
	}
	for _, c := range cases {
		if got := ListFromStringREP(c.input); !reflect.DeepEqual(got, c.want) {
			t.Errorf("ListFromStringREP(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
