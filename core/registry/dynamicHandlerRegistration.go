package registry

import (
	"errors"
	"reflect"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
)

type Registry struct {
	commandHandlers map[reflect.Type]handlers.ICommandHandler
}

type CommandServices struct {
	Command command.ICommand
	Handler handlers.ICommandHandler
}

func NewRegistry() *Registry {
	return &Registry{
		commandHandlers: make(map[reflect.Type]handlers.ICommandHandler),
	}
}

func (r *Registry) RegisterCommandHandlers(handlerList []CommandServices) *Registry {
	for _, h := range handlerList {
		r.commandHandlers[reflect.TypeOf(h.Command)] = h.Handler
	}

	return r
}

func (r *Registry) Resolve(T any) (handlers.IHandler, error) {
	handler, exists := r.commandHandlers[reflect.TypeOf(T)]
	if !exists {
		return nil, errors.New("handler not registered for command: " + reflect.TypeOf(T).Name())
	}

	return handler.(handlers.IHandler), nil
}
