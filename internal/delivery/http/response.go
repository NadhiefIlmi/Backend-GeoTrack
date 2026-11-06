package http

import (
	"encoding/json"
	"net/http"
)

// MetaPagination represents pagination metadata
type MetaPagination struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
}

// SuccessResponse defines standard successful response envelope
type SuccessResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    interface{}     `json:"data,omitempty"`
	Meta    *MetaPagination `json:"meta,omitempty"`
}

// ErrorDetail defines itemized error field
type ErrorDetail struct {
	Field string `json:"field,omitempty"`
	Issue string `json:"issue"`
}

// ErrorBody defines error body payload
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorResponse defines standard error response envelope
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

// JSON writes a JSON response with status code
func JSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// SendSuccess sends a structured success response
func SendSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}, meta *MetaPagination) {
	JSON(w, statusCode, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// SendError sends a structured error response
func SendError(w http.ResponseWriter, statusCode int, code string, message string, details interface{}) {
	JSON(w, statusCode, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

