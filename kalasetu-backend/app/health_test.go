package app_test

import (
	"testing"

	"kalasetu/testutil"
)

func TestHealthThroughHTTPHandler(t *testing.T) {
	h := testutil.NewHarness(t)

	var data struct {
		Health string `json:"health"`
	}
	h.GraphQL(t, "", `{ health }`, nil, &data)

	if data.Health != "OK" {
		t.Fatalf("health = %q, want OK", data.Health)
	}
}

func TestCreateUserWithRoleReturnsUsableToken(t *testing.T) {
	h := testutil.NewHarness(t)
	u := h.CreateUser(t, "Artist")

	// myApplications requires authentication.
	var data struct {
		MyApplications []struct{ ID string } `json:"myApplications"`
	}
	h.GraphQL(t, u.Token, `{ myApplications { id } }`, nil, &data)

	if len(data.MyApplications) != 0 {
		t.Fatalf("expected no applications for a fresh user, got %d", len(data.MyApplications))
	}

	var roles []string
	rows, err := h.DB.Query(`SELECT r.role FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = $1`, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			t.Fatal(err)
		}
		roles = append(roles, r)
	}
	if len(roles) != 1 || roles[0] != "Artist" {
		t.Fatalf("roles = %v, want [Artist]", roles)
	}
}
