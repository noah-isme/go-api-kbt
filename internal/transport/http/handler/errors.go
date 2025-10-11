package handler

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// tryRespondValidation inspects err for validator.ValidationErrors and, if
// present, writes a 400 with a structured errors map. Returns true if it
// handled the response.
func tryRespondValidation(w http.ResponseWriter, err error) bool {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := map[string]string{}
		for _, fe := range ve {
			out[fe.Field()] = fmt.Sprintf("failed on '%s'", fe.Tag())
		}
		respondJSON(w, http.StatusBadRequest, map[string]any{"errors": out})
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

	out := map[string]string{}
	for _, fe := range ve {
		fld := fe.Field()
		jsonName, ok := tagMap[fld]
		if !ok {
			jsonName = strings.ToLower(fld)
		}
		out[jsonName] = fmt.Sprintf("failed on '%s'", fe.Tag())
	}

	respondJSON(w, http.StatusBadRequest, map[string]any{"errors": out})
	return true
}
