package decorators

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"log"
	"reflect"
)

type LoggingDecorator[T command.ICommand] struct {
	logger *log.Logger
	next   command.ICommandHandler[T]
}

func (l *LoggingDecorator[T]) Handle(cmd T) error {
	l.logger.Printf("Executing command type %s\n", reflect.TypeOf(cmd).Name())
	return l.next.Handle(cmd)
}

func UseLoggingDecorator[T command.ICommand](handler command.ICommandHandler[T]) command.ICommandHandler[T] {
	return &LoggingDecorator[T]{next: handler, logger: log.Default()}
}
