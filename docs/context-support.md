# Context Support

Full `context.Context` support in go-cqrs for request tracing, timeouts, cancellation, and request-scoped data.

## 📚 Table of Contents

- [Why Context Matters](#why-context-matters)
- [Request Tracing](#request-tracing)
- [Timeouts & Cancellation](#timeouts--cancellation)
- [Request-Scoped Data](#request-scoped-data)
- [Context Patterns](#context-patterns)
- [Best Practices](#best-practices)
- [Production Examples](#production-examples)

## 🎯 Why Context Matters

Context provides crucial capabilities for production applications:

- **Request Tracing**: Track requests across service boundaries
- **Timeouts**: Prevent operations from running indefinitely
- **Cancellation**: Stop processing when clients disconnect
- **Request-Scoped Data**: Share data across handler calls
- **Observability**: Add logging, metrics, and monitoring

```go
// Every handler receives context
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Context flows through all operations
    if err := h.userRepo.Save(ctx, user); err != nil {
        return err
    }
    
    return h.emailService.SendWelcome(ctx, user.Email)
}
```

## 🔍 Request Tracing

### Correlation IDs

Track requests across services and handlers:

```go
type ContextKey string

const (
    CorrelationIDKey ContextKey = "correlation_id"
    UserIDKey       ContextKey = "user_id"
    RequestIDKey    ContextKey = "request_id"
)

// Middleware adds correlation ID
func AddCorrelationID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        correlationID := r.Header.Get("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }
        
        ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
        r = r.WithContext(ctx)
        
        w.Header().Set("X-Correlation-ID", correlationID)
        next.ServeHTTP(w, r)
    })
}

// Handler uses correlation ID
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    correlationID := ctx.Value(CorrelationIDKey).(string)
    
    h.logger.Info("Creating user", map[string]interface{}{
        "correlation_id": correlationID,
        "email":         cmd.Email,
        "operation":     "create_user",
    })
    
    user := &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        h.logger.Error("Failed to save user", map[string]interface{}{
            "correlation_id": correlationID,
            "error":         err.Error(),
        })
        return err
    }
    
    return nil
}
```

### Structured Logging

```go
type LoggingHandler struct {
    baseHandler CommandHandler
    logger      Logger
}

func (h *LoggingHandler) Handle(ctx context.Context, cmd Command) error {
    start := time.Now()
    correlationID := getCorrelationID(ctx)
    
    h.logger.Info("Command started", map[string]interface{}{
        "correlation_id": correlationID,
        "command_type":   reflect.TypeOf(cmd).Name(),
        "timestamp":      start,
    })
    
    err := h.baseHandler.Handle(ctx, cmd)
    
    duration := time.Since(start)
    
    if err != nil {
        h.logger.Error("Command failed", map[string]interface{}{
            "correlation_id": correlationID,
            "command_type":   reflect.TypeOf(cmd).Name(),
            "duration_ms":    duration.Milliseconds(),
            "error":         err.Error(),
        })
    } else {
        h.logger.Info("Command completed", map[string]interface{}{
            "correlation_id": correlationID,
            "command_type":   reflect.TypeOf(cmd).Name(),
            "duration_ms":    duration.Milliseconds(),
        })
    }
    
    return err
}
```

## ⏰ Timeouts & Cancellation

### Command Timeouts

```go
// HTTP handler with timeout
func CreateUserHTTPHandler(w http.ResponseWriter, r *http.Request) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    
    var cmd CreateUserCommand
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Execute command with timeout
    err := cqrs.ExecuteCommand(ctx, &cmd)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Request timeout", http.StatusRequestTimeout)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
}

// Handler respects timeout
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Check if context is already cancelled
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Long-running operation that checks context
    return h.processUserCreation(ctx, cmd)
}

func (h *CreateUserHandler) processUserCreation(ctx context.Context, cmd *CreateUserCommand) error {
    // Simulate some processing time
    for i := 0; i < 10; i++ {
        // Check for cancellation periodically
        select {
        case <-ctx.Done():
            h.logger.Info("User creation cancelled")
            return ctx.Err()
        case <-time.After(500 * time.Millisecond):
            // Continue processing
        }
    }
    
    // Save user (repository should also respect context)
    return h.userRepo.Save(ctx, &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    })
}
```

### Query Timeouts

```go
func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    // Set a shorter timeout for database queries
    dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    user, err := h.userRepo.GetByID(dbCtx, q.UserID)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("database query timeout: %w", err)
        }
        return nil, err
    }
    
    return user, nil
}
```

### Graceful Cancellation

```go
func (h *ProcessOrderHandler) Handle(ctx context.Context, cmd *ProcessOrderCommand) error {
    // Step 1: Validate inventory (fast operation)
    if err := h.validateInventory(ctx, cmd.Items); err != nil {
        return err
    }
    
    // Check cancellation before expensive operations
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Step 2: Process payment (can be slow)
    paymentResult, err := h.processPaymentWithCancellation(ctx, cmd.Payment)
    if err != nil {
        return err
    }
    
    // Step 3: Save order (final step)
    return h.saveOrder(ctx, cmd, paymentResult)
}

func (h *ProcessOrderHandler) processPaymentWithCancellation(ctx context.Context, payment Payment) (*PaymentResult, error) {
    // Create a channel for payment result
    resultChan := make(chan *PaymentResult)
    errorChan := make(chan error)
    
    // Start payment processing in goroutine
    go func() {
        result, err := h.paymentService.ProcessPayment(payment)
        if err != nil {
            errorChan <- err
            return
        }
        resultChan <- result
    }()
    
    // Wait for either completion or cancellation
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errorChan:
        return nil, err
    case <-ctx.Done():
        // Payment was cancelled - we might need to handle cleanup
        h.logger.Warn("Payment processing cancelled", map[string]interface{}{
            "order_id": payment.OrderID,
            "reason":   ctx.Err().Error(),
        })
        return nil, ctx.Err()
    }
}
```

## 📦 Request-Scoped Data

### User Authentication

```go
type AuthenticatedUser struct {
    ID    int
    Name  string
    Email string
    Roles []string
}

const AuthUserKey ContextKey = "auth_user"

// Middleware adds authenticated user to context
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        user, err := validateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        ctx := context.WithValue(r.Context(), AuthUserKey, user)
        r = r.WithContext(ctx)
        
        next.ServeHTTP(w, r)
    })
}

// Handler uses authenticated user
func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    authUser := ctx.Value(AuthUserKey).(*AuthenticatedUser)
    
    // Verify user can create orders for this customer
    if !h.canUserCreateOrderForCustomer(authUser, cmd.CustomerID) {
        return errors.New("unauthorized: cannot create order for this customer")
    }
    
    h.logger.Info("Creating order", map[string]interface{}{
        "user_id":     authUser.ID,
        "customer_id": cmd.CustomerID,
        "order_total": cmd.Total,
    })
    
    // Continue with order creation
    return h.processOrder(ctx, cmd)
}
```

### Tenant Isolation

```go
const TenantIDKey ContextKey = "tenant_id"

// Multi-tenant handler
func (h *GetUsersHandler) Handle(ctx context.Context, q GetUsersQuery) (*UserListResult, error) {
    tenantID := ctx.Value(TenantIDKey).(string)
    
    // All queries are scoped to the tenant
    users, err := h.userRepo.GetByTenant(ctx, tenantID, q.Filter)
    if err != nil {
        return nil, err
    }
    
    return &UserListResult{
        Users:    users,
        TenantID: tenantID,
    }, nil
}

// Repository implementation respects tenant context
func (r *PostgreSQLUserRepository) GetByTenant(ctx context.Context, tenantID string, filter UserFilter) ([]*User, error) {
    query := `
        SELECT id, name, email, created_at 
        FROM users 
        WHERE tenant_id = $1
    `
    
    // Add additional filters...
    
    return r.queryWithContext(ctx, query, tenantID)
}
```

### Request Metadata

```go
type RequestMetadata struct {
    UserAgent    string
    IPAddress    string
    APIVersion   string
    ClientID     string
    RequestTime  time.Time
}

const RequestMetadataKey ContextKey = "request_metadata"

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    metadata := ctx.Value(RequestMetadataKey).(*RequestMetadata)
    
    user := &User{
        Name:      cmd.Name,
        Email:     cmd.Email,
        CreatedAt: metadata.RequestTime,
        UserAgent: metadata.UserAgent,
        IPAddress: metadata.IPAddress,
    }
    
    return h.userRepo.Save(ctx, user)
}
```

## 🔄 Context Patterns

### Context Enrichment

```go
func EnrichContext(ctx context.Context, userID int) context.Context {
    // Add multiple values to context
    ctx = context.WithValue(ctx, UserIDKey, userID)
    ctx = context.WithValue(ctx, RequestIDKey, uuid.New().String())
    ctx = context.WithValue(ctx, RequestTimeKey, time.Now())
    
    return ctx
}

// Usage
func CreateUserEndpoint(w http.ResponseWriter, r *http.Request) {
    ctx := EnrichContext(r.Context(), getCurrentUserID(r))
    
    var cmd CreateUserCommand
    json.NewDecoder(r.Body).Decode(&cmd)
    
    err := cqrs.ExecuteCommand(ctx, &cmd)
    // Handle response...
}
```

### Context Propagation

```go
func (h *OrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    // Context automatically flows to all subsequent operations
    
    // 1. Validate customer (context flows to repository)
    customer, err := h.customerRepo.GetByID(ctx, cmd.CustomerID)
    if err != nil {
        return err
    }
    
    // 2. Check inventory (context flows to external service)
    available, err := h.inventoryService.CheckAvailability(ctx, cmd.Items)
    if err != nil {
        return err
    }
    
    // 3. Process payment (context flows with correlation ID)
    payment, err := h.paymentService.ProcessPayment(ctx, cmd.Payment)
    if err != nil {
        return err
    }
    
    // 4. Publish event (context flows to event handlers)
    return cqrs.PublishEvent(ctx, OrderCreatedEvent{
        OrderID:    generateID(),
        CustomerID: customer.ID,
        Total:      cmd.Total,
    })
}
```

### Context with Deadlines

```go
func (h *ComplexProcessHandler) Handle(ctx context.Context, cmd *ComplexProcessCommand) error {
    // Different timeouts for different operations
    
    // Fast validation - 1 second
    validationCtx, cancel1 := context.WithTimeout(ctx, 1*time.Second)
    defer cancel1()
    
    if err := h.validateCommand(validationCtx, cmd); err != nil {
        return err
    }
    
    // Medium operation - 10 seconds
    processCtx, cancel2 := context.WithTimeout(ctx, 10*time.Second)
    defer cancel2()
    
    if err := h.processData(processCtx, cmd); err != nil {
        return err
    }
    
    // Slow operation - 30 seconds
    saveCtx, cancel3 := context.WithTimeout(ctx, 30*time.Second)
    defer cancel3()
    
    return h.saveResults(saveCtx, cmd)
}
```

## 🎓 Best Practices

### 1. Always Check Context

```go
// ✅ Good - check context in long operations
func (h *ProcessDataHandler) Handle(ctx context.Context, cmd *ProcessDataCommand) error {
    for _, item := range cmd.Items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := h.processItem(ctx, item); err != nil {
                return err
            }
        }
    }
    return nil
}

// ❌ Bad - ignore context
func (h *ProcessDataHandler) Handle(ctx context.Context, cmd *ProcessDataCommand) error {
    for _, item := range cmd.Items {
        h.processItem(ctx, item) // No cancellation check
    }
    return nil
}
```

### 2. Use Context Values Sparingly

```go
// ✅ Good - use for request-scoped data
type RequestContext struct {
    UserID        int
    CorrelationID string
    RequestTime   time.Time
}

const RequestContextKey ContextKey = "request_context"

// ❌ Bad - overuse context values
ctx = context.WithValue(ctx, "user_name", userName)
ctx = context.WithValue(ctx, "user_email", userEmail)
ctx = context.WithValue(ctx, "user_role", userRole)
// ... too many values
```

### 3. Graceful Degradation

```go
func (h *NotificationHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    // Don't fail critical operations for non-critical context timeouts
    
    notificationCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    err := h.emailService.SendWelcome(notificationCtx, e.Email)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            // Log timeout but don't fail the event
            h.logger.Warn("Welcome email timeout", map[string]interface{}{
                "user_id": e.UserID,
                "email":   e.Email,
            })
            // Maybe queue for retry
            h.queueEmailForRetry(e.Email, "welcome")
            return nil
        }
        return err
    }
    
    return nil
}
```

## 🏭 Production Examples

### HTTP API Integration

```go
func CreateUserAPI(w http.ResponseWriter, r *http.Request) {
    // Create request context with timeout
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    
    // Add request metadata
    ctx = addRequestMetadata(ctx, r)
    
    // Parse command
    var cmd CreateUserCommand
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Execute with context
    err := cqrs.ExecuteCommand(ctx, &cmd)
    
    // Handle context-specific errors
    if err != nil {
        switch {
        case errors.Is(err, context.DeadlineExceeded):
            http.Error(w, "Request timeout", http.StatusRequestTimeout)
        case errors.Is(err, context.Canceled):
            http.Error(w, "Request cancelled", http.StatusGone)
        default:
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }
    
    w.WriteHeader(http.StatusCreated)
}

func addRequestMetadata(ctx context.Context, r *http.Request) context.Context {
    metadata := &RequestMetadata{
        UserAgent:   r.UserAgent(),
        IPAddress:   getClientIP(r),
        APIVersion:  r.Header.Get("API-Version"),
        RequestTime: time.Now(),
    }
    
    return context.WithValue(ctx, RequestMetadataKey, metadata)
}
```

### Background Job Processing

```go
func ProcessBackgroundJob(job Job) {
    // Create context with job timeout
    ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
    defer cancel()
    
    // Add job metadata
    ctx = context.WithValue(ctx, "job_id", job.ID)
    ctx = context.WithValue(ctx, "job_type", job.Type)
    
    // Execute job command
    err := cqrs.ExecuteCommand(ctx, job.Command)
    
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            // Handle job timeout
            markJobAsTimedOut(job.ID)
        } else {
            // Handle other errors
            markJobAsFailed(job.ID, err)
        }
        return
    }
    
    markJobAsCompleted(job.ID)
}
```

## 🚀 Next Steps

1. **Learn Dependency Injection**: [Dependency Injection](./dependency-injection.md)
2. **Add Auto-Registration**: [Auto-Registration Guide](./auto-registration.md)
3. **Use Decorators**: [Decorators & Middleware](./decorators.md)
4. **Testing with Context**: [Testing Strategies](./testing.md)

---

**Ready for advanced features? Continue with [Dependency Injection](./dependency-injection.md)! 🚀** 