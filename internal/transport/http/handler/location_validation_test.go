package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	service "go-api-kbt/internal/service/location"
	transport "go-api-kbt/internal/transport/http"
	handlerpkg "go-api-kbt/internal/transport/http/handler"

	"log/slog"
)

func TestLocationCreateValidationReturnsJSONTags(t *testing.T) {
	svc := service.NewService(nil)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	lh := handlerpkg.NewLocationHandler(svc, logger)
	ts := httptest.NewServer(transport.NewRouterBuilder(nil).WithLocationHandler(lh).Build())
	defer ts.Close()

	// invalid payload: missing required fields
	payload := []byte(`{"name":"","address":""}`)
	resp, err := http.Post(ts.URL+"/api/v1/locations", "application/json", bytes.NewReader(payload))
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
	if v, ok := problem.Fields["name"]; !ok {
		t.Fatalf("expected name in errors")
	} else if v != "failed on 'required'" {
		t.Fatalf("unexpected name error msg: %s", v)
	}
	if v, ok := problem.Fields["address"]; !ok {
		t.Fatalf("expected address in errors")
	} else if v != "failed on 'required'" {
		t.Fatalf("unexpected address error msg: %s", v)
	}
}
