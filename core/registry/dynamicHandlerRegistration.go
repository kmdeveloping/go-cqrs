package registry

import (
	"errors"
	"reflect"

	"github.com/kmdeveloping/go-cqrs/core/handlers"
)

type Registry struct {
	commandHandlers map[string]handlers.IHandler
}

type Service struct {
	TContract any
	THandler  handlers.IHandler
}

func NewRegistry() *Registry {
	return &Registry{
		commandHandlers: make(map[string]handlers.IHandler),
	}
}

func (r *Registry) RegisterHandlers(services []Service) *Registry {
	for _, svc := range services {
		r.commandHandlers[reflect.TypeOf(svc.TContract).Name()] = svc.THandler
	}
	return r
}

func (r *Registry) Resolve(T reflect.Type) (handlers.IHandler, error) {
	handler, exists := r.commandHandlers[T.Name()]
	if !exists {
		return nil, errors.New("handler not registered for command: " + T.Name())
	}

	return handler, nil
}
