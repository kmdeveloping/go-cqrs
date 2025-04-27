package decorators

import (
	"log"
	"reflect"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type LoggingDecorator struct {
	logger *log.Logger
	next   handlers.IHandler
}

var _ handlers.IHandler = (*LoggingDecorator)(nil)

// Publish implements handlers.IHandler.
func (l *LoggingDecorator) Publish(TEvent event.IEvent) error {
	l.logger.Println("i am the event logger decorator")
	return l.next.Publish(TEvent)
}

// Validate implements handlers.IHandler.
func (l *LoggingDecorator) Validate(TValidator validator.IValidator) error {
	l.logger.Println("i am the validator logger decorator")
	return l.next.Validate(TValidator)
}

func (l *LoggingDecorator) Get(TQuery query.IQuery) (any, error) {
	qry := reflect.TypeOf(TQuery).Name()
	l.logger.Printf("Executing query type %s\n", qry)
	return l.next.Get(TQuery)
}

func (l *LoggingDecorator) Execute(TCommand command.ICommand) error {
	cmd := reflect.TypeOf(TCommand).Name()
	l.logger.Printf("Executing command type %s\n", cmd)
	return l.next.Execute(TCommand)
}

func (l *LoggingDecorator) ExecuteWithResult(TCommandWithResult command.ICommandWithResult) error {
	cmd := reflect.TypeOf(TCommandWithResult).Name()
	l.logger.Printf("Executing command type %s\n", cmd)
	return l.next.Execute(TCommandWithResult)
}

func UseLoggingDecorator(handler handlers.IHandler) handlers.IHandler {
	return &LoggingDecorator{next: handler, logger: log.Default()}
}
