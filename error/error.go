package error

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error is a generic error container for gis.
type Error struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
	Cause     string `json:"Cause,omitempty"`
}

// Internal errors. Not the user's fault.
const (
	ErrContractionHierarchies = 101
	ErrMatrixComputation      = 102
)

// External errors. User error.
const (
	ReqErrMatrixNotFound = 200
)

var internalErrors = map[int]string{
	ErrContractionHierarchies: "Error while using contraction hierarchies.",
	ErrMatrixComputation:      "Error in matrix computation",
}

var requestErrors = map[int]string{}

var statusCodes = map[int]int{
	ErrContractionHierarchies: http.StatusInternalServerError,
	ErrMatrixComputation:      http.StatusInternalServerError,
}

// NewError creates a new error.
func NewError(errorCode int, cause string) *Error {
	return &Error{
		ErrorCode: errorCode,
		Message:   internalErrors[errorCode],
		Cause:     cause,
	}
}

// NewRequestError creates a new error with error code errorCode and reason specified by cause.
func NewRequestError(errorCode int, cause string) *Error {
	return &Error{
		ErrorCode: errorCode,
		Message:   requestErrors[errorCode],
		Cause:     cause,
	}
}

// Error formats the error.
func (e Error) Error() string {
	return e.Message + " (" + e.Cause + ")"
}

func (e Error) toJSONString() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e Error) statusCode() int {
	status, ok := statusCodes[e.ErrorCode]
	if !ok {
		status = http.StatusBadRequest
	}
	return status
}

// WriteTo writes the error to a responsewriter.
func (e Error) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.statusCode())
	fmt.Fprintln(w, e.toJSONString())
}
