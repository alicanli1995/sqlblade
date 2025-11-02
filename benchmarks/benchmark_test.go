package benchmarks

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/alicanli1995/sqlblade/sqlblade"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// BenchmarkUser represents a test user model
type BenchmarkUser struct {
	ID    int    `db:"id" gorm:"column:id"`
	Email string `db:"email" gorm:"column:email"`
	Name  string `db:"name" gorm:"column:name"`
	Age   int    `db:"age" gorm:"column:age"`
}

// TableName returns the table name for BenchmarkUser
func (BenchmarkUser) TableName() string {
	return "benchmark_users"
}

// GORM model
type GormUser struct {
	ID    int `gorm:"primaryKey"`
	Email string
	Name  string
	Age   int
}

func (GormUser) TableName() string {
	return "benchmark_users"
}

var testDB *sql.DB
var sqlxDB *sqlx.DB
var gormDB *gorm.DB
var ctx context.Context

func init() {
	// Initialize test database connection
	connStr := os.Getenv("DB_CONN")
	if connStr == "" {
		connStr = "postgres://benchmark_user:benchmark_pass@localhost:5433/benchmark_db?sslmode=disable"
	}

	// SQL DB
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	testDB = db

	// Enable prepared statement cache for SQLBlade
	sqlblade.PreparedStatementCache(db)

	// SQLX DB
	sqlxDB = sqlx.NewDb(db, "postgres")

	// GORM DB
	gormDB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic(err)
	}

	ctx = context.Background()

	// Ensure table exists
	testDB.Exec(`
		CREATE TABLE IF NOT EXISTS benchmark_users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255),
			name VARCHAR(255),
			age INTEGER
		)
	`)

	// Insert some test data if table is empty
	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM benchmark_users").Scan(&count)
	if count == 0 {
		// Insert sample data for testing
		for i := 0; i < 100; i++ {
			testDB.Exec(`
				INSERT INTO benchmark_users (email, name, age) 
				VALUES ($1, $2, $3)
			`, fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("User %d", i), 20+i%50)
		}
	}
}

// ========== SQLBlade Benchmarks ==========

func BenchmarkSQLBlade_Select(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = sqlblade.Query[BenchmarkUser](testDB).
			Where("id", ">", 0).
			Limit(10).
			Execute(ctx)
	}
}

func BenchmarkSQLBlade_SelectComplex(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = sqlblade.Query[BenchmarkUser](testDB).
			Where("age", ">", 18).
			Where("age", "<", 65).
			OrderBy("id", 0).
			Limit(10).
			Offset(0).
			Execute(ctx)
	}
}

func BenchmarkSQLBlade_Insert(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		user := BenchmarkUser{
			Email: fmt.Sprintf("test%d@example.com", i),
			Name:  "Test User",
			Age:   25,
		}
		_, _ = sqlblade.Insert(testDB, user).Execute(ctx)
	}
}

func BenchmarkSQLBlade_Update(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = sqlblade.Update[BenchmarkUser](testDB).
			Set("age", 30).
			Where("id", "=", 1).
			Execute(ctx)
	}
}

func BenchmarkSQLBlade_Count(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = sqlblade.Query[BenchmarkUser](testDB).
			Where("age", ">", 18).
			Count(ctx)
	}
}

// ========== GORM Benchmarks ==========

func BenchmarkGORM_Select(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var users []GormUser
		_ = gormDB.WithContext(ctx).
			Where("id > ?", 0).
			Limit(10).
			Find(&users)
	}
}

func BenchmarkGORM_SelectComplex(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var users []GormUser
		_ = gormDB.WithContext(ctx).
			Where("age > ?", 18).
			Where("age < ?", 65).
			Order("id ASC").
			Limit(10).
			Offset(0).
			Find(&users)
	}
}

func BenchmarkGORM_Insert(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		user := GormUser{
			Email: fmt.Sprintf("test%d@example.com", i),
			Name:  "Test User",
			Age:   25,
		}
		_ = gormDB.WithContext(ctx).Create(&user)
	}
}

func BenchmarkGORM_Update(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gormDB.WithContext(ctx).
			Model(&GormUser{}).
			Where("id = ?", 1).
			Update("age", 30)
	}
}

func BenchmarkGORM_Count(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var count int64
		_ = gormDB.WithContext(ctx).
			Model(&GormUser{}).
			Where("age > ?", 18).
			Count(&count)
	}
}

// ========== SQLX Benchmarks ==========

func BenchmarkSQLX_Select(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var users []BenchmarkUser
		_ = sqlxDB.SelectContext(ctx, &users, "SELECT * FROM benchmark_users WHERE id > $1 LIMIT 10", 0)
	}
}

func BenchmarkSQLX_SelectComplex(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var users []BenchmarkUser
		_ = sqlxDB.SelectContext(ctx, &users,
			"SELECT * FROM benchmark_users WHERE age > $1 AND age < $2 ORDER BY id ASC LIMIT $3 OFFSET $4",
			18, 65, 10, 0)
	}
}

func BenchmarkSQLX_Insert(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		user := BenchmarkUser{
			Email: fmt.Sprintf("test%d@example.com", i),
			Name:  "Test User",
			Age:   25,
		}
		_, _ = sqlxDB.NamedExecContext(ctx,
			"INSERT INTO benchmark_users (email, name, age) VALUES (:email, :name, :age)",
			user)
	}
}

func BenchmarkSQLX_Update(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = sqlxDB.ExecContext(ctx,
			"UPDATE benchmark_users SET age = $1 WHERE id = $2",
			30, 1)
	}
}

func BenchmarkSQLX_Count(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var count int
		_ = sqlxDB.GetContext(ctx, &count, "SELECT COUNT(*) FROM benchmark_users WHERE age > $1", 18)
	}
}

// ========== Standard Library Benchmarks ==========

func BenchmarkStdlib_Select(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, err := testDB.QueryContext(ctx, "SELECT * FROM benchmark_users WHERE id > $1 LIMIT 10", 0)
		if err != nil {
			b.Fatal(err)
		}
		var users []BenchmarkUser
		for rows.Next() {
			var u BenchmarkUser
			if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Age); err != nil {
				rows.Close()
				b.Fatal(err)
			}
			users = append(users, u)
		}
		rows.Close()
		_ = users
	}
}

func BenchmarkStdlib_SelectComplex(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, err := testDB.QueryContext(ctx,
			"SELECT * FROM benchmark_users WHERE age > $1 AND age < $2 ORDER BY id ASC LIMIT $3 OFFSET $4",
			18, 65, 10, 0)
		if err != nil {
			b.Fatal(err)
		}
		var users []BenchmarkUser
		for rows.Next() {
			var u BenchmarkUser
			if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Age); err != nil {
				rows.Close()
				b.Fatal(err)
			}
			users = append(users, u)
		}
		rows.Close()
		_ = users
	}
}

func BenchmarkStdlib_Insert(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = testDB.ExecContext(ctx,
			"INSERT INTO benchmark_users (email, name, age) VALUES ($1, $2, $3)",
			fmt.Sprintf("test%d@example.com", i), "Test User", 25)
	}
}

func BenchmarkStdlib_Update(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = testDB.ExecContext(ctx,
			"UPDATE benchmark_users SET age = $1 WHERE id = $2",
			30, 1)
	}
}

func BenchmarkStdlib_Count(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var count int
		_ = testDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM benchmark_users WHERE age > $1", 18).Scan(&count)
	}
}
