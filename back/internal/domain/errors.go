package domain

import (
	"errors"
)

// FailedPreconditionError ошибка не выполнения предварительного условия.
type FailedPreconditionError struct {
	Message string
}

// Error реализует ошибку.
func (e *FailedPreconditionError) Error() string {
	return e.Message
}

// NewFailedPreconditionError создает новую ошибку.
func NewFailedPreconditionError(message string) error {
	return &FailedPreconditionError{Message: message}
}

// IsFailedPreconditionError проверяет, имеет ли ошибка тип FailedPreconditionError.
func IsFailedPreconditionError(err error) bool {
	var e *FailedPreconditionError
	return errors.As(err, &e)
}
