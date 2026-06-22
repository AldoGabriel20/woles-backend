package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/woles/woles-backend/internal/migration"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down|status|reset|version>")
	}
	command := os.Args[1]

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	goose.SetBaseFS(migration.FS)

	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	if err = goose.RunContext(nil, command, db, "."); err != nil { //nolint:staticcheck
		log.Fatalf("goose %s: %v", command, err)
	}
}
