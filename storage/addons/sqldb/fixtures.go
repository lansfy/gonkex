package sqldb

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	actionExtend    = "$extend"
	actionEval      = "$eval"
	virtualFileName = "fake.yml"
)

type tableRow map[string]interface{}
type table []tableRow

type fixture struct {
	Inherits  []string
	Tables    yaml.MapSlice
	Templates yaml.MapSlice
}

type loadedTable struct {
	name string
	rows table
}

type loadContext struct {
	loader         ContentLoader
	tables         []loadedTable
	refsDefinition map[string]tableRow
	refsInserted   map[string]tableRow
}

func ConvertToTestFixtures(loader ContentLoader, names []string) ([]byte, error) {
	ctx := &loadContext{
		loader:         loader,
		refsDefinition: map[string]tableRow{},
		refsInserted:   map[string]tableRow{},
	}

	// gather data from files
	for _, name := range names {
		err := ctx.loadFile(name)
		if err != nil {
			return nil, fmt.Errorf("parse file for fixture %q: %w", name, err)
		}
	}

	data, err := ctx.generateTestFixtures()
	if err != nil {
		return nil, fmt.Errorf("generate global fixtures: %w", err)
	}

	return data, nil
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
		name := template.Key.(string)
		if _, ok := ctx.refsDefinition[name]; ok {
			return fmt.Errorf("unable to load template %s: duplicating ref name", name)
		}

		fields := template.Value.(yaml.MapSlice)
		row := make(tableRow, len(fields))
		for _, field := range fields {
			key := field.Key.(string)
			row[key] = field.Value
		}

		if base, ok := row[actionExtend]; ok {
			base := base.(string)
			baseRow, err := resolveReference(ctx.refsDefinition, base)
			if err != nil {
				return err
			}
			for k, v := range row {
				baseRow[k] = v
			}
			row = baseRow
		}

		ctx.refsDefinition[name] = row
	}

	for _, sourceTable := range loadedFixture.Tables {
		sourceRows, ok := sourceTable.Value.([]interface{})
		if !ok {
			return errors.New("expected array at root level")
		}
		rows := make(table, len(sourceRows))
		for i := range sourceRows {
			sourceFields := sourceRows[i].(yaml.MapSlice)
			fields := make(tableRow, len(sourceFields))
			for j := range sourceFields {
				fields[sourceFields[j].Key.(string)] = sourceFields[j].Value
			}
			rows[i] = fields
		}
		lt := loadedTable{
			name: sourceTable.Key.(string),
			rows: rows,
		}
		ctx.tables = append(ctx.tables, lt)
	}

	return nil
}

func (ctx *loadContext) generateTestFixtures() ([]byte, error) {
	tables := []tableContent{}
	for _, lt := range ctx.tables {
		items, err := ctx.processTableContent(lt.rows)
		if err != nil {
			return nil, fmt.Errorf("processing table '%s': %w", lt.name, err)
		}
		// append rows to global tables
		found := false
		for idx := range tables {
			if tables[idx].name == lt.name {
				tables[idx].items = append(tables[idx].items, items...)
				found = true
				break
			}
		}

		if !found {
			tables = append(tables, tableContent{name: lt.name, items: items})
		}
	}

	yamlTables := yaml.MapSlice{}
	for _, t := range tables {
		yamlTables = append(yamlTables, yaml.MapItem{
			Key:   t.name,
			Value: t.items,
		})
	}

	out, err := yaml.Marshal(yamlTables)
	if err != nil {
		return nil, err
	}

	return out, nil
}

type tableContent struct {
	name  string
	items []yaml.MapSlice
}

func (ctx *loadContext) processTableContent(rows table) ([]yaml.MapSlice, error) {
	// $extend keyword allows to import values from a named row
	for i, row := range rows {
		if _, ok := row[actionExtend]; !ok {
			continue
		}
		base := row[actionExtend].(string)
		baseRow, err := resolveReference(ctx.refsDefinition, base)
		if err != nil {
			return nil, err
		}
		for k, v := range row {
			baseRow[k] = v
		}
		rows[i] = baseRow
	}

	items := []yaml.MapSlice{}
	for _, row := range rows {
		values, err := ctx.loadRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, values)
	}
	return items, nil
}

func (ctx *loadContext) loadRow(row tableRow) (yaml.MapSlice, error) {
	fields := make([]string, 0, len(row))
	for name := range row {
		if !strings.HasPrefix(name, "$") {
			fields = append(fields, name)
		}
	}

	sort.Strings(fields)

	rowValues := tableRow{}
	values := yaml.MapSlice{}
	for _, name := range fields {
		val, err := ctx.resolveExpression(row[name])
		if err != nil {
			return values, err
		}
		values = append(values, yaml.MapItem{
			Key:   name,
			Value: val,
		})
		rowValues[name] = val
	}

	if name, ok := row["$name"]; ok {
		name := name.(string)
		if _, ok := ctx.refsDefinition[name]; ok {
			return nil, fmt.Errorf("duplicating ref name %s", name)
		}
		// add to references
		ctx.refsDefinition[name] = row
		ctx.refsInserted[name] = rowValues
	}

	return values, nil
}

// resolveExpression converts expressions starting with dollar sign into a value
// supporting expressions:
// - $eval()               - executes an SQL expression, e.g. $eval(CURRENT_DATE)
// - $recordName.fieldName - using value of previously inserted named record
func (ctx *loadContext) resolveExpression(value interface{}) (interface{}, error) {
	expr, ok := value.(string)
	if !ok || !strings.HasPrefix(expr, "$") {
		return value, nil
	}

	if strings.HasPrefix(expr, actionEval+"(") {
		if !strings.HasSuffix(expr, ")") {
			return "", fmt.Errorf("incorrect %s() usage: %s", actionEval, expr)
		}
		return "RAW=" + expr[len(actionEval)+1:len(expr)-1], nil
	}

	value, err := resolveFieldReference(ctx.refsInserted, expr)
	if err != nil {
		return "", err
	}

	return value, nil
}

// resolveReference finds previously stored reference by its name
func resolveReference(refs map[string]tableRow, refName string) (tableRow, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference '%s'", refName)
	}
	// make a copy of referencing data to prevent spoiling the source
	// by the way removing $-records from base row
	targetCopy := make(tableRow, len(target))
	for k, v := range target {
		if k == "" || k[0] != '$' {
			targetCopy[k] = v
		}
	}

	return targetCopy, nil
}

// resolveFieldReference finds previously stored reference by name
// and return value of its field
func resolveFieldReference(refs map[string]tableRow, ref string) (interface{}, error) {
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
