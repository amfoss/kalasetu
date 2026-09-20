package testutil

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"kalasetu/app"
	"kalasetu/payments/paymentstest"

	"github.com/gin-gonic/gin"
)

// Harness is an in-process App on a fresh database with a fake PaymentProvider.
type Harness struct {
	App      *app.App
	DB       *sql.DB
	Payments *paymentstest.Fake
}

// User is a registered user and a valid access token for them.
type User struct {
	ID    int
	Email string
	Token string
}

var userSeq atomic.Int64

// NewHarness builds a Harness. It sets the JWT secrets via t.Setenv, so tests
// using it cannot call t.Parallel.
func NewHarness(t testing.TB) *Harness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")

	db := NewDB(t)
	fake := paymentstest.NewFake()
	return &Harness{App: app.New(db, fake), DB: db, Payments: fake}
}

// CreateUser registers a new user through the real auth endpoint and grants
// them role (creating the role row if needed). Roles are not seeded by
// migrations, so the helper does it.
func (h *Harness) CreateUser(t testing.TB, role string) User {
	t.Helper()

	n := userSeq.Add(1)
	email := fmt.Sprintf("user%d@example.test", n)
	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "name": fmt.Sprintf("Test User %d", n),
	})

	rec := httptest.NewRecorder()
	h.App.Router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: status %d: %s", rec.Code, rec.Body)
	}

	var res struct {
		AccessToken string `json:"access_token"`
		User        struct {
			ID int `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode register response: %v", err)
	}

	if _, err := h.DB.Exec(`INSERT INTO roles (role) VALUES ($1) ON CONFLICT (role) DO NOTHING`, role); err != nil {
		t.Fatalf("ensure role: %v", err)
	}
	if _, err := h.DB.Exec(
		`INSERT INTO user_roles (user_id, role_id) SELECT $1, id FROM roles WHERE role = $2`,
		res.User.ID, role,
	); err != nil {
		t.Fatalf("assign role: %v", err)
	}

	return User{ID: res.User.ID, Email: email, Token: res.AccessToken}
}

// GraphQLResponse is the raw GraphQL envelope.
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// GraphQLRaw posts a query over HTTP and returns the envelope, for tests that
// expect errors. An empty token sends no Authorization header.
func (h *Harness) GraphQLRaw(t testing.TB, token, query string, vars map[string]any) GraphQLResponse {
	t.Helper()

	body, _ := json.Marshal(map[string]any{"query": query, "variables": vars})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	h.App.Router.ServeHTTP(rec, req)

	var res GraphQLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode graphql response (status %d): %v\n%s", rec.Code, err, rec.Body)
	}
	return res
}

// GraphQL posts a query and decodes its data into out, failing the test on any
// GraphQL error.
func (h *Harness) GraphQL(t testing.TB, token, query string, vars map[string]any, out any) {
	t.Helper()

	res := h.GraphQLRaw(t, token, query, vars)
	if len(res.Errors) > 0 {
		t.Fatalf("graphql errors: %+v", res.Errors)
	}
	if out != nil {
		if err := json.Unmarshal(res.Data, out); err != nil {
			t.Fatalf("decode data: %v\n%s", err, res.Data)
		}
	}
}
