package router

import (
	"reflect"

	ut "github.com/go-playground/universal-translator"
)

type FieldError struct {
	tag   string
	field string
}

func (f FieldError) Tag() string {
	return f.tag
}

func (f FieldError) ActualTag() string {
	return f.tag
}

func (f FieldError) Namespace() string {
	return ""
}

func (f FieldError) StructNamespace() string {
	return ""
}

func (f FieldError) Field() string {
	return f.field
}

func (f FieldError) StructField() string {
	return f.field
}

func (f FieldError) Value() interface{} {
	return nil
}

func (f FieldError) Param() string {
	return ""
}

func (f FieldError) Kind() reflect.Kind {
	return 0
}

func (f FieldError) Type() reflect.Type {
	return reflect.TypeOf(f)
}

func (f FieldError) Translate(ut ut.Translator) string {
	return ""
}

func (f FieldError) Error() string {
	return f.tag
}

func MakeFieldError(field, tag string) *FieldError {
	return &FieldError{tag: tag, field: field}
}
