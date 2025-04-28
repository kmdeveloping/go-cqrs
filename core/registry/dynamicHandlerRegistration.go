package registry

import (
	"errors"
	"fmt"
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"reflect"
	"sync"
)

type Registry struct {
	Handlers sync.Map
}

func NewRegistry() Registry {
	return Registry{sync.Map{}}
}

type cmd[TCommand command.ICommand] struct{}

func RegisterCommand[TCommand command.ICommand](handler handlers.ICommandHandler[TCommand], registry *Registry) {
	k := cmd[TCommand]{}
	_, existed := registry.Handlers.LoadOrStore(reflect.TypeOf(k), handler)
	if existed {
		fmt.Println("handler already registered")
	}
}

func LoadCommand[TCommand command.ICommand](registry *Registry) (handlers.ICommandHandler[TCommand], error) {
	var k cmd[TCommand]
	handler, ok := registry.Handlers.Load(reflect.TypeOf(k))
	if !ok {
		return nil, errors.New("handler not registered")
	}

	return handler.(handlers.ICommandHandler[TCommand]), nil
}
