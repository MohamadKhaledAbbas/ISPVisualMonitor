package utils

import (
	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator instance
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// Validator returns the underlying validator instance
func (v *Validator) Validator() *validator.Validate {
	return v.validator
}
