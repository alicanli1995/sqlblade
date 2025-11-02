package sqlblade

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"
)

// columnMapCache caches column name to index mappings for faster scanning
type columnMapCache struct {
	mu    sync.RWMutex
	store map[string]map[string]int // map[columnsKey]map[columnName]index
}

var columnMapCacheInst = &columnMapCache{
	store: make(map[string]map[string]int),
}

// getColumnMap returns cached column map or creates a new one
func (cmc *columnMapCache) getColumnMap(columns []string) map[string]int {
	// Create key from columns
	key := strings.Join(columns, ",")

	// Try cache
	cmc.mu.RLock()
	if cached, ok := cmc.store[key]; ok {
		cmc.mu.RUnlock()
		return cached
	}
	cmc.mu.RUnlock()

	// Build new map
	columnMap := make(map[string]int, len(columns))
	for i, col := range columns {
		columnMap[strings.ToLower(col)] = i
	}

	// Cache it
	cmc.mu.Lock()
	cmc.store[key] = columnMap
	cmc.mu.Unlock()

	return columnMap
}

// scanRowsOptimized scans database rows with optimized reflection and caching
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

	// Use cached column map
	columnMap := columnMapCacheInst.getColumnMap(columns)

	// Pre-allocate result slice with estimated capacity
	// Most queries return 10-100 rows, so we start with 10
	result = make([]T, 0, 10)

	// Pre-allocate scan values slice
	scanValues := make([]interface{}, len(columns))
	scanPtrs := make([]interface{}, len(columns))
	for i := range scanValues {
		scanPtrs[i] = &scanValues[i]
	}

	for rows.Next() {
		var val T
		ptrVal := reflect.ValueOf(&val).Elem()

		// Scan into pre-allocated slice
		if err := rows.Scan(scanPtrs...); err != nil {
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

			scanVal := scanValues[colIdx]
			if scanVal == nil {
				if field.isPtr {
					fieldVal.Set(reflect.Zero(fieldVal.Type()))
				}
				continue
			}

			// Convert and set the value
			if err := setFieldValue(fieldVal, scanVal, field.fieldType); err != nil {
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
