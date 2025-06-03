# Handlers & Validators

Implementing business logic with handlers and ensuring data integrity with validators in go-cqrs.

## 📚 Table of Contents

- [Command Handlers](#command-handlers)
- [Query Handlers](#query-handlers)
- [Event Handlers](#event-handlers)
- [Validators](#validators)
- [Error Handling](#error-handling)
- [Handler Patterns](#handler-patterns)
- [Best Practices](#best-practices)

## ⚡ Command Handlers

Command handlers contain the business logic for processing commands. They should be focused, testable, and handle one specific command type.

### Basic Command Handler

```go
type CreateUserHandler struct {
    userRepo     UserRepository
    emailService EmailService
    logger       Logger
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // 1. Validate business rules
    if err := h.validateBusinessRules(ctx, cmd); err != nil {
        return err
    }
    
    // 2. Create domain entity
    user := &User{
        ID:       generateID(),
        Name:     cmd.Name,
        Email:    cmd.Email,
        Password: hashPassword(cmd.Password),
        Status:   "Active",
        CreatedAt: time.Now(),
    }
    
    // 3. Persist changes
    if err := h.userRepo.Save(ctx, user); err != nil {
        return fmt.Errorf("failed to save user: %w", err)
    }
    
    // 4. Publish domain events
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
    
    if err := cqrs.PublishEvent(ctx, event); err != nil {
        h.logger.Error("Failed to publish UserCreatedEvent", err)
        // Don't fail the command for event publishing errors
    }
    
    return nil
}

func (h *CreateUserHandler) validateBusinessRules(ctx context.Context, cmd *CreateUserCommand) error {
    // Check if user already exists
    existingUser, err := h.userRepo.GetByEmail(ctx, cmd.Email)
    if err != nil {
        return fmt.Errorf("failed to check existing user: %w", err)
    }
    
    if existingUser != nil {
        return errors.New("user with this email already exists")
    }
    
    return nil
}
```

### Command Handler with Dependencies

```go
type PlaceOrderHandler struct {
    orderRepo       OrderRepository
    inventoryRepo   InventoryRepository
    paymentService  PaymentService
    shippingService ShippingService
    eventBus        EventBus
}

func (h *PlaceOrderHandler) Handle(ctx context.Context, cmd *PlaceOrderCommand) error {
    // 1. Validate inventory
    for _, item := range cmd.Items {
        available, err := h.inventoryRepo.CheckAvailability(ctx, item.ProductID, item.Quantity)
        if err != nil {
            return fmt.Errorf("failed to check inventory: %w", err)
        }
        if !available {
            return fmt.Errorf("insufficient inventory for product %d", item.ProductID)
        }
    }
    
    // 2. Reserve inventory
    reservationID, err := h.inventoryRepo.Reserve(ctx, cmd.Items)
    if err != nil {
        return fmt.Errorf("failed to reserve inventory: %w", err)
    }
    
    // 3. Process payment
    paymentResult, err := h.paymentService.ProcessPayment(ctx, PaymentRequest{
        Amount:        cmd.Total,
        PaymentMethod: cmd.PaymentMethod,
        CustomerID:    cmd.CustomerID,
    })
    if err != nil {
        // Rollback inventory reservation
        h.inventoryRepo.CancelReservation(ctx, reservationID)
        return fmt.Errorf("payment failed: %w", err)
    }
    
    // 4. Create order
    order := &Order{
        ID:            generateOrderID(),
        CustomerID:    cmd.CustomerID,
        Items:         cmd.Items,
        Total:         cmd.Total,
        Status:        "Confirmed",
        PaymentID:     paymentResult.ID,
        ReservationID: reservationID,
        CreatedAt:     time.Now(),
    }
    
    if err := h.orderRepo.Save(ctx, order); err != nil {
        // Rollback payment and inventory
        h.paymentService.RefundPayment(ctx, paymentResult.ID)
        h.inventoryRepo.CancelReservation(ctx, reservationID)
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    // 5. Publish events
    h.publishOrderEvents(ctx, order)
    
    return nil
}
```

## 🔍 Query Handlers

Query handlers retrieve and format data for display. They should be optimized for read performance and never modify state.

### Simple Query Handler

```go
type GetUserHandler struct {
    userRepo UserRepository
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*UserResult, error) {
    user, err := h.userRepo.GetByID(ctx, q.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    if user == nil {
        return nil, ErrUserNotFound
    }
    
    // Transform to result DTO
    result := &UserResult{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Status:    user.Status,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
    
    return result, nil
}
```

### Complex Query Handler with Aggregation

```go
type GetOrderSummaryHandler struct {
    orderRepo    OrderRepository
    customerRepo CustomerRepository
    cache        Cache
}

func (h *GetOrderSummaryHandler) Handle(ctx context.Context, q GetOrderSummaryQuery) (*OrderSummaryResult, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("order-summary:%d:%s", q.CustomerID, q.Period)
    if cached, found := h.cache.Get(cacheKey); found {
        return cached.(*OrderSummaryResult), nil
    }
    
    // Get customer info
    customer, err := h.customerRepo.GetByID(ctx, q.CustomerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get customer: %w", err)
    }
    
    // Get orders for period
    orders, err := h.orderRepo.GetByCustomerAndPeriod(ctx, q.CustomerID, q.FromDate, q.ToDate)
    if err != nil {
        return nil, fmt.Errorf("failed to get orders: %w", err)
    }
    
    // Calculate aggregations
    result := &OrderSummaryResult{
        CustomerID:   customer.ID,
        CustomerName: customer.Name,
        Period:       q.Period,
        TotalOrders:  len(orders),
        TotalAmount:  calculateTotalAmount(orders),
        AverageOrder: calculateAverageOrder(orders),
        TopProducts:  calculateTopProducts(orders),
        OrdersByStatus: groupOrdersByStatus(orders),
    }
    
    // Cache result
    h.cache.Set(cacheKey, result, 5*time.Minute)
    
    return result, nil
}
```

### Paginated Query Handler

```go
type ListUsersHandler struct {
    userRepo UserRepository
}

func (h *ListUsersHandler) Handle(ctx context.Context, q ListUsersQuery) (*UserListResult, error) {
    // Validate pagination parameters
    if q.PageSize <= 0 || q.PageSize > 100 {
        q.PageSize = 20 // Default page size
    }
    if q.Page <= 0 {
        q.Page = 1
    }
    
    // Build filter criteria
    filter := UserFilter{
        Name:   q.NameFilter,
        Email:  q.EmailFilter,
        Status: q.StatusFilter,
    }
    
    // Get total count
    totalCount, err := h.userRepo.Count(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to count users: %w", err)
    }
    
    // Get paginated results
    offset := (q.Page - 1) * q.PageSize
    users, err := h.userRepo.List(ctx, filter, offset, q.PageSize)
    if err != nil {
        return nil, fmt.Errorf("failed to list users: %w", err)
    }
    
    // Transform to DTOs
    userDTOs := make([]*UserSummary, len(users))
    for i, user := range users {
        userDTOs[i] = &UserSummary{
            ID:        user.ID,
            Name:      user.Name,
            Email:     user.Email,
            Status:    user.Status,
            CreatedAt: user.CreatedAt,
        }
    }
    
    // Calculate pagination info
    totalPages := int(math.Ceil(float64(totalCount) / float64(q.PageSize)))
    hasNext := q.Page < totalPages
    hasPrevious := q.Page > 1
    
    return &UserListResult{
        Users:       userDTOs,
        TotalCount:  totalCount,
        Page:        q.Page,
        PageSize:    q.PageSize,
        TotalPages:  totalPages,
        HasNext:     hasNext,
        HasPrevious: hasPrevious,
    }, nil
}
```

## 📢 Event Handlers

Event handlers process domain events and handle side effects. They should be idempotent and handle failures gracefully.

### Simple Event Handler

```go
type UserCreatedHandler struct {
    emailService EmailService
    logger       Logger
}

func (h *UserCreatedHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    h.logger.Info("Processing UserCreatedEvent", map[string]interface{}{
        "userID": e.UserID,
        "email":  e.Email,
    })
    
    // Send welcome email
    err := h.emailService.SendWelcomeEmail(ctx, WelcomeEmailRequest{
        ToEmail:   e.Email,
        UserName:  e.Name,
        UserID:    e.UserID,
        Template:  "welcome-template",
    })
    
    if err != nil {
        h.logger.Error("Failed to send welcome email", err)
        // Don't return error - email failure shouldn't fail the event processing
        // Consider using a retry mechanism or dead letter queue
    }
    
    return nil
}
```

### Event Handler with State Management

```go
type OrderPlacedHandler struct {
    readModelRepo OrderSummaryRepository
    analytics     AnalyticsService
    logger        Logger
}

func (h *OrderPlacedHandler) Handle(ctx context.Context, e OrderPlacedEvent) error {
    // Update read model for customer order summary
    err := h.updateCustomerOrderSummary(ctx, e)
    if err != nil {
        return fmt.Errorf("failed to update order summary: %w", err)
    }
    
    // Track analytics
    err = h.analytics.TrackOrderPlaced(ctx, AnalyticsEvent{
        EventType:  "order_placed",
        CustomerID: e.CustomerID,
        OrderID:    e.OrderID,
        Amount:     e.Total,
        Timestamp:  e.ExecutionTime,
        Properties: map[string]interface{}{
            "item_count":     len(e.Items),
            "payment_method": e.PaymentMethod,
        },
    })
    
    if err != nil {
        h.logger.Error("Failed to track analytics", err)
        // Don't fail - analytics is not critical
    }
    
    return nil
}

func (h *OrderPlacedHandler) updateCustomerOrderSummary(ctx context.Context, e OrderPlacedEvent) error {
    summary, err := h.readModelRepo.GetByCustomerID(ctx, e.CustomerID)
    if err != nil && !errors.Is(err, ErrNotFound) {
        return err
    }
    
    if summary == nil {
        summary = &OrderSummary{
            CustomerID: e.CustomerID,
        }
    }
    
    // Update summary
    summary.TotalOrders++
    summary.TotalAmount = summary.TotalAmount.Add(e.Total)
    summary.LastOrderDate = e.ExecutionTime
    
    return h.readModelRepo.Save(ctx, summary)
}
```

## ✅ Validators

Validators ensure data integrity and business rule compliance before command execution.

### Basic Validator

```go
type CreateUserValidator struct {
    userRepo UserRepository
}

func (v *CreateUserValidator) Validate(ctx context.Context, cmd *CreateUserCommand) error {
    var validationErrors []error
    
    // Required field validation
    if strings.TrimSpace(cmd.Name) == "" {
        validationErrors = append(validationErrors, errors.New("name is required"))
    }
    
    if strings.TrimSpace(cmd.Email) == "" {
        validationErrors = append(validationErrors, errors.New("email is required"))
    }
    
    if strings.TrimSpace(cmd.Password) == "" {
        validationErrors = append(validationErrors, errors.New("password is required"))
    }
    
    // Format validation
    if cmd.Email != "" && !isValidEmail(cmd.Email) {
        validationErrors = append(validationErrors, errors.New("invalid email format"))
    }
    
    // Business rule validation
    if len(cmd.Password) < 8 {
        validationErrors = append(validationErrors, errors.New("password must be at least 8 characters"))
    }
    
    if len(cmd.Name) < 2 {
        validationErrors = append(validationErrors, errors.New("name must be at least 2 characters"))
    }
    
    // Database validation
    if cmd.Email != "" {
        exists, err := v.userRepo.ExistsByEmail(ctx, cmd.Email)
        if err != nil {
            return fmt.Errorf("failed to check email uniqueness: %w", err)
        }
        if exists {
            validationErrors = append(validationErrors, errors.New("email already exists"))
        }
    }
    
    if len(validationErrors) > 0 {
        return &ValidationError{Errors: validationErrors}
    }
    
    return nil
}

func isValidEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}
```

### Complex Business Rule Validator

```go
type PlaceOrderValidator struct {
    customerRepo  CustomerRepository
    productRepo   ProductRepository
    inventoryRepo InventoryRepository
}

func (v *PlaceOrderValidator) Validate(ctx context.Context, cmd *PlaceOrderCommand) error {
    // Validate customer exists and is active
    customer, err := v.customerRepo.GetByID(ctx, cmd.CustomerID)
    if err != nil {
        return fmt.Errorf("failed to validate customer: %w", err)
    }
    if customer == nil {
        return errors.New("customer not found")
    }
    if customer.Status != "Active" {
        return errors.New("customer account is not active")
    }
    
    // Validate order has items
    if len(cmd.Items) == 0 {
        return errors.New("order must have at least one item")
    }
    
    var totalAmount decimal.Decimal
    
    // Validate each item
    for i, item := range cmd.Items {
        if item.ProductID <= 0 {
            return fmt.Errorf("invalid product ID in item %d", i)
        }
        
        if item.Quantity <= 0 {
            return fmt.Errorf("invalid quantity in item %d", i)
        }
        
        // Validate product exists
        product, err := v.productRepo.GetByID(ctx, item.ProductID)
        if err != nil {
            return fmt.Errorf("failed to validate product %d: %w", item.ProductID, err)
        }
        if product == nil {
            return fmt.Errorf("product %d not found", item.ProductID)
        }
        if !product.IsActive {
            return fmt.Errorf("product %d is not available", item.ProductID)
        }
        
        // Validate inventory
        available, err := v.inventoryRepo.CheckAvailability(ctx, item.ProductID, item.Quantity)
        if err != nil {
            return fmt.Errorf("failed to check inventory for product %d: %w", item.ProductID, err)
        }
        if !available {
            return fmt.Errorf("insufficient inventory for product %d", item.ProductID)
        }
        
        // Calculate expected price
        expectedTotal := product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
        if !item.Price.Equal(expectedTotal) {
            return fmt.Errorf("price mismatch for product %d", item.ProductID)
        }
        
        totalAmount = totalAmount.Add(item.Price)
    }
    
    // Validate total amount
    if !cmd.Total.Equal(totalAmount) {
        return errors.New("order total does not match item prices")
    }
    
    // Validate minimum order amount
    minOrderAmount := decimal.NewFromFloat(10.00)
    if cmd.Total.LessThan(minOrderAmount) {
        return fmt.Errorf("minimum order amount is %s", minOrderAmount)
    }
    
    return nil
}
```

## ❌ Error Handling

### Custom Error Types

```go
type ValidationError struct {
    Errors []error
}

func (e *ValidationError) Error() string {
    var messages []string
    for _, err := range e.Errors {
        messages = append(messages, err.Error())
    }
    return "validation failed: " + strings.Join(messages, ", ")
}

type BusinessRuleError struct {
    Rule    string
    Message string
}

func (e *BusinessRuleError) Error() string {
    return fmt.Sprintf("business rule violation [%s]: %s", e.Rule, e.Message)
}

type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}
```

### Error Handling in Handlers

```go
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    user := &User{
        Name:  cmd.Name,
        Email: cmd.Email,
    }
    
    err := h.userRepo.Save(ctx, user)
    if err != nil {
        // Check for specific error types
        if isDuplicateKeyError(err) {
            return &BusinessRuleError{
                Rule:    "unique_email",
                Message: "A user with this email already exists",
            }
        }
        
        if isConnectionError(err) {
            return fmt.Errorf("database connection failed: %w", err)
        }
        
        return fmt.Errorf("failed to save user: %w", err)
    }
    
    return nil
}
```

## 🔄 Handler Patterns

### Transactional Handler

```go
type TransactionalOrderHandler struct {
    txManager TransactionManager
    orderRepo OrderRepository
    // ... other repos
}

func (h *TransactionalOrderHandler) Handle(ctx context.Context, cmd *PlaceOrderCommand) error {
    return h.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        // All operations within this transaction
        order := createOrder(cmd)
        
        if err := h.orderRepo.Save(txCtx, order); err != nil {
            return err
        }
        
        if err := h.inventoryRepo.Reserve(txCtx, cmd.Items); err != nil {
            return err
        }
        
        return h.paymentRepo.ProcessPayment(txCtx, cmd.Payment)
    })
}
```

### Retry Handler

```go
type RetryableHandler struct {
    baseHandler CommandHandler
    maxRetries  int
    backoff     time.Duration
}

func (h *RetryableHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    var lastErr error
    
    for attempt := 0; attempt <= h.maxRetries; attempt++ {
        err := h.baseHandler.Handle(ctx, cmd)
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Don't retry on validation errors
        if isValidationError(err) {
            return err
        }
        
        if attempt < h.maxRetries {
            time.Sleep(h.backoff * time.Duration(attempt+1))
        }
    }
    
    return fmt.Errorf("command failed after %d attempts: %w", h.maxRetries, lastErr)
}
```

## 🎓 Best Practices

### 1. Single Responsibility

```go
// ✅ Good - each handler has one responsibility
type CreateUserHandler struct {}
type UpdateUserEmailHandler struct {}
type ActivateUserHandler struct {}

// ❌ Bad - handler does too many things
type UserManagementHandler struct {} // Handles all user operations
```

### 2. Dependency Injection

```go
// ✅ Good - dependencies injected
type CreateUserHandler struct {
    userRepo     UserRepository
    emailService EmailService
    logger       Logger
}

// ❌ Bad - hard dependencies
type CreateUserHandler struct {}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    db := sql.Open("postgres", "...") // Hard-coded dependency
}
```

### 3. Error Wrapping

```go
// ✅ Good - errors wrapped with context
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    if err := h.userRepo.Save(ctx, user); err != nil {
        return fmt.Errorf("failed to save user %s: %w", cmd.Email, err)
    }
}

// ❌ Bad - errors not wrapped
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    return h.userRepo.Save(ctx, user) // Original error context lost
}
```

### 4. Validation Separation

```go
// ✅ Good - validation separated from business logic
type CreateUserValidator struct {}
type CreateUserHandler struct {}

// ❌ Bad - validation mixed with business logic
func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    if cmd.Name == "" { return errors.New("name required") }
    if cmd.Email == "" { return errors.New("email required") }
    // ... business logic mixed with validation
}
```

## 🚀 Next Steps

1. **Learn Context Usage**: [Context Support](./context-support.md)
2. **Add Auto-Registration**: [Auto-Registration Guide](./auto-registration.md)
3. **Use Dependency Injection**: [Dependency Injection](./dependency-injection.md)
4. **Add Decorators**: [Decorators & Middleware](./decorators.md)

---

**Ready for advanced features? Explore [Auto-Registration Guide](./auto-registration.md)! 🚀** 