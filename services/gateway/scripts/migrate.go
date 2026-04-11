package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		log.Fatal("POSTGRES_URL environment variable is not set")
	}

	migrationPath := "/app/migrations/20260410_auth_provisioning.sql"
	if len(os.Args) > 1 {
		migrationPath = os.Args[1]
	}

	fmt.Printf("Applying migration: %s\n", migrationPath)

	content, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("Migration applied successfully!")
}
