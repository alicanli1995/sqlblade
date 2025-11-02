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

// Post represents a blog post model
type Post struct {
	ID        int       `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	AuthorID  int       `db:"author_id"`
	Published bool      `db:"published"`
	Views     int       `db:"views"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Author represents an author model
type Author struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Bio       string    `db:"bio"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	// Connect to PostgreSQL database
	connStr := "postgres://user:password@localhost/mydb?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}(db)

	ctx := context.Background()

	// Example 1: SELECT with complex WHERE conditions
	fmt.Println("=== Example 1: Complex SELECT Query ===")
	posts, err := sqlblade.Query[Post](db).
		Where("published", "=", true).
		Where("views", ">", 100).
		OrWhere("author_id", "=", 1).
		OrderBy("created_at", dialect.DESC).
		Limit(20).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d published posts\n", len(posts))
		for _, post := range posts {
			fmt.Printf("  - %s (%d views)\n", post.Title, post.Views)
		}
	}

	// Example 2: INSERT with RETURNING (PostgreSQL specific)
	fmt.Println("\n=== Example 2: INSERT with RETURNING ===")
	newPost := Post{
		Title:     "Getting Started with SQLBlade",
		Content:   "SQLBlade is a modern, type-safe query builder...",
		AuthorID:  1,
		Published: true,
		Views:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result, err := sqlblade.Insert(db, newPost).
		Returning("id").
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		// In PostgreSQL, we can get the ID from the result
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("Inserted post, rows affected: %d\n", rowsAffected)
	}

	// Example 3: Batch INSERT
	fmt.Println("\n=== Example 3: Batch INSERT ===")
	postsBatch := []Post{
		{Title: "Post 1", Content: "Content 1", AuthorID: 1, Published: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "Post 2", Content: "Content 2", AuthorID: 1, Published: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "Post 3", Content: "Content 3", AuthorID: 2, Published: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	_, err = sqlblade.InsertBatch(db, postsBatch).Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Batch insert completed successfully")
	}

	// Example 4: UPDATE with RETURNING
	fmt.Println("\n=== Example 4: UPDATE with RETURNING ===")
	_, err = sqlblade.Update[Post](db).
		Set("views", 0).
		Set("updated_at", time.Now()).
		Where("published", "=", false).
		Returning("id", "title").
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Updated posts successfully")
	}

	// Example 5: JOIN with multiple tables
	fmt.Println("\n=== Example 5: JOIN Query ===")
	postsWithAuthors, err := sqlblade.Query[Post](db).
		LeftJoin("authors", "posts.author_id = authors.id").
		Where("posts.published", "=", true).
		Select("posts.*, authors.name as author_name").
		OrderBy("posts.created_at", dialect.DESC).
		Limit(10).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d posts with author information\n", len(postsWithAuthors))
	}

	// Example 6: GROUP BY and Aggregate functions
	fmt.Println("\n=== Example 6: Aggregate Functions ===")

	// Count posts by author
	totalPosts, err := sqlblade.Query[Post](db).
		Where("published", "=", true).
		Count(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Total published posts: %d\n", totalPosts)
	}

	// Average views
	avgViews, err := sqlblade.Query[Post](db).
		Where("published", "=", true).
		Avg(ctx, "views")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Average views: %.2f\n", avgViews)
	}

	// Max views
	maxViews, err := sqlblade.Query[Post](db).
		Where("published", "=", true).
		Max(ctx, "views")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Maximum views: %v\n", maxViews)
	}

	// Example 7: Complex WHERE with IN clause
	fmt.Println("\n=== Example 7: WHERE with IN ===")
	authorIDs := []interface{}{1, 2, 3}
	postsByAuthors, err := sqlblade.Query[Post](db).
		Where("author_id", "IN", authorIDs).
		Where("published", "=", true).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d posts from specified authors\n", len(postsByAuthors))
	}

	// Example 8: Transaction with multiple operations
	fmt.Println("\n=== Example 8: Complex Transaction ===")
	err = sqlblade.WithTransactionContext(ctx, db, func(tx *sql.Tx) error {
		// Insert new post
		newPost := Post{
			Title:     "Transaction Example",
			Content:   "This is a transaction example",
			AuthorID:  1,
			Published: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := sqlblade.InsertTx(tx, newPost).Execute(ctx)
		if err != nil {
			return err
		}

		// Update author's post count (simulated)
		_, err = sqlblade.UpdateTx[Post](tx).
			Set("published", true).
			Where("title", "=", "Transaction Example").
			Execute(ctx)
		if err != nil {
			return err
		}

		// Delete old unpublished posts
		_, err = sqlblade.DeleteTx[Post](tx).
			Where("published", "=", false).
			Where("created_at", "<", time.Now().AddDate(0, -1, 0)).
			Execute(ctx)
		return err
	})
	if err != nil {
		log.Printf("Transaction error: %v", err)
	} else {
		fmt.Println("Complex transaction completed successfully")
	}

	// Example 9: Raw query with complex SQL
	fmt.Println("\n=== Example 9: Raw Query ===")
	rawPosts, err := sqlblade.Raw[Post](db, `
		SELECT * FROM posts 
		WHERE published = $1 
		AND views > $2 
		ORDER BY created_at DESC 
		LIMIT $3
	`, true, 50, 10).Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d posts via raw query\n", len(rawPosts))
	}

	// Example 10: DISTINCT query
	fmt.Println("\n=== Example 10: DISTINCT Query ===")
	uniqueAuthors, err := sqlblade.Query[Post](db).
		Distinct().
		Select("author_id").
		Where("published", "=", true).
		Execute(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d unique authors with published posts\n", len(uniqueAuthors))
	}
}
