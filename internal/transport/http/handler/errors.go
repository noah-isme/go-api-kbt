package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-playground/validator/v10"
)

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]any{
		"success": status >= 200 && status < 300,
		"data":    payload,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// response writing errors are logged but not returned to the client
		slog.Default().Error("encode response", slog.String("error", err.Error()))
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	problem := dto.NewProblem(context.Background(), status, http.StatusText(status), message)
	problem.Type = "about:blank"

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(problem); err != nil {
		slog.Default().Error("encode problem response", slog.String("error", err.Error()))
	}
}

// tryRespondValidation inspects err for validator.ValidationErrors and, if
// present, writes a 400 with a structured errors map. Returns true if it
// handled the response.
func tryRespondValidation(w http.ResponseWriter, err error) bool {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		problem := dto.NewProblem(context.Background(), http.StatusBadRequest, "Validation Failed", "One or more fields failed validation.")
		problem.Type = "/errors/validation"
		for _, fe := range ve {
			problem = problem.WithField(fe.Field(), fmt.Sprintf("failed on '%s'", fe.Tag()))
		}

		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(problem); err != nil {
			slog.Default().Error("encode problem response", slog.String("error", err.Error()))
		}
		return true
	}
	return false
}

// respondValidationWithJSONTags validates using validator.ValidationErrors
// and maps struct field names to their json tag equivalents before
// responding with a structured 400 payload.
func respondValidationWithJSONTags(w http.ResponseWriter, in any, err error) bool {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return false
	}

	// Build map of struct field name -> json tag
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tagMap := map[string]string{}
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := f.Name
			tag := f.Tag.Get("json")
			if tag == "" {
				tagMap[name] = strings.ToLower(name)
				continue
			}
			// tag could be 'name,omitempty'
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" || tagParts[0] == "" {
				tagMap[name] = strings.ToLower(name)
			} else {
				tagMap[name] = tagParts[0]
			}
		}
	}

	problem := dto.NewProblem(context.Background(), http.StatusBadRequest, "Validation Failed", "One or more fields failed validation.")
	problem.Type = "/errors/validation"
	for _, fe := range ve {
		fld := fe.Field()
		jsonName, ok := tagMap[fld]
		if !ok {
			jsonName = strings.ToLower(fld)
		}
		problem = problem.WithField(jsonName, fmt.Sprintf("failed on '%s'", fe.Tag()))
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(problem); err != nil {
		slog.Default().Error("encode problem response", slog.String("error", err.Error()))
	}
	return true
}