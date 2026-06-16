package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var requestValidator = newRequestValidator()

func newRequestValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(jsonTagName)
	_ = v.RegisterValidation("notblank", validateNotBlank)
	return v
}

func decodeAndValidateRequest(r *http.Request, dst any) string {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return "invalid request body"
	}

	if err := requestValidator.Struct(dst); err != nil {
		return validationErrorMessage(err)
	}

	return ""
}

func jsonTagName(field reflect.StructField) string {
	name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func validateNotBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func validationErrorMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return "invalid request body"
	}

	fieldErr := validationErrors[0]
	return fmt.Sprintf("%s failed validation: %s", fieldErr.Field(), fieldErr.Tag())
}
