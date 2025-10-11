package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "go-api-kbt/internal/domain/user"
	repo "go-api-kbt/internal/repository/user"
	service "go-api-kbt/internal/service/user"
	transport "go-api-kbt/internal/transport/http"
	handlerpkg "go-api-kbt/internal/transport/http/handler"

	"log/slog"
)

// mock repository minimal to satisfy interface
type mockUserRepo struct{}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.Entity) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*domain.Entity, error) {
	return nil, repo.ErrNotFound
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.Entity, error) {
	return nil, repo.ErrNotFound
}
func (m *mockUserRepo) List(ctx context.Context, limit, offset int) ([]domain.Entity, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.Entity) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id uint) error             { return nil }

func TestUserCreateValidationReturnsJSONTags(t *testing.T) {
	// create service with a mock repository
	repo := &mockUserRepo{}
	svc := service.NewService(repo, 4)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	uh := handlerpkg.NewUserHandler(svc, logger)
	ts := httptest.NewServer(transport.NewRouterBuilder(uh).Build())
	defer ts.Close()

	// send invalid payload (username too short -> min, email invalid -> email, password too short -> min)
	payload := []byte(`{"username":"a","email":"not-an-email","password":"short","role":"admin"}`)
	resp, err := http.Post(ts.URL+"/api/v1/users", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body map[string]map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errs, ok := body["errors"]
	if !ok {
		t.Fatalf("expected errors key")
	}
	// expect keys and specific messages
	if v, ok := errs["email"]; !ok {
		t.Fatalf("expected email in errors")
	} else if v != "failed on 'email'" {
		t.Fatalf("unexpected email error msg: %s", v)
	}
	if v, ok := errs["password"]; !ok {
		t.Fatalf("expected password in errors")
	} else if v != "failed on 'min'" {
		t.Fatalf("unexpected password error msg: %s", v)
	}
	if v, ok := errs["username"]; !ok {
		t.Fatalf("expected username in errors")
	} else if v != "failed on 'min'" {
		t.Fatalf("unexpected username error msg: %s", v)
	}
}
