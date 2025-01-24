package decorators

import (
	"log"

	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
)

type LoggingDecorator struct {
	logger  *log.Logger
	handler handlers.IHandler
}

var _ handlers.IHandler = (*LoggingDecorator)(nil)

// Publish implements handlers.IHandler.
func (l *LoggingDecorator) Publish(TEvent event.IEvent) error {
	l.logger.Println("i am the event logger decorator")
	return l.handler.Publish(TEvent)
}

// Validate implements handlers.IHandler.
func (l *LoggingDecorator) Validate(TValidator validator.IValidator) error {
	l.logger.Println("i am the validator logger decorator")
	return l.handler.Validate(TValidator)
}

func (l *LoggingDecorator) Get(TQuery query.IQuery) error {
	l.logger.Println("i am the query logger decorator")
	return l.handler.Get(TQuery)
}

func (l *LoggingDecorator) Run(TCommand command.ICommand) error {
	l.logger.Println("i am the logging decorator")
	return l.handler.Run(TCommand)
}

func UseLoggingDecorator(handler handlers.IHandler) handlers.IHandler {
	return &LoggingDecorator{handler: handler, logger: log.Default()}
}
