# Architectural Patterns

Common architectural patterns and best practices for go-cqrs implementation.

## 🏗️ **CQRS Patterns**

### **Basic CQRS Pattern**

The fundamental separation of command and query responsibilities:

```go
// Command side - Write operations
type CreateOrderCommand struct {
    command.Base
    CustomerID int
    Items      []OrderItem
}

type CreateOrderHandler struct {
    OrderRepo OrderRepository `inject:""`
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    // Write-optimized logic
    order := &Order{
        CustomerID: cmd.CustomerID,
        Items:      cmd.Items,
        Status:     "pending",
        Created:    time.Now(),
    }
    
    return h.OrderRepo.Create(ctx, order)
}

// Query side - Read operations
type GetOrdersQuery struct {
    query.Base
    CustomerID int
    Status     string
    Page       int
    PageSize   int
}

type GetOrdersHandler struct {
    OrderReadRepo OrderReadRepository `inject:""`
}

func (h *GetOrdersHandler) Handle(ctx context.Context, q GetOrdersQuery) ([]OrderDTO, error) {
    // Read-optimized logic with denormalized data
    return h.OrderReadRepo.GetOrdersByCustomer(ctx, OrderFilters{
        CustomerID: q.CustomerID,
        Status:     q.Status,
        Pagination: Pagination{Page: q.Page, Size: q.PageSize},
    })
}
```

### **Event Sourcing Pattern**

Store events instead of current state:

```go
// Event-sourced aggregate
type Account struct {
    ID       int
    Balance  float64
    Version  int
    Events   []IEvent
}

func (a *Account) Deposit(amount float64) {
    event := MoneyDepositedEvent{
        AccountID: a.ID,
        Amount:    amount,
        Timestamp: time.Now(),
    }
    
    a.apply(event)
    a.Events = append(a.Events, event)
}

func (a *Account) apply(event IEvent) {
    switch e := event.(type) {
    case MoneyDepositedEvent:
        a.Balance += e.Amount
    case MoneyWithdrawnEvent:
        a.Balance -= e.Amount
    }
    a.Version++
}

// Event store handler
type DepositMoneyHandler struct {
    EventStore EventStore `inject:""`
}

func (h *DepositMoneyHandler) Handle(ctx context.Context, cmd *DepositMoneyCommand) error {
    // Load aggregate from events
    account, err := h.loadAccount(ctx, cmd.AccountID)
    if err != nil {
        return err
    }
    
    // Execute business logic
    if err := account.Deposit(cmd.Amount); err != nil {
        return err
    }
    
    // Save new events
    return h.EventStore.SaveEvents(ctx, account.ID, account.Events, account.Version-len(account.Events))
}
```

## 📊 **Read Model Patterns**

### **Projection Pattern**

Maintain read-optimized views from events:

```go
type OrderProjection struct {
    OrderRepo     OrderRepository     `inject:""`
    CustomerRepo  CustomerRepository  `inject:""`
    ProductRepo   ProductRepository   `inject:""`
}

func (p *OrderProjection) Handle(ctx context.Context, e OrderCreatedEvent) error {
    // Build denormalized read model
    customer, _ := p.CustomerRepo.GetByID(ctx, e.CustomerID)
    
    var items []OrderItemView
    for _, item := range e.Items {
        product, _ := p.ProductRepo.GetByID(ctx, item.ProductID)
        items = append(items, OrderItemView{
            ProductID:   item.ProductID,
            ProductName: product.Name,
            Quantity:    item.Quantity,
            Price:       item.Price,
            Total:       item.Price * float64(item.Quantity),
        })
    }
    
    orderView := OrderView{
        ID:           e.OrderID,
        CustomerID:   e.CustomerID,
        CustomerName: customer.Name,
        Items:        items,
        Total:        e.Total,
        Status:       "pending",
        Created:      e.Timestamp,
    }
    
    return p.OrderRepo.SaveView(ctx, orderView)
}

// Separate read models for different use cases
type OrderListView struct {
    ID           int       `json:"id"`
    CustomerName string    `json:"customer_name"`
    ItemCount    int       `json:"item_count"`
    Total        float64   `json:"total"`
    Status       string    `json:"status"`
    Created      time.Time `json:"created"`
}

type OrderDetailView struct {
    ID           int             `json:"id"`
    CustomerID   int             `json:"customer_id"`
    CustomerName string          `json:"customer_name"`
    Items        []OrderItemView `json:"items"`
    Total        float64         `json:"total"`
    Status       string          `json:"status"`
    Created      time.Time       `json:"created"`
    Updated      time.Time       `json:"updated"`
}
```

### **Materialized View Pattern**

Pre-computed aggregations for complex queries:

```go
type SalesReportProjection struct {
    ReportRepo SalesReportRepository `inject:""`
}

func (p *SalesReportProjection) Handle(ctx context.Context, e OrderCompletedEvent) error {
    // Update various materialized views
    
    // Daily sales
    if err := p.updateDailySales(ctx, e); err != nil {
        return err
    }
    
    // Product sales
    if err := p.updateProductSales(ctx, e); err != nil {
        return err
    }
    
    // Customer metrics
    if err := p.updateCustomerMetrics(ctx, e); err != nil {
        return err
    }
    
    return nil
}

func (p *SalesReportProjection) updateDailySales(ctx context.Context, e OrderCompletedEvent) error {
    date := e.CompletedAt.Format("2006-01-02")
    
    return p.ReportRepo.UpdateDailySales(ctx, DailySalesUpdate{
        Date:        date,
        OrderCount:  1,
        TotalAmount: e.Total,
        ItemCount:   len(e.Items),
    })
}

// Query handler for materialized views
type GetSalesReportHandler struct {
    ReportRepo SalesReportRepository `inject:""`
}

func (h *GetSalesReportHandler) Handle(ctx context.Context, q GetSalesReportQuery) (*SalesReport, error) {
    // Fast query from pre-computed data
    return h.ReportRepo.GetSalesReport(ctx, q.StartDate, q.EndDate, q.GroupBy)
}
```

## 🔄 **Saga Patterns**

### **Orchestrator Pattern**

Central coordinator manages the saga:

```go
type OrderProcessingSaga struct {
    SagaID      string
    OrderID     int
    CustomerID  int
    Steps       []SagaStep
    CurrentStep int
    Status      string
    Data        map[string]interface{}
}

type OrderSagaOrchestrator struct {
    SagaRepo    SagaRepository    `inject:""`
    PaymentSvc  PaymentService    `inject:""`
    InventorySvc InventoryService `inject:""`
    ShippingSvc ShippingService   `inject:""`
}

func (o *OrderSagaOrchestrator) Handle(ctx context.Context, e OrderCreatedEvent) error {
    saga := &OrderProcessingSaga{
        SagaID:     generateSagaID(),
        OrderID:    e.OrderID,
        CustomerID: e.CustomerID,
        Steps: []SagaStep{
            {Name: "reserve_inventory", Status: "pending"},
            {Name: "process_payment", Status: "pending"},
            {Name: "arrange_shipping", Status: "pending"},
            {Name: "confirm_order", Status: "pending"},
        },
        Status: "running",
        Data:   make(map[string]interface{}),
    }
    
    return o.executeNextStep(ctx, saga)
}

func (o *OrderSagaOrchestrator) executeNextStep(ctx context.Context, saga *OrderProcessingSaga) error {
    if saga.CurrentStep >= len(saga.Steps) {
        return o.completeSaga(ctx, saga)
    }
    
    step := saga.Steps[saga.CurrentStep]
    
    switch step.Name {
    case "reserve_inventory":
        return o.reserveInventory(ctx, saga)
    case "process_payment":
        return o.processPayment(ctx, saga)
    case "arrange_shipping":
        return o.arrangeShipping(ctx, saga)
    case "confirm_order":
        return o.confirmOrder(ctx, saga)
    }
    
    return nil
}

func (o *OrderSagaOrchestrator) reserveInventory(ctx context.Context, saga *OrderProcessingSaga) error {
    // Execute step
    reservationID, err := o.InventorySvc.ReserveItems(ctx, saga.OrderID)
    if err != nil {
        return o.compensateSaga(ctx, saga, err)
    }
    
    // Update saga state
    saga.Data["reservation_id"] = reservationID
    saga.Steps[saga.CurrentStep].Status = "completed"
    saga.CurrentStep++
    
    // Save and continue
    if err := o.SagaRepo.Save(ctx, saga); err != nil {
        return err
    }
    
    return o.executeNextStep(ctx, saga)
}

func (o *OrderSagaOrchestrator) compensateSaga(ctx context.Context, saga *OrderProcessingSaga, originalErr error) error {
    // Compensate completed steps in reverse order
    for i := saga.CurrentStep - 1; i >= 0; i-- {
        step := saga.Steps[i]
        if step.Status == "completed" {
            o.compensateStep(ctx, saga, step)
        }
    }
    
    saga.Status = "failed"
    o.SagaRepo.Save(ctx, saga)
    
    // Publish failure event
    return cqrs.PublishEvent(ctx, OrderProcessingFailedEvent{
        OrderID: saga.OrderID,
        Reason:  originalErr.Error(),
    })
}
```

### **Choreography Pattern**

Distributed coordination through events:

```go
// Each service listens to events and decides what to do

type InventoryService struct {
    InventoryRepo InventoryRepository `inject:""`
}

func (s *InventoryService) Handle(ctx context.Context, e OrderCreatedEvent) error {
    // Reserve inventory
    reservationID, err := s.reserveItems(ctx, e.OrderID, e.Items)
    if err != nil {
        // Publish failure event
        return cqrs.PublishEvent(ctx, InventoryReservationFailedEvent{
            OrderID: e.OrderID,
            Reason:  err.Error(),
        })
    }
    
    // Publish success event
    return cqrs.PublishEvent(ctx, InventoryReservedEvent{
        OrderID:       e.OrderID,
        ReservationID: reservationID,
        Items:         e.Items,
    })
}

type PaymentService struct {
    PaymentGateway PaymentGateway `inject:""`
}

func (s *PaymentService) Handle(ctx context.Context, e InventoryReservedEvent) error {
    // Process payment only after inventory is reserved
    paymentID, err := s.processPayment(ctx, e.OrderID)
    if err != nil {
        // Publish compensation event
        cqrs.PublishEvent(ctx, PaymentFailedEvent{
            OrderID: e.OrderID,
            Reason:  err.Error(),
        })
        
        // Trigger inventory release
        return cqrs.PublishEvent(ctx, ReleaseInventoryEvent{
            ReservationID: e.ReservationID,
        })
    }
    
    return cqrs.PublishEvent(ctx, PaymentProcessedEvent{
        OrderID:   e.OrderID,
        PaymentID: paymentID,
    })
}
```

## 🎭 **Decorator Patterns**

### **Cross-Cutting Concerns**

```go
// Security decorator
func AuthorizationDecorator(requiredRole string) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            userRole, ok := ctx.Value("userRole").(string)
            if !ok || !hasRole(userRole, requiredRole) {
                return nil, errors.New("insufficient permissions")
            }
            
            return next.Handle(ctx, message)
        })
    }
}

// Caching decorator
func CacheDecorator(ttl time.Duration) decorators.HandlerDecorator {
    cache := make(map[string]interface{})
    mutex := sync.RWMutex{}
    
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            // Only cache queries
            if !isQuery(message) {
                return next.Handle(ctx, message)
            }
            
            key := generateCacheKey(message)
            
            // Check cache
            mutex.RLock()
            if cached, exists := cache[key]; exists {
                mutex.RUnlock()
                return cached, nil
            }
            mutex.RUnlock()
            
            // Execute and cache
            result, err := next.Handle(ctx, message)
            if err == nil {
                mutex.Lock()
                cache[key] = result
                mutex.Unlock()
                
                // TTL cleanup
                go func() {
                    time.Sleep(ttl)
                    mutex.Lock()
                    delete(cache, key)
                    mutex.Unlock()
                }()
            }
            
            return result, err
        })
    }
}

// Composition of decorators
func SetupDecoratorChain() {
    manager := cqrs.NewCqrsManager()
    
    // Order matters - outer decorators execute first
    manager.AddDecorator(AuthorizationDecorator("user"))
    manager.AddDecorator(CacheDecorator(time.Minute * 5))
    manager.AddDecorator(MetricsDecorator())
    manager.AddDecorator(LoggingDecorator())
    
    cqrs.SetManager(manager)
}
```

### **Conditional Decorators**

```go
func ConditionalDecorator(condition func(ctx context.Context, message any) bool, decorator decorators.HandlerDecorator) decorators.HandlerDecorator {
    return func(next decorators.IHandlerDecorator) decorators.IHandlerDecorator {
        decoratedNext := decorator(next)
        
        return decorators.HandlerDecoratorFunc(func(ctx context.Context, message any) (any, error) {
            if condition(ctx, message) {
                return decoratedNext.Handle(ctx, message)
            }
            return next.Handle(ctx, message)
        })
    }
}

// Usage examples
func SetupConditionalDecorators() {
    manager := cqrs.NewCqrsManager()
    
    // Cache only expensive queries
    manager.AddDecorator(ConditionalDecorator(
        func(ctx context.Context, message any) bool {
            return isExpensiveQuery(message)
        },
        CacheDecorator(time.Minute * 10),
    ))
    
    // Audit only sensitive commands
    manager.AddDecorator(ConditionalDecorator(
        func(ctx context.Context, message any) bool {
            return isSensitiveCommand(message)
        },
        AuditDecorator(),
    ))
    
    cqrs.SetManager(manager)
}
```

## 📝 **Validation Patterns**

### **Layered Validation**

```go
// Structural validation (automatic via tags)
type CreateUserCommand struct {
    command.Base
    Email    string `validate:"required,email"`
    Name     string `validate:"required,min=2,max=50"`
    Age      int    `validate:"required,min=18,max=120"`
}

// Business validation
type CreateUserValidator struct {
    UserRepo      UserRepository      `inject:""`
    BlacklistRepo BlacklistRepository `inject:""`
    DomainService DomainService       `inject:""`
}

func (v *CreateUserValidator) Validate(ctx context.Context, cmd *CreateUserCommand) error {
    // Business rule validations
    if exists, err := v.UserRepo.EmailExists(ctx, cmd.Email); err != nil {
        return err
    } else if exists {
        return errors.New("email already registered")
    }
    
    if blacklisted, err := v.BlacklistRepo.IsBlacklisted(ctx, cmd.Email); err != nil {
        return err
    } else if blacklisted {
        return errors.New("email is blacklisted")
    }
    
    if !v.DomainService.IsValidEmailDomain(cmd.Email) {
        return errors.New("email domain not allowed")
    }
    
    return nil
}

// Contextual validation
type DepositMoneyValidator struct {
    AccountRepo AccountRepository `inject:""`
    ComplianceService ComplianceService `inject:""`
}

func (v *DepositMoneyValidator) Validate(ctx context.Context, cmd *DepositMoneyCommand) error {
    account, err := v.AccountRepo.GetByID(ctx, cmd.AccountID)
    if err != nil {
        return err
    }
    
    // Context-aware validation
    userID, _ := ctx.Value("userID").(int)
    if account.OwnerID != userID {
        return errors.New("cannot deposit to account you don't own")
    }
    
    if account.Status == "frozen" {
        return errors.New("account is frozen")
    }
    
    // Compliance checks based on amount
    if cmd.Amount > 10000 {
        if err := v.ComplianceService.CheckLargeDeposit(ctx, userID, cmd.Amount); err != nil {
            return fmt.Errorf("compliance check failed: %w", err)
        }
    }
    
    return nil
}
```

### **Validation Pipeline**

```go
type ValidationPipeline struct {
    validators []func(ctx context.Context, cmd interface{}) error
}

func (p *ValidationPipeline) Add(validator func(ctx context.Context, cmd interface{}) error) {
    p.validators = append(p.validators, validator)
}

func (p *ValidationPipeline) Validate(ctx context.Context, cmd interface{}) error {
    for _, validator := range p.validators {
        if err := validator(ctx, cmd); err != nil {
            return err
        }
    }
    return nil
}

// Complex validation composition
func SetupCreateOrderValidation() *ValidationPipeline {
    pipeline := &ValidationPipeline{}
    
    // Add validators in order
    pipeline.Add(validateStructuralConstraints)
    pipeline.Add(validateCustomerStatus)
    pipeline.Add(validateProductAvailability)
    pipeline.Add(validateOrderLimits)
    pipeline.Add(validatePaymentMethod)
    
    return pipeline
}
```

## 🔄 **Event Processing Patterns**

### **Event Handler Ordering**

```go
type OrderedEventProcessor struct {
    handlers map[string][]PrioritizedHandler
}

type PrioritizedHandler struct {
    Handler  IEventHandler
    Priority int // Lower number = higher priority
}

func (p *OrderedEventProcessor) RegisterHandler(eventType string, handler IEventHandler, priority int) {
    if p.handlers == nil {
        p.handlers = make(map[string][]PrioritizedHandler)
    }
    
    p.handlers[eventType] = append(p.handlers[eventType], PrioritizedHandler{
        Handler:  handler,
        Priority: priority,
    })
    
    // Sort by priority
    sort.Slice(p.handlers[eventType], func(i, j int) bool {
        return p.handlers[eventType][i].Priority < p.handlers[eventType][j].Priority
    })
}

func (p *OrderedEventProcessor) ProcessEvent(ctx context.Context, event IEvent) error {
    eventType := reflect.TypeOf(event).Name()
    handlers := p.handlers[eventType]
    
    for _, ph := range handlers {
        if err := ph.Handler.Handle(ctx, event); err != nil {
            // Decide on error handling strategy
            return err
        }
    }
    
    return nil
}

// Usage
func SetupEventHandlerPriorities() {
    processor := &OrderedEventProcessor{}
    
    // Critical handlers first
    processor.RegisterHandler("UserCreatedEvent", &AuditHandler{}, 1)
    processor.RegisterHandler("UserCreatedEvent", &SecurityHandler{}, 2)
    
    // Business logic handlers
    processor.RegisterHandler("UserCreatedEvent", &WelcomeEmailHandler{}, 10)
    processor.RegisterHandler("UserCreatedEvent", &AnalyticsHandler{}, 11)
    
    // Optional handlers last
    processor.RegisterHandler("UserCreatedEvent", &RecommendationHandler{}, 20)
}
```

### **Event Replay Pattern**

```go
type EventReplayService struct {
    EventStore EventStore `inject:""`
    Projections map[string]IEventHandler
}

func (s *EventReplayService) ReplayEvents(ctx context.Context, fromTimestamp time.Time, projection string) error {
    handler, exists := s.Projections[projection]
    if !exists {
        return fmt.Errorf("projection %s not found", projection)
    }
    
    // Get events from store
    events, err := s.EventStore.GetEventsFromTimestamp(ctx, fromTimestamp)
    if err != nil {
        return err
    }
    
    // Replay events in order
    for _, event := range events {
        if err := handler.Handle(ctx, event); err != nil {
            return fmt.Errorf("replay failed at event %s: %w", event.GetID(), err)
        }
    }
    
    return nil
}

// Projection versioning
type ProjectionVersion struct {
    Name    string
    Version int
    Handler IEventHandler
}

func (s *EventReplayService) UpgradeProjection(ctx context.Context, projection ProjectionVersion) error {
    // Clear existing projection
    if err := s.clearProjection(ctx, projection.Name); err != nil {
        return err
    }
    
    // Replay all events with new handler
    return s.ReplayEvents(ctx, time.Time{}, projection.Name)
}
```

## 🏷️ **Domain Event Patterns**

### **Event Enrichment**

```go
type EnrichedEvent struct {
    OriginalEvent IEvent
    Metadata      map[string]interface{}
    Context       EventContext
}

type EventContext struct {
    UserID        int
    TenantID      int
    CorrelationID string
    CausationID   string
    Timestamp     time.Time
    IPAddress     string
    UserAgent     string
}

type EventEnricher struct {
    UserRepo   UserRepository   `inject:""`
    TenantRepo TenantRepository `inject:""`
}

func (e *EventEnricher) EnrichEvent(ctx context.Context, event IEvent) (*EnrichedEvent, error) {
    userID, _ := ctx.Value("userID").(int)
    tenantID, _ := ctx.Value("tenantID").(int)
    
    // Gather additional context
    user, _ := e.UserRepo.GetByID(ctx, userID)
    tenant, _ := e.TenantRepo.GetByID(ctx, tenantID)
    
    enriched := &EnrichedEvent{
        OriginalEvent: event,
        Metadata: map[string]interface{}{
            "user_name":    user.Name,
            "user_email":   user.Email,
            "tenant_name":  tenant.Name,
            "event_source": "api",
        },
        Context: EventContext{
            UserID:        userID,
            TenantID:      tenantID,
            CorrelationID: getCorrelationID(ctx),
            CausationID:   getCausationID(ctx),
            Timestamp:     time.Now(),
            IPAddress:     getIPAddress(ctx),
            UserAgent:     getUserAgent(ctx),
        },
    }
    
    return enriched, nil
}
```

### **Event Versioning**

```go
type VersionedEvent struct {
    Version int         `json:"version"`
    Type    string      `json:"type"`
    Data    interface{} `json:"data"`
}

type EventMigrator struct {
    migrations map[string]map[int]func(interface{}) (interface{}, error)
}

func (m *EventMigrator) MigrateEvent(eventType string, version int, data interface{}) (interface{}, error) {
    typeMigrations, exists := m.migrations[eventType]
    if !exists {
        return data, nil
    }
    
    currentData := data
    for v := version; v < m.getLatestVersion(eventType); v++ {
        migration, exists := typeMigrations[v+1]
        if !exists {
            continue
        }
        
        migrated, err := migration(currentData)
        if err != nil {
            return nil, fmt.Errorf("migration from v%d to v%d failed: %w", v, v+1, err)
        }
        currentData = migrated
    }
    
    return currentData, nil
}

// Example migration
func init() {
    migrator := &EventMigrator{
        migrations: map[string]map[int]func(interface{}) (interface{}, error){
            "UserCreatedEvent": {
                2: func(data interface{}) (interface{}, error) {
                    // v1 -> v2: Add email verification status
                    v1 := data.(UserCreatedEventV1)
                    return UserCreatedEventV2{
                        UserID:           v1.UserID,
                        Name:             v1.Name,
                        Email:            v1.Email,
                        EmailVerified:    false, // Default for migrated events
                        RegistrationDate: v1.CreatedAt,
                    }, nil
                },
                3: func(data interface{}) (interface{}, error) {
                    // v2 -> v3: Add user preferences
                    v2 := data.(UserCreatedEventV2)
                    return UserCreatedEventV3{
                        UserID:           v2.UserID,
                        Name:             v2.Name,
                        Email:            v2.Email,
                        EmailVerified:    v2.EmailVerified,
                        RegistrationDate: v2.RegistrationDate,
                        Preferences:      map[string]interface{}{}, // Empty default
                    }, nil
                },
            },
        },
    }
}
```

## 🧩 **Integration Patterns**

### **Anti-Corruption Layer**

```go
type LegacySystemAdapter struct {
    LegacyClient LegacySystemClient `inject:""`
}

func (a *LegacySystemAdapter) Handle(ctx context.Context, cmd *SyncUserToLegacyCommand) error {
    // Translate modern domain model to legacy format
    legacyUser := LegacyUser{
        ID:        fmt.Sprintf("USR_%d", cmd.UserID),
        FullName:  cmd.User.FirstName + " " + cmd.User.LastName,
        EmailAddr: cmd.User.Email,
        Status:    a.translateStatus(cmd.User.Status),
        CreateDt:  cmd.User.CreatedAt.Format("2006-01-02 15:04:05"),
    }
    
    // Call legacy system
    response, err := a.LegacyClient.CreateOrUpdateUser(legacyUser)
    if err != nil {
        return fmt.Errorf("legacy system error: %w", err)
    }
    
    // Translate response back if needed
    if response.Status == "ERROR" {
        return errors.New(response.ErrorMessage)
    }
    
    return nil
}

func (a *LegacySystemAdapter) translateStatus(modernStatus string) string {
    switch modernStatus {
    case "active":
        return "A"
    case "inactive":
        return "I"
    case "suspended":
        return "S"
    default:
        return "P" // Pending
    }
}
```

### **Gateway Pattern**

```go
type ExternalServiceGateway struct {
    EmailService    EmailServiceClient    `inject:""`
    PaymentService  PaymentServiceClient  `inject:""`
    AnalyticsService AnalyticsServiceClient `inject:""`
    CircuitBreaker  CircuitBreaker        `inject:""`
    Cache          Cache                  `inject:""`
}

func (g *ExternalServiceGateway) SendEmail(ctx context.Context, req EmailRequest) error {
    return g.CircuitBreaker.Execute(func() error {
        return g.EmailService.Send(ctx, req)
    })
}

func (g *ExternalServiceGateway) ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
    // Check cache for duplicate requests
    cacheKey := fmt.Sprintf("payment:%s", req.IdempotencyKey)
    if cached, err := g.Cache.Get(ctx, cacheKey); err == nil {
        return cached.(*PaymentResponse), nil
    }
    
    var response *PaymentResponse
    err := g.CircuitBreaker.Execute(func() error {
        var err error
        response, err = g.PaymentService.Process(ctx, req)
        return err
    })
    
    if err == nil {
        // Cache successful response
        g.Cache.Set(ctx, cacheKey, response, time.Minute*5)
    }
    
    return response, err
}
```

---

These patterns provide proven approaches for:

- **Clean Architecture** with clear separation of concerns
- **Scalability** through event-driven design
- **Reliability** with saga patterns and error handling
- **Maintainability** through decorator composition
- **Integration** with external systems and legacy code

For implementation examples, see [Real-World Examples](./examples.md) and [Integration Guide](./integrations.md). 