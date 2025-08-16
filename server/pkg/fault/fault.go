// Package fault implements a custom error type that enhances the standard Go error
// by allowing the inclusion of additional context, such as HTTP status codes
// and categorical kinds, to facilitate more robust error handling across different
// application layers.
package fault

import "fmt"

type Error struct {
	Message string
	Code    int
	Kind    string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("message: %s, kind: %s, original_error: %v", e.Message, e.Kind, e.Err)
	}
	return fmt.Sprintf("message: %s, kind: %s", e.Message, e.Kind)
}

func (e *Error) Unwrap() error {
	return e.Err
}

type Option func(*Error)

func New(message string, options ...Option) *Error {
	err := &Error{
		Message: message,
	}
	for _, opt := range options {
		opt(err)
	}
	return err
}

func WithHTTPCode(code int) Option {
	return func(e *Error) {
		e.Code = code
	}
}

func WithKind(kind string) Option {
	return func(e *Error) {
		e.Kind = kind
	}
}

func WithError(err error) Option {
	return func(e *Error) {
		e.Err = err
	}
}

const (
	KindNotFound        = "NotFound"
	KindValidation      = "Validation"
	KindUnexpected      = "Unexpected"
	KindConflict        = "Conflict"
	KindUnauthenticated = "Unauthenticated"
	KindForbidden       = "Forbidden"
)
