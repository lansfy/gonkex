package fixtures

import (
	"fmt"
)

type ContentLoader interface {
	Load(name string) (string, []byte, error)
}

type Item map[string]interface{}

type Collection struct {
	Name  string
	Type  string
	Items []Item
}

type LoadDataOpts struct {
	AllowedTypes  []string
	CustomActions map[string]func(string) string
}

func LoadData(loader ContentLoader, names []string, opts *LoadDataOpts) ([]*Collection, error) {
	var config LoadDataOpts
	if opts != nil {
		config = *opts
	}

	if config.CustomActions == nil {
		config.CustomActions = map[string]func(string) string{}
	}

	ctx := &loadContext{
		loader:         loader,
		refsDefinition: map[string]Item{},
		refsInserted:   map[string]Item{},
		opts:           config,
		allowedTypes:   map[string]bool{},
	}

	for _, name := range config.AllowedTypes {
		ctx.allowedTypes[name] = true
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
