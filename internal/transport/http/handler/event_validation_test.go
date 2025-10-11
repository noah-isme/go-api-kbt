package handler_test

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"

    handlerpkg "go-api-kbt/internal/transport/http/handler"
    transport "go-api-kbt/internal/transport/http"
    service "go-api-kbt/internal/service/event"

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
    if err != nil { t.Fatalf("request failed: %v", err) }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusBadRequest { t.Fatalf("expected 400, got %d", resp.StatusCode) }
    var body map[string]map[string]string
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { t.Fatalf("decode: %v", err) }
    errs, ok := body["errors"]
    if !ok { t.Fatalf("expected errors key") }
    if v, ok := errs["name"]; !ok {
        t.Fatalf("expected name in errors")
    } else if v != "failed on 'min'" && v != "failed on 'required'" {
        t.Fatalf("unexpected name error msg: %s", v)
    }
}
