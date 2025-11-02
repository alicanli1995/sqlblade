package sqlblade

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"
)

// structInfo caches reflection information for structs
type structInfo struct {
	fields    []fieldInfo
	tableName string
}

// fieldInfo contains information about a struct field
type fieldInfo struct {
	name      string
	dbColumn  string
	index     int
	isPtr     bool
	fieldType reflect.Type
}

var structCache sync.Map // map[reflect.Type]*structInfo

// getStructInfo returns cached struct information or builds it
func getStructInfo(typ reflect.Type) (*structInfo, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, ErrInvalidModel
	}

	// Check cache
	if cached, ok := structCache.Load(typ); ok {
		return cached.(*structInfo), nil
	}

	info := &structInfo{
		fields: make([]fieldInfo, 0),
	}

	// Try to get table name from TableName() method
	if method, ok := typ.MethodByName("TableName"); ok {
		val := reflect.New(typ).Interface()
		if tableNamer, ok := val.(interface{ TableName() string }); ok {
			info.tableName = tableNamer.TableName()
		}
	}

	// If no table name method, use snake_case of struct name
	if info.tableName == "" {
		info.tableName = toSnakeCase(typ.Name())
	}

	// Iterate through struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check for db tag
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Parse db tag (supports "column" or "column,option")
		parts := strings.Split(dbTag, ",")
		columnName := parts[0]

		fieldType := field.Type
		isPtr := fieldType.Kind() == reflect.Ptr
		if isPtr {
			fieldType = fieldType.Elem()
		}

		info.fields = append(info.fields, fieldInfo{
			name:      field.Name,
			dbColumn:  columnName,
			index:     i,
			isPtr:     isPtr,
			fieldType: fieldType,
		})
	}

	// Cache the result
	structCache.Store(typ, info)
	return info, nil
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// scanRows scans database rows into a slice of type T
func scanRows[T any](rows *sql.Rows) ([]T, error) {
	var result []T
	typ := reflect.TypeOf((*T)(nil)).Elem()

	info, err := getStructInfo(typ)
	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Create a map of column name to index
	columnMap := make(map[string]int)
	for i, col := range columns {
		columnMap[strings.ToLower(col)] = i
	}

	for rows.Next() {
		var val T
		ptrVal := reflect.ValueOf(&val).Elem()

		// Prepare scan values
		scanValues := make([]interface{}, len(columns))
		for i := range scanValues {
			var v interface{}
			scanValues[i] = &v
		}

		if err := rows.Scan(scanValues...); err != nil {
			return nil, err
		}

		// Map columns to struct fields
		for _, field := range info.fields {
			colIdx, ok := columnMap[strings.ToLower(field.dbColumn)]
			if !ok {
				continue
			}

			fieldVal := ptrVal.Field(field.index)
			if !fieldVal.IsValid() || !fieldVal.CanSet() {
				continue
			}

			scanVal := scanValues[colIdx].(*interface{})
			if *scanVal == nil {
				if field.isPtr {
					fieldVal.Set(reflect.Zero(fieldVal.Type()))
				}
				continue
			}

			// Convert and set the value
			if err := setFieldValue(fieldVal, *scanVal, field.fieldType); err != nil {
				return nil, err
			}
		}

		result = append(result, val)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// scanRow scans a single database row into type T
func scanRow[T any](rows *sql.Rows) (T, error) {
	var zero T
	results, err := scanRows[T](rows)
	if err != nil {
		return zero, err
	}
	if len(results) == 0 {
		return zero, ErrNoRows
	}
	return results[0], nil
}

// setFieldValue sets a value to a struct field with type conversion
func setFieldValue(field reflect.Value, value interface{}, fieldType reflect.Type) error {
	val := reflect.ValueOf(value)

	// Handle NULL values
	if !val.IsValid() {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
		}
		return nil
	}

	// Direct assignment if types match
	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
		return nil
	}

	// Handle pointer fields
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType))
		}
		field = field.Elem()
	}

	// Type conversion
	switch fieldType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Kind() == reflect.Int64 || val.Kind() == reflect.Int {
			field.SetInt(val.Int())
		} else if val.Kind() == reflect.Float64 {
			field.SetInt(int64(val.Float()))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Kind() == reflect.Int64 || val.Kind() == reflect.Int {
			field.SetUint(uint64(val.Int()))
		} else if val.Kind() == reflect.Float64 {
			field.SetUint(uint64(val.Float()))
		}
	case reflect.Float32, reflect.Float64:
		if val.Kind() == reflect.Float64 || val.Kind() == reflect.Float32 {
			field.SetFloat(val.Float())
		} else if val.Kind() == reflect.Int64 {
			field.SetFloat(float64(val.Int()))
		}
	case reflect.String:
		field.SetString(val.String())
	case reflect.Bool:
		field.SetBool(val.Bool())
	default:
		// Try direct conversion
		if val.Type().ConvertibleTo(fieldType) {
			field.Set(val.Convert(fieldType))
		}
	}

	return nil
}

