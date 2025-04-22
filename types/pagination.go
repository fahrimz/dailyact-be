package types

// PaginationQuery represents the query parameters for pagination
type PaginationQuery struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
}

// PaginationResponse represents the pagination metadata in responses
type PaginationResponse struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	HasMore     bool  `json:"has_more"`
}

// NewPaginationResponse creates a new pagination response
func NewPaginationResponse(currentPage, pageSize int, totalItems int64) PaginationResponse {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	
	return PaginationResponse{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasMore:     currentPage < totalPages,
	}
}
