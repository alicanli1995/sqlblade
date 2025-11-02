package sqlblade

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type columnMapCache struct {
	mu    sync.RWMutex
	store map[string]map[string]int
}

var columnMapCacheInst = &columnMapCache{
	store: make(map[string]map[string]int),
}

func (cmc *columnMapCache) getColumnMap(columns []string) map[string]int {
	key := strings.Join(columns, ",")

	cmc.mu.RLock()
	if cached, ok := cmc.store[key]; ok {
		cmc.mu.RUnlock()
		return cached
	}
	cmc.mu.RUnlock()

	columnMap := make(map[string]int, len(columns))
	for i, col := range columns {
		columnMap[strings.ToLower(col)] = i
	}

	cmc.mu.Lock()
	cmc.store[key] = columnMap
	cmc.mu.Unlock()

	return columnMap
}

func scanRowsOptimized[T any](rows *sql.Rows) ([]T, error) {
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

	columnMap := columnMapCacheInst.getColumnMap(columns)

	result = make([]T, 0, 10)

	scanValues := make([]interface{}, len(columns))
	scanPtrs := make([]interface{}, len(columns))
	for i := range scanValues {
		scanPtrs[i] = &scanValues[i]
	}

	for rows.Next() {
		var val T
		ptrVal := reflect.ValueOf(&val).Elem()

		if err := rows.Scan(scanPtrs...); err != nil {
			return nil, fmt.Errorf("sqlblade: failed to scan row: %w", err)
		}

		for _, field := range info.fields {
			colIdx, ok := columnMap[strings.ToLower(field.dbColumn)]
			if !ok {
				continue
			}

			fieldVal := ptrVal.Field(field.index)
			if !fieldVal.IsValid() || !fieldVal.CanSet() {
				continue
			}

			scanVal := scanValues[colIdx]
			if scanVal == nil {
				if field.isPtr {
					fieldVal.Set(reflect.Zero(fieldVal.Type()))
				}
				continue
			}

			if err := setFieldValue(fieldVal, scanVal, field.fieldType); err != nil {
				return nil, fmt.Errorf("sqlblade: failed to set field %s: %w", field.name, err)
			}
		}

		result = append(result, val)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
