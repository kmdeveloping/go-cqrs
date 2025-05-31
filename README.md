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

### 1. Define Your Domain Models

**Command Example:**
```go
package commands

import "github.com/kmdeveloping/go-cqrs/command"

type CreateOrderCommand struct {
    command.Base  // or command.BaseWithResult if you need to return data
    ProductID string
    Quantity  int
    Price     float64
}

var _ command.ICommand = (*CreateOrderCommand)(nil)
```

**Query Example:**
```go
package queries

import "github.com/kmdeveloping/go-cqrs/query"

type GetOrderQuery struct {
    query.Base
    OrderID string
}

type GetOrderResponse struct {
    ID       string
    UserID   string
    Status   string
    Total    float64
}

var _ query.IQuery = (*GetOrderQuery)(nil)
```

**Event Example:**
```go
package events

import "github.com/kmdeveloping/go-cqrs/event"

type OrderCreatedEvent struct {
    event.Base
    OrderID   string
    UserID    string
    Total     float64
}

var _ event.IEvent = (*OrderCreatedEvent)(nil)
```

### 2. Implement Handlers

All handlers accept `context.Context` as the first parameter for timeout, cancellation, and request-scoped data.

**Command Handler:**
```go
package handlers

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/command"
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

type CreateOrderHandler struct {
    db *sql.DB
}

var _ command.ICommandHandler[commands.CreateOrderCommand] = (*CreateOrderHandler)(nil)

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *commands.CreateOrderCommand) error {
    // Access request-scoped data
    userID := ctx.Value("userID").(string)
    
    // Use context for database operations with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Create order
    orderID, err := h.createOrder(ctx, userID, cmd)
    if err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    
    // Publish event with context propagation
    return cqrs.PublishEvent(ctx, events.OrderCreatedEvent{
        OrderID: orderID,
        UserID:  userID,
        Total:   cmd.Price * float64(cmd.Quantity),
    })
}
```

**Query Handler:**
```go
type GetOrderHandler struct {
    db *sql.DB
}

var _ query.IQueryHandler[queries.GetOrderQuery, queries.GetOrderResponse] = (*GetOrderHandler)(nil)

func (h *GetOrderHandler) Handle(ctx context.Context, qry queries.GetOrderQuery) (queries.GetOrderResponse, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return queries.GetOrderResponse{}, ctx.Err()
    default:
    }
    
    // Query with context and timeout
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    
    var result queries.GetOrderResponse
    err := h.db.QueryRowContext(ctx, 
        "SELECT id, user_id, status, total FROM orders WHERE id = ?", 
        qry.OrderID).Scan(&result.ID, &result.UserID, &result.Status, &result.Total)
    
    return result, err
}
```

**Event Handler:**
```go
type OrderCreatedHandler struct {
    notificationService NotificationService
}

var _ event.IEventHandler[events.OrderCreatedEvent] = (*OrderCreatedHandler)(nil)

func (h *OrderCreatedHandler) Handle(ctx context.Context, event events.OrderCreatedEvent) error {
    // Propagate context to external services
    return h.notificationService.SendOrderConfirmation(ctx, event.UserID, event.OrderID)
}
```

**Validator:**
```go
type CreateOrderValidator struct {
    db *sql.DB
}

var _ validator.IValidatorHandler[commands.CreateOrderCommand] = (*CreateOrderValidator)(nil)

func (v *CreateOrderValidator) Validate(ctx context.Context, cmd *commands.CreateOrderCommand) error {
    // Basic validation
    if cmd.Quantity <= 0 {
        return fmt.Errorf("quantity must be greater than 0")
    }
    
    if cmd.Price <= 0 {
        return fmt.Errorf("price must be greater than 0")
    }
    
    // Database validation with context and timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    userID := ctx.Value("userID").(string)
    return v.validateUserCanOrder(ctx, userID, cmd)
}
```

### 3. Initialize and Register

```go
package main

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

func init() {
    // Initialize CQRS manager
    manager := cqrs.NewCqrsManager()
    
    // Add built-in decorators
    manager.AddLoggingDecorator()
    manager.AddMetricsDecorator()
    
    // Register validators (run before handlers)
    cqrs.RegisterValidator(&handlers.CreateOrderValidator{})
    
    // Register handlers
    cqrs.RegisterCommandHandler(&handlers.CreateOrderHandler{})
    cqrs.RegisterQueryHandler(&handlers.GetOrderHandler{})
    cqrs.RegisterEventHandler(&handlers.OrderCreatedHandler{})
}
```

### 4. Execute with Context

```go
func main() {
    // Create context with request-scoped data
    ctx := context.Background()
    ctx = context.WithValue(ctx, "userID", "user-123")
    ctx = context.WithValue(ctx, "requestID", "req-456")
    
    // Execute command (validators run first, then handler)
    cmd := &commands.CreateOrderCommand{
        ProductID: "product-123",
        Quantity:  2,
        Price:     99.99,
    }
    
    if err := cqrs.ExecuteCommand(ctx, cmd); err != nil {
        log.Fatalf("Command failed: %v", err)
    }
    
    // Execute query
    qry := queries.GetOrderQuery{OrderID: "order-123"}
    result, err := cqrs.ExecuteQuery[queries.GetOrderQuery, queries.GetOrderResponse](ctx, qry)
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    
    fmt.Printf("Order: %+v\n", result)
}
```

## Command Execution Flow

Commands follow this execution flow:

1. **Context Creation** - Request context with scoped data
2. **Validator Execution** - All registered validators run with context
3. **Validation Failure** - If any validator fails, execution stops
4. **Handler Execution** - If validation passes, command handler runs
5. **Event Publishing** - Events published with context propagation

```go
ctx := context.WithValue(context.Background(), "userID", "user-123")
err := cqrs.ExecuteCommand(ctx, &commands.CreateOrderCommand{...})
// 1. CreateOrderValidator.Validate(ctx, cmd) - runs first
// 2. CreateOrderHandler.Handle(ctx, cmd) - runs if validation passes
// 3. Events published with context if handler succeeds
```

## Context Benefits

### Request Tracing & Observability
```go
func (h *OrderHandler) Handle(ctx context.Context, cmd *commands.CreateOrderCommand) error {
    span, ctx := opentracing.StartSpanFromContext(ctx, "create-order")
    defer span.Finish()
    
    return h.processOrder(ctx, cmd)
}
```

### Timeout & Cancellation
```go
func (h *ReportHandler) Handle(ctx context.Context, qry *queries.GenerateReportQuery) (*Report, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        return h.generateReport(ctx, qry)
    }
}
```

### Request-Scoped Data
```go
func (h *OrderHandler) Handle(ctx context.Context, cmd *commands.CreateOrderCommand) error {
    userID := ctx.Value("userID").(string)
    tenantID := ctx.Value("tenantID").(string)
    
    return h.createOrder(ctx, cmd, userID, tenantID)
}
```

### Database Transactions
```go
func (h *PaymentHandler) Handle(ctx context.Context, cmd *commands.ProcessPaymentCommand) error {
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Use transaction context
    ctx = context.WithValue(ctx, "transaction", tx)
    
    if err := h.processPayment(ctx, cmd); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Advanced Patterns

### Custom Decorators

Decorators add cross-cutting concerns to all handlers:

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

func RequestIDDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            requestID := generateRequestID()
            ctx = context.WithValue(ctx, "requestID", requestID)
            
            log.Printf("[%s] Processing %T", requestID, message)
            return next.Handle(ctx, message)
        })
    }
}

// Register decorators
manager.AddDecorator(RequestIDDecorator())
manager.AddDecorator(TimeoutDecorator(30 * time.Second))
```

### Complex Validation

Validators can perform database checks and external service calls:

```go
func (v *CreateOrderValidator) Validate(ctx context.Context, cmd *commands.CreateOrderCommand) error {
    userID := ctx.Value("userID").(string)
    
    // Database validation with timeout
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Check user exists and is active
    var count int
    err := v.db.QueryRowContext(ctx, 
        "SELECT COUNT(*) FROM users WHERE id = ? AND active = 1", 
        userID).Scan(&count)
    if err != nil {
        return fmt.Errorf("failed to validate user: %w", err)
    }
    
    if count == 0 {
        return fmt.Errorf("user %s not found or inactive", userID)
    }
    
    // Check product availability
    var stock int
    err = v.db.QueryRowContext(ctx,
        "SELECT stock_quantity FROM products WHERE id = ?",
        cmd.ProductID).Scan(&stock)
    if err != nil {
        return fmt.Errorf("product not found: %w", err)
    }
    
    if stock < cmd.Quantity {
        return fmt.Errorf("insufficient stock: need %d, have %d", cmd.Quantity, stock)
    }
    
    return nil
}
```

### Transaction Handlers

For complex operations requiring transactions:

```go
func (h *TransferHandler) Handle(ctx context.Context, cmd *commands.TransferMoneyCommand) error {
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Add transaction to context
    ctx = context.WithValue(ctx, "transaction", tx)
    
    // Debit from account
    if err := h.debitAccount(ctx, cmd.FromAccount, cmd.Amount); err != nil {
        return err
    }
    
    // Credit to account
    if err := h.creditAccount(ctx, cmd.ToAccount, cmd.Amount); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Handler Auto-Registration

The go-cqrs library includes a powerful code generation tool that automatically scans your project for handlers and generates registration code. This eliminates the need for manual handler registration and ensures you never forget to register a new handler.

### Installation

Install the handler registry generator tool globally:

```bash
go install github.com/kmdeveloping/go-cqrs/tools/gen-handler-registry@latest
```

Or install from source:
```bash
git clone https://github.com/kmdeveloping/go-cqrs.git
cd go-cqrs
go install ./tools/gen-handler-registry
```

### Usage

1. **Organize your handlers** in a `handlers` directory (case-insensitive):
```
your-project/
├── go.mod
├── main.go
└── handlers/
    ├── CreateOrderHandler.go
    ├── GetOrderHandler.go
    ├── OrderCreatedHandler.go
    └── CreateOrderValidator.go
```

2. **Run the generator** from your project root:
```bash
gen-handler-registry
```

3. **Use the generated code** in your main function:
```go
//go:generate gen-handler-registry

package main

import (
    "context"
    "github.com/kmdeveloping/go-cqrs/cqrs"
)

func main() {
    // Initialize CQRS manager
    manager := cqrs.NewCqrsManager()
    
    // Call the generated registration function
    registerHandlers()
    
    // Your application logic...
}
```

### What Gets Generated

The tool generates a `registry_gen.go` file with proper imports and registration calls:

```go
// Code generated by gen-handler-registry. DO NOT EDIT.

package main

import (
    "github.com/kmdeveloping/go-cqrs/cqrs"
    "your-module/handlers"
)

func registerHandlers() {
    // Register handlers
    cqrs.RegisterCommandHandler(&handlers.CreateOrderHandler{})
    cqrs.RegisterQueryHandler(&handlers.GetOrderHandler{})
    cqrs.RegisterEventHandler(&handlers.OrderCreatedHandler{})
    cqrs.RegisterValidator(&handlers.CreateOrderValidator{})
}
```

### Supported Handler Types

The tool automatically detects handlers based on naming conventions:

- **Command Handlers:** Structs ending with `CommandHandler`
- **Query Handlers:** Structs ending with `QueryHandler`
- **Event Handlers:** Structs ending with `EventHandler`
- **Validators:** Structs ending with `Validator`

### Integration with Go Generate

Add the generator to your go generate workflow:

```go
//go:generate gen-handler-registry

package main

func main() {
    registerHandlers() // Generated function
    // ...
}
```

Then run:
```bash
go generate ./...
```

### Advanced Features

- **Automatic Module Detection:** Reads module path from `go.mod`
- **Cross-Platform Support:** Works on Windows, macOS, and Linux
- **Import Path Resolution:** Automatically determines correct import paths
- **Case-Insensitive Directory Search:** Finds "handlers", "Handlers", or "HANDLERS"

For complete documentation and examples, see: [tools/gen-handler-registry/README.md](tools/gen-handler-registry/README.md)

## Testing

### Basic Testing
```go
func TestCreateOrderHandler(t *testing.T) {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "userID", "test-user")
    
    handler := &CreateOrderHandler{db: testDB}
    cmd := &commands.CreateOrderCommand{
        ProductID: "test-product",
        Quantity:  1,
        Price:     10.0,
    }
    
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
    err := handler.Handle(ctx, &SlowCommand{})
    assert.Equal(t, context.DeadlineExceeded, err)
}
```

### Validator Testing
```go
func TestValidator(t *testing.T) {
    ctx := context.WithValue(context.Background(), "userID", "test-user")
    
    validator := &CreateOrderValidator{db: testDB}
    cmd := &commands.CreateOrderCommand{Quantity: -1} // Invalid
    
    err := validator.Validate(ctx, cmd)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "quantity must be greater than 0")
}
```

## Migration from Previous Versions

### Context Support Added (v1.1+)

If you're upgrading from a version without context support:

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
   ctx := context.Background() // or your request context
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

For easier migration, convenience methods are available:

```go
// These use context.Background() internally
err := cqrs.ExecuteCommandWithBackground(cmd)
result, err := cqrs.ExecuteQueryWithBackground[Query, Response](qry)
err := cqrs.PublishEventWithBackground(event)
```

## Performance Considerations

- **Context Overhead:** Minimal - context passing is lightweight
- **Memory Usage:** Use context values sparingly for request-scoped data only
- **Cancellation:** Check `ctx.Done()` in long-running operations
- **Timeouts:** Set reasonable timeouts to prevent resource exhaustion

## Security Considerations

- **Context Values:** Never store sensitive data without proper access controls
- **Timeout Attacks:** Use timeouts to prevent resource exhaustion
- **Request Isolation:** Ensure proper isolation between tenants/users

## Use Cases

Perfect for:

- **Microservices** with complex business logic
- **Multi-tenant applications** requiring request isolation
- **Event-driven architectures** with reliable event publishing
- **APIs** requiring request tracing and timeout management
- **CQRS implementations** with advanced validation and observability

## License

[MIT License](LICENSE)
