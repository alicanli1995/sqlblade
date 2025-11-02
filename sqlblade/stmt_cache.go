package sqlblade

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"sync"
)

// stmtCache caches prepared statements by SQL query hash
type stmtCache struct {
	mu    sync.RWMutex
	store map[string]*sql.Stmt
	db    *sql.DB
}

var globalStmtCache *stmtCache
var stmtCacheOnce sync.Once

// initStmtCache initializes the global statement cache
func initStmtCache(db *sql.DB) *stmtCache {
	stmtCacheOnce.Do(func() {
		globalStmtCache = &stmtCache{
			store: make(map[string]*sql.Stmt),
			db:    db,
		}
	})
	return globalStmtCache
}

// getStmt returns a cached prepared statement or creates a new one
func (sc *stmtCache) getStmt(ctx context.Context, sqlStr string) (*sql.Stmt, error) {
	// Hash SQL string to use as cache key
	hash := hashSQL(sqlStr)

	// Try to get from cache
	sc.mu.RLock()
	if stmt, ok := sc.store[hash]; ok {
		sc.mu.RUnlock()
		return stmt, nil
	}
	sc.mu.RUnlock()

	// Create new prepared statement
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Double-check after acquiring write lock
	if stmt, ok := sc.store[hash]; ok {
		return stmt, nil
	}

	stmt, err := sc.db.PrepareContext(ctx, sqlStr)
	if err != nil {
		return nil, err
	}

	sc.store[hash] = stmt
	return stmt, nil
}

// hashSQL creates a SHA256 hash of SQL string for cache key
func hashSQL(sqlStr string) string {
	h := sha256.Sum256([]byte(sqlStr))
	return hex.EncodeToString(h[:])
}

// clearStmtCache clears all cached statements
func (sc *stmtCache) clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, stmt := range sc.store {
		stmt.Close()
	}
	sc.store = make(map[string]*sql.Stmt)
}

// PreparedStatementCache enables prepared statement caching for a database connection
// This should be called once per database connection for optimal performance
func PreparedStatementCache(db *sql.DB) {
	initStmtCache(db)
}
