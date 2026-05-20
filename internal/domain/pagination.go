package domain

type PaginatedRequests struct {
	Items      []Request `json:"items"`
	NextCursor string    `json:"nextCursor,omitempty"`
	HasMore    bool      `json:"hasMore"`
}

type ListRequestsParams struct {
	Limit  int32
	Cursor string
}
