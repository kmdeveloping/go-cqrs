# CQRS Fundamentals

Understanding Command Query Responsibility Segregation and how go-cqrs implements these patterns.

## 📚 Table of Contents

- [What is CQRS?](#what-is-cqrs)
- [Core Principles](#core-principles)
- [Benefits of CQRS](#benefits-of-cqrs)
- [When to Use CQRS](#when-to-use-cqrs)
- [CQRS vs Traditional CRUD](#cqrs-vs-traditional-crud)
- [Event Sourcing vs CQRS](#event-sourcing-vs-cqrs)
- [go-cqrs Implementation](#go-cqrs-implementation)
- [Architecture Patterns](#architecture-patterns)

## 🎯 What is CQRS?

**Command Query Responsibility Segregation (CQRS)** is an architectural pattern that separates read and write operations into different models:

- **Commands**: Change state (write operations) - return no data
- **Queries**: Return data (read operations) - never change state
- **Events**: Notify about state changes (optional)

### Simple Example

```go
// ❌ Traditional approach - mixed responsibility
type UserService struct{}

func (s *UserService) CreateUser(name string) (*User, error) {
    // Creates AND returns data - mixed responsibility
}

func (s *UserService) GetUser(id int) (*User, error) {
    // Only reads - this is fine
}

// ✅ CQRS approach - separated responsibilities
type CreateUserCommand struct {
    Name string
}

type CreateUserHandler struct{}
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Only creates - no return data
}

type GetUserQuery struct {
    ID int
}

type GetUserHandler struct{}
func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    // Only reads and returns data
}
```

## 🏗️ Core Principles

### 1. **Command Query Separation**

Commands and queries are handled by different code paths:

```go
// Commands modify state
type PlaceOrderCommand struct {
    CustomerID int
    Items      []OrderItem
}

// Queries read state  
type GetOrderQuery struct {
    OrderID int
}
```

### 2. **Single Responsibility**

Each handler has one clear purpose:

```go
type PlaceOrderHandler struct {
    orderRepo OrderRepository
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd *PlaceOrderCommand) error {
    // Only responsible for placing orders
    order := &Order{
        CustomerID: cmd.CustomerID,
        Items:      cmd.Items,
        Status:     "Pending",
    }
    return h.orderRepo.Save(ctx, order)
}
```

### 3. **Event-Driven Communication**

State changes can publish events:

```go
func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd *PlaceOrderCommand) error {
    // Create order
    order := createOrder(cmd)
    
    // Publish event
    return cqrs.PublishEvent(ctx, OrderPlacedEvent{
        OrderID:    order.ID,
        CustomerID: order.CustomerID,
        Total:      order.Total,
    })
}
```

## ✨ Benefits of CQRS

### 🎯 **Scalability**
```go
// Different scaling for reads vs writes
type UserRepository struct {
    writeDB *sql.DB      // Master database
    readDB  *sql.DB      // Read replicas
}
```

### 🔒 **Security**
```go
// Different security models
type AdminCommandHandler struct {} // Requires admin role
type PublicQueryHandler struct {}  // Public access
```

### ⚡ **Performance**
```go
// Optimized data models
type WriteModel struct {
    // Normalized for writes
    ID   int
    Data NormalizedData
}

type ReadModel struct {
    // Denormalized for reads
    ID            int
    DisplayName   string
    CachedSummary string
}
```

### 🧪 **Testability**
```go
func TestCreateUser(t *testing.T) {
    handler := &CreateUserHandler{}
    
    // Test only the write logic
    err := handler.Handle(ctx, &CreateUserCommand{Name: "Test"})
    assert.NoError(t, err)
}

func TestGetUser(t *testing.T) {
    handler := &GetUserHandler{}
    
    // Test only the read logic
    user, err := handler.Handle(ctx, GetUserQuery{ID: 1})
    assert.NoError(t, err)
    assert.Equal(t, "Test", user.Name)
}
```

## 🎯 When to Use CQRS

### ✅ **Good Fit**

- **Complex business logic** with different read/write requirements
- **High-traffic applications** needing independent scaling
- **Event-driven architectures** with eventual consistency
- **Different team ownership** of read vs write operations
- **Audit requirements** with event tracking

```go
// E-commerce example - perfect for CQRS
type PlaceOrderCommand struct {
    CustomerID int
    Items      []OrderItem    // Complex validation needed
    Payments   []Payment      // Multiple payment methods
    Shipping   ShippingInfo   // Address validation
}

type OrderSummaryQuery struct {
    CustomerID int
}

type OrderSummaryResult struct {
    TotalOrders    int       // Aggregated data
    LastOrderDate  time.Time // Computed from events
    FavoriteItems  []Item    // Analytics data
}
```

### ❌ **Poor Fit**

- **Simple CRUD applications** with basic read/write operations
- **Small applications** with limited complexity
- **Strict consistency requirements** (ACID transactions)
- **Real-time reporting** needs

```go
// Simple user profile - traditional CRUD is better
type User struct {
    ID    int
    Name  string
    Email string
}

// No complex business logic, no scaling needs
func UpdateUserProfile(userID int, name, email string) error {
    return db.Update("users", userID, name, email)
}
```

## 🆚 CQRS vs Traditional CRUD

### Traditional CRUD
```go
type UserService struct {
    db *sql.DB
}

// Single model for read and write
func (s *UserService) CreateUser(user *User) error {
    return s.db.Insert(user)
}

func (s *UserService) GetUser(id int) (*User, error) {
    return s.db.Get(id)
}

func (s *UserService) UpdateUser(user *User) error {
    return s.db.Update(user)
}
```

### CQRS Approach
```go
// Separate command and query models
type CreateUserCommand struct {
    Name  string
    Email string
    // Only fields needed for creation
}

type UserSummaryQuery struct {
    ID int
}

type UserSummary struct {
    ID            int
    Name          string
    OrderCount    int       // Computed field
    LastActive    time.Time // Aggregated data
    // Optimized for display
}
```

## 🔄 Event Sourcing vs CQRS

### CQRS (This Library)
```go
// Commands change current state
type TransferMoneyCommand struct {
    FromAccount int
    ToAccount   int
    Amount      decimal.Decimal
}

// Current state is stored
type Account struct {
    ID      int
    Balance decimal.Decimal // Current balance
}
```

### Event Sourcing + CQRS
```go
// Events represent state changes
type MoneyTransferredEvent struct {
    FromAccount int
    ToAccount   int
    Amount      decimal.Decimal
    Timestamp   time.Time
}

// State is computed from events
func (a *Account) ApplyEvent(event MoneyTransferredEvent) {
    if event.FromAccount == a.ID {
        a.Balance = a.Balance.Sub(event.Amount)
    }
    if event.ToAccount == a.ID {
        a.Balance = a.Balance.Add(event.Amount)
    }
}
```

> **Note**: go-cqrs supports both approaches but doesn't require event sourcing.

## 🏛️ go-cqrs Implementation

### Type-Safe Design
```go
// Commands are strongly typed
type CreateOrderCommand struct {
    command.Base
    CustomerID int
    Items      []OrderItem
}

// Handlers use generics for type safety
type CreateOrderHandler struct{}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    // cmd is guaranteed to be *CreateOrderCommand
    return h.processOrder(cmd.CustomerID, cmd.Items)
}
```

### Context Support
```go
func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    // Full context support for:
    // - Request tracing
    // - Timeouts
    // - Cancellation
    // - Request-scoped data
    
    userID := ctx.Value("userID")
    
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        return h.processOrder(cmd)
    }
}
```

### Event Publishing
```go
func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    order := createOrder(cmd)
    
    // Publish domain event
    return cqrs.PublishEvent(ctx, OrderCreatedEvent{
        OrderID:    order.ID,
        CustomerID: order.CustomerID,
    })
}
```

## 🏗️ Architecture Patterns

### 1. **Simple CQRS**
```go
// Single database, separate read/write models
type WriteModel struct {
    ID   int
    Data string
}

type ReadModel struct {
    ID          int
    DisplayData string
    ComputedValue int
}
```

### 2. **CQRS with Read Replicas**
```go
type OrderService struct {
    writeDB *sql.DB // Master
    readDB  *sql.DB // Replica
}

// Commands use write DB
func (s *OrderService) HandleCreateOrder(cmd *CreateOrderCommand) error {
    return s.writeDB.Insert(order)
}

// Queries use read DB
func (s *OrderService) HandleGetOrder(q GetOrderQuery) (*Order, error) {
    return s.readDB.Get(q.ID)
}
```

### 3. **Event-Driven CQRS**
```go
// Commands publish events
func HandleCreateOrder(ctx context.Context, cmd *CreateOrderCommand) error {
    order := createOrder(cmd)
    
    return cqrs.PublishEvent(ctx, OrderCreatedEvent{
        OrderID: order.ID,
    })
}

// Events update read models
func HandleOrderCreated(ctx context.Context, e OrderCreatedEvent) error {
    return updateOrderSummary(e.OrderID)
}
```

### 4. **Microservices CQRS**
```go
// Each service handles specific commands/queries
type OrderService struct{}   // Handles order commands
type InventoryService struct{} // Handles inventory queries
type PaymentService struct{} // Handles payment commands

// Services communicate via events
func (s *OrderService) HandleCreateOrder(cmd *CreateOrderCommand) error {
    // ... create order
    
    // Notify other services
    return cqrs.PublishEvent(ctx, OrderCreatedEvent{
        OrderID: order.ID,
        Items:   order.Items,
    })
}
```

## 🎓 Best Practices

### 1. **Keep Commands Simple**
```go
// ✅ Good - single responsibility
type CreateUserCommand struct {
    Name  string
    Email string
}

// ❌ Bad - multiple responsibilities  
type ManageUserCommand struct {
    Action string // "create", "update", "delete"
    Name   string
    Email  string
    ID     int
}
```

### 2. **Make Queries Efficient**
```go
// ✅ Good - optimized for display
type UserListQuery struct {
    Page     int
    PageSize int
    Filter   string
}

type UserListResult struct {
    Users      []UserSummary
    TotalCount int
    HasMore    bool
}
```

### 3. **Use Domain Events**
```go
// ✅ Good - meaningful domain events
type CustomerRegisteredEvent struct {
    CustomerID int
    Name       string
    Email      string
}

// ❌ Bad - technical events
type UserTableUpdatedEvent struct {
    Table  string
    Action string
}
```

## 🚀 Next Steps

1. **Build Your First Handler**: [Commands, Queries & Events](./commands-queries-events.md)
2. **See Practical Examples**: [Basic Examples](./basic-examples.md)
3. **Add Validation**: [Handlers & Validators](./handlers-validators.md)
4. **Learn Auto-Registration**: [Auto-Registration Guide](./auto-registration.md)

---

**Ready to implement CQRS? Start with [Commands, Queries & Events](./commands-queries-events.md)! 🚀** 