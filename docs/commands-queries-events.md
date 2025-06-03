# Commands, Queries & Events

The three core building blocks of go-cqrs: Commands change state, Queries read data, and Events notify about changes.

## 📚 Table of Contents

- [Commands](#commands)
- [Queries](#queries)
- [Events](#events)
- [Base Types](#base-types)
- [Naming Conventions](#naming-conventions)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)

## ⚡ Commands 

Commands represent **intentions to change state**. They should be named as imperatives (verbs) and return no data.

### Basic Command Structure

```go
import "github.com/kmdeveloping/go-cqrs/command"

type CreateUserCommand struct {
    command.Base  // Embeds common fields
    Name         string
    Email        string
    Password     string
}
```

### Command Examples

```go
// User management
type CreateUserCommand struct {
    command.Base
    Name     string
    Email    string
    Password string
}

type UpdateUserCommand struct {
    command.Base
    UserID int
    Name   string
    Email  string
}

type DeleteUserCommand struct {
    command.Base
    UserID int
}

// E-commerce
type PlaceOrderCommand struct {
    command.Base
    CustomerID     int
    Items          []OrderItem
    ShippingAddr   Address
    PaymentMethod  string
}

type CancelOrderCommand struct {
    command.Base
    OrderID int
    Reason  string
}
```

### Command Handlers

```go
type CreateUserHandler struct {
    userRepo     UserRepository
    emailService EmailService
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Validate business rules
    if err := h.validateUser(cmd); err != nil {
        return err
    }
    
    // Create user
    user := &User{
        Name:     cmd.Name,
        Email:    cmd.Email,
        Password: hashPassword(cmd.Password),
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        return err
    }
    
    // Publish domain event
    return cqrs.PublishEvent(ctx, UserCreatedEvent{
        UserID: user.ID,
        Name:   user.Name,
        Email:  user.Email,
    })
}
```

### Commands with Results

Some commands need to return data (like IDs). Use `BaseWithResult`:

```go
type CreateUserCommand struct {
    command.BaseWithResult
    Name  string
    Email string
}

type CreateUserHandler struct {
    userRepo UserRepository
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    user := &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        return err
    }
    
    // Set result
    cmd.Result = user.ID
    return nil
}

// Usage
func CreateNewUser(ctx context.Context, name, email string) (int, error) {
    cmd := &CreateUserCommand{Name: name, Email: email}
    
    err := cqrs.ExecuteCommand(ctx, cmd)
    if err != nil {
        return 0, err
    }
    
    return cmd.Result.(int), nil
}
```

## 🔍 Queries

Queries represent **requests for data**. They should be named as questions and never change state.

### Basic Query Structure

```go
import "github.com/kmdeveloping/go-cqrs/query"

type GetUserQuery struct {
    query.Base  // Embeds common fields
    UserID     int
}
```

### Query Examples

```go
// Single item queries
type GetUserQuery struct {
    query.Base
    UserID int
}

type GetOrderQuery struct {
    query.Base
    OrderID int
}

// List queries
type ListUsersQuery struct {
    query.Base
    Page     int
    PageSize int
    Filter   string
}

type SearchProductsQuery struct {
    query.Base
    Query    string
    Category string
    MinPrice decimal.Decimal
    MaxPrice decimal.Decimal
    Limit    int
}

// Aggregation queries
type GetOrderStatsQuery struct {
    query.Base
    CustomerID int
    FromDate   time.Time
    ToDate     time.Time
}
```

### Query Handlers

```go
type GetUserHandler struct {
    userRepo UserRepository
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    user, err := h.userRepo.GetByID(ctx, q.UserID)
    if err != nil {
        return nil, err
    }
    
    if user == nil {
        return nil, errors.New("user not found")
    }
    
    return user, nil
}

type ListUsersHandler struct {
    userRepo UserRepository
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (*UserListResult, error) {
    users, total, err := h.userRepo.List(ctx, q.Page, q.PageSize, q.Filter)
    if err != nil {
        return nil, err
    }
    
    return &UserListResult{
        Users:      users,
        TotalCount: total,
        Page:       q.Page,
        PageSize:   q.PageSize,
    }, nil
}
```

### Query Result Types

Define specific result types for better API design:

```go
type UserListResult struct {
    Users      []*UserSummary `json:"users"`
    TotalCount int            `json:"total_count"`
    Page       int            `json:"page"`
    PageSize   int            `json:"page_size"`
    HasMore    bool           `json:"has_more"`
}

type UserSummary struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Email       string    `json:"email"`
    CreatedAt   time.Time `json:"created_at"`
    LastLogin   time.Time `json:"last_login"`
    OrderCount  int       `json:"order_count"`
    TotalSpent  decimal.Decimal `json:"total_spent"`
}

type OrderStats struct {
    TotalOrders    int             `json:"total_orders"`
    TotalAmount    decimal.Decimal `json:"total_amount"`
    AverageOrder   decimal.Decimal `json:"average_order"`
    LastOrderDate  time.Time       `json:"last_order_date"`
    TopProducts    []ProductStat   `json:"top_products"`
}
```

## 📢 Events

Events represent **things that have happened**. They should be named in past tense and are immutable.

### Basic Event Structure

```go
import (
    "github.com/kmdeveloping/go-cqrs/event"
    "github.com/google/uuid"
    "time"
)

type UserCreatedEvent struct {
    event.Base  // Includes ExecutionTime, CorrelationUid, MetaData
    UserID     int
    Name       string
    Email      string
}
```

### Event Examples

```go
// User events
type UserCreatedEvent struct {
    event.Base
    UserID int
    Name   string
    Email  string
}

type UserUpdatedEvent struct {
    event.Base
    UserID   int
    Changes  map[string]interface{}
}

type UserDeletedEvent struct {
    event.Base
    UserID int
    Reason string
}

// Order events
type OrderPlacedEvent struct {
    event.Base
    OrderID    int
    CustomerID int
    Items      []OrderItem
    Total      decimal.Decimal
}

type OrderShippedEvent struct {
    event.Base
    OrderID      int
    TrackingCode string
    Carrier      string
}

type OrderCancelledEvent struct {
    event.Base
    OrderID int
    Reason  string
}
```

### Event Handlers

```go
type UserCreatedHandler struct {
    emailService     EmailService
    analyticsService AnalyticsService
}

func (h *UserCreatedHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    // Send welcome email
    if err := h.emailService.SendWelcomeEmail(ctx, e.Email, e.Name); err != nil {
        // Log error but don't fail - email is not critical
        log.Printf("Failed to send welcome email: %v", err)
    }
    
    // Update analytics
    return h.analyticsService.TrackUserRegistration(ctx, e.UserID)
}

// Multiple handlers for same event
type UserCreatedNotificationHandler struct {
    slackService SlackService
}

func (h *UserCreatedNotificationHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    return h.slackService.NotifyNewUser(ctx, e.Name, e.Email)
}
```

### Publishing Events

```go
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    user := createUser(cmd)
    
    // Publish event with correlation data
    event := UserCreatedEvent{
        Base: event.Base{
            ExecutionTime:  time.Now(),
            CorrelationUid: uuid.New(),
            MetaData:       "user-registration",
        },
        UserID: user.ID,
        Name:   user.Name,
        Email:  user.Email,
    }
    
    return cqrs.PublishEvent(ctx, event)
}
```

## 🏗️ Base Types

### Command Base Types

```go
// For commands that don't return data
type Base struct{}

// For commands that need to return data
type BaseWithResult struct {
    Result any
}
```

### Query Base Type

```go
type Base struct{}
```

### Event Base Type

```go
type Base struct {
    ExecutionTime  time.Time   // When the event occurred
    CorrelationUid uuid.UUID   // For tracing requests
    MetaData       string      // Additional context
}
```

## 📝 Naming Conventions

### Commands (Imperative - Actions)
```go
// ✅ Good
type CreateUserCommand struct{}
type UpdateOrderCommand struct{}
type DeleteProductCommand struct{}
type SendEmailCommand struct{}
type ProcessPaymentCommand struct{}

// ❌ Bad
type UserCommand struct{}           // Too generic
type UserCreation struct{}          // Not a command
type CreateUserRequest struct{}     // This is a command, not request
```

### Queries (Questions - What you want to know)
```go
// ✅ Good
type GetUserQuery struct{}
type ListOrdersQuery struct{}
type FindProductsQuery struct{}
type SearchCustomersQuery struct{}
type GetOrderStatsQuery struct{}

// ❌ Bad
type UserQuery struct{}            // Too generic
type Users struct{}                // Not descriptive
type UserRequest struct{}          // This is a query, not request
```

### Events (Past tense - What happened)
```go
// ✅ Good
type UserCreatedEvent struct{}
type OrderPlacedEvent struct{}
type PaymentProcessedEvent struct{}
type EmailSentEvent struct{}
type ProductUpdatedEvent struct{}

// ❌ Bad
type CreateUserEvent struct{}      // Not past tense
type UserEvent struct{}            // Too generic
type NewUser struct{}              // Not descriptive enough
```

## ✨ Best Practices

### 1. Keep Commands Focused

```go
// ✅ Good - single responsibility
type CreateUserCommand struct {
    command.Base
    Name     string
    Email    string
    Password string
}

type UpdateUserEmailCommand struct {
    command.Base
    UserID   int
    NewEmail string
}

// ❌ Bad - multiple responsibilities
type ManageUserCommand struct {
    command.Base
    Action   string // "create", "update", "delete"
    UserID   int
    Name     string
    Email    string
    Password string
}
```

### 2. Design Queries for UI

```go
// ✅ Good - designed for display needs
type GetUserProfileQuery struct {
    query.Base
    UserID int
}

type UserProfileResult struct {
    ID              int             `json:"id"`
    Name            string          `json:"name"`
    Email           string          `json:"email"`
    Avatar          string          `json:"avatar"`
    LastLogin       time.Time       `json:"last_login"`
    OrderCount      int             `json:"order_count"`
    TotalSpent      decimal.Decimal `json:"total_spent"`
    FavoriteItems   []Item          `json:"favorite_items"`
}

// ❌ Bad - just returns raw entity
type GetUserQuery struct {
    UserID int
}
// Returns: User{} // Raw database entity
```

### 3. Make Events Rich

```go
// ✅ Good - rich domain event
type OrderPlacedEvent struct {
    event.Base
    OrderID         int
    CustomerID      int
    CustomerEmail   string
    Items           []OrderItem
    TotalAmount     decimal.Decimal
    ShippingAddress Address
    PaymentMethod   string
}

// ❌ Bad - minimal technical event
type OrderCreatedEvent struct {
    event.Base
    OrderID int
}
```

### 4. Use Value Objects

```go
type Address struct {
    Street     string
    City       string
    State      string
    PostalCode string
    Country    string
}

type Money struct {
    Amount   decimal.Decimal
    Currency string
}

type PlaceOrderCommand struct {
    command.Base
    CustomerID      int
    Items           []OrderItem
    ShippingAddress Address
    Total           Money
}
```

## 🔄 Common Patterns

### 1. Command with Validation

```go
type CreateUserCommand struct {
    command.Base
    Name     string
    Email    string
    Password string
}

type CreateUserValidator struct{}

func (v *CreateUserValidator) Validate(ctx context.Context, cmd *CreateUserCommand) error {
    if cmd.Name == "" {
        return errors.New("name is required")
    }
    
    if !isValidEmail(cmd.Email) {
        return errors.New("invalid email")
    }
    
    if len(cmd.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    return nil
}
```

### 2. Query with Pagination

```go
type ListUsersQuery struct {
    query.Base
    Page     int
    PageSize int
    SortBy   string
    SortDir  string
    Filter   UserFilter
}

type UserFilter struct {
    Name      string
    Email     string
    Status    string
    CreatedAt DateRange
}

type DateRange struct {
    From time.Time
    To   time.Time
}
```

### 3. Event Chain

```go
// Order placement triggers multiple events
func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd *PlaceOrderCommand) error {
    order := createOrder(cmd)
    
    // Primary event
    if err := cqrs.PublishEvent(ctx, OrderPlacedEvent{
        OrderID:    order.ID,
        CustomerID: order.CustomerID,
        Total:      order.Total,
    }); err != nil {
        return err
    }
    
    // Secondary events (handled by other handlers)
    // - InventoryReservedEvent (from inventory service)
    // - PaymentAuthorizedEvent (from payment service)
    // - EmailQueuedEvent (from notification service)
    
    return nil
}
```

## 🚀 Next Steps

1. **Learn Handler Implementation**: [Handlers & Validators](./handlers-validators.md)
2. **See Full Examples**: [Basic Examples](./basic-examples.md)
3. **Add Auto-Registration**: [Auto-Registration Guide](./auto-registration.md)
4. **Production Features**: [Production Readiness](./production-ready.md)

---

**Ready to implement handlers? Continue with [Handlers & Validators](./handlers-validators.md)! 🚀** 