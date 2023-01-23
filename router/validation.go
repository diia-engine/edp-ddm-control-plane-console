package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
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

func (r *Router) AddValidator(tag string, valid validator.Func) error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation(tag, valid); err != nil {
			return errors.Wrap(err, "unable to register validator")
		}
	}

	return nil
}
