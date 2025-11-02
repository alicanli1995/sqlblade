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

// fastColumnKey generates a cache key faster than strings.Join
func fastColumnKey(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	if len(columns) == 1 {
		return columns[0]
	}

	// Pre-allocate buffer with approximate size
	estimatedSize := len(columns) * 10
	var buf strings.Builder
	buf.Grow(estimatedSize)

	for i, col := range columns {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(col)
	}
	return buf.String()
}

// toLowerCache caches lowercase conversions
var toLowerCache = sync.Map{}

func cachedToLower(s string) string {
	if cached, ok := toLowerCache.Load(s); ok {
		return cached.(string)
	}
	lower := strings.ToLower(s)
	toLowerCache.Store(s, lower)
	return lower
}

func (cmc *columnMapCache) getColumnMap(columns []string) map[string]int {
	key := fastColumnKey(columns)

	cmc.mu.RLock()
	if cached, ok := cmc.store[key]; ok {
		cmc.mu.RUnlock()
		return cached
	}
	cmc.mu.RUnlock()

	columnMap := make(map[string]int, len(columns))
	for i, col := range columns {
		columnMap[cachedToLower(col)] = i
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

	result = make([]T, 0, resultInitialCapacity)

	scanBuf := globalScanBufferPool.Get(len(columns))
	defer globalScanBufferPool.Put(scanBuf)

	for rows.Next() {
		var val T
		ptrVal := reflect.ValueOf(&val).Elem()

		if err := rows.Scan(scanBuf.ptrs...); err != nil {
			return nil, fmt.Errorf("sqlblade: failed to scan row: %w", err)
		}

		for _, field := range info.fields {
			colIdx, ok := columnMap[field.dbColumn]
			if !ok {
				continue
			}

			fieldVal := ptrVal.Field(field.index)
			if !fieldVal.IsValid() || !fieldVal.CanSet() {
				continue
			}

			scanVal := scanBuf.values[colIdx]
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
