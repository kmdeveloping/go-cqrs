# go-cqrs

A lightweight, type-safe CQRS (Command Query Responsibility Segregation) implementation for Go applications.

## Overview

This library provides a clean, type-safe way to implement the CQRS pattern in Go applications. It separates operations into:

- **Commands**: Write operations that change state
- **Queries**: Read operations that return data
- **Events**: Notifications that something has happened
- **Validators**: Validation rules for commands

## Features

- Type-safe command, query, and event handling using Go generics
- Command validation
- Middleware/decorator pattern for cross-cutting concerns
- Auto-registration of handlers using code generation
- Thread-safe handler registry

## Installation

```bash
go get github.com/kmdeveloping/go-cqrs
```

## Quick Start

### 1. Define Commands, Queries, and Events

**Command Example:**
```go
package commands

import "github.com/kmdeveloping/go-cqrs/command"

type DoSomethingCommand struct {
    Something string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
```

**Query Example:**
```go
package queries

import "github.com/kmdeveloping/go-cqrs/query"

type GetNameQuery struct {
    ID string
}

var _ query.IQuery = (*GetNameQuery)(nil)
```

**Event Example:**
```go
package events

import "github.com/kmdeveloping/go-cqrs/event"

type SomeEvent struct {
    Message string
}

var _ event.IEvent = (*SomeEvent)(nil)
```

### 2. Implement Handlers

**Command Handler Example:**
```go
package handlers

import (
    "github.com/kmdeveloping/go-cqrs/example/commands"
)

type DoThatCommandHandler struct{}

func (h *DoThatCommandHandler) Handle(cmd commands.DoSomethingCommand) error {
    // Handle the command
    return nil
}

type DoSomeCommandWithEventPublishingHandler struct {}

func (h *DoSomeCommandWithEventPublishingHandler) Handle(cmd commands.DoSomeCommandWithEvent) error {
    // handle the command

    // events can be published within command and query handlers as well as from the main app
    return cqrs.PublishEvent(events.SomeEventToPublish{
        SomeParam: "I am an event"
    })    
}
```

**Query Handler Example:**
```go
package handlers

import (
    "github.com/kmdeveloping/go-cqrs/example/queries"
)

type GetNameQueryHandler struct{}

func (h *GetNameQueryHandler) Handle(query queries.GetNameQuery) (string, error) {
    // Execute query and return result
    return "SomeName", nil
}
```

**Event Handler Example:**
```go
package handlers

import (
    "github.com/kmdeveloping/go-cqrs/example/events"
)

type SomeEventHandler struct{}

func (h *SomeEventHandler) Handle(event events.SomeEvent) error {
    // Handle the event
    return nil
}
```

**Validator Example:**
```go
package handlers

import (
    "fmt"
    "github.com/kmdeveloping/go-cqrs/example/commands"
)

type DoSomethingCommandValidator struct{}

func (v *DoSomethingCommandValidator) Validate(cmd commands.DoSomethingCommand) error {
    if len(cmd.Something) < 6 {
        return fmt.Errorf("parameter [Something] must have at least 6 characters")
    }
    return nil
}
```

### 3. Bootstrap the CQRS Manager

Initialize the CQRS manager in your application:

```go
package main

import "github.com/kmdeveloping/go-cqrs/cqrs"

func init() {
    // Initialize CQRS manager
    manager := cqrs.NewCqrsManager()
    
    // Add default decorators (metrics, logging, error handling)    
    // manager.AddLoggingDecorator()
    // manager.AddMetricsDecorator()

    // Or add custom decorators
    // manager.AddDecorator(myCustomDecorator.SomeDecorator())
    
    // Register handlers manually or use auto-registration
    RegisterHandlers()
}

func RegisterHandlers() {
    // Example manual registration
    cqrs.RegisterCommandHandler(&handlers.DoThatCommandHandler{})
    cqrs.RegisterQueryHandler(&handlers.GetNameQueryHandler{})
    cqrs.RegisterEventHandler(&handlers.SomeEventHandler{})
    cqrs.RegisterValidator(&handlers.DoSomethingCommandValidator{})
}
```

### 4. Execute Commands, Queries and Events

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/example/commands"
    "github.com/kmdeveloping/go-cqrs/example/events"
    "github.com/kmdeveloping/go-cqrs/example/queries"
)

func main() {
    // Initialize CQRS (as shown above)
    // ...
    
    // Execute a command
    cmd := commands.DoSomethingCommand{Something: "example"}
    if err := cqrs.ExecuteCommand(cmd); err != nil {
        log.Fatalf("Command execution failed: %v", err)
    }
    
    // Execute a query
    query := queries.GetNameQuery{ID: "123"}
    result, err := cqrs.ExecuteQuery[queries.GetNameQuery, string](query)
    if err != nil {
        log.Fatalf("Query execution failed: %v", err)
    }
    fmt.Printf("Query result: %s\n", result)
    
    // Publish an event
    event := events.SomeEvent{Message: "Something happened"}
    if err := cqrs.PublishEvent(event); err != nil {
        log.Fatalf("Event publishing failed: %v", err)
    }
}
```

## Handler Auto-Registration

This project includes a code generation tool to automatically register handlers:

1. Use the `tools/gen-handler-registry/main.go` tool to scan your handlers directory
2. Add the generated registration code to your application bootstrap

Example:
```go
//go:generate go run ../tools/gen-handler-registry/main.go
```

## Advanced Usage

### Custom Decorators

You can create custom decorators to add cross-cutting concerns:

```go
package mydecoratos

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/decorators"
)

func CustomDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(ctx context.Context, message any) (any, error) {
            // start your decorators actions here
            // this logic will run before the base handler is run

            // return the next decorator in line
            // return next.Handle(ctx, message)

            // if your decorator needs to evaluate the base handler response then you can split the 
            // next call by assigning variables to the handle call which returns (any, error)
            // response, err := next.Handle(ctx, message)

            // do something with the response var if needed 
            // return the outputs after your logic completes so other decorators can complete their action
            // return response, err
        }
    }
}
```

## License

[MIT License](LICENSE)
