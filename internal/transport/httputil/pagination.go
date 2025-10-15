package httputil

import (
	"net/http"
	"strconv"

	"go-api-kbt/internal/transport/http/dto"
)

// ParsePagination extracts limit and offset query parameters with sane defaults.
func ParsePagination(r *http.Request) (limit, offset int) {
	query := r.URL.Query()

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err = strconv.Atoi(query.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	return
}

// ParsePaginationParams parses advanced pagination parameters including page and filters.
func ParsePaginationParams(r *http.Request) dto.PaginationParams {
	query := r.URL.Query()

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	return dto.PaginationParams{
		Page:  page,
		Limit: limit,
		Query: query.Get("q"),
		Sort:  query.Get("sort"),
	}
}

// ParseUintParam converts a string path parameter to uint.
func ParseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
