package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/alicanli1995/sqlblade/sqlblade"
	"github.com/alicanli1995/sqlblade/sqlblade/dialect"
	_ "github.com/go-sql-driver/mysql"
)

// User represents a user model
type User struct {
	ID        int       `db:"id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func main() {
	// Connect to database
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Example 1: Simple SELECT query
	fmt.Println("=== Example 1: Simple SELECT ===")
	users, err := sqlblade.Query[User](db).
		Where("age", ">", 18).
		Where("status", "=", "active").
		OrderBy("created_at", dialect.DESC).
		Limit(10).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d users\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
		}
	}

	// Example 2: COUNT query
	fmt.Println("\n=== Example 2: COUNT ===")
	count, err := sqlblade.Query[User](db).
		Where("age", ">", 18).
		Count(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Total users: %d\n", count)
	}

	// Example 3: INSERT
	fmt.Println("\n=== Example 3: INSERT ===")
	newUser := User{
		Email:     "john@example.com",
		Name:      "John Doe",
		Age:       25,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result, err := sqlblade.Insert(db, newUser).Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		id, _ := result.LastInsertId()
		fmt.Printf("Inserted user with ID: %d\n", id)
	}

	// Example 4: UPDATE
	fmt.Println("\n=== Example 4: UPDATE ===")
	_, err = sqlblade.Update[User](db).
		Set("status", "inactive").
		Where("last_login", "<", time.Now().AddDate(0, -6, 0)).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Updated users successfully")
	}

	// Example 5: DELETE
	fmt.Println("\n=== Example 5: DELETE ===")
	_, err = sqlblade.Delete[User](db).
		Where("status", "=", "deleted").
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Deleted users successfully")
	}

	// Example 6: Transaction
	fmt.Println("\n=== Example 6: Transaction ===")
	err = sqlblade.WithTransaction(db, func(tx *sql.Tx) error {
		user := User{
			Email:     "jane@example.com",
			Name:      "Jane Doe",
			Age:       30,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := sqlblade.InsertTx(tx, user).Execute(ctx)
		if err != nil {
			return err
		}

		_, err = sqlblade.UpdateTx[User](tx).
			Set("status", "verified").
			Where("email", "=", "jane@example.com").
			Execute(ctx)
		return err
	})
	if err != nil {
		log.Printf("Transaction error: %v", err)
	} else {
		fmt.Println("Transaction completed successfully")
	}

	// Example 7: JOIN
	fmt.Println("\n=== Example 7: JOIN ===")
	usersWithPosts, err := sqlblade.Query[User](db).
		Join("posts", "users.id = posts.user_id").
		Where("posts.published", "=", true).
		Select("users.*").
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d users with published posts\n", len(usersWithPosts))
	}

	// Example 8: Raw query
	fmt.Println("\n=== Example 8: Raw Query ===")
	rawUsers, err := sqlblade.Raw[User](db, "SELECT * FROM users WHERE age > ?", 18).Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d users via raw query\n", len(rawUsers))
	}

	// Example 9: Aggregate functions
	fmt.Println("\n=== Example 9: Aggregate Functions ===")
	avgAge, err := sqlblade.Query[User](db).
		Where("status", "=", "active").
		Avg(ctx, "age")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Average age: %.2f\n", avgAge)
	}

	// Example 10: Query Preview - See SQL without executing
	fmt.Println("\n=== Example 10: Query Preview ===")
	previewQuery := sqlblade.Query[User](db).
		Where("email", "LIKE", "%@example.com%").
		Join("profiles", "profiles.user_id = users.id").
		OrderBy("created_at", dialect.DESC)

	preview := previewQuery.Preview()
	fmt.Println("Generated SQL:", preview.SQL())
	fmt.Println("\nSQL with substituted args:", preview.SQLWithArgs())
	fmt.Println("\nQuery arguments:", preview.Args())

	// Example 11: Query Fragments - Reusable query parts
	fmt.Println("\n=== Example 11: Query Fragments ===")
	// Create a reusable fragment
	activeUsersFragment := sqlblade.NewQueryFragment().
		Where("status", "=", "active").
		OrderBy("created_at", dialect.DESC)

	// Apply fragment to multiple queries
	recentActiveUsers, err := sqlblade.Query[User](db).
		Apply(activeUsersFragment).
		Limit(5).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d recent active users\n", len(recentActiveUsers))
	}

	// Example 12: Subqueries - Powerful WHERE conditions
	fmt.Println("\n=== Example 12: Subqueries ===")
	// Find users who have recent posts (example subquery)
	usersWithRecentPosts, err := sqlblade.Query[User](db).
		WhereSubquery("id", "IN", sqlblade.NewSubquery(
			sqlblade.Query[struct {
				UserID int `db:"user_id"`
			}](db).
				Select("user_id").
				Where("created_at", ">", time.Now().AddDate(0, -1, 0)),
		)).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d users with recent posts\n", len(usersWithRecentPosts))
	}

	// Example 13: EXISTS - Check existence efficiently
	fmt.Println("\n=== Example 13: EXISTS Queries ===")
	hasActiveUsers, err := sqlblade.Query[User](db).
		Where("status", "=", "active").
		Limit(1).
		Exists(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		if hasActiveUsers {
			fmt.Println("✅ There are active users!")
		} else {
			fmt.Println("❌ No active users found")
		}
	}

	// Example 14: SQL Query Debugging
	fmt.Println("\n=== Example 14: SQL Query Debugging ===")
	fmt.Println("Enable debug mode to see beautiful SQL logging:")
	fmt.Println("  sqlblade.EnableDebug()")
	fmt.Println("  sqlblade.ConfigureDebug(func(qd *sqlblade.QueryDebugger) { ... })")
	fmt.Println("\nAll queries will be logged with timing, parameters, and formatted SQL!")
}
