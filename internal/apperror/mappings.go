package apperror

import (
	"context"
	"errors"
)

func Public(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	if errors.Is(err, context.Canceled) {
		return Wrap("cancelled", "operation cancelled", err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return Wrap("cancelled", "operation timed out", err)
	}
	return Wrap("internal_error", "operation failed", err)
}
