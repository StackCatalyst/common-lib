package main

import (
	"context"
	"log"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/database"
	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	// Example 1: Database setup
	config := database.DefaultConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.Database = "myapp"
	config.User = "admin"
	config.Password = "password"
	config.MaxConns = 10
	config.MinConns = 2

	metricsReporter := metrics.New(metrics.DefaultOptions())
	db, err := database.New(config, metricsReporter)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Example 2: Basic query execution
	rows, err := db.Query(ctx, "SELECT id, name, email, created_at FROM users WHERE active = $1", true)
	if err != nil {
		log.Printf("Query failed: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		users = append(users, user)
	}

	// Example 3: Single row query
	row := db.QueryRow(ctx, "SELECT id, name, email, created_at FROM users WHERE id = $1", "user123")
	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
		log.Printf("Failed to find user: %v", err)
	}

	// Example 4: Transaction handling
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
	}

	// Execute queries within transaction
	_, err = tx.Exec(ctx, "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
		"new-user", "New User", "new@example.com")
	if err != nil {
		tx.Rollback(ctx)
		log.Printf("Failed to insert user: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	}

	// Example 5: Batch operations
	batch := &pgx.Batch{}
	batch.Queue("UPDATE users SET last_login = $1 WHERE id = $2", []interface{}{time.Now(), "user1"})
	batch.Queue("UPDATE users SET last_login = $1 WHERE id = $2", []interface{}{time.Now(), "user2"})

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Execute each query in the batch
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			log.Printf("Failed to execute batch query %d: %v", i, err)
		}
	}

	// Example 6: Connection pool metrics
	db.UpdatePoolStats()
}
