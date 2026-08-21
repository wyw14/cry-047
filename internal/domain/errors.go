package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidState   = errors.New("invalid state")
	ErrDuplicate      = errors.New("duplicate command")
	ErrEvidence       = errors.New("invalid evidence")
	ErrVersionChanged = errors.New("version changed")
)

type FieldViolation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Code       string           `json:"code"`
	Message    string           `json:"message"`
	Violations []FieldViolation `json:"violations"`
}

func (e *ValidationError) Error() string { return e.Message }

func Invalid(field, message string) error {
	return &ValidationError{
		Code: "validation_failed", Message: "提交内容不符合业务规则",
		Violations: []FieldViolation{{Field: field, Message: message}},
	}
}

func Wrap(kind error, resource, id string) error {
	return fmt.Errorf("%s %s: %w", resource, id, kind)
}
