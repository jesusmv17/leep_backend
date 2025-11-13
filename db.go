package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

// InitDB loads .env, connects to Supabase Postgres, and stores the pool in `db`.
func InitDB() {
	// Load local env vars (DATABASE_URL=...)
	err := godotenv.Load()
	if err != nil {
		// not fatal in production, but locally we expect .env to exist
		log.Println("⚠️  No .env file found, continuing anyway")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL is not set in environment (.env)")
	}

	// Create a connection pool
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to create DB pool: %v", err)
	}

	// Ping to verify connection works
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	db = pool
	fmt.Println("✅ Connected to Supabase Postgres successfully!")
}
