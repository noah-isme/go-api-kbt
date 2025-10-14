package dto

import (
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

// Problem represents an RFC 7807 compliant error response.
type Problem struct {
	Type     string            `json:"type,omitempty"`
	Title    string            `json:"title"`
	Status   int               `json:"status"`
	Detail   string            `json:"detail,omitempty"`
	Instance string            `json:"instance,omitempty"`
	TraceID  string            `json:"trace_id,omitempty"`
	Code     string            `json:"code,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"`
}

// NewProblem creates a new Problem instance.
func NewProblem(ctx context.Context, status int, title, detail string) Problem {
	spanCtx := trace.SpanContextFromContext(ctx)
	traceID := ""
	if spanCtx.IsValid() {
		traceID = spanCtx.TraceID().String()
	}

	return Problem{
		Title:   title,
		Status:  status,
		Detail:  detail,
		TraceID: traceID,
	}
}

// WithCode sets the error code.
func (p Problem) WithCode(code string) Problem {
	p.Code = code
	return p
}

// WithInstance sets the instance URI.
func (p Problem) WithInstance(instance string) Problem {
	p.Instance = instance
	return p
}

// WithField adds a field error.
func (p Problem) WithField(field, message string) Problem {
	if p.Fields == nil {
		p.Fields = make(map[string]string)
	}
	p.Fields[field] = message
	return p
}

func (p Problem) Error() string {
	return fmt.Sprintf("status %d: %s - %s", p.Status, p.Title, p.Detail)
}
