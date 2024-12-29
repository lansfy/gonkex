package colorize

import (
	"fmt"
	"strings"

	"github.com/kylelemons/godebug/diff"
)

func MakeColorDiff(chunks []diff.Chunk) []*Part {
	parts := []*Part{}
	for _, c := range chunks {
		for _, line := range c.Added {
			parts = append(parts, Red(fmt.Sprintf("+%s\n", line)))
		}
		for _, line := range c.Deleted {
			parts = append(parts, Green(fmt.Sprintf("-%s\n", line)))
		}
		for _, line := range c.Equal {
			parts = append(parts, None(fmt.Sprintf(" %s\n", line)))
		}
	}
	if len(parts) != 0 {
		parts[len(parts)-1].value = strings.TrimRight(parts[len(parts)-1].value, "\n")
	}
	return parts
}
