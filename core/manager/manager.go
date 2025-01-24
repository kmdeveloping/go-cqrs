package manager

import (
	"errors"
	"reflect"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

var defaultHandlerName = "Handler"

type ICmdQry interface {
	command.ICommand
	query.IQuery
}

type ICqrsManager interface {
	Execute(T ICmdQry) error
	Publish(T event.IEvent) error
	Validate(T validator.IValidator) (any, error)
}

type CqrsManager struct {
}

var _ ICqrsManager = (*CqrsManager)(nil)

func NewCqrsManager() ICqrsManager {
	return &CqrsManager{}
}

func (m *CqrsManager) Execute(T ICmdQry) error {
	task := []reflect.Value{reflect.ValueOf(T)}
	h := reflect.ValueOf(T).MethodByName(defaultHandlerName)
	res := h.Call(task)
	if err := res[0].Interface(); err != nil {
		return errors.New(err.(string))
	}

	return nil
}

func (m *CqrsManager) Publish(T event.IEvent) error {
	v := reflect.ValueOf(T)
	event := []reflect.Value{v}

	for i := 0; i < v.Type().NumMethod(); i++ {
		res := v.Method(i).Call(event)
		if err := res[0].Interface(); err != nil {
			return errors.New(err.(string))
		}
	}

	return nil
}

func (m *CqrsManager) Validate(T validator.IValidator) (any, error) {
	return nil, nil
}
