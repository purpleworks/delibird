package delibird

import "fmt"

// Error json model
type ApiError struct {
	Code    string
	Message string
}

// String() implement
func (e ApiError) String() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewApiError creates ApiError object
func NewApiError(code, message string) *ApiError {
	return &ApiError{code, message}
}

type Track struct {
	TrackingNumber string         `json:"tracking_number"`
	CompanyCode    string         `json:"company_code"`
	CompanyName    string         `json:"company_name"`
	Sender         string         `json:"sender"`
	Receiver       string         `json:"receiver"`
	Signer         string         `json:"signer"`
	StatusCode     TrackingStatus `json:"status_code"`
	StatusText     string         `json:"status_text"`
	History        []History      `json:"history"`
}

type History struct {
	Area       string         `json:"area"`
	Tel        string         `json:"tel,omitempty"`
	Date       int64          `json:"date"`
	DateText   string         `json:"date_text"`
	StatusCode TrackingStatus `json:"status_code"`
	StatusText string         `json:"status_text"`
}
