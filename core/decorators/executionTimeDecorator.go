package decorators

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"log"
	"time"
)

type ExecutionTimeDecorator[T command.ICommand] struct {
	next command.ICommandHandler[T]
}

func (e *ExecutionTimeDecorator[T]) Handle(cmd T) error {
	start := time.Now()
	log.Printf("Execution started @ %s\n", start.Format(time.RFC3339Nano))

	c := e.next.Handle(cmd)

	stop := time.Now()
	log.Printf("Execution completed @ %s\t total time: %s\n", stop.Format(time.RFC3339Nano), stop.Sub(start).String())

	return c
}

func UseExecutionTimeDecorator[T command.ICommand](handler command.ICommandHandler[T]) command.ICommandHandler[T] {
	return &ExecutionTimeDecorator[T]{next: handler}
}
