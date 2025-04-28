package decorators

import (
	"github.com/kmdeveloping/go-cqrs/core/command"
	"github.com/kmdeveloping/go-cqrs/core/event"
	"github.com/kmdeveloping/go-cqrs/core/handlers"
	"github.com/kmdeveloping/go-cqrs/core/query"
	"github.com/kmdeveloping/go-cqrs/core/validator"
	"log"
	"time"
)

type ExecutionTimeDecorator struct {
	next handlers.IHandler
}

var _ handlers.IHandler = (*ExecutionTimeDecorator)(nil)

func (e *ExecutionTimeDecorator) Execute(TCommand command.ICommand) error {
	start := time.Now()
	log.Printf("Execution started @ %s\n", start.Format(time.RFC3339Nano))

	cmd := e.next.Execute(TCommand)

	stop := time.Now()
	log.Printf("Execution completed @ %s\t total time: %s\n", stop.Format(time.RFC3339Nano), stop.Sub(start).String())

	return cmd
}

func (e *ExecutionTimeDecorator) Get(TQuery query.IQuery) (any, error) {
	start := time.Now()
	log.Printf("Execution started @ %s\n", start.Format(time.RFC3339Nano))

	qry, err := e.next.Get(TQuery)

	stop := time.Now()
	log.Printf("Execution completed @ %s\t total time: %s\n", stop.Format(time.RFC3339Nano), stop.Sub(start).String())

	return qry, err
}

func (e *ExecutionTimeDecorator) Publish(TEvent event.IEvent) error {
	start := time.Now()
	log.Printf("Execution started @ %s\n", start.Format(time.RFC3339Nano))

	evt := e.next.Publish(TEvent)

	stop := time.Now()
	log.Printf("Execution completed @ %s\t total time: %s\n", stop.Format(time.RFC3339Nano), stop.Sub(start).String())

	return evt
}

func (e *ExecutionTimeDecorator) Validate(TValidator validator.IValidator) error {
	start := time.Now()
	log.Printf("Execution started @ %s\n", start.Format(time.RFC3339Nano))

	val := e.next.Validate(TValidator)

	stop := time.Now()
	log.Printf("Execution completed @ %s\t total time: %s\n", stop.Format(time.RFC3339Nano), stop.Sub(start).String())

	return val
}

func UseExecutionTimeDecorator(handler handlers.IHandler) handlers.IHandler {
	return &ExecutionTimeDecorator{handler}
}
