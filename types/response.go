package types

// Response is the standard API response structure
type Response struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message"`
	Data       interface{}        `json:"data,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
	Error      *ErrorInfo         `json:"error,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}, pagination *PaginationResponse) Response {
	return Response{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message, detail string) Response {
	return Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Detail:  detail,
		},
	}
}
