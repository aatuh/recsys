package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"recsys/internal/migrator"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Build with: go build -o bin/migrate ./cmd/migrate
// Usage:
//
//	DATABASE_URL=postgres://... bin/migrate up
//	DATABASE_URL=postgres://... bin/migrate down
//	DATABASE_URL=postgres://... bin/migrate status
func main() {
	var (
		dir    = flag.String("dir", "", "migrations dir override")
		table  = flag.String("table", "", "schema_migrations table")
		lock   = flag.Int64("lock", 0, "advisory lock key")
		allowD = flag.Bool("allow-down", false, "enable down")
	)
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("command required: up | down | status")
	}
	cmd := strings.ToLower(flag.Args()[0])

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL env is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	opts := migrator.Options{
		MigrationsDir:      *dir,
		TableName:          *table,
		LockKey:            *lock,
		AllowDangerousDown: *allowD,
		Logger: func(f string, a ...any) {
			log.Printf(f, a...)
		},
	}
	r := migrator.New(db, embedded(), opts)

	ctx, cancel := context.WithTimeout(context.Background(),
		15*time.Minute)
	defer cancel()

	switch cmd {
	case "up":
		if err := r.Up(ctx); err != nil {
			log.Fatal(err)
		}
	case "down":
		if err := r.Down(ctx); err != nil {
			log.Fatal(err)
		}
	case "status":
		s, err := r.Status(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(s)
	default:
		log.Fatalf("unknown command: %s", cmd)
	}
}

// embedded returns the embedded FS with migrations. If you prefer to
// always load from disk, return nil here and use -dir or MIGRATIONS_DIR.
func embedded() fs.FS {
	// Uncomment to embed. Place files under internal/migrator/migrations/.
	//
	// //go:embed ../../internal/migrator/migrations/*.sql
	// var mfs embed.FS
	// return mfs
	return nil
}
