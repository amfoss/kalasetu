package migrations

import "testing"

// TestSourceLoads guards the embedded migration set. golang-migrate rejects a
// source holding two migrations at the same version, and app startup only logs
// a migration failure as a warning, so a duplicate version disables every
// migration while the server still boots. This test deliberately needs no
// database: it runs in exactly the environment where the database-backed tests
// skip, which is where that slip went unnoticed.
func TestSourceLoads(t *testing.T) {
	d, err := newSource()
	if err != nil {
		t.Fatalf("load embedded migration set: %v", err)
	}
	defer d.Close()

	if _, err := d.First(); err != nil {
		t.Fatalf("migration set is empty: %v", err)
	}
}
