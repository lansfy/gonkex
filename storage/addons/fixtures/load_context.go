package fixtures

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	actionExtend = "$extend"
	actionName   = "$name"
)

type fixture struct {
	Inherits    []string            `yaml:"inherits"`
	Templates   MapSlice            `yaml:"templates"`
	Collections map[string]MapSlice `yaml:",inline"`
}

type loadContext struct {
	loader         ContentLoader
	tables         []*Collection
	refsDefinition map[string]Item
	refsInserted   map[string]Item
	opts           LoadDataOpts
	allowedTypes   map[string]bool
}

func (ctx *loadContext) loadFile(name string) error {
	_, data, err := ctx.loader.Load(name)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return ctx.loadYml(data)
}

func (ctx *loadContext) loadYml(data []byte) error {
	// read yml into struct
	var loadedFixture fixture
	if err := yaml.Unmarshal(data, &loadedFixture); err != nil {
		return err
	}

	// load inherits
	for _, inheritFile := range loadedFixture.Inherits {
		if err := ctx.loadFile(inheritFile); err != nil {
			return err
		}
	}

	for _, template := range loadedFixture.Templates {
		name := template.Key
		wrap := func(err error) error {
			return fmt.Errorf("parsing template '%s': %w", name, err)
		}

		if _, ok := ctx.refsDefinition[name]; ok {
			return wrap(errors.New("duplicate template name"))
		}

		row, ok := template.Value.(map[string]interface{})
		if !ok {
			return wrap(fmt.Errorf("map expected, found value with type '%T'", template.Value))
		}

		if actionValue, ok := row[actionExtend]; ok {
			base, ok := actionValue.(string)
			if !ok {
				return wrap(fmt.Errorf("key '%s' has non-string value '%v'", actionExtend, actionValue))
			}

			baseRow, err := resolveItemReference(ctx.refsDefinition, base)
			if err != nil {
				return wrap(err)
			}
			for k, v := range row {
				baseRow[k] = v
			}
			row = baseRow
		}

		ctx.refsDefinition[name] = row
	}

	for collType, collections := range loadedFixture.Collections {
		if !ctx.allowedTypes[collType] {
			return fmt.Errorf("unknown item type '%s' found", collType)
		}

		for _, sourceTable := range collections {
			wrap := func(err error) error {
				return fmt.Errorf("parsing %s '%s': %w", singularNumber(collType), sourceTable.Key, err)
			}

			sourceRows, ok := sourceTable.Value.([]interface{})
			if !ok {
				return wrap(errors.New("expected array at root level"))
			}
			rows := make([]Item, len(sourceRows))
			for i := range sourceRows {
				sourceFields, ok := sourceRows[i].(map[string]interface{})
				if !ok {
					return wrap(fmt.Errorf("array of map expected, found value with type '%T'", sourceRows[i]))
				}
				rows[i] = sourceFields
			}
			lt := &Collection{
				Name:  sourceTable.Key,
				Type:  collType,
				Items: rows,
			}
			ctx.tables = append(ctx.tables, lt)
		}
	}

	return nil
}

func (ctx *loadContext) generateSummary() ([]*Collection, error) {
	tables := []*Collection{}
	for _, lt := range ctx.tables {
		items, err := ctx.processTableContent(lt.Items)
		if err != nil {
			return nil, fmt.Errorf("processing %s '%s': %w", singularNumber(lt.Type), lt.Name, err)
		}
		// append rows to global tables
		found := false
		for idx := range tables {
			if tables[idx].Name == lt.Name && tables[idx].Type == lt.Type {
				tables[idx].Items = append(tables[idx].Items, items...)
				found = true
				break
			}
		}

		if !found {
			tables = append(tables, &Collection{
				Name:  lt.Name,
				Type:  lt.Type,
				Items: items,
			})
		}
	}
	return tables, nil
}

func (ctx *loadContext) processTableContent(rows []Item) ([]Item, error) {
	// $extend keyword allows to import values from a named row
	for i, row := range rows {
		if _, ok := row[actionExtend]; !ok {
			continue
		}
		base, ok := row[actionExtend].(string)
		if !ok {
			return nil, fmt.Errorf("key '%s' has non-string value '%v'", actionExtend, row[actionExtend])
		}

		baseRow, err := resolveItemReference(ctx.refsDefinition, base)
		if err != nil {
			return nil, err
		}
		for k, v := range row {
			baseRow[k] = v
		}
		rows[i] = baseRow
	}

	items := []Item{}
	for _, row := range rows {
		values, err := ctx.loadRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, values)
	}
	return items, nil
}

func (ctx *loadContext) loadRow(row Item) (Item, error) {
	rowValues := Item{}
	for name := range row {
		if strings.HasPrefix(name, "$") {
			continue
		}
		val, err := ctx.resolveExpression(row[name])
		if err != nil {
			return rowValues, err
		}
		rowValues[name] = val
	}

	if actionValue, ok := row[actionName]; ok {
		name, ok := actionValue.(string)
		if !ok {
			return nil, fmt.Errorf("key '%s' has non-string value '%v'", actionName, actionValue)
		}

		if _, ok := ctx.refsDefinition[name]; ok {
			return nil, fmt.Errorf("duplicate $name '%s'", name)
		}
		// add to references
		ctx.refsDefinition[name] = row
		ctx.refsInserted[name] = rowValues
	}

	return rowValues, nil
}

// resolveExpression converts expressions starting with dollar sign into a value
// supporting expressions:
// - $some_action(...)     - transform value in bracked with specified action's function
// - $recordName.fieldName - using value of previously inserted named record
func (ctx *loadContext) resolveExpression(value interface{}) (interface{}, error) {
	expr, ok := value.(string)
	if !ok || !strings.HasPrefix(expr, "$") {
		return value, nil
	}

	idxStart := strings.Index(expr, "(")
	if idxStart != -1 {
		if !strings.HasSuffix(expr, ")") {
			return "", fmt.Errorf("incorrect action usage '$someaction(...)' for '%s'", expr)
		}
		action := expr[1:idxStart]
		actionValue := expr[idxStart+1 : len(expr)-1]
		f, ok := ctx.opts.CustomActions[action]
		if !ok {
			return "", fmt.Errorf("unknown action '%s' in '%s'", action, expr)
		}
		return f(actionValue), nil
	}

	value, err := resolveFieldReference(ctx.refsInserted, expr)
	if err != nil {
		return "", err
	}

	return value, nil
}

// resolveItemReference finds previously stored reference by its name
func resolveItemReference(refs map[string]Item, refName string) (Item, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference '%s'", refName)
	}
	// make a copy of referencing data to prevent spoiling the source
	// by the way removing $-records from base row
	targetCopy := make(Item, len(target))
	for k, v := range target {
		if k == "" || k[0] != '$' {
			targetCopy[k] = v
		}
	}

	return targetCopy, nil
}

// resolveFieldReference finds previously stored reference by name and return value of its field
func resolveFieldReference(refs map[string]Item, ref string) (interface{}, error) {
	parts := strings.SplitN(ref, ".", 2)
	if len(parts) < 2 || len(parts[0]) < 2 || len(parts[1]) < 1 {
		return nil, fmt.Errorf("invalid reference '%s', correct form is '$refName.field'", ref)
	}

	// remove leading $
	refName := parts[0][1:]
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference '%s' in '%s'", refName, ref)
	}

	value, ok := target[parts[1]]
	if !ok {
		return nil, fmt.Errorf("undefined reference field '%s'", parts[1])
	}

	return value, nil
}

func singularNumber(name string) string {
	return strings.TrimSuffix(name, "s")
}
