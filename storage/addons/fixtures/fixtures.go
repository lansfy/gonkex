package fixtures

import (
	"fmt"
)

const (
	actionExtend = "$extend"
)

type Item map[string]interface{}

type Collection struct {
	Name  string
	Items []Item
}

type LoadDataOpts struct {
	CustomActions map[string]func(string) string
}

func LoadData(loader ContentLoader, names []string, opts *LoadDataOpts) ([]*Collection, error) {
	ctx := &loadContext{
		loader:         loader,
		refsDefinition: map[string]Item{},
		refsInserted:   map[string]Item{},
	}

	if opts != nil {
		ctx.customActions = opts.CustomActions
	}

	if ctx.customActions == nil {
		ctx.customActions = map[string]func(string) string{}
	}

	// gather data from files
	for _, name := range names {
		err := ctx.loadFile(name)
		if err != nil {
			return nil, fmt.Errorf("parse file for fixture %q: %w", name, err)
		}
	}

	return ctx.generateSummary()
}

type ContentLoader interface {
	Load(name string) (string, []byte, error)
}
