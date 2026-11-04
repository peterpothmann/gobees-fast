package input

import "strings"

// ListFromStringREP splits a string at blanks into a list of strings [LIST].
//
// Multiple and leading blanks are skipped. Unlike strings.Fields, it keeps two NetLogo edge cases:
// "" gives [""] (one empty item), and a trailing blank gives an empty last item ("a b " -> ["a" "b" ""]).
func ListFromStringREP(input string) []string {
	result := []string{}
	remaining := input
	for strings.Contains(remaining, " ") {
		cutHere := strings.Index(remaining, " ")
		if cutHere > 0 {
			result = append(result, remaining[:cutHere])
			remaining = remaining[cutHere+1:]
		} else {
			remaining = remaining[1:]
		}
	}
	return append(result, remaining)
}
