package registry

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

func isRegistryName(fl validator.FieldLevel) bool {
	reg := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?([a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return reg.MatchString(value)
}

func registryAdmins(fl validator.FieldLevel) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9@._,-]+$`)
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return reg.MatchString(value)
}

func (a *App) registerCustomValidators() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("registry-name", isRegistryName); err != nil {
			return errors.Wrap(err, "unable to register custom validator")
		}

		if err := v.RegisterValidation("registry-admins", registryAdmins); err != nil {
			return errors.Wrap(err, "unable to register custom validator")
		}
	}

	return nil
}
