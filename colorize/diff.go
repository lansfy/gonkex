package colorize

import (
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

func joinChanges(result *[]*Part, data []string, sep byte, colorer func(v string) *Part) {
	if len(data) == 0 {
		return
	}
	var builder strings.Builder
	for _, line := range data {
		_ = builder.WriteByte(sep)
		_, _ = builder.WriteString(line)
		_, _ = builder.WriteString("\n")
	}

	*result = append(*result, colorer(builder.String()))
}

func MakeColorDiff(title string, expected, actual []string) []*Part {
	matcher := difflib.NewMatcher(expected, actual)
	opcodes := matcher.GetOpCodes()

	parts := []*Part{None(title)}
	for _, opcode := range opcodes {
		switch opcode.Tag {
		case 'r': // replace
			// Handle replace as delete + insert
			deleted := expected[opcode.I1:opcode.I2]
			added := actual[opcode.J1:opcode.J2]
			joinChanges(&parts, deleted, '-', Green)
			joinChanges(&parts, added, '+', Red)
		case 'd': // delete
			deleted := expected[opcode.I1:opcode.I2]
			joinChanges(&parts, deleted, '-', Green)
		case 'i': // insert
			added := actual[opcode.J1:opcode.J2]
			joinChanges(&parts, added, '+', Red)
		case 'e': // equal
			equal := expected[opcode.I1:opcode.I2]
			joinChanges(&parts, equal, ' ', None)
		}
	}
	return parts
}
