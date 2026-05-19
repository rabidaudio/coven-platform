// Package api manages the low-level operations for an HTTP api.
//
// It has helpers for parsing and generating JSON response bodies,
// and defines a JSON response format convention.
package api

import (
	"net/http"
)

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalItems int `json:"total_items"`
}

type apiResponse struct {
	Status     string         `json:"status"`
	Error      *errorResponse `json:"error,omitempty"`
	Data       any            `json:"data,omitempty"`
	Pagination *Pagination    `json:"pagination,omitempty"`
}

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Respond serializes a single data object into a response object
// and writes the JSON to w. If not provided, the status string will be "OK".
func Respond[T any](w http.ResponseWriter, data *T, status string) {
	if status == "" {
		status = "OK"
	}
	r := apiResponse{
		Status: status,
		Data:   data,
	}
	MarshalBody(w, &r)
}

// RespondPage serializes an array of data objects into a response object
// and writes the JSON to w. If the results represent only one page of a
// paginated endpoint, [Pagination] can be provided indicating the page metadata.
func RespondPage[T any](w http.ResponseWriter, data []T, p *Pagination, status string) {
	r := apiResponse{
		Status:     status,
		Data:       data,
		Pagination: p,
	}
	MarshalBody(w, &r)
}

// RespondError serializes an error into a response object and writes the JSON to w,
// and sets a non-success status code.
//
// If the error is a [HttpResponseError], the status code attached is used. You can
// implement this interface on your error, or use [RespondErrorStatus] or one of
// the shortcut methods for it to wrap your error with a particular status code.
// For any other error, 500 ("Server Error") is used.
//
// In non-production environments, the error message is included in the response. These
// are obfuscated in production.
func RespondError(w http.ResponseWriter, err error) {
	r := errorResponse{
		Message: err.Error(),
		Code:    http.StatusInternalServerError,
	}
	if herr, ok := err.(HTTPResponseError); ok {
		r.Message = herr.Error()
		r.Code = herr.Status()
	}
	if IsEnv(Production) {
		// hide the full error
		r.Message = "server error"
	}
	ar := apiResponse{
		Status: "ERROR",
		Error:  &r,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(r.Code) // send the error code
	MarshalBody(w, &ar)
}

// RespondErrorStatus serializes an error into a response object and writes the JSON to w,
// and sets a non-success status code status.
func RespondErrorStatus(w http.ResponseWriter, cause error, status int) {
	RespondError(w, httpError{cause: cause, status: http.StatusBadRequest})
}

// RespondBadFormat responds with a 400 ("Bad Format") error
func RespondBadFormat(w http.ResponseWriter, cause error) {
	RespondErrorStatus(w, cause, http.StatusBadRequest)
}

// RespondUnauthorizedError responds with a 401 ("Unauthorized") error
func RespondUnauthorizedError(w http.ResponseWriter, cause error) {
	RespondErrorStatus(w, cause, http.StatusUnauthorized)
}

// RespondForbiddenError responds with a 403 ("Forbidden") error
func RespondForbiddenError(w http.ResponseWriter, cause error) {
	RespondErrorStatus(w, cause, http.StatusForbidden)
}

// RespondNotFoundError responds with a 404 ("Not Found") error
func RespondNotFoundError(w http.ResponseWriter, cause error) {
	RespondErrorStatus(w, cause, http.StatusNotFound)
}
