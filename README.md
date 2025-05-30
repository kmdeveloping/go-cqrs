# go-cqrs

A lightweight, type-safe CQRS (Command Query Responsibility Segregation) implementation for Go applications with full `context.Context` support.

## Overview

This library provides a clean, type-safe way to implement the CQRS pattern in Go applications. It separates operations into:

- **Commands**: Write operations that change state
- **Queries**: Read operations that return data
- **Events**: Notifications that something has happened
- **Validators**: Validation rules for commands

## Features

- ✅ **Full context.Context support** - Request tracing, timeouts, cancellation, and request-scoped data
- ✅ **Type-safe handlers** using Go generics
- ✅ **Command validation** with context-aware validators
- ✅ **Decorator pattern** for cross-cutting concerns
- ✅ **Auto-registration** of handlers using code generation
- ✅ **Thread-safe** handler registry
- ✅ **Production-ready** with timeout, cancellation, and observability support

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
    // you can set a Result value in the command handler 
    // this can be accessed from the executeCommand caller
    // command.BaseWithResult 
    command.Base
    Something string
}

var _ command.ICommand = (*DoSomethingCommand)(nil)
```

**Query Example:**
```go
package queries

import "github.com/kmdeveloping/go-cqrs/query"

type GetNameQuery struct {
    query.Base
    ID string
}

var _ query.IQuery = (*GetNameQuery)(nil)
```

**Event Example:**
```go
package events

import "github.com/kmdeveloping/go-cqrs/event"

type SomeEvent struct {
    event.Base
    Message string
}

var _ event.IEvent = (*SomeEvent)(nil)
```

### 2. Implement Context-Aware Handlers

All handlers now accept `context.Context` as the first parameter for modern Go patterns:

**Command Handler Example:**
```go
package handlers

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/command"
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/example/commands"
    "github.com/kmdeveloping/go-cqrs/example/events"
)

type DoThatCommandHandler struct{}

var _ command.ICommandHandler[commands.DoSomethingCommand] = (*DoThatCommandHandler)(nil)
func (h *DoThatCommandHandler) Handle(cmd *commands.DoSomethingCommand) error {
    // Handle the command
    return nil
}

// reference the interface to make sure your implementing it correctly
var _ command.ICommandHandler[commands.DoSomthingCommand] = (*DoThatCommandHandler)(nil)

// or publish an event from a handler
type DoSomeCommandWithEventPublishingHandler struct {}

func (h *DoThatCommandHandler) Handle(ctx context.Context, cmd *commands.DoSomethingCommand) error {
    // Access request-scoped data
    userID := ctx.Value("userID").(string)
    
    // Handle the command with context
    // ... business logic ...
    
    // Publish events with context propagation
    return cqrs.PublishEvent(ctx, events.SomeEvent{
        Message: cmd.Something,
    })
func (h *DoSomeCommandWithEventPublishingHandler) Handle(cmd *commands.DoSomeCommandWithEvent) error {
    // handle the command and publish an event to do something next

    // events will be completed in sequence of the publish call
    // the command handler will return once all events complete
    return cqrs.PublishEvent(events.SomeEventToPublish{
        SomeParam: "I am an event"
    })
}

// reference the interface ...
var _ command.ICommandHandler[commands.DoSomeCommandWithEvent] = (*DoSomeCommandWithEventPublishingHandler)(nil)
```

**Query Handler Example:**
```go
package handlers

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/query"
    "github.com/kmdeveloping/go-cqrs/example/queries"
)

type GetNameQueryHandler struct{}

var _ query.IQueryHandler[queries.GetNameQuery, queries.GetNameQueryResponse] = (*GetNameQueryHandler)(nil)

func (h *GetNameQueryHandler) Handle(ctx context.Context, qry queries.GetNameQuery) (queries.GetNameQueryResponse, error) {
    // Use context for database operations with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Check for cancellation
    select {
    case <-ctx.Done():
        return queries.GetNameQueryResponse{}, ctx.Err()
    default:
    }
    
    // Query with context
    return h.db.QueryWithContext(ctx, qry.ID)
}
```

**Event Handler Example:**
```go
package handlers

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/event"
    "github.com/kmdeveloping/go-cqrs/example/events"
)

type SomeEventHandler struct{}

var _ event.IEventHandler[events.SomeEvent] = (*SomeEventHandler)(nil)

func (h *SomeEventHandler) Handle(ctx context.Context, event events.SomeEvent) error {
    // Propagate context to external services
    return h.notificationService.SendNotification(ctx, event.Message)
}
```

**Validator Example:**
```go
package handlers

import (
    "context"
    "fmt"
    "github.com/kmdeveloping/go-cqrs/validator"
    "github.com/kmdeveloping/go-cqrs/example/commands"
)

type DoSomethingCommandValidator struct{
    db *sql.DB
}

var _ validator.IValidatorHandler[commands.DoSomethingCommand] = (*DoSomethingCommandValidator)(nil)

func (v *DoSomethingCommandValidator) Validate(ctx context.Context, cmd *commands.DoSomethingCommand) error {
    // Basic validation
    if len(cmd.Something) < 6 {
        return fmt.Errorf("parameter [Something] must have at least 6 characters")
    }
    
    // Database validation with context and timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Check business rules with context
    userID := ctx.Value("userID").(string)
    return v.validateUserPermissions(ctx, userID, cmd)
}
```

### 3. Bootstrap the CQRS Manager

Initialize the CQRS manager with context-aware decorators:

```go
package main

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

func init() {
    // Initialize CQRS manager
    manager := cqrs.NewCqrsManager()
    
    // Add context-aware decorators
    manager.AddLoggingDecorator()
    manager.AddMetricsDecorator()
    
    // Add custom decorators
    manager.AddDecorator(TimeoutDecorator(30 * time.Second))
    manager.AddDecorator(RequestIDDecorator())
    
    // Register handlers
    RegisterHandlers()
}

func RegisterHandlers() {
    // Register validators (run before handlers)
    cqrs.RegisterValidator(&handlers.DoSomethingCommandValidator{})
    
    // Register handlers
    cqrs.RegisterCommandHandler(&handlers.DoThatCommandHandler{})
    cqrs.RegisterQueryHandler(&handlers.GetNameQueryHandler{})
    cqrs.RegisterEventHandler(&handlers.SomeEventHandler{})
}
```

### 4. Execute Commands, Queries and Events with Context

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "github.com/kmdeveloping/go-cqrs/example/commands"
    "github.com/kmdeveloping/go-cqrs/example/events"
    "github.com/kmdeveloping/go-cqrs/example/queries"
)

func main() {
    // Create context with request-scoped data
    ctx := context.Background()
    ctx = context.WithValue(ctx, "userID", "user-123")
    ctx = context.WithValue(ctx, "tenantID", "tenant-456")
    ctx = context.WithValue(ctx, "requestID", "req-789")
    
    // Execute a command with context (validators run first)
    cmd := &commands.DoSomethingCommand{Something: "example"}
    if err := cqrs.ExecuteCommand(ctx, cmd); err != nil {
    // Initialize CQRS (as shown above)
    // ...
    
    // Execute a command using a pointer
    cmd := &commands.DoSomethingCommand{Something: "example"}
    if err := cqrs.ExecuteCommand(cmd); err != nil {
        log.Fatalf("Command execution failed: %v", err)
    }
    
    // Execute a query with context
    query := queries.GetNameQuery{ID: "123"}
    result, err := cqrs.ExecuteQuery[queries.GetNameQuery, queries.GetNameQueryResponse](ctx, query)
    if err != nil {
        log.Fatalf("Query execution failed: %v", err)
    }
    fmt.Printf("Query result: %+v\n", result)
    
    // Publish an event with context
    event := events.SomeEvent{Message: "Something happened"}
    if err := cqrs.PublishEvent(ctx, event); err != nil {
        log.Fatalf("Event publishing failed: %v", err)
    }
}
```

## Context Support Benefits

### 1. Request Tracing & Observability

```go
func (h *UserCommandHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    span, ctx := opentracing.StartSpanFromContext(ctx, "create-user")
    defer span.Finish()
    
    // Implementation with distributed tracing
    return h.userService.CreateUser(ctx, cmd)
}
```

### 2. Timeout & Cancellation

```go
func (h *ReportQueryHandler) Handle(ctx context.Context, qry *GenerateReportQuery) (*Report, error) {
    // Set timeout for long-running operations
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err() // Handle cancellation/timeout
    default:
        return h.generateReport(ctx, qry)
    }
}
```

### 3. Request-Scoped Data

```go
func (h *OrderCommandHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    userID := ctx.Value("userID").(string)
    tenantID := ctx.Value("tenantID").(string)
    
    // Use request-scoped data for business logic
    return h.createOrder(ctx, cmd, userID, tenantID)
}
```

### 4. Database Transaction Context

```go
func (h *PaymentCommandHandler) Handle(ctx context.Context, cmd *ProcessPaymentCommand) error {
    // Start transaction with context
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Use transaction context for nested operations
    ctx = context.WithValue(ctx, "transaction", tx)
    
    if err := h.processPayment(ctx, cmd); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Advanced Context Patterns

### Real-World Handler Examples

#### Complex Command Handler with Transaction Context
```go
type ProcessPaymentHandler struct {
    db *sql.DB
}

func (h *ProcessPaymentHandler) Handle(ctx context.Context, cmd *ProcessPaymentCommand) error {
    // Start transaction with context
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Add transaction to context for nested operations
    ctx = context.WithValue(ctx, "transaction", tx)

    // Process payment with transaction context
    if err := h.processPayment(ctx, cmd); err != nil {
        return fmt.Errorf("payment processing failed: %w", err)
    }

    // Update order status
    if err := h.updateOrderStatus(ctx, cmd.OrderID, "paid"); err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }

    // Commit transaction
    return tx.Commit()
}

func (h *ProcessPaymentHandler) processPayment(ctx context.Context, cmd *ProcessPaymentCommand) error {
    tx, ok := ctx.Value("transaction").(*sql.Tx)
    if !ok {
        return fmt.Errorf("transaction not found in context")
    }

    // Simulate payment processing with timeout
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    query := `INSERT INTO payments (order_id, amount, status) VALUES (?, ?, 'completed')`
    _, err := tx.ExecContext(ctx, query, cmd.OrderID, cmd.Amount)
    return err
}
```

#### Cancellation-Aware Query Handler
```go
type GetOrderHandler struct {
    db *sql.DB
}

func (h *GetOrderHandler) Handle(ctx context.Context, qry GetOrderQuery) (OrderResponse, error) {
    var response OrderResponse

    // Check for cancellation before starting
    select {
    case <-ctx.Done():
        return response, ctx.Err()
    default:
    }

    // Query with context and timeout
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    query := `SELECT id, user_id, tenant_id, status FROM orders WHERE id = ?`
    row := h.db.QueryRowContext(ctx, query, qry.OrderID)

    err := row.Scan(&response.ID, &response.UserID, &response.TenantID, &response.Status)
    if err != nil {
        return response, fmt.Errorf("failed to get order: %w", err)
    }

    return response, nil
}
```

### Context-Aware Decorators

#### Timeout Decorator
```go
func TimeoutDecorator(timeout time.Duration) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            ctx, cancel := context.WithTimeout(ctx, timeout)
            defer cancel()
            
            return next.Handle(ctx, message)
        })
    }
}
```

#### Request ID Decorator
```go
func RequestIDDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            requestID, exists := ctx.Value("requestID").(string)
            if !exists {
                requestID = generateRequestID()
                ctx = context.WithValue(ctx, "requestID", requestID)
            }

            log.Printf("[%s] Processing %T", requestID, message)
            result, err := next.Handle(ctx, message)

            if err != nil {
                log.Printf("[%s] Error processing %T: %v", requestID, message, err)
            } else {
                log.Printf("[%s] Successfully processed %T", requestID, message)
            }

            return result, err
        })
    }
}

func generateRequestID() string {
    return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
```

#### Tenant Isolation Decorator
```go
func TenantIsolationDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            tenantID, ok := ctx.Value("tenantID").(string)
            if !ok {
                return nil, fmt.Errorf("tenant ID required but not found in context")
            }
            
            log.Printf("Processing %T for tenant %s", message, tenantID)
            return next.Handle(ctx, message)
        })
    }
}
```

#### Circuit Breaker Decorator
```go
func CircuitBreakerDecorator(maxFailures int, timeout time.Duration) decorators.HandlerDecorator {
    cb := NewCircuitBreaker(maxFailures, timeout)
    
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            return cb.ExecuteWithContext(ctx, func() (any, error) {
                return next.Handle(ctx, message)
            })
        })
    }
}

// Simple circuit breaker implementation
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures: maxFailures,
        timeout:     timeout,
        state:       "closed",
    }
}

func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func() (any, error)) (any, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Simple circuit breaker logic
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = "half-open"
        } else {
            return nil, fmt.Errorf("circuit breaker is open")
        }
    }

    result, err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return nil, err
    }

    // Reset on success
    cb.failures = 0
    cb.state = "closed"
    return result, nil
}
```

### Advanced Validation Patterns

#### Database Validation with Context
```go
type CreateOrderValidator struct {
    db *sql.DB
}

func (v *CreateOrderValidator) Validate(ctx context.Context, cmd *CreateOrderCommand) error {
    // Extract request-scoped data
    userID, ok := ctx.Value("userID").(string)
    if !ok {
        return fmt.Errorf("user ID required for validation")
    }
    
    tenantID, ok := ctx.Value("tenantID").(string)
    if !ok {
        return fmt.Errorf("tenant ID required for validation")
    }
    
    // Basic validation
    if cmd.Quantity <= 0 {
        return fmt.Errorf("quantity must be greater than 0")
    }

    if cmd.Price <= 0 {
        return fmt.Errorf("price must be greater than 0")
    }
    
    // Database validation with timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Validate user in tenant
    if err := v.validateUserInTenant(ctx, userID, tenantID); err != nil {
        return fmt.Errorf("user validation failed: %w", err)
    }
    
    // Validate product availability
    if err := v.validateProductAvailability(ctx, cmd.ProductID, cmd.Quantity, tenantID); err != nil {
        return fmt.Errorf("product validation failed: %w", err)
    }
    
    return nil
}

func (v *CreateOrderValidator) validateUserInTenant(ctx context.Context, userID, tenantID string) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    query := `SELECT COUNT(*) FROM users WHERE id = ? AND tenant_id = ? AND active = 1`
    var count int
    err := v.db.QueryRowContext(ctx, query, userID, tenantID).Scan(&count)
    if err != nil {
        return fmt.Errorf("failed to validate user: %w", err)
    }

    if count == 0 {
        return fmt.Errorf("user %s not found or not active in tenant %s", userID, tenantID)
    }

    return nil
}

func (v *CreateOrderValidator) validateProductAvailability(ctx context.Context, productID string, quantity int, tenantID string) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    query := `SELECT stock_quantity FROM products WHERE id = ? AND tenant_id = ? AND active = 1`
    var stock int
    err := v.db.QueryRowContext(ctx, query, productID, tenantID).Scan(&stock)
    if err != nil {
        if err == sql.ErrNoRows {
            return fmt.Errorf("product %s not found in tenant %s", productID, tenantID)
        }
        return fmt.Errorf("failed to check product availability: %w", err)
    }

    if stock < quantity {
        return fmt.Errorf("insufficient stock for product %s: requested %d, available %d", productID, quantity, stock)
    }

    return nil
}
```

#### External Service Validation
```go
type ProcessPaymentValidator struct {
    paymentService PaymentValidationService
}

func (v *ProcessPaymentValidator) Validate(ctx context.Context, cmd *ProcessPaymentCommand) error {
    // Validate amount
    if cmd.Amount <= 0 {
        return fmt.Errorf("payment amount must be greater than 0")
    }
    
    // Use context for external service validation with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    userID, _ := ctx.Value("userID").(string)
    return v.paymentService.ValidatePaymentLimits(ctx, userID, cmd.Amount)
}

type PaymentValidationService interface {
    ValidatePaymentLimits(ctx context.Context, userID string, amount float64) error
}
```

### Complete Example Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

func main() {
    // Initialize CQRS with context-aware decorators
    manager := cqrs.NewCqrsManager()

    // Add decorators in order
    manager.AddDecorator(RequestIDDecorator())
    manager.AddDecorator(TenantIsolationDecorator())
    manager.AddDecorator(TimeoutDecorator(30 * time.Second))
    manager.AddDecorator(CircuitBreakerDecorator(5, 1*time.Minute))

    // Register validators (these run before handlers)
    cqrs.RegisterValidator(&CreateOrderValidator{})
    cqrs.RegisterValidator(&ProcessPaymentValidator{})

    // Register handlers
    cqrs.RegisterCommandHandler(&CreateOrderHandler{})
    cqrs.RegisterCommandHandler(&ProcessPaymentHandler{})
    cqrs.RegisterQueryHandler(&GetOrderHandler{})
    cqrs.RegisterEventHandler(&OrderCreatedHandler{})

    // Create context with request-scoped data
    ctx := context.Background()
    ctx = context.WithValue(ctx, "requestID", "req-123")
    ctx = context.WithValue(ctx, "userID", "user-456")
    ctx = context.WithValue(ctx, "tenantID", "tenant-789")

    // Execute command with context (validators will run first)
    cmd := &CreateOrderCommand{
        ProductID: "product-123",
        Quantity:  2,
        Price:     99.99,
    }

    if err := cqrs.ExecuteCommand(ctx, cmd); err != nil {
        log.Printf("Command execution failed: %v", err)
        // This could be a validation error or handler error
    }

    // Execute query with context
    qry := GetOrderQuery{OrderID: "order-123"}
    result, err := cqrs.ExecuteQuery[GetOrderQuery, OrderResponse](ctx, qry)
    if err != nil {
        log.Printf("Query execution failed: %v", err)
    } else {
        log.Printf("Order: %+v", result)
    }
}
```

## Command Execution Flow

The library now follows this execution flow for commands:

1. **Context Creation** - Request context with scoped data
2. **Validator Execution** - All registered validators run with context
3. **Validation Failure** - If any validator fails, command execution stops
4. **Handler Execution** - If validation passes, command handler runs
5. **Event Publishing** - Events published with context propagation

```go
// Example execution flow
ctx := context.Background()
ctx = context.WithValue(ctx, "userID", "user-123")
ctx = context.WithValue(ctx, "tenantID", "tenant-456")

cmd := &CreateOrderCommand{
    ProductID: "product-123",
    Quantity:  2,
    Price:     99.99,
}

// This will:
// 1. Run CreateOrderValidator.Validate(ctx, cmd)
// 2. If validation passes, run CreateOrderHandler.Handle(ctx, cmd)
// 3. If handler succeeds, publish any events with context
err := cqrs.ExecuteCommand(ctx, cmd)
```

## Migration from v1.x

### Breaking Changes
- All handler interfaces now require `context.Context` as the first parameter
- Execution methods require context parameter
- Validator interface signature changed

### Migration Steps

1. **Update Handler Signatures:**
   ```go
   // Before
   func (h *MyHandler) Handle(cmd *MyCommand) error
   
   // After
   func (h *MyHandler) Handle(ctx context.Context, cmd *MyCommand) error
   ```

2. **Update Execution Calls:**
   ```go
   // Before
   err := cqrs.ExecuteCommand(cmd)
   
   // After
   ctx := context.Background() // or request context
   err := cqrs.ExecuteCommand(ctx, cmd)
   ```

3. **Update Validators:**
   ```go
   // Before
   func (v *MyValidator) Validate(cmd *MyCommand) error
   
   // After
   func (v *MyValidator) Validate(ctx context.Context, cmd *MyCommand) error
   ```

### Backward Compatibility

For easier migration, convenience methods are provided:

```go
func ExecuteCommandWithBackground[T any](cmd *T) error
func ExecuteQueryWithBackground[T query.IQuery, R any](qry T) (R, error)
func PublishEventWithBackground[T event.IEvent](e T) error
```

## Context Support Benefits

### 1. Request Tracing & Observability

```go
func (h *UserCommandHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    span, ctx := opentracing.StartSpanFromContext(ctx, "create-user")
    defer span.Finish()
    
    // Implementation with distributed tracing
    return h.userService.CreateUser(ctx, cmd)
}
```

### 2. Timeout & Cancellation

```go
func (h *ReportQueryHandler) Handle(ctx context.Context, qry *GenerateReportQuery) (*Report, error) {
    // Set timeout for long-running operations
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err() // Handle cancellation/timeout
    default:
        return h.generateReport(ctx, qry)
    }
}
```

### 3. Request-Scoped Data

```go
func (h *OrderCommandHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    userID := ctx.Value("userID").(string)
    tenantID := ctx.Value("tenantID").(string)
    
    // Use request-scoped data for business logic
    return h.createOrder(ctx, cmd, userID, tenantID)
}
```

### 4. Database Transaction Context

```go
func (h *PaymentCommandHandler) Handle(ctx context.Context, cmd *ProcessPaymentCommand) error {
    // Start transaction with context
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Use transaction context for nested operations
    ctx = context.WithValue(ctx, "transaction", tx)
    
    if err := h.processPayment(ctx, cmd); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Advanced Context Patterns

### Real-World Handler Examples

#### Complex Command Handler with Transaction Context
```go
type ProcessPaymentHandler struct {
    db *sql.DB
}

func (h *ProcessPaymentHandler) Handle(ctx context.Context, cmd *ProcessPaymentCommand) error {
    // Start transaction with context
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Add transaction to context for nested operations
    ctx = context.WithValue(ctx, "transaction", tx)

    // Process payment with transaction context
    if err := h.processPayment(ctx, cmd); err != nil {
        return fmt.Errorf("payment processing failed: %w", err)
    }

    // Update order status
    if err := h.updateOrderStatus(ctx, cmd.OrderID, "paid"); err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }

    // Commit transaction
    return tx.Commit()
}

func (h *ProcessPaymentHandler) processPayment(ctx context.Context, cmd *ProcessPaymentCommand) error {
    tx, ok := ctx.Value("transaction").(*sql.Tx)
    if !ok {
        return fmt.Errorf("transaction not found in context")
    }

    // Simulate payment processing with timeout
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    query := `INSERT INTO payments (order_id, amount, status) VALUES (?, ?, 'completed')`
    _, err := tx.ExecContext(ctx, query, cmd.OrderID, cmd.Amount)
    return err
}
```

#### Cancellation-Aware Query Handler
```go
type GetOrderHandler struct {
    db *sql.DB
}

func (h *GetOrderHandler) Handle(ctx context.Context, qry GetOrderQuery) (OrderResponse, error) {
    var response OrderResponse

    // Check for cancellation before starting
    select {
    case <-ctx.Done():
        return response, ctx.Err()
    default:
    }

    // Query with context and timeout
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    query := `SELECT id, user_id, tenant_id, status FROM orders WHERE id = ?`
    row := h.db.QueryRowContext(ctx, query, qry.OrderID)

    err := row.Scan(&response.ID, &response.UserID, &response.TenantID, &response.Status)
    if err != nil {
        return response, fmt.Errorf("failed to get order: %w", err)
    }

    return response, nil
}
```

### Context-Aware Decorators

#### Timeout Decorator
```go
func TimeoutDecorator(timeout time.Duration) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            ctx, cancel := context.WithTimeout(ctx, timeout)
            defer cancel()
            
            return next.Handle(ctx, message)
        })
    }
}
```

#### Request ID Decorator
```go
func RequestIDDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            requestID, exists := ctx.Value("requestID").(string)
            if !exists {
                requestID = generateRequestID()
                ctx = context.WithValue(ctx, "requestID", requestID)
            }

            log.Printf("[%s] Processing %T", requestID, message)
            result, err := next.Handle(ctx, message)

            if err != nil {
                log.Printf("[%s] Error processing %T: %v", requestID, message, err)
            } else {
                log.Printf("[%s] Successfully processed %T", requestID, message)
            }

            return result, err
        })
    }
}

func generateRequestID() string {
    return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
```

#### Tenant Isolation Decorator
```go
func TenantIsolationDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            tenantID, ok := ctx.Value("tenantID").(string)
            if !ok {
                return nil, fmt.Errorf("tenant ID required but not found in context")
            }
            
            log.Printf("Processing %T for tenant %s", message, tenantID)
            return next.Handle(ctx, message)
        })
    }
}
```

#### Circuit Breaker Decorator
```go
func CircuitBreakerDecorator(maxFailures int, timeout time.Duration) decorators.HandlerDecorator {
    cb := NewCircuitBreaker(maxFailures, timeout)
    
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            return cb.ExecuteWithContext(ctx, func() (any, error) {
                return next.Handle(ctx, message)
            })
        })
    }
}

// Simple circuit breaker implementation
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures: maxFailures,
        timeout:     timeout,
        state:       "closed",
    }
}

func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func() (any, error)) (any, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Simple circuit breaker logic
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = "half-open"
        } else {
            return nil, fmt.Errorf("circuit breaker is open")
        }
    }

    result, err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return nil, err
    }

    // Reset on success
    cb.failures = 0
    cb.state = "closed"
    return result, nil
}
```

### Advanced Validation Patterns

#### Database Validation with Context
```go
type CreateOrderValidator struct {
    db *sql.DB
}

func (v *CreateOrderValidator) Validate(ctx context.Context, cmd *CreateOrderCommand) error {
    // Extract request-scoped data
    userID, ok := ctx.Value("userID").(string)
    if !ok {
        return fmt.Errorf("user ID required for validation")
    }
    
    tenantID, ok := ctx.Value("tenantID").(string)
    if !ok {
        return fmt.Errorf("tenant ID required for validation")
    }
    
    // Basic validation
    if cmd.Quantity <= 0 {
        return fmt.Errorf("quantity must be greater than 0")
    }

    if cmd.Price <= 0 {
        return fmt.Errorf("price must be greater than 0")
    }
    
    // Database validation with timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Validate user in tenant
    if err := v.validateUserInTenant(ctx, userID, tenantID); err != nil {
        return fmt.Errorf("user validation failed: %w", err)
    }
    
    // Validate product availability
    if err := v.validateProductAvailability(ctx, cmd.ProductID, cmd.Quantity, tenantID); err != nil {
        return fmt.Errorf("product validation failed: %w", err)
    }
    
    return nil
}

func (v *CreateOrderValidator) validateUserInTenant(ctx context.Context, userID, tenantID string) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    query := `SELECT COUNT(*) FROM users WHERE id = ? AND tenant_id = ? AND active = 1`
    var count int
    err := v.db.QueryRowContext(ctx, query, userID, tenantID).Scan(&count)
    if err != nil {
        return fmt.Errorf("failed to validate user: %w", err)
    }

    if count == 0 {
        return fmt.Errorf("user %s not found or not active in tenant %s", userID, tenantID)
    }

    return nil
}

func (v *CreateOrderValidator) validateProductAvailability(ctx context.Context, productID string, quantity int, tenantID string) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    query := `SELECT stock_quantity FROM products WHERE id = ? AND tenant_id = ? AND active = 1`
    var stock int
    err := v.db.QueryRowContext(ctx, query, productID, tenantID).Scan(&stock)
    if err != nil {
        if err == sql.ErrNoRows {
            return fmt.Errorf("product %s not found in tenant %s", productID, tenantID)
        }
        return fmt.Errorf("failed to check product availability: %w", err)
    }

    if stock < quantity {
        return fmt.Errorf("insufficient stock for product %s: requested %d, available %d", productID, quantity, stock)
    }

    return nil
}
```

#### External Service Validation
```go
type ProcessPaymentValidator struct {
    paymentService PaymentValidationService
}

func (v *ProcessPaymentValidator) Validate(ctx context.Context, cmd *ProcessPaymentCommand) error {
    // Validate amount
    if cmd.Amount <= 0 {
        return fmt.Errorf("payment amount must be greater than 0")
    }
    
    // Use context for external service validation with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    userID, _ := ctx.Value("userID").(string)
    return v.paymentService.ValidatePaymentLimits(ctx, userID, cmd.Amount)
}

type PaymentValidationService interface {
    ValidatePaymentLimits(ctx context.Context, userID string, amount float64) error
}
```

### Complete Example Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

func main() {
    // Initialize CQRS with context-aware decorators
    manager := cqrs.NewCqrsManager()

    // Add decorators in order
    manager.AddDecorator(RequestIDDecorator())
    manager.AddDecorator(TenantIsolationDecorator())
    manager.AddDecorator(TimeoutDecorator(30 * time.Second))
    manager.AddDecorator(CircuitBreakerDecorator(5, 1*time.Minute))

    // Register validators (these run before handlers)
    cqrs.RegisterValidator(&CreateOrderValidator{})
    cqrs.RegisterValidator(&ProcessPaymentValidator{})

    // Register handlers
    cqrs.RegisterCommandHandler(&CreateOrderHandler{})
    cqrs.RegisterCommandHandler(&ProcessPaymentHandler{})
    cqrs.RegisterQueryHandler(&GetOrderHandler{})
    cqrs.RegisterEventHandler(&OrderCreatedHandler{})

    // Create context with request-scoped data
    ctx := context.Background()
    ctx = context.WithValue(ctx, "requestID", "req-123")
    ctx = context.WithValue(ctx, "userID", "user-456")
    ctx = context.WithValue(ctx, "tenantID", "tenant-789")

    // Execute command with context (validators will run first)
    cmd := &CreateOrderCommand{
        ProductID: "product-123",
        Quantity:  2,
        Price:     99.99,
    }

    if err := cqrs.ExecuteCommand(ctx, cmd); err != nil {
        log.Printf("Command execution failed: %v", err)
        // This could be a validation error or handler error
    }

    // Execute query with context
    qry := GetOrderQuery{OrderID: "order-123"}
    result, err := cqrs.ExecuteQuery[GetOrderQuery, OrderResponse](ctx, qry)
    if err != nil {
        log.Printf("Query execution failed: %v", err)
    } else {
        log.Printf("Order: %+v", result)
    }
}
```

## Command Execution Flow

The library now follows this execution flow for commands:

1. **Context Creation** - Request context with scoped data
2. **Validator Execution** - All registered validators run with context
3. **Validation Failure** - If any validator fails, command execution stops
4. **Handler Execution** - If validation passes, command handler runs
5. **Event Publishing** - Events published with context propagation

```go
// Example execution flow
ctx := context.Background()
ctx = context.WithValue(ctx, "userID", "user-123")
ctx = context.WithValue(ctx, "tenantID", "tenant-456")

cmd := &CreateOrderCommand{
    ProductID: "product-123",
    Quantity:  2,
    Price:     99.99,
}

// This will:
// 1. Run CreateOrderValidator.Validate(ctx, cmd)
// 2. If validation passes, run CreateOrderHandler.Handle(ctx, cmd)
// 3. If handler succeeds, publish any events with context
err := cqrs.ExecuteCommand(ctx, cmd)
```

## Migration from v1.x

### Breaking Changes
- All handler interfaces now require `context.Context` as the first parameter
- Execution methods require context parameter
- Validator interface signature changed

### Migration Steps

1. **Update Handler Signatures:**
   ```go
   // Before
   func (h *MyHandler) Handle(cmd *MyCommand) error
   
   // After
   func (h *MyHandler) Handle(ctx context.Context, cmd *MyCommand) error
   ```

2. **Update Execution Calls:**
   ```go
   // Before
   err := cqrs.ExecuteCommand(cmd)
   
   // After
   ctx := context.Background() // or request context
   err := cqrs.ExecuteCommand(ctx, cmd)
   ```

3. **Update Validators:**
   ```go
   // Before
   func (v *MyValidator) Validate(cmd *MyCommand) error
   
   // After
   func (v *MyValidator) Validate(ctx context.Context, cmd *MyCommand) error
   ```

### Backward Compatibility

For easier migration, convenience methods are provided:

```go
func ExecuteCommandWithBackground[T any](cmd *T) error
func ExecuteQueryWithBackground[T query.IQuery, R any](qry T) (R, error)
func PublishEventWithBackground[T event.IEvent](e T) error
```
## Handler Auto-Registration

### Work in progress
This project includes a code generation tool to automatically register handlers:

1. Use the `tools/gen-handler-registry/main.go` tool to scan your handlers directory
2. Add the generated registration code to your application bootstrap

Example:
```go
//go:generate go run ../tools/gen-handler-registry/main.go
```

## Testing with Context

### Basic Handler Testing
```go
func TestMyHandler(t *testing.T) {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "userID", "test-user")
    
    handler := &MyHandler{}
    cmd := &MyCommand{Data: "test"}
    
    err := handler.Handle(ctx, cmd)
    assert.NoError(t, err)
}
```

### Timeout Testing
```go
func TestHandlerTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    handler := &SlowHandler{}
    cmd := &SlowCommand{}
    
    err := handler.Handle(ctx, cmd)
    assert.Equal(t, context.DeadlineExceeded, err)
}
```

### Validator Testing
```go
func TestValidator(t *testing.T) {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "userID", "test-user")
    ctx = context.WithValue(ctx, "tenantID", "test-tenant")
    
    validator := &CreateOrderValidator{db: testDB}
    cmd := &CreateOrderCommand{
        ProductID: "test-product",
        Quantity:  1,
        Price:     10.0,
    }
    
    err := validator.Validate(ctx, cmd)
    assert.NoError(t, err)
}
```

## Performance Considerations

- **Context Overhead:** Minimal - context passing is very lightweight in Go
- **Memory Usage:** Context values should be used sparingly and for request-scoped data only
- **Cancellation:** Handlers should check `ctx.Done()` for long-running operations
- **Timeouts:** Set reasonable timeouts to prevent resource exhaustion

## Security Considerations

- **Context Values:** Never store sensitive data in context values without proper access controls
- **Timeout Attacks:** Set reasonable timeouts to prevent resource exhaustion
- **Request Isolation:** Use context to ensure proper request isolation between tenants/users

## Use Cases

This library is perfect for:

- **Microservices** with complex business logic
- **Multi-tenant applications** requiring request isolation
- **Event-driven architectures** with reliable event publishing
- **APIs** requiring request tracing and timeout management
- **CQRS implementations** with advanced validation and observability
package mydecoratos

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/decorators"
)

func CustomDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
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
        })
    }
}
```

## License

[MIT License](LICENSE)
