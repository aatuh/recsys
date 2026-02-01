package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type migrationFile struct {
	Version  int64
	Name     string
	Dir      string
	Checksum string
}

type appliedRow struct {
	Version  int64
	Name     string
	Checksum string
	Success  bool
}

func runPreflight(args []string) {
	flags := flag.NewFlagSet("preflight", flag.ExitOnError)
	dir := flags.String("dir", "", "migrations dir override")
	table := flags.String("table", "schema_migrations", "schema_migrations table")
	if err := flags.Parse(args); err != nil {
		log.Fatalf("parse flags: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dsn) == "" {
		log.Fatal("DATABASE_URL env is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	applied, err := loadApplied(ctx, db, *table)
	if err != nil {
		log.Fatalf("load applied: %v", err)
	}
	migrationsUp, err := loadMigrations(*dir)
	if err != nil {
		log.Fatalf("load migrations: %v", err)
	}

	var failed []appliedRow
	for _, row := range applied {
		if !row.Success {
			failed = append(failed, row)
		}
	}
	if len(failed) > 0 {
		for _, row := range failed {
			log.Printf("failed migration: %d %s", row.Version, row.Name)
		}
		log.Fatal("preflight failed: database contains failed migrations")
	}

	var mismatches []string
	appliedSet := map[string]struct{}{}
	for _, row := range applied {
		key := migrationKey(row.Version, row.Name)
		appliedSet[key] = struct{}{}
		want, ok := migrationsUp[key]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("missing migration file for %d %s", row.Version, row.Name))
			continue
		}
		if row.Checksum != want.Checksum {
			mismatches = append(mismatches, fmt.Sprintf("checksum mismatch for %d %s", row.Version, row.Name))
		}
	}
	if len(mismatches) > 0 {
		for _, msg := range mismatches {
			log.Printf("preflight: %s", msg)
		}
		log.Fatal("preflight failed: migration drift detected")
	}

	var pending []migrationFile
	for key, mig := range migrationsUp {
		if _, ok := appliedSet[key]; !ok {
			pending = append(pending, mig)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	log.Printf("preflight ok: applied=%d pending=%d", len(applied), len(pending))
	if len(pending) > 0 {
		for _, mig := range pending {
			log.Printf("pending: %d %s", mig.Version, mig.Name)
		}
	}
}

func loadApplied(ctx context.Context, db *sql.DB, table string) ([]appliedRow, error) {
	if strings.TrimSpace(table) == "" {
		table = "schema_migrations"
	}
	q := fmt.Sprintf("SELECT version, name, checksum, success FROM %s ORDER BY applied_at ASC, version ASC;", pq(table))
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []appliedRow
	for rows.Next() {
		var row appliedRow
		if err := rows.Scan(&row.Version, &row.Name, &row.Checksum, &row.Success); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func loadMigrations(dir string) (map[string]migrationFile, error) {
	var roots []fs.FS
	dir = strings.TrimSpace(dir)
	if dir != "" && dir != "-" {
		roots = []fs.FS{os.DirFS(dir)}
	} else {
		roots = []fs.FS{migrations.Migrations}
	}

	out := map[string]migrationFile{}
	for _, root := range roots {
		entries, err := fs.ReadDir(root, ".")
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			mig, ok := parseFileName(name)
			if !ok || mig.Dir != "up" {
				continue
			}
			b, err := fs.ReadFile(root, name)
			if err != nil {
				return nil, err
			}
			mig.Checksum = checksum(strings.TrimSpace(string(b)))
			key := migrationKey(mig.Version, mig.Name)
			if _, exists := out[key]; exists {
				return nil, fmt.Errorf("duplicate migration %s", key)
			}
			out[key] = mig
		}
	}
	return out, nil
}

var fileRe = regexp.MustCompile(`^(\\d{8,14})_([a-zA-Z0-9_\\-]+)\\.(up|down)\\.sql$`)

func parseFileName(name string) (migrationFile, bool) {
	m := fileRe.FindStringSubmatch(filepath.Base(name))
	if m == nil {
		return migrationFile{}, false
	}
	ver, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return migrationFile{}, false
	}
	return migrationFile{
		Version: ver,
		Name:    m[2],
		Dir:     m[3],
	}, true
}

func migrationKey(version int64, name string) string {
	return fmt.Sprintf("%d:%s", version, name)
}

func checksum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:])
}

func pq(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}
