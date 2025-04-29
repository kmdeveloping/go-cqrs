package cqrs

import (
	"fmt"
	"reflect"
)

type HandlerBase struct {
	Manager *Manager
}

var base = &HandlerBase{}

func RegisterCqrsManager(mgr *Manager) {
	base.Manager = mgr
}

func NewHandler[T any]() *T {
	handler := new(T)

	v := reflect.ValueOf(handler).Elem()
	baseField := v.FieldByName("HandlerBase")

	if baseField.IsValid() && baseField.CanSet() && baseField.Type().AssignableTo(reflect.TypeOf(&HandlerBase{})) {
		baseField.Set(reflect.ValueOf(base))
	} else {
		panic(fmt.Sprintf("NewHandler: %T does not embed *HandlerBase or field is not settable", handler))
	}

	return handler
}
