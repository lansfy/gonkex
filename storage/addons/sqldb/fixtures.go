package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/lansfy/gonkex/storage/addons/sqldb/testfixtures"

	"gopkg.in/yaml.v2"
)

const (
	actionExtend = "$extend"
	actionEval   = "$eval"
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
	files          map[string]bool
	tables         []loadedTable
	refsDefinition map[string]tableRow
	refsInserted   map[string]tableRow
}

func LoadFixtures(dialect SQLType, db *sql.DB, location string, names []string) error {
	ctx := &loadContext{
		files:          map[string]bool{},
		refsDefinition: map[string]tableRow{},
		refsInserted:   map[string]tableRow{},
	}

	// gather data from files
	for _, name := range names {
		err := ctx.loadFile(location, name)
		if err != nil {
			return fmt.Errorf("parse file for fixture %q: %w", name, err)
		}
	}

	data, err := ctx.generateTestFixtures()
	if err != nil {
		return fmt.Errorf("generate global fixtures: %w", err)
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect(string(dialect)),
		testfixtures.FS(NewOneFileFS(data)),
		testfixtures.FilesMultiTables("fake.yml"),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.SkipTableChecksumComputation(),
		testfixtures.ResetSequencesTo(1),
	)
	if err != nil {
		return err
	}
	return fixtures.Load()
}

func findFixturePath(location, name string) (string, error) {
	candidates := []string{
		location + "/" + name,
		location + "/" + name + ".yml",
		location + "/" + name + ".yaml",
	}

	var err error
	for _, candidate := range candidates {
		if _, err = os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	if os.IsNotExist(err) {
		return "", errors.New("file not exists")
	}
	return "", err
}

func (ctx *loadContext) loadFile(location, name string) error {
	file, err := findFixturePath(location, name)
	if err != nil {
		return err
	}

	// skip previously loaded files
	if ctx.files[file] {
		return nil
	}
	ctx.files[file] = true

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return ctx.loadYml(location, data)
}

func (ctx *loadContext) loadYml(location string, data []byte) error {
	// read yml into struct
	var loadedFixture fixture
	if err := yaml.Unmarshal(data, &loadedFixture); err != nil {
		return err
	}

	// load inherits
	for _, inheritFile := range loadedFixture.Inherits {
		if err := ctx.loadFile(location, inheritFile); err != nil {
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
		items, err := ctx.processTableContent(lt.name, lt.rows)
		if err != nil {
			return nil, err
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
		yamlTables = append(yamlTables, yaml.MapItem{t.name, t.items})
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

func (ctx *loadContext) processTableContent(tableName string, rows table) ([]yaml.MapSlice, error) {
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
		values, err := loadRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, values)
	}
	return items, nil
}

func loadRow(row tableRow) (yaml.MapSlice, error) {
	fields := make([]string, 0, len(row))
	for name := range row {
		if !strings.HasPrefix(name, "$") {
			fields = append(fields, name)
		}
	}

	sort.Strings(fields)

	values := yaml.MapSlice{}
	for _, name := range fields {
		val, err := resolveExpression(row[name])
		if err != nil {
			return values, err
		}
		values = append(values, yaml.MapItem{name, val})
	}
	return values, nil
}

// resolveExpression converts expressions starting with dollar sign into a value
// supporting expressions:
// $eval() - executes an SQL expression, e.g. $eval(CURRENT_DATE)
func resolveExpression(value interface{}) (interface{}, error) {
	expr, ok := value.(string)
	if !ok || !strings.HasPrefix(expr, "$") {
		return value, nil
	}

	if !strings.HasPrefix(expr, actionEval+"(") {
		return "", fmt.Errorf("incorrect $ prefix: %s", expr)
	}

	if !strings.HasSuffix(expr, ")") {
		return "", fmt.Errorf("incorrect %s() usage: %s", actionEval, expr)
	}

	return "RAW=" + expr[len(actionEval)+1:len(expr)-1], nil
}

// resolveReference finds previously stored reference by its name
func resolveReference(refs map[string]tableRow, refName string) (tableRow, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference %s", refName)
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
