# Decorators & Middleware

Implementing cross-cutting concerns with decorators and middleware in go-cqrs applications.

## 📚 Table of Contents

- [Decorator Pattern](#decorator-pattern)
- [Built-in Decorators](#built-in-decorators)
- [Custom Decorators](#custom-decorators)
- [Chaining Decorators](#chaining-decorators)
- [Performance Decorators](#performance-decorators)
- [Security Decorators](#security-decorators)
- [Resilience Decorators](#resilience-decorators)
- [Best Practices](#best-practices)

## 🎯 Decorator Pattern

Decorators wrap handlers to add cross-cutting functionality without modifying the core business logic:

```go
// Core business logic (unchanged)
type CreateUserHandler struct {
    userRepo UserRepository
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Pure business logic
    user := &User{Name: cmd.Name, Email: cmd.Email}
    return h.userRepo.Save(ctx, user)
}

// Decorator adds logging (wrapper)
type LoggingDecorator struct {
    handler CommandHandler
    logger  Logger
}

func (d *LoggingDecorator) Handle(ctx context.Context, cmd Command) error {
    d.logger.Info("Command started", getCommandInfo(cmd))
    
    err := d.handler.Handle(ctx, cmd)
    
    if err != nil {
        d.logger.Error("Command failed", err)
    } else {
        d.logger.Info("Command completed")
    }
    
    return err
}
```

## 🏗️ Built-in Decorators

go-cqrs provides several built-in decorators:

### Logging Decorator

```go
// Enable logging for all commands
manager := cqrs.NewCqrsManager()
manager.AddLoggingDecorator()
cqrs.SetManager(manager)

// Custom logging configuration
logger := &StructuredLogger{
    level: "INFO",
    format: "json",
}

manager.AddLoggingDecorator(WithLogger(logger))
```

### Metrics Decorator

```go
// Enable Prometheus metrics
manager.AddMetricsDecorator()

// Custom metrics configuration
metricsConfig := MetricsConfig{
    Namespace: "myapp",
    Subsystem: "cqrs",
    Labels:    []string{"service", "version"},
}

manager.AddMetricsDecorator(WithMetrics(metricsConfig))
```

### Tracing Decorator

```go
// Enable OpenTelemetry tracing
manager.AddTracingDecorator()

// Custom tracing configuration
tracingConfig := TracingConfig{
    ServiceName:    "user-service",
    ServiceVersion: "v1.0.0",
    Environment:    "production",
}

manager.AddTracingDecorator(WithTracing(tracingConfig))
```

## 🛠️ Custom Decorators

### Validation Decorator

```go
type ValidationDecorator struct {
    handler   CommandHandler
    validator Validator
}

func NewValidationDecorator(handler CommandHandler, validator Validator) *ValidationDecorator {
    return &ValidationDecorator{
        handler:   handler,
        validator: validator,
    }
}

func (d *ValidationDecorator) Handle(ctx context.Context, cmd Command) error {
    // Validate command before execution
    if err := d.validator.Validate(ctx, cmd); err != nil {
        return &ValidationError{
            Message: "Command validation failed",
            Errors:  []error{err},
        }
    }
    
    // Execute if validation passes
    return d.handler.Handle(ctx, cmd)
}

// Usage
func setupHandlerWithValidation() {
    baseHandler := &CreateUserHandler{userRepo: userRepo}
    validator := &CreateUserValidator{userRepo: userRepo}
    
    decoratedHandler := NewValidationDecorator(baseHandler, validator)
    cqrs.RegisterCommandHandler(decoratedHandler)
}
```

### Caching Decorator

```go
type CachingDecorator struct {
    handler QueryHandler
    cache   CacheService
    ttl     time.Duration
}

func NewCachingDecorator(handler QueryHandler, cache CacheService, ttl time.Duration) *CachingDecorator {
    return &CachingDecorator{
        handler: handler,
        cache:   cache,
        ttl:     ttl,
    }
}

func (d *CachingDecorator) Handle(ctx context.Context, query Query) (interface{}, error) {
    // Generate cache key
    cacheKey := d.generateCacheKey(query)
    
    // Try to get from cache first
    if cached, found := d.cache.Get(ctx, cacheKey); found {
        cacheHits.Inc()
        return cached, nil
    }
    
    // Execute query
    result, err := d.handler.Handle(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    d.cache.Set(ctx, cacheKey, result, d.ttl)
    cacheMisses.Inc()
    
    return result, nil
}

func (d *CachingDecorator) generateCacheKey(query Query) string {
    queryType := reflect.TypeOf(query).Name()
    queryData := fmt.Sprintf("%+v", query)
    hasher := sha256.New()
    hasher.Write([]byte(queryType + queryData))
    return fmt.Sprintf("query:%s:%x", queryType, hasher.Sum(nil)[:8])
}
```

### Transaction Decorator

```go
type TransactionDecorator struct {
    handler   CommandHandler
    txManager TransactionManager
}

func NewTransactionDecorator(handler CommandHandler, txManager TransactionManager) *TransactionDecorator {
    return &TransactionDecorator{
        handler:   handler,
        txManager: txManager,
    }
}

func (d *TransactionDecorator) Handle(ctx context.Context, cmd Command) error {
    return d.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        return d.handler.Handle(txCtx, cmd)
    })
}

// Transaction manager interface
type TransactionManager interface {
    WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

// PostgreSQL implementation
type PostgreSQLTransactionManager struct {
    db *sql.DB
}

func (tm *PostgreSQLTransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
    tx, err := tm.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()
    
    // Create context with transaction
    txCtx := context.WithValue(ctx, "tx", tx)
    
    if err := fn(txCtx); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit()
}
```

## 🔗 Chaining Decorators

Decorators can be chained to combine multiple concerns:

### Manual Chaining

```go
func setupDecoratedHandler() {
    // Base handler
    baseHandler := &CreateUserHandler{userRepo: userRepo}
    
    // Chain decorators
    handler := NewLoggingDecorator(
        NewMetricsDecorator(
            NewValidationDecorator(
                NewTransactionDecorator(
                    baseHandler,
                    txManager,
                ),
                validator,
            ),
            metricsCollector,
        ),
        logger,
    )
    
    cqrs.RegisterCommandHandler(handler)
}
```

### Builder Pattern

```go
type HandlerBuilder struct {
    handler    CommandHandler
    decorators []DecoratorFunc
}

type DecoratorFunc func(CommandHandler) CommandHandler

func NewHandlerBuilder(handler CommandHandler) *HandlerBuilder {
    return &HandlerBuilder{handler: handler}
}

func (b *HandlerBuilder) WithLogging(logger Logger) *HandlerBuilder {
    b.decorators = append(b.decorators, func(h CommandHandler) CommandHandler {
        return NewLoggingDecorator(h, logger)
    })
    return b
}

func (b *HandlerBuilder) WithMetrics(collector MetricsCollector) *HandlerBuilder {
    b.decorators = append(b.decorators, func(h CommandHandler) CommandHandler {
        return NewMetricsDecorator(h, collector)
    })
    return b
}

func (b *HandlerBuilder) WithValidation(validator Validator) *HandlerBuilder {
    b.decorators = append(b.decorators, func(h CommandHandler) CommandHandler {
        return NewValidationDecorator(h, validator)
    })
    return b
}

func (b *HandlerBuilder) WithTransaction(txManager TransactionManager) *HandlerBuilder {
    b.decorators = append(b.decorators, func(h CommandHandler) CommandHandler {
        return NewTransactionDecorator(h, txManager)
    })
    return b
}

func (b *HandlerBuilder) Build() CommandHandler {
    result := b.handler
    
    // Apply decorators in reverse order (innermost first)
    for i := len(b.decorators) - 1; i >= 0; i-- {
        result = b.decorators[i](result)
    }
    
    return result
}

// Usage
func setupWithBuilder() {
    handler := NewHandlerBuilder(&CreateUserHandler{userRepo: userRepo}).
        WithTransaction(txManager).
        WithValidation(validator).
        WithMetrics(metricsCollector).
        WithLogging(logger).
        Build()
    
    cqrs.RegisterCommandHandler(handler)
}
```

### Automatic Decoration

```go
type DecoratorRegistry struct {
    decorators []DecoratorConfig
}

type DecoratorConfig struct {
    Name     string
    Priority int
    Factory  func(CommandHandler) CommandHandler
    Filter   func(Command) bool
}

func (r *DecoratorRegistry) RegisterDecorator(config DecoratorConfig) {
    r.decorators = append(r.decorators, config)
    sort.Slice(r.decorators, func(i, j int) bool {
        return r.decorators[i].Priority < r.decorators[j].Priority
    })
}

func (r *DecoratorRegistry) DecorateHandler(handler CommandHandler, cmdType reflect.Type) CommandHandler {
    result := handler
    
    for _, decorator := range r.decorators {
        if decorator.Filter == nil || decorator.Filter(reflect.New(cmdType).Interface().(Command)) {
            result = decorator.Factory(result)
        }
    }
    
    return result
}

// Setup automatic decoration
func setupAutoDecoration() {
    registry := &DecoratorRegistry{}
    
    // Register decorators by priority
    registry.RegisterDecorator(DecoratorConfig{
        Name:     "transaction",
        Priority: 1,
        Factory:  func(h CommandHandler) CommandHandler { return NewTransactionDecorator(h, txManager) },
        Filter:   isTransactionalCommand,
    })
    
    registry.RegisterDecorator(DecoratorConfig{
        Name:     "validation",
        Priority: 2,
        Factory:  func(h CommandHandler) CommandHandler { return NewValidationDecorator(h, validator) },
    })
    
    registry.RegisterDecorator(DecoratorConfig{
        Name:     "metrics",
        Priority: 3,
        Factory:  func(h CommandHandler) CommandHandler { return NewMetricsDecorator(h, metrics) },
    })
    
    registry.RegisterDecorator(DecoratorConfig{
        Name:     "logging",
        Priority: 4,
        Factory:  func(h CommandHandler) CommandHandler { return NewLoggingDecorator(h, logger) },
    })
}
```

## ⚡ Performance Decorators

### Timeout Decorator

```go
type TimeoutDecorator struct {
    handler CommandHandler
    timeout time.Duration
}

func NewTimeoutDecorator(handler CommandHandler, timeout time.Duration) *TimeoutDecorator {
    return &TimeoutDecorator{
        handler: handler,
        timeout: timeout,
    }
}

func (d *TimeoutDecorator) Handle(ctx context.Context, cmd Command) error {
    ctx, cancel := context.WithTimeout(ctx, d.timeout)
    defer cancel()
    
    // Execute in goroutine to handle cancellation
    errChan := make(chan error, 1)
    
    go func() {
        errChan <- d.handler.Handle(ctx, cmd)
    }()
    
    select {
    case err := <-errChan:
        return err
    case <-ctx.Done():
        return &TimeoutError{
            Timeout: d.timeout,
            Command: reflect.TypeOf(cmd).Name(),
        }
    }
}
```

### Circuit Breaker Decorator

```go
type CircuitBreakerDecorator struct {
    handler     CommandHandler
    breaker     CircuitBreaker
    fallbackFn  func(context.Context, Command) error
}

type CircuitBreaker interface {
    Execute(func() error) error
    State() string
    Metrics() CircuitBreakerMetrics
}

type CircuitBreakerMetrics struct {
    Requests      int64
    TotalFailures int64
    TotalSuccesses int64
    ConsecutiveFailures int64
}

func NewCircuitBreakerDecorator(
    handler CommandHandler, 
    breaker CircuitBreaker,
    fallbackFn func(context.Context, Command) error,
) *CircuitBreakerDecorator {
    return &CircuitBreakerDecorator{
        handler:    handler,
        breaker:    breaker,
        fallbackFn: fallbackFn,
    }
}

func (d *CircuitBreakerDecorator) Handle(ctx context.Context, cmd Command) error {
    err := d.breaker.Execute(func() error {
        return d.handler.Handle(ctx, cmd)
    })
    
    if err != nil && d.breaker.State() == "open" && d.fallbackFn != nil {
        // Circuit is open, try fallback
        return d.fallbackFn(ctx, cmd)
    }
    
    return err
}
```

### Rate Limiting Decorator

```go
type RateLimitDecorator struct {
    handler CommandHandler
    limiter RateLimiter
}

type RateLimiter interface {
    Allow(ctx context.Context, key string) bool
    Wait(ctx context.Context, key string) error
}

func NewRateLimitDecorator(handler CommandHandler, limiter RateLimiter) *RateLimitDecorator {
    return &RateLimitDecorator{
        handler: handler,
        limiter: limiter,
    }
}

func (d *RateLimitDecorator) Handle(ctx context.Context, cmd Command) error {
    // Generate rate limit key (could be user ID, IP, command type, etc.)
    key := d.generateRateLimitKey(ctx, cmd)
    
    if !d.limiter.Allow(ctx, key) {
        return &RateLimitError{
            Key:     key,
            Command: reflect.TypeOf(cmd).Name(),
        }
    }
    
    return d.handler.Handle(ctx, cmd)
}

func (d *RateLimitDecorator) generateRateLimitKey(ctx context.Context, cmd Command) string {
    userID := getUserID(ctx)
    commandType := reflect.TypeOf(cmd).Name()
    return fmt.Sprintf("ratelimit:%d:%s", userID, commandType)
}
```

## 🔒 Security Decorators

### Authorization Decorator

```go
type AuthorizationDecorator struct {
    handler    CommandHandler
    authorizer Authorizer
}

type Authorizer interface {
    Authorize(ctx context.Context, cmd Command) error
}

func NewAuthorizationDecorator(handler CommandHandler, authorizer Authorizer) *AuthorizationDecorator {
    return &AuthorizationDecorator{
        handler:    handler,
        authorizer: authorizer,
    }
}

func (d *AuthorizationDecorator) Handle(ctx context.Context, cmd Command) error {
    // Check authorization first
    if err := d.authorizer.Authorize(ctx, cmd); err != nil {
        return &AuthorizationError{
            UserID:  getUserID(ctx),
            Command: reflect.TypeOf(cmd).Name(),
            Reason:  err.Error(),
        }
    }
    
    return d.handler.Handle(ctx, cmd)
}

// Role-based authorizer
type RoleBasedAuthorizer struct {
    permissions map[string][]string // command -> required roles
}

func (a *RoleBasedAuthorizer) Authorize(ctx context.Context, cmd Command) error {
    user := getAuthenticatedUser(ctx)
    if user == nil {
        return errors.New("user not authenticated")
    }
    
    commandType := reflect.TypeOf(cmd).Name()
    requiredRoles := a.permissions[commandType]
    
    if len(requiredRoles) == 0 {
        return nil // No authorization required
    }
    
    for _, requiredRole := range requiredRoles {
        if user.HasRole(requiredRole) {
            return nil
        }
    }
    
    return fmt.Errorf("user lacks required role for command %s", commandType)
}
```

### Audit Decorator

```go
type AuditDecorator struct {
    handler  CommandHandler
    auditor  AuditLogger
}

type AuditEvent struct {
    Timestamp   time.Time
    UserID      int
    Command     string
    CommandData interface{}
    Success     bool
    Error       string
    Duration    time.Duration
    IPAddress   string
    UserAgent   string
}

type AuditLogger interface {
    LogAuditEvent(ctx context.Context, event AuditEvent) error
}

func NewAuditDecorator(handler CommandHandler, auditor AuditLogger) *AuditDecorator {
    return &AuditDecorator{
        handler: handler,
        auditor: auditor,
    }
}

func (d *AuditDecorator) Handle(ctx context.Context, cmd Command) error {
    start := time.Now()
    
    auditEvent := AuditEvent{
        Timestamp:   start,
        UserID:      getUserID(ctx),
        Command:     reflect.TypeOf(cmd).Name(),
        CommandData: cmd,
        IPAddress:   getClientIP(ctx),
        UserAgent:   getUserAgent(ctx),
    }
    
    err := d.handler.Handle(ctx, cmd)
    
    auditEvent.Duration = time.Since(start)
    auditEvent.Success = err == nil
    if err != nil {
        auditEvent.Error = err.Error()
    }
    
    // Log audit event (fire and forget)
    go d.auditor.LogAuditEvent(ctx, auditEvent)
    
    return err
}
```

## 🛡️ Resilience Decorators

### Retry Decorator

```go
type RetryDecorator struct {
    handler     CommandHandler
    maxRetries  int
    backoff     BackoffStrategy
    retryFilter func(error) bool
}

type BackoffStrategy interface {
    NextDelay(attempt int) time.Duration
}

type ExponentialBackoff struct {
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func (b *ExponentialBackoff) NextDelay(attempt int) time.Duration {
    delay := time.Duration(float64(b.InitialDelay) * math.Pow(b.Multiplier, float64(attempt)))
    if delay > b.MaxDelay {
        delay = b.MaxDelay
    }
    return delay
}

func NewRetryDecorator(
    handler CommandHandler, 
    maxRetries int, 
    backoff BackoffStrategy,
    retryFilter func(error) bool,
) *RetryDecorator {
    return &RetryDecorator{
        handler:     handler,
        maxRetries:  maxRetries,
        backoff:     backoff,
        retryFilter: retryFilter,
    }
}

func (d *RetryDecorator) Handle(ctx context.Context, cmd Command) error {
    var lastErr error
    
    for attempt := 0; attempt <= d.maxRetries; attempt++ {
        if attempt > 0 {
            delay := d.backoff.NextDelay(attempt - 1)
            
            select {
            case <-time.After(delay):
                // Continue with retry
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        
        err := d.handler.Handle(ctx, cmd)
        if err == nil {
            if attempt > 0 {
                // Log successful retry
                retrySuccesses.WithLabelValues(reflect.TypeOf(cmd).Name()).Inc()
            }
            return nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if d.retryFilter != nil && !d.retryFilter(err) {
            break
        }
        
        // Check if we have more attempts
        if attempt == d.maxRetries {
            break
        }
        
        // Log retry attempt
        retryAttempts.WithLabelValues(reflect.TypeOf(cmd).Name()).Inc()
    }
    
    // All retries failed
    retryFailures.WithLabelValues(reflect.TypeOf(cmd).Name()).Inc()
    return &RetryExhaustedError{
        Attempts:    d.maxRetries + 1,
        LastError:   lastErr,
        Command:     reflect.TypeOf(cmd).Name(),
    }
}

// Example retry filter
func isRetryableError(err error) bool {
    // Retry on temporary network errors, database connection issues, etc.
    if isNetworkError(err) || isDatabaseConnectionError(err) {
        return true
    }
    
    // Don't retry on validation errors, authorization errors, etc.
    if isValidationError(err) || isAuthorizationError(err) {
        return false
    }
    
    return false
}
```

## 🎓 Best Practices

### 1. Order Matters

```go
// ✅ Good - logical order
handler := NewLoggingDecorator(          // Outermost - logs everything
    NewMetricsDecorator(                 // Metrics include all attempts
        NewRetryDecorator(               // Retry wraps execution
            NewAuthorizationDecorator(   // Auth before business logic
                NewValidationDecorator(  // Validate before processing
                    NewTransactionDecorator( // Transaction around business logic
                        baseHandler,     // Core business logic
                    ),
                ),
            ),
        ),
    ),
)

// ❌ Bad - authorization after transaction
handler := NewTransactionDecorator(
    NewAuthorizationDecorator(
        baseHandler, // Auth happens inside transaction
    ),
)
```

### 2. Keep Decorators Focused

```go
// ✅ Good - single responsibility
type LoggingDecorator struct {
    handler CommandHandler
    logger  Logger
}

// ❌ Bad - multiple responsibilities
type LoggingAndMetricsDecorator struct {
    handler CommandHandler
    logger  Logger
    metrics MetricsCollector
}
```

### 3. Handle Errors Properly

```go
// ✅ Good - preserve original error information
func (d *ValidationDecorator) Handle(ctx context.Context, cmd Command) error {
    if err := d.validator.Validate(ctx, cmd); err != nil {
        return &ValidationError{
            Command:       reflect.TypeOf(cmd).Name(),
            OriginalError: err,
            ValidationDetails: getValidationDetails(err),
        }
    }
    
    return d.handler.Handle(ctx, cmd)
}

// ❌ Bad - lose error context
func (d *ValidationDecorator) Handle(ctx context.Context, cmd Command) error {
    if err := d.validator.Validate(ctx, cmd); err != nil {
        return errors.New("validation failed") // Lost original error
    }
    
    return d.handler.Handle(ctx, cmd)
}
```

### 4. Make Decorators Configurable

```go
// ✅ Good - configurable decorator
type CachingDecoratorConfig struct {
    TTL                time.Duration
    KeyGenerator       func(Query) string
    ShouldCache        func(Query) bool
    ShouldCacheResult  func(interface{}) bool
}

func NewCachingDecorator(handler QueryHandler, cache CacheService, config CachingDecoratorConfig) *CachingDecorator {
    return &CachingDecorator{
        handler: handler,
        cache:   cache,
        config:  config,
    }
}
```

## 🚀 Next Steps

1. **Error Handling**: [Error Handling](./error-handling.md)
2. **Configuration**: [Configuration Options](./configuration.md)
3. **API Reference**: [API Documentation](./api-reference.md)
4. **Migration Guide**: [Migration Guide](./migration.md)

---

**Ready for robust error handling? Continue with [Error Handling](./error-handling.md)! 🚀** 