package sqlblade

import (
	"sync"
)

// ScanBufferPool provides reusable scan buffers to reduce allocations
type ScanBufferPool struct {
	pool *sync.Pool
}

var globalScanBufferPool = &ScanBufferPool{
	pool: &sync.Pool{
		New: func() interface{} {
			return &scanBuffer{
				values: make([]interface{}, 0, 16),
				ptrs:   make([]interface{}, 0, 16),
			}
		},
	},
}

type scanBuffer struct {
	values []interface{}
	ptrs   []interface{}
}

func (sbp *ScanBufferPool) Get(size int) *scanBuffer {
	bufInterface := sbp.pool.Get()
	buf, ok := bufInterface.(*scanBuffer)
	if !ok {
		buf = &scanBuffer{
			values: make([]interface{}, 0, 16),
			ptrs:   make([]interface{}, 0, 16),
		}
	}
	if cap(buf.values) < size {
		buf.values = make([]interface{}, size)
		buf.ptrs = make([]interface{}, size)
	} else {
		buf.values = buf.values[:size]
		buf.ptrs = buf.ptrs[:size]
	}
	for i := range buf.values {
		buf.ptrs[i] = &buf.values[i]
	}
	return buf
}

func (sbp *ScanBufferPool) Put(buf *scanBuffer) {
	for i := range buf.values {
		buf.values[i] = nil
	}
	sbp.pool.Put(buf)
}

// TableNameCache caches TableName() method results
type tableNameCache struct {
	mu    sync.RWMutex
	cache map[string]string
}

var globalTableNameCache = &tableNameCache{
	cache: make(map[string]string),
}

func (tnc *tableNameCache) get(structTypeName string) (string, bool) {
	tnc.mu.RLock()
	defer tnc.mu.RUnlock()
	val, ok := tnc.cache[structTypeName]
	return val, ok
}

func (tnc *tableNameCache) set(structTypeName, tableName string) {
	tnc.mu.Lock()
	defer tnc.mu.Unlock()
	tnc.cache[structTypeName] = tableName
}
