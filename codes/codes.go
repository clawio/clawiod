package codes

import (
	"fmt"
	"net/http"

	"github.com/clawio/clawiod/helpers"
)

// A Code is an unsigned 32-bit error code.
type Code uint32

const (
	// To add new coded always add them in the end, to not break iota

	// Success indicates no error.
	Success Code = iota

	// InvalidToken is returned when the auth token is invalid or has expired
	InvalidToken

	// Unauthenticated is returned when authentication is needed for execution.
	Unauthenticated

	// BadAuthenticationData is returned when the authentication fails.
	BadAuthenticationData

	// BadInputData is returned when the input parameters are not valid.
	BadInputData

	// Internal is returned when there is an unexpected/undesired problem.
	Internal

	// NotFound is returned when something cannot be found.
	NotFound

	// BadChecksum is returned when two checksum differs.
	BadChecksum

	// TooBig is returned when something is too big to be processed.
	TooBig
)

// String returns a string representation of the Code
func (c Code) String() string {
	switch c {
	case InvalidToken:
		return "invalid or expired token"
	case Unauthenticated:
		return "unauthenticated request"
	case BadAuthenticationData:
		return "bad authentication data"
	case BadInputData:
		return "bad input data"
	case Internal:
		return "internal error"
	case NotFound:
		return "not found"
	case BadChecksum:
		return "checksums differ"
	case TooBig:
		return "too big"
	default:
		return "FIXME: this should be a helpful message"
	}
}

// Response is a ClawIO API response.  This wraps the standard http.Response
// returned from ClawIO and provides convenient access to future things like
// pagination links.
type Response struct {
	*http.Response
}

func (r *Response) String() string {
	return fmt.Sprintf("%v %v: %d",
		r.Response.Request.Method, helpers.SanitizeURL(r.Response.Request.URL),
		r.Response.StatusCode)

}

// NewResponse creates a new Response for the provided http.Response.
func NewResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// An ErrorResponse reports one or more errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:"-"` // HTTP response that caused this error
	*Err     `json:"error"` // more detail on individual errors
}

// NewErrorResponse wraps a Response with an error.
func NewErrorResponse(res *http.Response, e *Err) *ErrorResponse {
	response := &ErrorResponse{Response: res, Err: e}
	return response
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %s: %d (%s)",
		r.Response.Request.Method, helpers.SanitizeURL(r.Response.Request.URL),
		r.Response.StatusCode, r.Err.Error())
}

// An Err reports more details on an individual error in an ErrorResponse.
type Err struct {
	Message string `json:"message"`
	Code    Code   `json:"code"`
}

// Error() implements the Error interface.
func (e *Err) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// NewErr is a usefull function to create Errs with the corresponding Code message.
// If no message is passed, the default code message will be used.
func NewErr(c Code, msg string) *Err {
	if msg == "" {
		msg = c.String()
	}
	return &Err{msg, c}
}
