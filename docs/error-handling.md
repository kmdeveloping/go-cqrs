# Error Handling

Comprehensive error patterns and troubleshooting strategies for go-cqrs applications.

## 📚 Table of Contents

- [Error Types](#error-types)
- [Error Patterns](#error-patterns)
- [Error Propagation](#error-propagation)
- [Recovery Strategies](#recovery-strategies)
- [Troubleshooting Guide](#troubleshooting-guide)
- [Custom Error Types](#custom-error-types)
- [Best Practices](#best-practices)

## 🚨 Error Types

go-cqrs applications deal with several categories of errors:

### System Errors
- Network failures
- Database connection issues
- Service unavailability
- Resource exhaustion

### Domain Errors
- Business rule violations
- Validation failures
- Authorization errors
- Not found errors

### Technical Errors
- Serialization failures
- Type conversion errors
- Configuration issues
- Programming errors

```go
// Example error categorization
func CategorizeError(err error) string {
    switch {
    case isNetworkError(err):
        return "network"
    case isDatabaseError(err):
        return "database"
    case isValidationError(err):
        return "validation"
    case isAuthorizationError(err):
        return "authorization"
    case isBusinessRuleError(err):
        return "business"
    default:
        return "unknown"
    }
}
```

## 🎯 Error Patterns

### Validation Errors

```go
type ValidationError struct {
    Field   string                 `json:"field"`
    Message string                 `json:"message"`
    Value   interface{}            `json:"value,omitempty"`
    Code    string                 `json:"code"`
}

type ValidationErrors struct {
    Command string            `json:"command"`
    Errors  []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
    if len(e.Errors) == 1 {
        return fmt.Sprintf("validation failed for %s: %s", e.Command, e.Errors[0].Message)
    }
    return fmt.Sprintf("validation failed for %s: %d errors", e.Command, len(e.Errors))
}

func (e *ValidationErrors) Add(field, message, code string, value interface{}) {
    e.Errors = append(e.Errors, ValidationError{
        Field:   field,
        Message: message,
        Value:   value,
        Code:    code,
    })
}

// Usage in validator
func (v *CreateUserValidator) Validate(ctx context.Context, cmd *CreateUserCommand) error {
    validationErrors := &ValidationErrors{Command: "CreateUserCommand"}
    
    if strings.TrimSpace(cmd.Name) == "" {
        validationErrors.Add("name", "Name is required", "REQUIRED", cmd.Name)
    }
    
    if len(cmd.Name) < 2 {
        validationErrors.Add("name", "Name must be at least 2 characters", "MIN_LENGTH", cmd.Name)
    }
    
    if !isValidEmail(cmd.Email) {
        validationErrors.Add("email", "Invalid email format", "INVALID_FORMAT", cmd.Email)
    }
    
    // Check email uniqueness
    exists, err := v.userRepo.ExistsByEmail(ctx, cmd.Email)
    if err != nil {
        return fmt.Errorf("failed to check email uniqueness: %w", err)
    }
    if exists {
        validationErrors.Add("email", "Email already exists", "ALREADY_EXISTS", cmd.Email)
    }
    
    if len(validationErrors.Errors) > 0 {
        return validationErrors
    }
    
    return nil
}
```

### Business Rule Errors

```go
type BusinessRuleError struct {
    Rule        string      `json:"rule"`
    Message     string      `json:"message"`
    Context     interface{} `json:"context,omitempty"`
    Suggestions []string    `json:"suggestions,omitempty"`
}

func (e *BusinessRuleError) Error() string {
    return fmt.Sprintf("business rule violation [%s]: %s", e.Rule, e.Message)
}

func NewBusinessRuleError(rule, message string) *BusinessRuleError {
    return &BusinessRuleError{
        Rule:    rule,
        Message: message,
    }
}

func (e *BusinessRuleError) WithContext(context interface{}) *BusinessRuleError {
    e.Context = context
    return e
}

func (e *BusinessRuleError) WithSuggestions(suggestions ...string) *BusinessRuleError {
    e.Suggestions = suggestions
    return e
}

// Usage in business logic
func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd *PlaceOrderCommand) error {
    customer, err := h.customerRepo.GetByID(ctx, cmd.CustomerID)
    if err != nil {
        return fmt.Errorf("failed to get customer: %w", err)
    }
    
    if customer.Status != "Active" {
        return NewBusinessRuleError("customer_must_be_active", 
            "Orders can only be placed by active customers").
            WithContext(map[string]interface{}{
                "customer_id": customer.ID,
                "status":      customer.Status,
            }).
            WithSuggestions("Contact customer support to activate account")
    }
    
    if cmd.Total.LessThan(decimal.NewFromFloat(10.00)) {
        return NewBusinessRuleError("minimum_order_amount", 
            "Order total must be at least $10.00").
            WithContext(map[string]interface{}{
                "current_total": cmd.Total,
                "minimum":       10.00,
            })
    }
    
    return h.processOrder(ctx, cmd)
}
```

### Not Found Errors

```go
type NotFoundError struct {
    Resource string      `json:"resource"`
    ID       string      `json:"id"`
    Context  interface{} `json:"context,omitempty"`
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

func NewNotFoundError(resource, id string) *NotFoundError {
    return &NotFoundError{
        Resource: resource,
        ID:       id,
    }
}

func (e *NotFoundError) WithContext(context interface{}) *NotFoundError {
    e.Context = context
    return e
}

// Usage in repository
func (r *PostgreSQLUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    user := &User{}
    
    err := r.db.QueryRowContext(ctx, 
        "SELECT id, name, email, created_at FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, NewNotFoundError("User", strconv.Itoa(id))
        }
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    
    return user, nil
}
```

### Authorization Errors

```go
type AuthorizationError struct {
    UserID      int         `json:"user_id"`
    Action      string      `json:"action"`
    Resource    string      `json:"resource"`
    Reason      string      `json:"reason"`
    Required    []string    `json:"required_permissions,omitempty"`
    Suggestions []string    `json:"suggestions,omitempty"`
}

func (e *AuthorizationError) Error() string {
    return fmt.Sprintf("authorization failed for user %d: %s", e.UserID, e.Reason)
}

func NewAuthorizationError(userID int, action, resource, reason string) *AuthorizationError {
    return &AuthorizationError{
        UserID:   userID,
        Action:   action,
        Resource: resource,
        Reason:   reason,
    }
}

// Usage in authorization
func (a *RoleBasedAuthorizer) Authorize(ctx context.Context, cmd Command) error {
    user := getAuthenticatedUser(ctx)
    if user == nil {
        return NewAuthorizationError(0, "execute", reflect.TypeOf(cmd).Name(), 
            "user not authenticated")
    }
    
    commandType := reflect.TypeOf(cmd).Name()
    requiredRoles := a.permissions[commandType]
    
    if len(requiredRoles) == 0 {
        return nil
    }
    
    for _, role := range requiredRoles {
        if user.HasRole(role) {
            return nil
        }
    }
    
    return NewAuthorizationError(user.ID, "execute", commandType, 
        "insufficient permissions").
        WithRequired(requiredRoles).
        WithSuggestions("Contact administrator to request additional permissions")
}

func (e *AuthorizationError) WithRequired(permissions []string) *AuthorizationError {
    e.Required = permissions
    return e
}

func (e *AuthorizationError) WithSuggestions(suggestions []string) *AuthorizationError {
    e.Suggestions = suggestions
    return e
}
```

## 🔄 Error Propagation

### Wrapping Errors

```go
// ✅ Good - preserve error context
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    user := &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        return fmt.Errorf("failed to save user %s: %w", cmd.Email, err)
    }
    
    if err := h.emailService.SendWelcome(ctx, user.Email, user.Name); err != nil {
        // Log warning but don't fail the command
        h.logger.Warn("Failed to send welcome email", map[string]interface{}{
            "user_id": user.ID,
            "email":   user.Email,
            "error":   err.Error(),
        })
    }
    
    return nil
}
```

### Error Context Enhancement

```go
type ErrorContext struct {
    Operation   string                 `json:"operation"`
    Timestamp   time.Time              `json:"timestamp"`
    UserID      int                    `json:"user_id,omitempty"`
    RequestID   string                 `json:"request_id,omitempty"`
    Command     string                 `json:"command,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ContextualError struct {
    Err     error        `json:"error"`
    Context ErrorContext `json:"context"`
}

func (e *ContextualError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Context.Operation, e.Err.Error())
}

func (e *ContextualError) Unwrap() error {
    return e.Err
}

func NewContextualError(err error, operation string) *ContextualError {
    return &ContextualError{
        Err: err,
        Context: ErrorContext{
            Operation: operation,
            Timestamp: time.Now(),
        },
    }
}

func (e *ContextualError) WithUser(userID int) *ContextualError {
    e.Context.UserID = userID
    return e
}

func (e *ContextualError) WithRequest(requestID string) *ContextualError {
    e.Context.RequestID = requestID
    return e
}

func (e *ContextualError) WithCommand(command string) *ContextualError {
    e.Context.Command = command
    return e
}

func (e *ContextualError) WithMetadata(key string, value interface{}) *ContextualError {
    if e.Context.Metadata == nil {
        e.Context.Metadata = make(map[string]interface{})
    }
    e.Context.Metadata[key] = value
    return e
}

// Usage in decorator
func (d *LoggingDecorator) Handle(ctx context.Context, cmd Command) error {
    err := d.handler.Handle(ctx, cmd)
    
    if err != nil {
        // Enhance error with context
        contextualErr := NewContextualError(err, "command_execution").
            WithUser(getUserID(ctx)).
            WithRequest(getRequestID(ctx)).
            WithCommand(reflect.TypeOf(cmd).Name()).
            WithMetadata("correlation_id", getCorrelationID(ctx))
        
        return contextualErr
    }
    
    return nil
}
```

## 🛠️ Recovery Strategies

### Graceful Degradation

```go
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Core functionality - must succeed
    user := &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    if err := h.userRepo.Save(ctx, user); err != nil {
        return fmt.Errorf("failed to save user: %w", err)
    }
    
    // Secondary functionality - can fail gracefully
    h.performSecondaryOperations(ctx, user)
    
    return nil
}

func (h *CreateUserHandler) performSecondaryOperations(ctx context.Context, user *User) {
    // Send welcome email (non-critical)
    if err := h.emailService.SendWelcome(ctx, user.Email, user.Name); err != nil {
        h.logger.Warn("Failed to send welcome email", map[string]interface{}{
            "user_id": user.ID,
            "error":   err.Error(),
        })
        
        // Queue for retry
        h.emailQueue.Enqueue(WelcomeEmailJob{
            UserID: user.ID,
            Email:  user.Email,
            Name:   user.Name,
        })
    }
    
    // Update analytics (non-critical)
    if err := h.analytics.TrackUserRegistration(ctx, user.ID); err != nil {
        h.logger.Warn("Failed to track user registration", map[string]interface{}{
            "user_id": user.ID,
            "error":   err.Error(),
        })
    }
    
    // Send notification (non-critical)
    go func() {
        if err := h.notificationService.NotifyUserRegistration(user); err != nil {
            h.logger.Warn("Failed to send registration notification", map[string]interface{}{
                "user_id": user.ID,
                "error":   err.Error(),
            })
        }
    }()
}
```

### Compensating Actions

```go
func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd *PlaceOrderCommand) error {
    var compensations []func() error
    
    // Reserve inventory
    reservationID, err := h.inventoryService.Reserve(ctx, cmd.Items)
    if err != nil {
        return fmt.Errorf("failed to reserve inventory: %w", err)
    }
    compensations = append(compensations, func() error {
        return h.inventoryService.CancelReservation(ctx, reservationID)
    })
    
    // Process payment
    paymentID, err := h.paymentService.ProcessPayment(ctx, cmd.Payment)
    if err != nil {
        h.executeCompensations(compensations)
        return fmt.Errorf("failed to process payment: %w", err)
    }
    compensations = append(compensations, func() error {
        return h.paymentService.RefundPayment(ctx, paymentID)
    })
    
    // Create order
    order := &Order{
        CustomerID:    cmd.CustomerID,
        Items:         cmd.Items,
        PaymentID:     paymentID,
        ReservationID: reservationID,
    }
    
    if err := h.orderRepo.Save(ctx, order); err != nil {
        h.executeCompensations(compensations)
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    return nil
}

func (h *PlaceOrderHandler) executeCompensations(compensations []func() error) {
    // Execute compensations in reverse order
    for i := len(compensations) - 1; i >= 0; i-- {
        if err := compensations[i](); err != nil {
            h.logger.Error("Compensation failed", err, map[string]interface{}{
                "compensation_index": i,
            })
        }
    }
}
```

### Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    state          string // "closed", "open", "half-open"
    failures       int
    lastFailureTime time.Time
    mutex          sync.RWMutex
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
    cb.mutex.RLock()
    state := cb.state
    cb.mutex.RUnlock()
    
    switch state {
    case "open":
        if time.Since(cb.lastFailureTime) > cb.resetTimeout {
            cb.mutex.Lock()
            cb.state = "half-open"
            cb.mutex.Unlock()
            return cb.attemptOperation(operation)
        }
        return &CircuitBreakerOpenError{
            ResetTime: cb.lastFailureTime.Add(cb.resetTimeout),
        }
    
    case "half-open":
        return cb.attemptOperation(operation)
    
    default: // closed
        return cb.attemptOperation(operation)
    }
}

func (cb *CircuitBreaker) attemptOperation(operation func() error) error {
    err := operation()
    
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailureTime = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        
        return err
    }
    
    // Success - reset circuit breaker
    cb.failures = 0
    cb.state = "closed"
    
    return nil
}

type CircuitBreakerOpenError struct {
    ResetTime time.Time
}

func (e *CircuitBreakerOpenError) Error() string {
    return fmt.Sprintf("circuit breaker is open, will reset at %s", e.ResetTime.Format(time.RFC3339))
}
```

## 🔍 Troubleshooting Guide

### Common Error Scenarios

#### Database Connection Issues

```go
func isDatabaseConnectionError(err error) bool {
    if err == nil {
        return false
    }
    
    errStr := err.Error()
    return strings.Contains(errStr, "connection refused") ||
           strings.Contains(errStr, "connection reset") ||
           strings.Contains(errStr, "no connection to the server") ||
           strings.Contains(errStr, "connection timeout")
}

func handleDatabaseError(ctx context.Context, err error, operation string) error {
    if isDatabaseConnectionError(err) {
        // Log for operations team
        logger.Error("Database connection issue detected", err, map[string]interface{}{
            "operation": operation,
            "timestamp": time.Now(),
            "context":   getContextInfo(ctx),
        })
        
        // Return user-friendly error
        return &ServiceUnavailableError{
            Service: "database",
            Reason:  "temporarily unavailable",
            RetryAfter: time.Minute * 5,
        }
    }
    
    return err
}
```

#### Timeout Handling

```go
func isTimeoutError(err error) bool {
    return errors.Is(err, context.DeadlineExceeded) ||
           strings.Contains(err.Error(), "timeout")
}

func handleTimeout(ctx context.Context, err error, operation string) error {
    if isTimeoutError(err) {
        logger.Warn("Operation timeout", map[string]interface{}{
            "operation": operation,
            "timeout":   getTimeoutFromContext(ctx),
            "elapsed":   getElapsedTime(ctx),
        })
        
        return &TimeoutError{
            Operation: operation,
            Timeout:   getTimeoutFromContext(ctx),
            Suggestion: "Consider increasing timeout or optimizing the operation",
        }
    }
    
    return err
}
```

#### Resource Exhaustion

```go
func isResourceExhaustionError(err error) bool {
    errStr := err.Error()
    return strings.Contains(errStr, "too many connections") ||
           strings.Contains(errStr, "out of memory") ||
           strings.Contains(errStr, "resource temporarily unavailable")
}

func handleResourceExhaustion(err error, resource string) error {
    if isResourceExhaustionError(err) {
        return &ResourceExhaustionError{
            Resource: resource,
            Reason:   err.Error(),
            Suggestions: []string{
                "Wait and retry",
                "Reduce load",
                "Scale up resources",
            },
        }
    }
    
    return err
}
```

### Error Recovery Patterns

```go
func (h *OrderProcessingHandler) Handle(ctx context.Context, cmd *ProcessOrderCommand) error {
    return h.withRetryAndCircuitBreaker(ctx, func() error {
        return h.processOrderInternal(ctx, cmd)
    })
}

func (h *OrderProcessingHandler) withRetryAndCircuitBreaker(ctx context.Context, operation func() error) error {
    return h.circuitBreaker.Execute(func() error {
        return h.retryWithBackoff(ctx, operation, 3, time.Second)
    })
}

func (h *OrderProcessingHandler) retryWithBackoff(ctx context.Context, operation func() error, maxRetries int, initialDelay time.Duration) error {
    var lastErr error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            delay := time.Duration(math.Pow(2, float64(attempt-1))) * initialDelay
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        
        err := operation()
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Don't retry certain types of errors
        if !isRetryableError(err) {
            break
        }
    }
    
    return lastErr
}

func isRetryableError(err error) bool {
    // Retry on temporary failures
    if isDatabaseConnectionError(err) ||
       isTimeoutError(err) ||
       isResourceExhaustionError(err) {
        return true
    }
    
    // Don't retry on permanent failures
    if isValidationError(err) ||
       isAuthorizationError(err) ||
       isNotFoundError(err) {
        return false
    }
    
    return false
}
```

## 🛡️ Custom Error Types

### Service Unavailable Error

```go
type ServiceUnavailableError struct {
    Service    string        `json:"service"`
    Reason     string        `json:"reason"`
    RetryAfter time.Duration `json:"retry_after"`
}

func (e *ServiceUnavailableError) Error() string {
    return fmt.Sprintf("service %s is unavailable: %s (retry after %s)", 
        e.Service, e.Reason, e.RetryAfter)
}
```

### Rate Limit Error

```go
type RateLimitError struct {
    Key        string        `json:"key"`
    Limit      int           `json:"limit"`
    Window     time.Duration `json:"window"`
    RetryAfter time.Duration `json:"retry_after"`
}

func (e *RateLimitError) Error() string {
    return fmt.Sprintf("rate limit exceeded for %s: %d requests per %s (retry after %s)",
        e.Key, e.Limit, e.Window, e.RetryAfter)
}
```

### Configuration Error

```go
type ConfigurationError struct {
    Component string `json:"component"`
    Setting   string `json:"setting"`
    Reason    string `json:"reason"`
    Example   string `json:"example,omitempty"`
}

func (e *ConfigurationError) Error() string {
    return fmt.Sprintf("configuration error in %s.%s: %s", 
        e.Component, e.Setting, e.Reason)
}
```

## 🎓 Best Practices

### 1. Error Classification

```go
// ✅ Good - classify errors appropriately
func ClassifyError(err error) ErrorClass {
    switch {
    case isValidationError(err):
        return ClientError    // 4xx
    case isAuthorizationError(err):
        return ClientError    // 4xx
    case isNotFoundError(err):
        return ClientError    // 4xx
    case isDatabaseError(err):
        return ServerError    // 5xx
    case isNetworkError(err):
        return ServerError    // 5xx
    default:
        return ServerError    // 5xx - safer default
    }
}
```

### 2. Meaningful Error Messages

```go
// ✅ Good - specific and actionable
return NewValidationError("email", 
    "Email address 'invalid-email' is not valid. Please provide a valid email address like 'user@example.com'",
    "INVALID_EMAIL_FORMAT")

// ❌ Bad - vague and unhelpful
return errors.New("invalid input")
```

### 3. Error Logging

```go
// ✅ Good - structured logging with context
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    if err := h.userRepo.Save(ctx, user); err != nil {
        h.logger.Error("Failed to save user", err, map[string]interface{}{
            "operation":      "save_user",
            "user_email":     cmd.Email,
            "correlation_id": getCorrelationID(ctx),
            "duration_ms":    time.Since(start).Milliseconds(),
        })
        return fmt.Errorf("failed to save user: %w", err)
    }
    return nil
}
```

### 4. Error Testing

```go
func TestCreateUserHandler_DatabaseError(t *testing.T) {
    // Setup
    userRepo := &MockUserRepository{}
    handler := &CreateUserHandler{userRepo: userRepo}
    
    // Mock database error
    userRepo.On("Save", mock.Anything, mock.AnythingOfType("*User")).
        Return(errors.New("database connection failed"))
    
    // Execute
    cmd := &CreateUserCommand{Name: "Test", Email: "test@example.com"}
    err := handler.Handle(context.Background(), cmd)
    
    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to save user")
    
    // Verify the original error is preserved
    assert.Contains(t, err.Error(), "database connection failed")
}
```

## 🚀 Next Steps

1. **Configuration**: [Configuration Options](./configuration.md)
2. **API Reference**: [API Documentation](./api-reference.md)
3. **Migration Guide**: [Migration Guide](./migration.md)
4. **Examples**: [Real-World Examples](./examples.md)

---

**Ready for configuration management? Continue with [Configuration Options](./configuration.md)! 🚀** 