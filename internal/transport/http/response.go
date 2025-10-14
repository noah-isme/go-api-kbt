package http

import (
	"net/http"
	"strconv"

	"go-api-kbt/internal/transport/http/dto"
)

func parsePagination(r *http.Request) (int, int) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func parseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

// ParsePaginationParams parses pagination and filtering parameters from the request.
func ParsePaginationParams(r *http.Request) dto.PaginationParams {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := query.Get("q")
	sort := query.Get("sort")

	return dto.PaginationParams{
		Page:  page,
		Limit: limit,
		Query: q,
		Sort:  sort,
	}
}