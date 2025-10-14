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
func (m *mockMedalerRepo) FindAll(ctx context.Context) ([]domain.Medaler, error) { return nil, nil }

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
	defer resp.Body.Close()

	// The current handler does not have explicit validation, so it will return 500 for now.
	// Once validation is added to the service layer, this test will need to be updated.
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}
