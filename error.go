package gobinance

import "fmt"

// errorDTO is the error structure returned by binance in the event of a non-200 response
type errorDTO struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// HttpError is an error type returned when a non-200 response is received from the binance API
type HttpError struct {
	HttpStatus int
	errorDTO
}

// StatusCode returns the HTTP status code received from binance that triggered the error
func (h *HttpError) StatusCode() int {
	return h.HttpStatus
}

// ErrorCode returns the value of the `code` field in binance's error response.
func (h *HttpError) ErrorCode() int {
	return h.Code
}

// Error implements the error interface and returns a human-readable description of the error
func (h *HttpError) Error() string {
	return fmt.Sprintf("got status %v from binance. error code was %v: %v", h.HttpStatus, h.Code, h.Msg)
}