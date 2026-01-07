package handlers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func validationError(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		e := errs[0]
		field := strings.ToLower(e.Field())

		switch e.Tag() {
		case "required":
			return fmt.Sprintf("%s is required", field)
		case "email":
			return "invalid email format"
		case "min":
			return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
		case "max":
			return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
		default:
			return fmt.Sprintf("invalid %s", field)
		}
	}
	return "invalid request payload"
}
