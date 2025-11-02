package sqlblade

import (
	"database/sql"
	"fmt"
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

	if cached, ok := structCache.Load(typ); ok {
		return cached.(*structInfo), nil
	}

	info := &structInfo{
		fields: make([]fieldInfo, 0),
	}

	if _, ok := typ.MethodByName("TableName"); ok {
		val := reflect.New(typ).Interface()
		if tableNamer, ok := val.(interface{ TableName() string }); ok {
			info.tableName = tableNamer.TableName()
		}
	}

	if info.tableName == "" {
		info.tableName = toSnakeCase(typ.Name())
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if !field.IsExported() {
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

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

func scanRows[T any](rows *sql.Rows) ([]T, error) {
	return scanRowsOptimized[T](rows)
}

// setFieldValue sets a value to a struct field with type conversion
func setFieldValue(field reflect.Value, value interface{}, fieldType reflect.Type) error {
	if !field.CanSet() {
		return fmt.Errorf("sqlblade: field cannot be set")
	}
	val := reflect.ValueOf(value)

	if !val.IsValid() {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
		}
		return nil
	}

	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
		return nil
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType))
		}
		field = field.Elem()
	}

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
		if val.Type().ConvertibleTo(fieldType) {
			field.Set(val.Convert(fieldType))
		}
	}

	return nil
}
