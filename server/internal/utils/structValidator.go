package utils

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func ValidateStruct(s interface{}) map[string]string {
	errors := make(map[string]string)

	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors[err.Field()] = msgForTag(err.Tag(), err.Param())
		}
	}
	return errors
}

func msgForTag(tag, param string) string {
	switch tag {
	case "required":
		return "this field is required"
	case "email":
		return "invalid email format"
	case "min":
		return "must be at least " + param + " characters"
	case "max":
		return "must be at most " + param + " characters"
	default:
		return "invalid value"
	}
}
