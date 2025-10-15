package dto

// PaginationParams defines parameters for pagination.
type PaginationParams struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Query string `json:"query"`
	Sort  string `json:"sort"`
}

// PaginationMeta defines metadata for pagination responses.
type PaginationMeta struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"total_records"`
}
