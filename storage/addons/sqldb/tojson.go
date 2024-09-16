package sqldb

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

func ExecuteQuery(_ SQLType, db *sql.DB, query string) ([]json.RawMessage, error) {
	if idx := strings.IndexByte(query, ';'); idx >= 0 {
		query = query[:idx]
	}

	var dbResponse []json.RawMessage
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		j, err := convertResultToJson(rows)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(j)
		if err != nil {
			return nil, err
		}

		dbResponse = append(dbResponse, data)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return dbResponse, nil
}

func convertResultToJson(rows *sql.Rows) (interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for i, v := range cols {
		cols[i] = strings.ToLower(v)
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	lenCols := len(cols)
	rawResult := make([]interface{}, lenCols)
	dest := make([]interface{}, lenCols)
	for i := range rawResult {
		dest[i] = &rawResult[i]
	}

	if err := rows.Scan(dest...); err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	for i, raw := range rawResult {
		result[cols[i]] = nil
		if raw != nil {
			val, err := convertRaw(raw, colTypes[i].DatabaseTypeName())
			if err != nil {
				return nil, err
			}
			result[cols[i]] = val
		}
	}
	return result, nil
}

func convertString(value, colType string) (interface{}, error) {
	switch colType {
	case "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT", "YEAR", "INT2", "INT4", "INT8":
		return strconv.Atoi(value)
	case "TINYINT", "BOOL", "BOOLEAN", "BIT":
		return strconv.ParseBool(value)
	case "FLOAT", "DOUBLE", "DECIMAL", "FLOAT4", "FLOAT8", "NUMERIC":
		return strconv.ParseFloat(value, 64)
	case "DATETIME", "TIMESTAMP":
		return normalizeTime(value, "2006-01-02 15:04:05", timeFormat)
	case "DATE":
		return normalizeTime(value, "2006-01-02", "2006-01-02")
	case "TIME":
		return normalizeTime(value, "15:04:05", "15:04:05")
	case "UUID":
		return value, nil
	case "NULL":
		return nil, nil
	case "JSONB":
		m := json.RawMessage{}
		err := json.Unmarshal([]byte(value), &m)
		if err != nil {
			return nil, err
		}
		return m, nil
	default:
		return value, nil
	}
}

const timeFormat = "2006-01-02T15:04:05-07:00"

func normalizeTime(value, pattern, newPattern string) (string, error) {
	t, err := time.Parse(pattern, value)
	if err != nil {
		return "", err
	}
	return t.Format(newPattern), nil
}

func convertBytes(v []byte, colType string) (interface{}, error) {
	value := string(v)
	if colType == "" || colType[0] != '_' {
		return convertString(value, colType)
	}

	// value is array
	value = value[1 : len(value)-1]
	colType = colType[1:]
	arr := []interface{}{}
	if value == "" {
		return arr, nil
	}
	for _, item := range strings.Split(value, ",") {
		t, err := convertString(item, colType)
		if err != nil {
			return nil, err
		}
		arr = append(arr, t)
	}
	return arr, nil
}

func convertRaw(raw interface{}, colType string) (interface{}, error) {
	switch v := raw.(type) {
	case time.Time:
		return v.Format(timeFormat), nil
	case []byte:
		return convertBytes(v, colType)
	default:
		return raw, nil
	}
}
