package registry

import (
	"errors"
	"reflect"
	"regexp"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type Registry struct {
	commandHandlers  map[reflect.Type]handlers.ICommandHandler
	queryHandlers    map[reflect.Type]handlers.IQueryHandler[any]
	eventHandlers    map[reflect.Type][]handlers.IEventHandler
	validateHandlers map[reflect.Type]handlers.IValidatorHandler
}

type CommandServices struct {
	Command command.ICommand
	Handler handlers.ICommandHandler
}

type QueryServices struct {
	Query   query.IQuery
	Handler handlers.IQueryHandler[any]
}

type EventServices struct {
	Event    event.IEvent
	Handlers []handlers.IEventHandler
}

type ValidatorServices struct {
	Validator validator.IValidator
	Handler   handlers.IValidatorHandler
}

func NewRegistry() *Registry {
	return &Registry{
		commandHandlers:  make(map[reflect.Type]handlers.ICommandHandler),
		queryHandlers:    make(map[reflect.Type]handlers.IQueryHandler[any]),
		eventHandlers:    make(map[reflect.Type][]handlers.IEventHandler),
		validateHandlers: make(map[reflect.Type]handlers.IValidatorHandler),
	}
}

func (r *Registry) RegisterCommandHandlers(handlerList []CommandServices) *Registry {
	for _, h := range handlerList {
		r.commandHandlers[reflect.TypeOf(h.Command)] = h.Handler
	}
	return r
}

func (r *Registry) RegisterQueryHandlers(handlerList []QueryServices) *Registry {
	for _, h := range handlerList {
		r.queryHandlers[reflect.TypeOf(h.Query)] = h.Handler
	}
	return r
}

func (r *Registry) RegisterEventHandlers(handlerList []EventServices) *Registry {
	for _, h := range handlerList {
		r.eventHandlers[reflect.TypeOf(h.Event)] = h.Handlers
	}
	return r
}

func (r *Registry) RegisterValidatorHandlers(handlerList []ValidatorServices) *Registry {
	for _, h := range handlerList {
		r.validateHandlers[reflect.TypeOf(h.Validator)] = h.Handler
	}
	return r
}

func (r *Registry) Resolve(T any) (handlers.IHandler, error) {
	t := reflect.TypeOf(T)
	if cmd, _ := regexp.MatchString("Command", t.Name()); cmd {
		handler, exists := r.commandHandlers[t]
		if !exists {
			return nil, errors.New("handler not registered for command: " + t.Name())
		}

		return handler.(handlers.IHandler), nil
	}

	if qry, _ := regexp.MatchString("Query", t.Name()); qry {
		handler, exists := r.queryHandlers[t]
		if !exists {
			return nil, errors.New("handler not registered for query: " + t.Name())
		}

		return handler.(handlers.IHandler), nil
	}

	if evt, _ := regexp.MatchString("Event", t.Name()); evt {
		//handlers, exists := r.eventHandlers[t]
		//if !exists {
		//return nil, errors.New("handler not registered for event: " + t.Name())
		//}

		return nil, errors.New("event handlers not supported yet")
	}

	if val, _ := regexp.MatchString("Validator", t.Name()); val {
		handler, exists := r.validateHandlers[t]
		if !exists {
			return nil, errors.New("handler not registered for validator: " + t.Name())
		}

		return handler.(handlers.IHandler), nil
	}

	return nil, errors.New("no handlers registered for: " + t.Name())
}
