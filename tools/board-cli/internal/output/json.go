package output

import (
	"encoding/json"
	"io"
)

// ErrorCode represents standardized error codes for JSON output
type ErrorCode string

const (
	NotFound        ErrorCode = "NOT_FOUND"
	InvalidID       ErrorCode = "INVALID_ID"
	InvalidStatus   ErrorCode = "INVALID_STATUS"
	CycleDetected   ErrorCode = "CYCLE_DETECTED"
	ValidationError ErrorCode = "VALIDATION_ERROR"
	InternalError   ErrorCode = "INTERNAL_ERROR"
)

// JSONError represents the error response structure
type JSONError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

// PrintJSON writes any data as JSON to the given writer
func PrintJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// PrintError writes a JSON error response to the given writer
func PrintError(w io.Writer, code ErrorCode, message string, details map[string]interface{}) error {
	if details == nil {
		details = make(map[string]interface{})
	}
	errResp := JSONError{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	return PrintJSON(w, errResp)
}

// NewErrorResponse creates a JSONError struct without printing
func NewErrorResponse(code ErrorCode, message string, details map[string]interface{}) JSONError {
	if details == nil {
		details = make(map[string]interface{})
	}
	return JSONError{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// MarshalJSON converts data to JSON bytes with indentation
func MarshalJSON(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
