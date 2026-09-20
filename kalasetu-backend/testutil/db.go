// Package testutil is the test seam for the backend: it builds the whole app
// in-process on a real, throwaway Postgres so tests can drive the GraphQL API
// over HTTP.
package testutil

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"kalasetu/migrations"

	_ "github.com/lib/pq"
)

// AdminURLEnv names the env var holding a Postgres connection URL for a role
// allowed to CREATE DATABASE, e.g. postgres://postgres:pw@localhost:5432/postgres?sslmode=disable
const AdminURLEnv = "KALASETU_TEST_DATABASE_URL"

// NewDB creates an empty database with a random name, applies all migrations,
// and drops the database when the test finishes. The test is skipped when
// AdminURLEnv is unset, so `go test ./...` still passes without a database.
func NewDB(t testing.TB) *sql.DB {
	t.Helper()

	adminURL := os.Getenv(AdminURLEnv)
	if adminURL == "" {
		t.Skipf("%s not set; run tests with `make test`", AdminURLEnv)
	}

	admin, err := sql.Open("postgres", adminURL)
	if err != nil {
		t.Fatalf("open admin connection: %v", err)
	}
	t.Cleanup(func() { admin.Close() })

	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	name := "kalasetu_test_" + hex.EncodeToString(suffix[:])

	if _, err := admin.Exec(`CREATE DATABASE ` + name); err != nil {
		t.Fatalf("create test database: %v", err)
	}

	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("parse %s: %v", AdminURLEnv, err)
	}
	u.Path = "/" + name

	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		if _, err := admin.Exec(`DROP DATABASE IF EXISTS ` + name + ` WITH (FORCE)`); err != nil {
			t.Errorf("drop test database %s: %v", name, err)
		}
	})

	if err := migrations.RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return db
}
