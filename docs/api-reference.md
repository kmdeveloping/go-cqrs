# API Reference

Complete API documentation for go-cqrs library.

## 📚 **Core Interfaces**

### **Command Interface**

```go
package command

import "context"

type ICommand any

// ICommandHandler interface handles commands of type T that implements ICommand
// Commands are always passed as pointers to handlers for consistency
type ICommandHandler[T ICommand] interface {
    Handle(ctx context.Context, cmd *T) error
}

type Base struct{}
type BaseWithResult struct {
    Result any
}
```

**Usage:**
```go
type CreateUserCommand struct {
    command.Base
    Name  string
    Email string
}

type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Implementation
    return nil
}
```

### **Query Interface**

```go
package query

import "context"

type IQuery any

type IQueryHandler[T IQuery, R any] interface {
    Handle(ctx context.Context, query T) (R, error)
}

type Base struct{}
```

**Usage:**
```go
type GetUserQuery struct {
    query.Base
    ID int
}

type GetUserHandler struct{}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    // Implementation
    return &User{}, nil
}
```

### **Event Interface**

```go
package event

import (
    "context"
    "time"
    "github.com/google/uuid"
)

type IEvent any

type IEventHandler[T IEvent] interface {
    Handle(ctx context.Context, event T) error
}

type Base struct {
    ExecutionTime  time.Time
    CorrelationUid uuid.UUID
    MetaData       string
}
```

**Usage:**
```go
type UserCreatedEvent struct {
    event.Base
    UserID int
    Name   string
}

type UserCreatedHandler struct{}

func (h *UserCreatedHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    // Implementation
    return nil
}
```

## 🔧 **Core CQRS Manager**

### **Manager Setup**

```go
func NewCqrsManager() *CqrsManager
func SetManager(manager *CqrsManager) error
func ResetManager() // For testing
```

**Example:**
```go
manager := cqrs.NewCqrsManager()
manager.AddLoggingDecorator()
manager.AddMetricsDecorator()

if err := cqrs.SetManager(manager); err != nil {
    log.Fatal(err)
}
```

### **Handler Registration**

```go
// Manual registration
func RegisterCommandHandler[T ICommand](handler ICommandHandler[T])
func RegisterQueryHandler[T IQuery, R any](handler IQueryHandler[T, R])
func RegisterEventHandler[T IEvent](handler IEventHandler[T])

// Auto-registration
func AutoRegisterHandlers(handlers ...any) *AutoRegistryResult
func AutoRegisterWithDependencies(container Container, handlers ...any) *AutoRegistryResult
```

**Example:**
```go
// Manual registration
cqrs.RegisterCommandHandler(&CreateUserHandler{})
cqrs.RegisterQueryHandler(&GetUserHandler{})
cqrs.RegisterEventHandler(&UserCreatedHandler{})

// Auto-registration
result := cqrs.AutoRegisterHandlers(
    &CreateUserHandler{},
    &GetUserHandler{},
    &UserCreatedHandler{},
)
```

### **Command Execution**

```go
func ExecuteCommand(ctx context.Context, cmd ICommand) error
func ExecuteCommandWithResult[T any](ctx context.Context, cmd ICommand) (T, error)
```

**Example:**
```go
ctx := context.Background()

// Simple command
err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{
    Name:  "John Doe",
    Email: "john@example.com",
})

// Command with result
result, err := cqrs.ExecuteCommandWithResult[int](ctx, &CreateUserCommand{
    Name:  "John Doe",
    Email: "john@example.com",
})
```

### **Query Execution**

```go
func ExecuteQuery[T IQuery, R any](ctx context.Context, query T) (R, error)
```

**Example:**
```go
user, err := cqrs.ExecuteQuery[GetUserQuery, *User](ctx, GetUserQuery{
    ID: 1,
})
```

### **Event Publishing**

```go
func PublishEvent(ctx context.Context, event IEvent) error
func PublishEvents(ctx context.Context, events ...IEvent) error
```

**Example:**
```go
err := cqrs.PublishEvent(ctx, UserCreatedEvent{
    UserID: 1,
    Name:   "John Doe",
})

// Multiple events
err := cqrs.PublishEvents(ctx,
    UserCreatedEvent{UserID: 1, Name: "John"},
    WelcomeEmailEvent{UserID: 1, Email: "john@example.com"},
)
```

## 🏗️ **Dependency Injection**

### **Container Interface**

```go
type Container interface {
    Register(key string, instance any)
    Resolve(key string) (any, error)
    ResolveType(t reflect.Type) (any, error)
}

func NewSimpleContainer() Container
```

### **Registration Functions**

```go
func Register[T any](container Container, instance T)
func Resolve[T any](container Container) (T, error)
```

**Example:**
```go
container := cqrs.NewSimpleContainer()

// Register dependencies
cqrs.Register[UserRepository](container, &PostgreSQLUserRepository{})
cqrs.Register[EmailService](container, &SMTPEmailService{})

// Auto-register with DI
result := cqrs.AutoRegisterWithDependencies(container,
    &CreateUserHandler{}, // Dependencies auto-injected
    &SendWelcomeEmailHandler{},
)
```

## 🎭 **Decorators**

### **Built-in Decorators**

```go
func (m *CqrsManager) AddLoggingDecorator()
func (m *CqrsManager) AddMetricsDecorator()
func (m *CqrsManager) AddDecorator(decorator HandlerDecorator)
```

### **Custom Decorators**

```go
type HandlerDecorator func(next IHandlerDecorator) IHandlerDecorator

type IHandlerDecorator interface {
    Handle(ctx context.Context, message any) (any, error)
}

type HandlerDecoratorFunc func(ctx context.Context, message any) (any, error)

func (f HandlerDecoratorFunc) Handle(ctx context.Context, message any) (any, error) {
    return f(ctx, message)
}
```

**Example:**
```go
func TimingDecorator() decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            start := time.Now()
            result, err := next.Handle(ctx, message)
            duration := time.Since(start)
            
            log.Printf("Handler executed in %v", duration)
            return result, err
        })
    }
}

// Use decorator
manager.AddDecorator(TimingDecorator())
```

## ✅ **Validation**

### **Validator Interface**

```go
type IValidator[T any] interface {
    Validate(ctx context.Context, obj T) error
}
```

### **Registration and Usage**

```go
func RegisterValidator[T any](validator IValidator[T])
```

**Example:**
```go
type CreateUserValidator struct{}

func (v *CreateUserValidator) Validate(ctx context.Context, cmd *CreateUserCommand) error {
    if cmd.Name == "" {
        return errors.New("name is required")
    }
    if cmd.Email == "" {
        return errors.New("email is required")
    }
    return nil
}

// Register validator
cqrs.RegisterValidator(&CreateUserValidator{})
```

## 🔍 **Monitoring & Health Checks**

### **Metrics Functions**

```go
func GetMetrics() map[string]interface{}
func GetCommandCount() int64
func GetQueryCount() int64
func GetEventCount() int64
func GetHandlerCounts() (commands, queries, events, validators int)
```

### **Health Check Functions**

```go
type HealthCheck struct {
    Status  string            `json:"status"`
    Checks  map[string]string `json:"checks"`
    Details map[string]any    `json:"details"`
}

func GetHealthCheck() HealthCheck
```

**Example:**
```go
// HTTP health endpoint
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    health := cqrs.GetHealthCheck()
    if health.Status != "healthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    json.NewEncoder(w).Encode(health)
})
```

## 🏭 **Production Configuration**

### **Production Config**

```go
type ProductionConfig struct {
    MaxRegistrationTime     time.Duration
    EnableRetry            bool
    MaxRetries             int
    RetryDelay             time.Duration
    EnableHealthChecks     bool
    HealthCheckInterval    time.Duration
    EnableMetrics          bool
    MetricsInterval        time.Duration
    EnableCircuitBreaker   bool
    CircuitBreakerThreshold int
    TimeoutDuration        time.Duration
    EnableVerboseLogging   bool
}

func DefaultProductionConfig() *ProductionConfig
func NewProductionAutoRegistry(manager *CqrsManager, config *ProductionConfig) *ProductionAutoRegistry
```

### **Production Registry**

```go
type ProductionAutoRegistry struct {
    // internal fields
}

func (r *ProductionAutoRegistry) RegisterHandlerInstancesWithConfig(handlers ...any) *AutoRegistryResult
```

**Example:**
```go
config := cqrs.DefaultProductionConfig()
config.EnableRetry = true
config.MaxRetries = 3

prodRegistry := cqrs.NewProductionAutoRegistry(manager, config)
result := prodRegistry.RegisterHandlerInstancesWithConfig(
    &CreateUserHandler{},
    &GetUserHandler{},
    &UserCreatedHandler{},
)
```

## 📊 **Auto-Registry Results**

### **Result Types**

```go
type AutoRegistryResult struct {
    RegisteredHandlers   int
    RegisteredValidators int
    Errors              []error
    HandlerTypes        []string
    ValidatorTypes      []string
    Duration            time.Duration
}

type RegistrationError struct {
    HandlerType string
    Error       error
}

func (e *RegistrationError) Error() string
```

**Example:**
```go
result := cqrs.AutoRegisterHandlers(handlers...)

if len(result.Errors) > 0 {
    for _, err := range result.Errors {
        log.Printf("Registration error: %v", err)
    }
    return fmt.Errorf("failed to register %d handlers", len(result.Errors))
}

log.Printf("Successfully registered %d handlers, %d validators in %v",
    result.RegisteredHandlers,
    result.RegisteredValidators, 
    result.Duration)
```

## 🧪 **Testing Utilities**

### **Test Functions**

```go
func ResetManager() // Reset global state for testing
func NewTestContainer() Container // Container for testing
```

### **Mock Helpers**

```go
type MockCommandHandler[T ICommand] struct {
    HandleFunc func(ctx context.Context, cmd *T) error
    CallCount  int
}

func (m *MockCommandHandler[T]) Handle(ctx context.Context, cmd *T) error {
    m.CallCount++
    if m.HandleFunc != nil {
        return m.HandleFunc(ctx, cmd)
    }
    return nil
}
```

**Example:**
```go
func TestUserCreation(t *testing.T) {
    // Clean state
    defer cqrs.ResetManager()
    
    // Setup test environment
    manager := cqrs.NewCqrsManager()
    cqrs.SetManager(manager)
    
    // Register mock handler
    mockHandler := &MockCommandHandler[CreateUserCommand]{
        HandleFunc: func(ctx context.Context, cmd *CreateUserCommand) error {
            return nil
        },
    }
    cqrs.RegisterCommandHandler(mockHandler)
    
    // Test execution
    err := cqrs.ExecuteCommand(context.Background(), &CreateUserCommand{
        Name: "Test User",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 1, mockHandler.CallCount)
}
```

## 🚨 **Error Types**

### **Common Errors**

```go
var (
    ErrHandlerNotFound     = errors.New("handler not found")
    ErrValidationFailed    = errors.New("validation failed")
    ErrManagerNotSet       = errors.New("CQRS manager not set")
    ErrRegistrationFailed  = errors.New("handler registration failed")
    ErrTimeout            = errors.New("operation timed out")
    ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)
```

### **Error Handling Patterns**

```go
func IsHandlerError(err error) bool
func IsValidationError(err error) bool
func IsTimeoutError(err error) bool
```

**Example:**
```go
err := cqrs.ExecuteCommand(ctx, cmd)
if err != nil {
    switch {
    case cqrs.IsHandlerError(err):
        log.Printf("Handler error: %v", err)
    case cqrs.IsValidationError(err):
        log.Printf("Validation error: %v", err)
    case cqrs.IsTimeoutError(err):
        log.Printf("Timeout error: %v", err)
    default:
        log.Printf("Unknown error: %v", err)
    }
}
```

## 🔗 **Context Support**

### **Context Keys**

```go
type ContextKey string

const (
    RequestIDKey    ContextKey = "requestID"
    UserIDKey      ContextKey = "userID"
    CorrelationKey ContextKey = "correlationID"
)
```

### **Context Utilities**

```go
func WithRequestID(ctx context.Context, requestID string) context.Context
func GetRequestID(ctx context.Context) (string, bool)
func WithUserID(ctx context.Context, userID string) context.Context
func GetUserID(ctx context.Context) (string, bool)
```

**Example:**
```go
ctx := context.Background()
ctx = cqrs.WithRequestID(ctx, "req-123")
ctx = cqrs.WithUserID(ctx, "user-456")

err := cqrs.ExecuteCommand(ctx, &CreateUserCommand{
    Name: "John Doe",
})

// In handler
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    requestID, _ := cqrs.GetRequestID(ctx)
    userID, _ := cqrs.GetUserID(ctx)
    
    log.Printf("Handling command for request %s by user %s", requestID, userID)
    return nil
}
```

---

## 📚 **Quick Reference**

| Category | Key Functions |
|----------|---------------|
| **Setup** | `NewCqrsManager()`, `SetManager()` |
| **Registration** | `RegisterCommandHandler()`, `AutoRegisterHandlers()` |
| **Execution** | `ExecuteCommand()`, `ExecuteQuery()`, `PublishEvent()` |
| **Decorators** | `AddLoggingDecorator()`, `AddMetricsDecorator()` |
| **Testing** | `ResetManager()`, `NewTestContainer()` |
| **Monitoring** | `GetMetrics()`, `GetHealthCheck()` |

For more examples and patterns, see the [Examples Guide](./examples.md) and [Integration Guide](./integrations.md). 