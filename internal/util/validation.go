package util

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func Validate[T any](data T) map[string]string {
	err := validator.New().Struct(data)
	res := map[string]string{}

	if err != nil {
		for _, v := range err.(validator.ValidationErrors) {
			res[v.StructField()] = TranslateTag(v)
		}
	}
	return res
}

func TranslateTag(fd validator.FieldError) string {
	switch fd.ActualTag() {
	case "required":
		return fmt.Sprintf("field %s must be filled", fd.StructField())
	case "min":
		return fmt.Sprintf("field %s size minimum is %s", fd.StructField(), fd.Param())
	case "unique":
		return fmt.Sprintf("field %s should be unique", fd.StructField())
	case "gt":
		return fmt.Sprintf("field %s should be greater than %s", fd.StructField(), fd.Param())
	case "max":
		return fmt.Sprintf("field %s size maximum is %s", fd.StructField(), fd.Param())
	case "uuid":
		return fmt.Sprintf("field %s should be uuid", fd.StructField())
	case "email":
		return fmt.Sprintf("field %s should be email", fd.StructField())
	}

	return "validasi gagal"
}
