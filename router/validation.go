package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func (r *Router) parseValidationErrors(params gin.H) gin.H {
	for k, v := range params {
		validationErrors, ok := v.(validator.ValidationErrors)
		if !ok {
			continue
		}

		errorMessages := make(map[string][]validator.FieldError)
		for _, vErr := range validationErrors {
			fieldParts := strings.Split(vErr.StructField(), ".")
			field := fieldParts[len(fieldParts)-1]

			fieldErrorMessages, ok := errorMessages[field]
			if !ok {
				fieldErrorMessages = make([]validator.FieldError, 0, 1)
			}
			fieldErrorMessages = append(fieldErrorMessages, vErr)

			errorMessages[field] = fieldErrorMessages
		}

		params[k] = errorMessages
	}

	return params
}
