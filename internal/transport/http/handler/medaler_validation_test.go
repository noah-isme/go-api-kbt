package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "go-api-kbt/internal/domain/medaler"
	service "go-api-kbt/internal/service/medaler"
	transport "go-api-kbt/internal/transport/http"
	handlerpkg "go-api-kbt/internal/transport/http/handler"

	"log/slog"

	"gorm.io/gorm"
)

// mock repository minimal to satisfy interface
type mockMedalerRepo struct{}

func (m *mockMedalerRepo) Create(ctx context.Context, medaler *domain.Medaler) error { return nil }
func (m *mockMedalerRepo) FindByID(ctx context.Context, id uint) (*domain.Medaler, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockMedalerRepo) FindAll(ctx context.Context, limit, offset int, query, sort string) ([]domain.Medaler, int, error) {
	return nil, 0, nil
}
func (m *mockMedalerRepo) Update(ctx context.Context, medaler *domain.Medaler) error { return nil }
func (m *mockMedalerRepo) Delete(ctx context.Context, id uint) error                 { return nil }

func TestMedalerCreateValidation(t *testing.T) {
	// create service with a mock repository
	repo := &mockMedalerRepo{}
	svc := service.NewService(repo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mh := handlerpkg.NewMedalerHandler(svc, logger)
	ts := httptest.NewServer(transport.NewRouterBuilder(nil).WithMedalerHandler(mh).Build())
	defer ts.Close()

	// send invalid payload (email invalid)
	payload := []byte(`{"name":"test","email":"not-an-email"}`)
	resp, err := http.Post(ts.URL+"/api/v1/medalers", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var problem struct {
		Fields map[string]string `json:"fields"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&problem); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if v, ok := problem.Fields["email"]; !ok {
		t.Fatalf("expected email validation error")
	} else if v != "failed on 'email'" {
		t.Fatalf("unexpected email validation message: %s", v)
	}
}
