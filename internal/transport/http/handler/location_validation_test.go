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
	payload := []byte(`{"latitude":0}`)
	resp, err := http.Post(ts.URL+"/api/v1/locations", "application/json", bytes.NewReader(payload))
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
	if v, ok := errs["user_id"]; !ok {
		t.Fatalf("expected user_id in errors")
	} else if v != "failed on 'required'" {
		t.Fatalf("unexpected user_id error msg: %s", v)
	}
}
