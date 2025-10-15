package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"
	"time"

	domain "go-api-kbt/internal/domain/user"
	repo "go-api-kbt/internal/repository/user"
	service "go-api-kbt/internal/service/user"
	handler "go-api-kbt/internal/transport/http/handler"

	"log/slog"
)

// inMemoryRepo is a simple thread-safe repository implementation for tests.
type inMemoryRepo struct {
	mu     sync.Mutex
	nextID uint
	data   map[uint]*domain.Entity
}

func newInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{
		nextID: 1,
		data:   make(map[uint]*domain.Entity),
	}
}

func (m *inMemoryRepo) Create(ctx context.Context, entity *domain.Entity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// check email uniqueness
	for _, e := range m.data {
		if e.Email == entity.Email {
			return repo.ErrNotFound // mimic not-found or conflict; service checks GetByEmail first
		}
	}

	entity.ID = m.nextID
	m.nextID++
	now := time.Now().UTC()
	entity.CreatedAt = now
	entity.UpdatedAt = now
	m.data[entity.ID] = entity
	return nil
}

func (m *inMemoryRepo) GetByID(ctx context.Context, id uint) (*domain.Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.data[id]; ok {
		// return a copy
		copy := *e
		return &copy, nil
	}
	return nil, repo.ErrNotFound
}

func (m *inMemoryRepo) GetByEmail(ctx context.Context, email string) (*domain.Entity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.data {
		if e.Email == email {
			copy := *e
			return &copy, nil
		}
	}
	return nil, repo.ErrNotFound
}

func (m *inMemoryRepo) List(ctx context.Context, limit, offset int) ([]domain.Entity, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	entries := make([]domain.Entity, 0, len(m.data))
	for _, e := range m.data {
		entries = append(entries, *e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })

	total := len(entries)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	page := make([]domain.Entity, end-offset)
	copy(page, entries[offset:end])

	return page, total, nil
}

func (m *inMemoryRepo) Update(ctx context.Context, entity *domain.Entity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[entity.ID]; !ok {
		return repo.ErrNotFound
	}
	entity.UpdatedAt = time.Now().UTC()
	m.data[entity.ID] = entity
	return nil
}

func (m *inMemoryRepo) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return repo.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	repo := newInMemoryRepo()
	svc := service.NewService(repo, 4)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	uh := handler.NewUserHandler(svc, logger)
	router := NewRouterBuilder(uh).Build()
	return httptest.NewServer(router)
}

func TestHealthz(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf("expected body 'ok', got %q", string(body))
	}
}

func TestUserCRUD(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create
	input := map[string]any{
		"username": "integuser",
		"email":    "integ@example.com",
		"password": "password123",
		"role":     string(domain.RoleAdmin),
	}
	b, _ := json.Marshal(input)
	resp, err := http.Post(ts.URL+"/api/v1/users", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(body))
	}
	var envelope map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	data, ok := envelope["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data envelope")
	}
	idf, ok := data["id"].(float64)
	if !ok {
		t.Fatalf("invalid id in response")
	}
	id := uint(idf)

	// List
	resp, err = http.Get(ts.URL + "/api/v1/users")
	if err != nil {
		t.Fatalf("list request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var listEnvelope struct {
		Data struct {
			Data []map[string]any `json:"data"`
			Meta struct {
				Page         float64 `json:"page"`
				Limit        float64 `json:"limit"`
				TotalRecords float64 `json:"total_records"`
			} `json:"meta"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listEnvelope); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listEnvelope.Data.Data) == 0 {
		t.Fatalf("expected at least one user in list response")
	}
	if listEnvelope.Data.Meta.TotalRecords < 1 {
		t.Fatalf("expected total records meta to be populated")
	}

	// Get
	resp, err = http.Get(ts.URL + "/api/v1/users/" + stringInt(id))
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Update
	update := map[string]any{"username": "updated"}
	ub, _ := json.Marshal(update)
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/users/"+stringInt(id), bytes.NewReader(ub))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("update request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Delete
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/users/"+stringInt(id), nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// Get after delete should be 404
	resp, err = http.Get(ts.URL + "/api/v1/users/" + stringInt(id))
	if err != nil {
		t.Fatalf("get after delete failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", resp.StatusCode)
	}
}

func stringInt(v uint) string {
	return fmt.Sprintf("%d", v)
}

// end
