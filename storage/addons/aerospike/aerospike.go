package aerospike

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/aerospike/aerospike-client-go/v5"
	"gopkg.in/yaml.v2"
)

type Storage struct {
	client *aerospikeClient
}

type StorageOpts struct {
}

func NewStorage(client *aerospike.Client, namespace string, opts *StorageOpts) *Storage {
	return &Storage{
		client: newClient(client, namespace),
	}
}

func (l *Storage) GetType() string {
	return "aerospike"
}

func (l *Storage) LoadFixtures(location string, names []string) error {
	ctx := &loadContext{
		location:       location,
		refsDefinition: make(set),
	}

	// Gather data from files.
	for _, name := range names {
		err := l.loadFile(name, ctx)
		if err != nil {
			return fmt.Errorf("unable to load fixture %s: %s", name, err.Error())
		}
	}

	return l.loadSets(ctx)
}

func (l *Storage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	return nil, errors.New("not implemented")
}

type (
	binMap map[string]interface{}
	set    map[string]binMap
)

type fixture struct {
	Inherits  []string
	Sets      yaml.MapSlice
	Templates yaml.MapSlice
}

type loadedSet struct {
	name string
	data set
}
type loadContext struct {
	location       string
	files          []string
	sets           []loadedSet
	refsDefinition set
}

func (l *Storage) loadFile(name string, ctx *loadContext) error {
	candidates := []string{
		ctx.location + "/" + name,
		ctx.location + "/" + name + ".yml",
		ctx.location + "/" + name + ".yaml",
	}
	var err error
	var file string
	for _, candidate := range candidates {
		if _, err = os.Stat(candidate); err == nil {
			file = candidate

			break
		}
	}
	if err != nil {
		return err
	}
	// skip previously loaded files
	if inArray(file, ctx.files) {
		return nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	ctx.files = append(ctx.files, file)

	return l.loadYml(data, ctx)
}

func (l *Storage) loadYml(data []byte, ctx *loadContext) error {
	// read yml into struct
	var loadedFixture fixture
	if err := yaml.Unmarshal(data, &loadedFixture); err != nil {
		return err
	}

	// load inherits
	for _, inheritFile := range loadedFixture.Inherits {
		if err := l.loadFile(inheritFile, ctx); err != nil {
			return err
		}
	}

	// loadedFixture.templates
	// yaml.MapSlice{
	//    string => yaml.MapSlice{  --- template name
	//        string => interface{} --- bin name: value
	//    }
	// }
	for _, template := range loadedFixture.Templates {
		name := template.Key.(string)
		if _, ok := ctx.refsDefinition[name]; ok {
			return fmt.Errorf("unable to load template %s: duplicating ref name", name)
		}

		binMap, err := binMapFromYaml(template)
		if err != nil {
			return err
		}

		if base, ok := binMap["$extend"]; ok {
			baseName := base.(string)
			baseBinMap, err := l.resolveReference(ctx.refsDefinition, baseName)
			if err != nil {
				return err
			}
			for k, v := range binMap {
				baseBinMap[k] = v
			}
			binMap = baseBinMap
		}
		ctx.refsDefinition[name] = binMap
	}

	// loadedFixture.sets
	// yaml.MapSlice{
	//    string => yaml.MapSlice{      --- set name
	//        string => yaml.MapSlice{  --- key name
	//            string => interface{} --- bin name: value
	//        }
	//    }
	// }
	for _, yamlSet := range loadedFixture.Sets {
		set, err := setFromYaml(yamlSet)
		if err != nil {
			return err
		}
		lt := loadedSet{
			name: yamlSet.Key.(string),
			data: set,
		}
		ctx.sets = append(ctx.sets, lt)
	}

	return nil
}

func setFromYaml(mapItem yaml.MapItem) (set, error) {
	entries, ok := mapItem.Value.(yaml.MapSlice)
	if !ok {
		return nil, errors.New("expected map/array as set")
	}

	set := make(set, len(entries))
	for _, e := range entries {
		key := e.Key.(string)
		binmap, err := binMapFromYaml(e)
		if err != nil {
			return nil, err
		}
		set[key] = binmap
	}

	return set, nil
}

func binMapFromYaml(mapItem yaml.MapItem) (binMap, error) {
	bins, ok := mapItem.Value.(yaml.MapSlice)
	if !ok {
		return nil, errors.New("expected map/array as binmap")
	}

	binmap := make(binMap, len(bins))
	for j := range bins {
		binmap[bins[j].Key.(string)] = bins[j].Value
	}

	return binmap, nil
}

func (l *Storage) loadSets(ctx *loadContext) error {
	// truncate first
	truncatedSets := make(map[string]bool)
	for _, s := range ctx.sets {
		if _, ok := truncatedSets[s.name]; ok {
			// already truncated
			continue
		}
		if err := l.truncateSet(s.name); err != nil {
			return err
		}
		truncatedSets[s.name] = true
	}

	// then load data
	for _, s := range ctx.sets {
		if len(s.data) == 0 {
			continue
		}
		if err := l.loadSet(ctx, s); err != nil {
			return fmt.Errorf("failed to load set '%s' because:\n%s", s.name, err)
		}
	}

	return nil
}

// truncateTable truncates table
func (l *Storage) truncateSet(name string) error {
	return l.client.Truncate(name)
}

func (l *Storage) loadSet(ctx *loadContext, set loadedSet) error {
	// $extend keyword allows, to import values from a named row
	for key, binMap := range set.data {
		if _, ok := binMap["$extend"]; !ok {
			continue
		}
		baseName := binMap["$extend"].(string)
		baseBinMap, err := l.resolveReference(ctx.refsDefinition, baseName)
		if err != nil {
			return err
		}
		for k, v := range binMap {
			baseBinMap[k] = v
		}
		set.data[key] = baseBinMap
	}

	for key, binmap := range set.data {
		err := l.client.InsertBinMap(set.name, key, binmap)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveReference finds previously stored reference by its name
func (l *Storage) resolveReference(refs set, refName string) (binMap, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference %s", refName)
	}
	// make a copy of referencing data to prevent spoiling the source
	// by the way removing $-records from base row
	targetCopy := make(binMap, len(target))
	for k, v := range target {
		if k == "" || k[0] != '$' {
			targetCopy[k] = v
		}
	}

	return targetCopy, nil
}

// inArray checks whether the needle is present in haystack slice
func inArray(needle string, haystack []string) bool {
	for _, e := range haystack {
		if needle == e {
			return true
		}
	}

	return false
}
