package colorize

import (
	"strings"

	"github.com/kylelemons/godebug/diff"
	"github.com/kylelemons/godebug/pretty"
)

func joinChanges(result *[]Part, data []string, sep byte, colorer func(v string) Part) {
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

func MakeColorDiff(expected, actual []string) []Part {
	diffCfg := *pretty.DefaultConfig
	diffCfg.Diffable = true
	chunks := diff.DiffChunks(expected, actual)

	parts := []Part{}
	for _, c := range chunks {
		joinChanges(&parts, c.Added, '+', Red)
		joinChanges(&parts, c.Deleted, '-', Green)
		joinChanges(&parts, c.Equal, ' ', None)
	}
	return parts
}
