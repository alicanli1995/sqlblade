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
		cachedInfo, ok := cached.(*structInfo)
		if !ok {
			return nil, ErrInvalidModel
		}
		return cachedInfo, nil
	}

	info := &structInfo{
		fields: make([]fieldInfo, 0),
	}

	structTypeName := typ.String()
	if cachedTableName, ok := globalTableNameCache.get(structTypeName); ok {
		info.tableName = cachedTableName
	} else if _, ok := typ.MethodByName("TableName"); ok {
		val := reflect.New(typ).Interface()
		if tableNamer, ok := val.(interface{ TableName() string }); ok {
			tableName := tableNamer.TableName()
			info.tableName = tableName
			globalTableNameCache.set(structTypeName, tableName)
		}
	}

	if info.tableName == "" {
		tableName := toSnakeCase(typ.Name())
		info.tableName = tableName
		globalTableNameCache.set(structTypeName, tableName)
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
		columnNameLower := strings.ToLower(columnName)

		fieldType := field.Type
		isPtr := fieldType.Kind() == reflect.Ptr
		if isPtr {
			fieldType = fieldType.Elem()
		}

		info.fields = append(info.fields, fieldInfo{
			name:      field.Name,
			dbColumn:  columnNameLower,
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

func setFieldValue(field reflect.Value, value interface{}, fieldType reflect.Type) error {
	if !field.CanSet() {
		return fmt.Errorf("sqlblade: field cannot be set")
	}
	if val, ok := value.(int64); ok {
		if field.Kind() == reflect.Int64 {
			field.SetInt(val)
			return nil
		}
		if field.Kind() == reflect.Int && val >= -9223372036854775808 && val <= 9223372036854775807 {
			field.SetInt(val)
			return nil
		}
	}

	if val, ok := value.(float64); ok {
		if field.Kind() == reflect.Float64 {
			field.SetFloat(val)
			return nil
		}
		if field.Kind() == reflect.Float32 {
			field.SetFloat(val)
			return nil
		}
	}

	if val, ok := value.(string); ok && field.Kind() == reflect.String {
		field.SetString(val)
		return nil
	}

	if val, ok := value.(bool); ok && field.Kind() == reflect.Bool {
		field.SetBool(val)
		return nil
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
			intVal := val.Int()
			if intVal >= 0 {
				field.SetUint(uint64(intVal))
			}
		} else if val.Kind() == reflect.Float64 {
			floatVal := val.Float()
			if floatVal >= 0 && floatVal <= float64(^uint64(0)) {
				field.SetUint(uint64(floatVal))
			}
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
