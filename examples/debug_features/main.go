package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/alicanli1995/sqlblade/sqlblade"
	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
	_ "github.com/lib/pq"
)

type User struct {
	ID        int       `db:"id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	// Connect to database (adjust connection string as needed)
	db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// ============================================
	// 1. SQL Query Debugging - Beautiful logging!
	// ============================================
	fmt.Println("=== 1. SQL Query Debugging ===")
	sqlblade.EnableDebug()

	// Configure debugger (optional)
	sqlblade.ConfigureDebug(func(qd *sqlblade.QueryDebugger) {
		qd.ShowArgs(true)
		qd.ShowTiming(true)
		qd.SetSlowQueryThreshold(50 * time.Millisecond)
	})

	// Now all queries will be logged beautifully!
	users, _ := sqlblade.Query[User](db).
		Where("age", ">", 18).
		Where("status", "=", "active").
		OrderBy("created_at", dialect.DESC).
		Limit(10).
		Execute(ctx)

	fmt.Printf("Found %d users\n\n", len(users))

	// ============================================
	// 2. Query Preview - See SQL without executing!
	// ============================================
	fmt.Println("=== 2. Query Preview ===")

	query := sqlblade.Query[User](db).
		Where("email", "LIKE", "%@example.com%").
		Join("profiles", "profiles.user_id = users.id").
		GroupBy("users.id").
		OrderBy("created_at", dialect.DESC)

	// Preview the SQL without executing
	preview := query.Preview()
	fmt.Println("SQL:", preview.SQL())
	fmt.Println("\nSQL with args:", preview.SQLWithArgs())
	fmt.Println("\nArgs:", preview.Args())

	// Pretty print the query
	fmt.Println("\nPretty print:")
	preview.PrettyPrint()

	// You can still execute it later!
	// results, _ := preview.Execute(ctx)

	// ============================================
	// 3. Query Fragments - Reusable query parts!
	// ============================================
	fmt.Println("\n=== 3. Query Fragments ===")

	// Create a reusable fragment
	activeUsersFragment := sqlblade.NewQueryFragment().
		Where("status", "=", "active").
		Where("email_verified", "=", true).
		OrderBy("created_at", dialect.DESC)

	// Apply fragment to multiple queries
	recentActiveUsers, _ := sqlblade.Query[User](db).
		Apply(activeUsersFragment).
		Limit(10).
		Execute(ctx)

	fmt.Printf("Found %d recent active users\n", len(recentActiveUsers))

	// Reuse the same fragment with different limits
	allActiveUsers, _ := sqlblade.Query[User](db).
		Apply(activeUsersFragment).
		Execute(ctx)

	fmt.Printf("Found %d total active users\n\n", len(allActiveUsers))

	// ============================================
	// 4. Subqueries - Powerful WHERE conditions!
	// ============================================
	fmt.Println("=== 4. Subqueries ===")

	// Find users who have placed orders
	usersWithOrders, _ := sqlblade.Query[User](db).
		WhereSubquery("id", "IN", sqlblade.NewSubquery(
			sqlblade.Query[struct {
				UserID int `db:"user_id"`
			}](db).
				Select("user_id").
				Where("status", "=", "completed"),
		)).
		Execute(ctx)

	fmt.Printf("Found %d users with orders\n\n", len(usersWithOrders))

	// ============================================
	// 5. EXISTS - Check existence efficiently!
	// ============================================
	fmt.Println("=== 5. EXISTS Queries ===")

	hasActiveUsers, _ := sqlblade.Query[User](db).
		Where("status", "=", "active").
		Limit(1).
		Exists(ctx)

	if hasActiveUsers {
		fmt.Println("✅ There are active users!")
	}

	// ============================================
	// 6. Custom Logging Hook
	// ============================================
	fmt.Println("\n=== 6. Custom Logging Hook ===")

	// Add a custom hook for query logging
	sqlblade.DefaultHooks.BeforeQuery(func(ctx context.Context, query string, args []interface{}) error {
		fmt.Printf("🔍 Executing query: %s\n", query)
		return nil
	})

	// Now all queries will trigger this hook
	_, _ = sqlblade.Query[User](db).
		Where("age", ">", 25).
		Execute(ctx)

	fmt.Println("\n✨ All features demonstrated!")
}
