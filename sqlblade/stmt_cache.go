package sqlblade

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"sync"
)

type stmtCache struct {
	mu    sync.RWMutex
	store map[string]*sql.Stmt
	db    *sql.DB
}

var (
	globalStmtCache *stmtCache
	stmtCacheOnce   sync.Once
)

func initStmtCache(db *sql.DB) *stmtCache {
	stmtCacheOnce.Do(func() {
		globalStmtCache = &stmtCache{
			store: make(map[string]*sql.Stmt),
			db:    db,
		}
	})
	return globalStmtCache
}

func (sc *stmtCache) getStmt(ctx context.Context, sqlStr string) (*sql.Stmt, error) {
	hash := hashSQL(sqlStr)

	sc.mu.RLock()
	if stmt, ok := sc.store[hash]; ok {
		sc.mu.RUnlock()
		return stmt, nil
	}
	sc.mu.RUnlock()

	sc.mu.Lock()
	defer sc.mu.Unlock()

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

func hashSQL(sqlStr string) string {
	h := sha256.Sum256([]byte(sqlStr))
	return hex.EncodeToString(h[:])
}

// ClearStmtCache clears all cached statements
func ClearStmtCache() {
	if globalStmtCache != nil {
		globalStmtCache.clear()
	}
}

func (sc *stmtCache) clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, stmt := range sc.store {
		_ = stmt.Close()
	}
	sc.store = make(map[string]*sql.Stmt)
}

func PreparedStatementCache(db *sql.DB) {
	initStmtCache(db)
}
