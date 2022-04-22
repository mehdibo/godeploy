package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func NewValidator() *Validator {
	v := validator.New()
	_ = v.RegisterValidation("fingerprint", fingerprint)
	return &Validator{validator: v}
}

func fingerprint(fl validator.FieldLevel) bool {
	return strings.HasPrefix(fl.Field().String(), "SHA256:")
}
