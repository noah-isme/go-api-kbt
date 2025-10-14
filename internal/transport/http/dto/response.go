package dto

// ListResponse is a generic response for list endpoints with pagination metadata.
type ListResponse struct {
	Data any `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// ActivityListResponse represents the response for listing activities.
type ActivityListResponse struct {
	Data []ActivityDTO `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

// MedalerListResponse represents the response for listing medalers.
type MedalerListResponse struct {
	Data []MedalerDTO `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

// UserListResponse represents the response for listing users.
type UserListResponse struct {
	Data []UserDTO `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}

// EventListResponse represents the response for listing events.
type EventListResponse struct {
	Data []EventDTO `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

// LocationListResponse represents the response for listing locations.
type LocationListResponse struct {
	Data []LocationDTO `json:"data"`
	Meta PaginationMeta      `json:"meta"`
}
