package sqlblade

import (
	"context"
	"database/sql"
)

// QueryHook defines a hook function that can be called before or after queries
type QueryHook func(ctx context.Context, query string, args []interface{}) error

// HookType represents the type of hook
type HookType int

const (
	// BeforeQuery hook is called before executing a query
	BeforeQuery HookType = iota
	// AfterQuery hook is called after executing a query successfully
	AfterQuery
)

// Hooks manages query hooks
type Hooks struct {
	beforeQuery []QueryHook
	afterQuery  []QueryHook
}

// NewHooks creates a new hooks manager
func NewHooks() *Hooks {
	return &Hooks{
		beforeQuery: make([]QueryHook, 0),
		afterQuery:  make([]QueryHook, 0),
	}
}

// BeforeQuery adds a hook to be called before query execution
func (h *Hooks) BeforeQuery(hook QueryHook) {
	h.beforeQuery = append(h.beforeQuery, hook)
}

// AfterQuery adds a hook to be called after query execution
func (h *Hooks) AfterQuery(hook QueryHook) {
	h.afterQuery = append(h.afterQuery, hook)
}

// executeBeforeHooks executes all before query hooks
func (h *Hooks) executeBeforeHooks(ctx context.Context, query string, args []interface{}) error {
	for _, hook := range h.beforeQuery {
		if err := hook(ctx, query, args); err != nil {
			return err
		}
	}
	return nil
}

// executeAfterHooks executes all after query hooks
func (h *Hooks) executeAfterHooks(ctx context.Context, query string, args []interface{}) error {
	for _, hook := range h.afterQuery {
		if err := hook(ctx, query, args); err != nil {
			return err
		}
	}
	return nil
}

// DefaultHooks is a global hooks instance
var DefaultHooks = NewHooks()

