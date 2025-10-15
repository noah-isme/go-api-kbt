package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	service "go-api-kbt/internal/service/event"
	transport "go-api-kbt/internal/transport/http"
	handlerpkg "go-api-kbt/internal/transport/http/handler"

	"log/slog"
)

func TestEventCreateValidationReturnsJSONTags(t *testing.T) {
	svc := service.NewService(nil)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	eh := handlerpkg.NewEventHandler(svc, logger)
	ts := httptest.NewServer(transport.NewRouterBuilder(nil).WithEventHandler(eh).Build())
	defer ts.Close()

	// invalid payload: empty name
	payload := []byte(`{"name":"","description":"x"}`)
	resp, err := http.Post(ts.URL+"/api/v1/events", "application/json", bytes.NewReader(payload))
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
	} else if v != "failed on 'min'" && v != "failed on 'required'" {
		t.Fatalf("unexpected name error msg: %s", v)
	}
}
