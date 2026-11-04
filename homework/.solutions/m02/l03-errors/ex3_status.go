package l03errors

import (
	"context"
	"errors"
)

func HTTPStatus(err error) int {
	var fe *FieldError
	switch {
	case err == nil:
		return 200
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.As(err, &fe):
		return 400
	case errors.Is(err, context.DeadlineExceeded):
		return 504
	case errors.Is(err, context.Canceled):
		return 499
	default:
		return 500
	}
}
