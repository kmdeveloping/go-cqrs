# Real-World Examples

Comprehensive examples demonstrating go-cqrs in production scenarios.

## 🏢 **E-Commerce System Example**

### **Domain Models**

```go
// User domain
type User struct {
    ID       int       `json:"id"`
    Email    string    `json:"email"`
    Name     string    `json:"name"`
    Status   string    `json:"status"`
    Created  time.Time `json:"created"`
    Updated  time.Time `json:"updated"`
}

// Product domain
type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
    Stock       int     `json:"stock"`
    CategoryID  int     `json:"category_id"`
}

// Order domain
type Order struct {
    ID         int         `json:"id"`
    UserID     int         `json:"user_id"`
    Items      []OrderItem `json:"items"`
    Total      float64     `json:"total"`
    Status     string      `json:"status"`
    Created    time.Time   `json:"created"`
}

type OrderItem struct {
    ProductID int     `json:"product_id"`
    Quantity  int     `json:"quantity"`
    Price     float64 `json:"price"`
}
```

### **Commands**

```go
// User commands
type CreateUserCommand struct {
    command.Base
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2"`
    Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserCommand struct {
    command.Base
    UserID int    `json:"user_id" validate:"required"`
    Name   string `json:"name" validate:"required,min=2"`
    Email  string `json:"email" validate:"required,email"`
}

type DeactivateUserCommand struct {
    command.Base
    UserID int    `json:"user_id" validate:"required"`
    Reason string `json:"reason" validate:"required"`
}

// Product commands
type CreateProductCommand struct {
    command.Base
    Name        string  `json:"name" validate:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Stock       int     `json:"stock" validate:"required,gte=0"`
    CategoryID  int     `json:"category_id" validate:"required"`
}

type UpdateStockCommand struct {
    command.Base
    ProductID int `json:"product_id" validate:"required"`
    Quantity  int `json:"quantity" validate:"required"`
}

// Order commands
type CreateOrderCommand struct {
    command.BaseWithResult
    UserID int                    `json:"user_id" validate:"required"`
    Items  []CreateOrderItemData `json:"items" validate:"required,min=1"`
}

type CreateOrderItemData struct {
    ProductID int `json:"product_id" validate:"required"`
    Quantity  int `json:"quantity" validate:"required,gt=0"`
}

type CancelOrderCommand struct {
    command.Base
    OrderID int    `json:"order_id" validate:"required"`
    Reason  string `json:"reason" validate:"required"`
}
```

### **Queries**

```go
// User queries
type GetUserQuery struct {
    query.Base
    UserID int `json:"user_id" validate:"required"`
}

type GetUserByEmailQuery struct {
    query.Base
    Email string `json:"email" validate:"required,email"`
}

type ListUsersQuery struct {
    query.Base
    Page     int    `json:"page" validate:"min=1"`
    PageSize int    `json:"page_size" validate:"min=1,max=100"`
    Status   string `json:"status,omitempty"`
}

// Product queries
type GetProductQuery struct {
    query.Base
    ProductID int `json:"product_id" validate:"required"`
}

type SearchProductsQuery struct {
    query.Base
    SearchTerm string  `json:"search_term"`
    CategoryID *int    `json:"category_id,omitempty"`
    MinPrice   *float64 `json:"min_price,omitempty"`
    MaxPrice   *float64 `json:"max_price,omitempty"`
    Page       int     `json:"page" validate:"min=1"`
    PageSize   int     `json:"page_size" validate:"min=1,max=100"`
}

// Order queries
type GetOrderQuery struct {
    query.Base
    OrderID int `json:"order_id" validate:"required"`
}

type GetUserOrdersQuery struct {
    query.Base
    UserID   int    `json:"user_id" validate:"required"`
    Status   string `json:"status,omitempty"`
    Page     int    `json:"page" validate:"min=1"`
    PageSize int    `json:"page_size" validate:"min=1,max=100"`
}
```

### **Events**

```go
// User events
type UserCreatedEvent struct {
    event.Base
    UserID int    `json:"user_id"`
    Email  string `json:"email"`
    Name   string `json:"name"`
}

type UserUpdatedEvent struct {
    event.Base
    UserID   int    `json:"user_id"`
    OldEmail string `json:"old_email"`
    NewEmail string `json:"new_email"`
    OldName  string `json:"old_name"`
    NewName  string `json:"new_name"`
}

type UserDeactivatedEvent struct {
    event.Base
    UserID int    `json:"user_id"`
    Reason string `json:"reason"`
}

// Product events
type ProductCreatedEvent struct {
    event.Base
    ProductID   int     `json:"product_id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    CategoryID  int     `json:"category_id"`
}

type StockUpdatedEvent struct {
    event.Base
    ProductID   int `json:"product_id"`
    OldStock    int `json:"old_stock"`
    NewStock    int `json:"new_stock"`
    Difference  int `json:"difference"`
}

type LowStockAlertEvent struct {
    event.Base
    ProductID   int    `json:"product_id"`
    ProductName string `json:"product_name"`
    CurrentStock int   `json:"current_stock"`
    Threshold   int    `json:"threshold"`
}

// Order events
type OrderCreatedEvent struct {
    event.Base
    OrderID int     `json:"order_id"`
    UserID  int     `json:"user_id"`
    Total   float64 `json:"total"`
    Items   []OrderItem `json:"items"`
}

type OrderCancelledEvent struct {
    event.Base
    OrderID int    `json:"order_id"`
    UserID  int    `json:"user_id"`
    Reason  string `json:"reason"`
}

type PaymentProcessedEvent struct {
    event.Base
    OrderID       int     `json:"order_id"`
    PaymentID     string  `json:"payment_id"`
    Amount        float64 `json:"amount"`
    PaymentMethod string  `json:"payment_method"`
}
```

### **Command Handlers**

```go
type CreateUserHandler struct {
    UserRepo     UserRepository     `inject:""`
    EmailService EmailService      `inject:""`
    Hasher       PasswordHasher    `inject:""`
    Logger       *logrus.Logger    `inject:""`
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) error {
    // Check if user already exists
    existingUser, _ := h.UserRepo.GetByEmail(ctx, cmd.Email)
    if existingUser != nil {
        return errors.New("user with this email already exists")
    }
    
    // Hash password
    hashedPassword, err := h.Hasher.Hash(cmd.Password)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    
    // Create user
    user := &User{
        Email:   cmd.Email,
        Name:    cmd.Name,
        Status:  "active",
        Created: time.Now(),
        Updated: time.Now(),
    }
    
    userID, err := h.UserRepo.Create(ctx, user, hashedPassword)
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    // Publish event
    if err := cqrs.PublishEvent(ctx, UserCreatedEvent{
        UserID: userID,
        Email:  cmd.Email,
        Name:   cmd.Name,
    }); err != nil {
        h.Logger.WithError(err).Error("Failed to publish UserCreatedEvent")
        // Don't fail the command for event publishing errors
    }
    
    h.Logger.WithFields(logrus.Fields{
        "user_id": userID,
        "email":   cmd.Email,
    }).Info("User created successfully")
    
    return nil
}

type CreateOrderHandler struct {
    OrderRepo   OrderRepository   `inject:""`
    ProductRepo ProductRepository `inject:""`
    UserRepo    UserRepository    `inject:""`
    Logger      *logrus.Logger    `inject:""`
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) error {
    // Validate user exists
    user, err := h.UserRepo.GetByID(ctx, cmd.UserID)
    if err != nil {
        return fmt.Errorf("invalid user: %w", err)
    }
    if user.Status != "active" {
        return errors.New("user account is not active")
    }
    
    // Validate products and calculate total
    var orderItems []OrderItem
    var total float64
    
    for _, item := range cmd.Items {
        product, err := h.ProductRepo.GetByID(ctx, item.ProductID)
        if err != nil {
            return fmt.Errorf("invalid product %d: %w", item.ProductID, err)
        }
        
        if product.Stock < item.Quantity {
            return fmt.Errorf("insufficient stock for product %s", product.Name)
        }
        
        itemTotal := product.Price * float64(item.Quantity)
        total += itemTotal
        
        orderItems = append(orderItems, OrderItem{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     product.Price,
        })
    }
    
    // Create order
    order := &Order{
        UserID:  cmd.UserID,
        Items:   orderItems,
        Total:   total,
        Status:  "pending",
        Created: time.Now(),
    }
    
    orderID, err := h.OrderRepo.Create(ctx, order)
    if err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    
    // Update stock for each product
    for _, item := range cmd.Items {
        if err := cqrs.ExecuteCommand(ctx, &UpdateStockCommand{
            ProductID: item.ProductID,
            Quantity:  -item.Quantity, // Decrease stock
        }); err != nil {
            h.Logger.WithError(err).Error("Failed to update stock")
            // In production, you might want to implement compensation
        }
    }
    
    // Publish event
    if err := cqrs.PublishEvent(ctx, OrderCreatedEvent{
        OrderID: orderID,
        UserID:  cmd.UserID,
        Total:   total,
        Items:   orderItems,
    }); err != nil {
        h.Logger.WithError(err).Error("Failed to publish OrderCreatedEvent")
    }
    
    // Set result for command with result
    cmd.Result = orderID
    
    h.Logger.WithFields(logrus.Fields{
        "order_id": orderID,
        "user_id":  cmd.UserID,
        "total":    total,
    }).Info("Order created successfully")
    
    return nil
}
```

### **Query Handlers**

```go
type GetUserHandler struct {
    UserRepo UserRepository `inject:""`
    Cache    CacheService   `inject:""`
    Logger   *logrus.Logger `inject:""`
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("user:%d", q.UserID)
    if cached, err := h.Cache.Get(ctx, cacheKey); err == nil {
        var user User
        if err := json.Unmarshal(cached, &user); err == nil {
            h.Logger.WithField("user_id", q.UserID).Debug("Cache hit for user")
            return &user, nil
        }
    }
    
    // Get from database
    user, err := h.UserRepo.GetByID(ctx, q.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    // Cache the result
    if userData, err := json.Marshal(user); err == nil {
        h.Cache.Set(ctx, cacheKey, userData, time.Minute*15)
    }
    
    return user, nil
}

type SearchProductsHandler struct {
    ProductRepo ProductRepository `inject:""`
    Logger      *logrus.Logger    `inject:""`
}

func (h *SearchProductsHandler) Handle(ctx context.Context, q SearchProductsQuery) (*ProductSearchResult, error) {
    filters := ProductFilters{
        SearchTerm: q.SearchTerm,
        CategoryID: q.CategoryID,
        MinPrice:   q.MinPrice,
        MaxPrice:   q.MaxPrice,
    }
    
    pagination := Pagination{
        Page:     q.Page,
        PageSize: q.PageSize,
    }
    
    result, err := h.ProductRepo.Search(ctx, filters, pagination)
    if err != nil {
        return nil, fmt.Errorf("failed to search products: %w", err)
    }
    
    h.Logger.WithFields(logrus.Fields{
        "search_term": q.SearchTerm,
        "results":     len(result.Products),
        "page":        q.Page,
    }).Info("Product search completed")
    
    return result, nil
}

type ProductSearchResult struct {
    Products    []Product `json:"products"`
    Total       int       `json:"total"`
    Page        int       `json:"page"`
    PageSize    int       `json:"page_size"`
    TotalPages  int       `json:"total_pages"`
}
```

### **Event Handlers**

```go
type UserCreatedEventHandler struct {
    EmailService     EmailService     `inject:""`
    WelcomeService   WelcomeService   `inject:""`
    AnalyticsService AnalyticsService `inject:""`
    Logger           *logrus.Logger   `inject:""`
}

func (h *UserCreatedEventHandler) Handle(ctx context.Context, e UserCreatedEvent) error {
    // Send welcome email
    if err := h.EmailService.SendWelcomeEmail(ctx, e.Email, e.Name); err != nil {
        h.Logger.WithError(err).Error("Failed to send welcome email")
        // Don't fail for email errors
    }
    
    // Create welcome package
    if err := h.WelcomeService.CreateWelcomePackage(ctx, e.UserID); err != nil {
        h.Logger.WithError(err).Error("Failed to create welcome package")
    }
    
    // Track analytics
    if err := h.AnalyticsService.TrackUserRegistration(ctx, e.UserID, e.Email); err != nil {
        h.Logger.WithError(err).Error("Failed to track user registration")
    }
    
    h.Logger.WithFields(logrus.Fields{
        "user_id": e.UserID,
        "email":   e.Email,
    }).Info("User created event processed")
    
    return nil
}

type OrderCreatedEventHandler struct {
    PaymentService     PaymentService     `inject:""`
    InventoryService   InventoryService   `inject:""`
    NotificationService NotificationService `inject:""`
    Logger             *logrus.Logger     `inject:""`
}

func (h *OrderCreatedEventHandler) Handle(ctx context.Context, e OrderCreatedEvent) error {
    // Reserve inventory
    if err := h.InventoryService.ReserveItems(ctx, e.OrderID, e.Items); err != nil {
        h.Logger.WithError(err).Error("Failed to reserve inventory")
        // In production, you might want to cancel the order
        return fmt.Errorf("failed to reserve inventory: %w", err)
    }
    
    // Send order confirmation
    if err := h.NotificationService.SendOrderConfirmation(ctx, e.UserID, e.OrderID); err != nil {
        h.Logger.WithError(err).Error("Failed to send order confirmation")
        // Don't fail for notification errors
    }
    
    // Process payment (async)
    go func() {
        paymentCtx := context.Background()
        if err := h.PaymentService.ProcessPayment(paymentCtx, e.OrderID, e.Total); err != nil {
            h.Logger.WithError(err).Error("Failed to process payment")
            // In production, implement proper error handling
        }
    }()
    
    h.Logger.WithFields(logrus.Fields{
        "order_id": e.OrderID,
        "user_id":  e.UserID,
        "total":    e.Total,
    }).Info("Order created event processed")
    
    return nil
}

type LowStockAlertEventHandler struct {
    EmailService        EmailService        `inject:""`
    InventoryService    InventoryService    `inject:""`
    NotificationService NotificationService `inject:""`
    Logger              *logrus.Logger      `inject:""`
}

func (h *LowStockAlertEventHandler) Handle(ctx context.Context, e LowStockAlertEvent) error {
    // Notify inventory team
    if err := h.EmailService.SendLowStockAlert(ctx, e.ProductID, e.ProductName, e.CurrentStock); err != nil {
        h.Logger.WithError(err).Error("Failed to send low stock alert email")
    }
    
    // Auto-reorder if configured
    if h.InventoryService.ShouldAutoReorder(ctx, e.ProductID) {
        if err := h.InventoryService.CreateReorderRequest(ctx, e.ProductID); err != nil {
            h.Logger.WithError(err).Error("Failed to create auto-reorder request")
        } else {
            h.Logger.WithField("product_id", e.ProductID).Info("Auto-reorder request created")
        }
    }
    
    // Send push notification to managers
    if err := h.NotificationService.SendPushNotification(ctx, "inventory_managers", fmt.Sprintf(
        "Low stock alert: %s has only %d items left", e.ProductName, e.CurrentStock,
    )); err != nil {
        h.Logger.WithError(err).Error("Failed to send push notification")
    }
    
    return nil
}
```

### **Validators**

```go
type CreateUserValidator struct {
    UserRepo   UserRepository `inject:""`
    BlacklistRepo BlacklistRepository `inject:""`
}

func (v *CreateUserValidator) Validate(ctx context.Context, cmd *CreateUserCommand) error {
    // Check email format (handled by struct tags, but additional checks)
    if !isValidEmailDomain(cmd.Email) {
        return errors.New("email domain is not allowed")
    }
    
    // Check if email is blacklisted
    if blacklisted, err := v.BlacklistRepo.IsEmailBlacklisted(ctx, cmd.Email); err != nil {
        return fmt.Errorf("failed to check email blacklist: %w", err)
    } else if blacklisted {
        return errors.New("email is blacklisted")
    }
    
    // Check password strength
    if err := validatePasswordStrength(cmd.Password); err != nil {
        return fmt.Errorf("password validation failed: %w", err)
    }
    
    // Check for profanity in name
    if containsProfanity(cmd.Name) {
        return errors.New("name contains inappropriate content")
    }
    
    return nil
}

type CreateOrderValidator struct {
    UserRepo    UserRepository    `inject:""`
    ProductRepo ProductRepository `inject:""`
    OrderRepo   OrderRepository   `inject:""`
}

func (v *CreateOrderValidator) Validate(ctx context.Context, cmd *CreateOrderCommand) error {
    // Check if user can place orders
    user, err := v.UserRepo.GetByID(ctx, cmd.UserID)
    if err != nil {
        return fmt.Errorf("invalid user: %w", err)
    }
    
    if user.Status != "active" {
        return errors.New("user account is not active")
    }
    
    // Check order limits
    todayOrders, err := v.OrderRepo.GetUserOrdersToday(ctx, cmd.UserID)
    if err != nil {
        return fmt.Errorf("failed to check order limits: %w", err)
    }
    
    if len(todayOrders) >= 10 { // Max 10 orders per day
        return errors.New("daily order limit exceeded")
    }
    
    // Validate each order item
    for _, item := range cmd.Items {
        product, err := v.ProductRepo.GetByID(ctx, item.ProductID)
        if err != nil {
            return fmt.Errorf("invalid product %d: %w", item.ProductID, err)
        }
        
        if product.Stock < item.Quantity {
            return fmt.Errorf("insufficient stock for product %s", product.Name)
        }
        
        if item.Quantity > 100 { // Max 100 items per product
            return fmt.Errorf("quantity limit exceeded for product %s", product.Name)
        }
    }
    
    return nil
}
```

## 🏦 **Banking System Example**

### **Account Management**

```go
// Commands
type CreateAccountCommand struct {
    command.Base
    CustomerID  int     `json:"customer_id" validate:"required"`
    AccountType string  `json:"account_type" validate:"required,oneof=checking savings"`
    InitialDeposit float64 `json:"initial_deposit" validate:"min=0"`
}

type TransferMoneyCommand struct {
    command.Base
    FromAccountID int     `json:"from_account_id" validate:"required"`
    ToAccountID   int     `json:"to_account_id" validate:"required"`
    Amount        float64 `json:"amount" validate:"required,gt=0"`
    Description   string  `json:"description"`
}

type FreezeAccountCommand struct {
    command.Base
    AccountID int    `json:"account_id" validate:"required"`
    Reason    string `json:"reason" validate:"required"`
}

// Queries
type GetAccountBalanceQuery struct {
    query.Base
    AccountID int `json:"account_id" validate:"required"`
}

type GetTransactionHistoryQuery struct {
    query.Base
    AccountID int       `json:"account_id" validate:"required"`
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
    Page      int       `json:"page" validate:"min=1"`
    PageSize  int       `json:"page_size" validate:"min=1,max=100"`
}

// Events
type AccountCreatedEvent struct {
    event.Base
    AccountID      int     `json:"account_id"`
    CustomerID     int     `json:"customer_id"`
    AccountType    string  `json:"account_type"`
    InitialDeposit float64 `json:"initial_deposit"`
}

type MoneyTransferredEvent struct {
    event.Base
    TransactionID int     `json:"transaction_id"`
    FromAccountID int     `json:"from_account_id"`
    ToAccountID   int     `json:"to_account_id"`
    Amount        float64 `json:"amount"`
    Description   string  `json:"description"`
}

type AccountFrozenEvent struct {
    event.Base
    AccountID int    `json:"account_id"`
    Reason    string `json:"reason"`
    FrozenBy  int    `json:"frozen_by"`
}
```

### **Transaction Handler with Saga Pattern**

```go
type TransferMoneyHandler struct {
    AccountRepo     AccountRepository     `inject:""`
    TransactionRepo TransactionRepository `inject:""`
    AuditService    AuditService         `inject:""`
    FraudService    FraudService         `inject:""`
    Logger          *logrus.Logger       `inject:""`
}

func (h *TransferMoneyHandler) Handle(ctx context.Context, cmd *TransferMoneyCommand) error {
    // Fraud detection
    if suspicious, err := h.FraudService.CheckTransaction(ctx, cmd.FromAccountID, cmd.ToAccountID, cmd.Amount); err != nil {
        return fmt.Errorf("fraud check failed: %w", err)
    } else if suspicious {
        // Publish suspicious activity event
        cqrs.PublishEvent(ctx, SuspiciousActivityEvent{
            AccountID: cmd.FromAccountID,
            Amount:    cmd.Amount,
            Reason:    "unusual_transfer_pattern",
        })
        return errors.New("transaction blocked due to suspicious activity")
    }
    
    // Begin transaction saga
    saga := NewTransferSaga(cmd.FromAccountID, cmd.ToAccountID, cmd.Amount, cmd.Description)
    
    // Step 1: Reserve funds
    if err := h.reserveFunds(ctx, saga); err != nil {
        h.auditFailure(ctx, saga, "reserve_funds", err)
        return fmt.Errorf("failed to reserve funds: %w", err)
    }
    
    // Step 2: Create transaction record
    if err := h.createTransaction(ctx, saga); err != nil {
        h.compensateReserveFunds(ctx, saga)
        h.auditFailure(ctx, saga, "create_transaction", err)
        return fmt.Errorf("failed to create transaction: %w", err)
    }
    
    // Step 3: Complete transfer
    if err := h.completeTransfer(ctx, saga); err != nil {
        h.compensateTransaction(ctx, saga)
        h.compensateReserveFunds(ctx, saga)
        h.auditFailure(ctx, saga, "complete_transfer", err)
        return fmt.Errorf("failed to complete transfer: %w", err)
    }
    
    // Publish success event
    if err := cqrs.PublishEvent(ctx, MoneyTransferredEvent{
        TransactionID: saga.TransactionID,
        FromAccountID: cmd.FromAccountID,
        ToAccountID:   cmd.ToAccountID,
        Amount:        cmd.Amount,
        Description:   cmd.Description,
    }); err != nil {
        h.Logger.WithError(err).Error("Failed to publish MoneyTransferredEvent")
    }
    
    h.auditSuccess(ctx, saga)
    return nil
}

type TransferSaga struct {
    FromAccountID int
    ToAccountID   int
    Amount        float64
    Description   string
    TransactionID int
    ReservationID string
    Steps         []string
}

func (h *TransferMoneyHandler) reserveFunds(ctx context.Context, saga *TransferSaga) error {
    reservationID, err := h.AccountRepo.ReserveFunds(ctx, saga.FromAccountID, saga.Amount)
    if err != nil {
        return err
    }
    saga.ReservationID = reservationID
    saga.Steps = append(saga.Steps, "funds_reserved")
    return nil
}

func (h *TransferMoneyHandler) compensateReserveFunds(ctx context.Context, saga *TransferSaga) {
    if saga.ReservationID != "" {
        h.AccountRepo.ReleaseFunds(ctx, saga.ReservationID)
    }
}
```

## 📚 **Library Management System**

### **Book and Member Management**

```go
// Domain Models
type Book struct {
    ID          int    `json:"id"`
    ISBN        string `json:"isbn"`
    Title       string `json:"title"`
    Author      string `json:"author"`
    Category    string `json:"category"`
    Available   bool   `json:"available"`
    Location    string `json:"location"`
}

type Member struct {
    ID            int       `json:"id"`
    MemberNumber  string    `json:"member_number"`
    Name          string    `json:"name"`
    Email         string    `json:"email"`
    MembershipType string   `json:"membership_type"`
    JoinDate      time.Time `json:"join_date"`
    Status        string    `json:"status"`
}

type BookLoan struct {
    ID         int       `json:"id"`
    BookID     int       `json:"book_id"`
    MemberID   int       `json:"member_id"`
    LoanDate   time.Time `json:"loan_date"`
    DueDate    time.Time `json:"due_date"`
    ReturnDate *time.Time `json:"return_date,omitempty"`
    Status     string    `json:"status"`
}

// Commands
type BorrowBookCommand struct {
    command.Base
    BookID   int `json:"book_id" validate:"required"`
    MemberID int `json:"member_id" validate:"required"`
}

type ReturnBookCommand struct {
    command.Base
    LoanID int `json:"loan_id" validate:"required"`
}

type ReserveBookCommand struct {
    command.Base
    BookID   int `json:"book_id" validate:"required"`
    MemberID int `json:"member_id" validate:"required"`
}

// Queries
type SearchBooksQuery struct {
    query.Base
    SearchTerm string `json:"search_term"`
    Category   string `json:"category,omitempty"`
    Available  *bool  `json:"available,omitempty"`
    Page       int    `json:"page" validate:"min=1"`
    PageSize   int    `json:"page_size" validate:"min=1,max=50"`
}

type GetMemberLoansQuery struct {
    query.Base
    MemberID int    `json:"member_id" validate:"required"`
    Status   string `json:"status,omitempty"`
}

// Handlers with business rules
type BorrowBookHandler struct {
    BookRepo   BookRepository   `inject:""`
    MemberRepo MemberRepository `inject:""`
    LoanRepo   LoanRepository   `inject:""`
    Logger     *logrus.Logger   `inject:""`
}

func (h *BorrowBookHandler) Handle(ctx context.Context, cmd *BorrowBookCommand) error {
    // Check member status and loan limits
    member, err := h.MemberRepo.GetByID(ctx, cmd.MemberID)
    if err != nil {
        return fmt.Errorf("invalid member: %w", err)
    }
    
    if member.Status != "active" {
        return errors.New("member account is not active")
    }
    
    // Check current loans
    currentLoans, err := h.LoanRepo.GetActiveLoansByMember(ctx, cmd.MemberID)
    if err != nil {
        return fmt.Errorf("failed to check current loans: %w", err)
    }
    
    // Apply membership type limits
    maxLoans := h.getMaxLoansForMembershipType(member.MembershipType)
    if len(currentLoans) >= maxLoans {
        return fmt.Errorf("loan limit exceeded for %s membership", member.MembershipType)
    }
    
    // Check if member has overdue books
    for _, loan := range currentLoans {
        if loan.DueDate.Before(time.Now()) {
            return errors.New("cannot borrow new books with overdue items")
        }
    }
    
    // Check book availability
    book, err := h.BookRepo.GetByID(ctx, cmd.BookID)
    if err != nil {
        return fmt.Errorf("invalid book: %w", err)
    }
    
    if !book.Available {
        return errors.New("book is not available")
    }
    
    // Calculate due date based on membership type
    loanPeriod := h.getLoanPeriodForMembershipType(member.MembershipType)
    dueDate := time.Now().Add(loanPeriod)
    
    // Create loan
    loan := &BookLoan{
        BookID:   cmd.BookID,
        MemberID: cmd.MemberID,
        LoanDate: time.Now(),
        DueDate:  dueDate,
        Status:   "active",
    }
    
    loanID, err := h.LoanRepo.Create(ctx, loan)
    if err != nil {
        return fmt.Errorf("failed to create loan: %w", err)
    }
    
    // Mark book as unavailable
    if err := h.BookRepo.SetAvailability(ctx, cmd.BookID, false); err != nil {
        // Compensate by deleting the loan
        h.LoanRepo.Delete(ctx, loanID)
        return fmt.Errorf("failed to update book availability: %w", err)
    }
    
    // Publish event
    if err := cqrs.PublishEvent(ctx, BookBorrowedEvent{
        LoanID:   loanID,
        BookID:   cmd.BookID,
        MemberID: cmd.MemberID,
        DueDate:  dueDate,
    }); err != nil {
        h.Logger.WithError(err).Error("Failed to publish BookBorrowedEvent")
    }
    
    return nil
}

func (h *BorrowBookHandler) getMaxLoansForMembershipType(membershipType string) int {
    switch membershipType {
    case "premium":
        return 10
    case "standard":
        return 5
    case "student":
        return 3
    default:
        return 2
    }
}

func (h *BorrowBookHandler) getLoanPeriodForMembershipType(membershipType string) time.Duration {
    switch membershipType {
    case "premium":
        return time.Hour * 24 * 30 // 30 days
    case "standard":
        return time.Hour * 24 * 21 // 21 days
    case "student":
        return time.Hour * 24 * 14 // 14 days
    default:
        return time.Hour * 24 * 7  // 7 days
    }
}
```

## 🎓 **University Course Management**

### **Student Enrollment System**

```go
// Commands for complex enrollment business logic
type EnrollStudentCommand struct {
    command.Base
    StudentID    int      `json:"student_id" validate:"required"`
    CourseID     int      `json:"course_id" validate:"required"`
    SemesterID   int      `json:"semester_id" validate:"required"`
    Prerequisites []int   `json:"prerequisites"`
}

type DropCourseCommand struct {
    command.Base
    EnrollmentID int    `json:"enrollment_id" validate:"required"`
    Reason       string `json:"reason" validate:"required"`
}

type GradeStudentCommand struct {
    command.Base
    EnrollmentID int    `json:"enrollment_id" validate:"required"`
    Grade        string `json:"grade" validate:"required,oneof=A B C D F"`
    Comments     string `json:"comments"`
}

// Complex validation with business rules
type EnrollStudentValidator struct {
    StudentRepo      StudentRepository      `inject:""`
    CourseRepo       CourseRepository       `inject:""`
    EnrollmentRepo   EnrollmentRepository   `inject:""`
    PrerequisiteRepo PrerequisiteRepository `inject:""`
}

func (v *EnrollStudentValidator) Validate(ctx context.Context, cmd *EnrollStudentCommand) error {
    // Check student eligibility
    student, err := v.StudentRepo.GetByID(ctx, cmd.StudentID)
    if err != nil {
        return fmt.Errorf("invalid student: %w", err)
    }
    
    if student.Status != "active" {
        return errors.New("student is not in active status")
    }
    
    if student.AccountBalance < 0 {
        return errors.New("student has outstanding balance")
    }
    
    // Check course capacity and timing
    course, err := v.CourseRepo.GetByID(ctx, cmd.CourseID)
    if err != nil {
        return fmt.Errorf("invalid course: %w", err)
    }
    
    currentEnrollments, err := v.EnrollmentRepo.GetEnrollmentCount(ctx, cmd.CourseID, cmd.SemesterID)
    if err != nil {
        return fmt.Errorf("failed to check enrollment count: %w", err)
    }
    
    if currentEnrollments >= course.MaxCapacity {
        return errors.New("course is at maximum capacity")
    }
    
    // Check schedule conflicts
    studentSchedule, err := v.EnrollmentRepo.GetStudentSchedule(ctx, cmd.StudentID, cmd.SemesterID)
    if err != nil {
        return fmt.Errorf("failed to check student schedule: %w", err)
    }
    
    if hasTimeConflict(studentSchedule, course.Schedule) {
        return errors.New("course schedule conflicts with existing enrollment")
    }
    
    // Check prerequisites
    completedCourses, err := v.EnrollmentRepo.GetCompletedCourses(ctx, cmd.StudentID)
    if err != nil {
        return fmt.Errorf("failed to check completed courses: %w", err)
    }
    
    prerequisites, err := v.PrerequisiteRepo.GetCoursePrerequisites(ctx, cmd.CourseID)
    if err != nil {
        return fmt.Errorf("failed to check prerequisites: %w", err)
    }
    
    for _, prereq := range prerequisites {
        if !contains(completedCourses, prereq.CourseID) {
            return fmt.Errorf("prerequisite course %s not completed", prereq.CourseName)
        }
    }
    
    // Check credit hour limits
    currentCredits, err := v.EnrollmentRepo.GetStudentCreditHours(ctx, cmd.StudentID, cmd.SemesterID)
    if err != nil {
        return fmt.Errorf("failed to check credit hours: %w", err)
    }
    
    maxCredits := v.getMaxCreditsForStudent(student)
    if currentCredits+course.CreditHours > maxCredits {
        return fmt.Errorf("enrollment would exceed maximum credit hours (%d)", maxCredits)
    }
    
    return nil
}
```

## 🏥 **Healthcare Appointment System**

### **Patient and Appointment Management**

```go
// Healthcare domain with compliance requirements
type ScheduleAppointmentCommand struct {
    command.Base
    PatientID     int       `json:"patient_id" validate:"required"`
    DoctorID      int       `json:"doctor_id" validate:"required"`
    AppointmentType string  `json:"appointment_type" validate:"required"`
    RequestedTime time.Time `json:"requested_time" validate:"required"`
    Duration      int       `json:"duration" validate:"required,min=15,max=240"`
    Reason        string    `json:"reason" validate:"required"`
    Priority      string    `json:"priority" validate:"oneof=routine urgent emergency"`
}

// HIPAA-compliant event handling
type AppointmentScheduledEventHandler struct {
    NotificationService NotificationService `inject:""`
    AuditService       AuditService       `inject:""`
    CalendarService    CalendarService    `inject:""`
    ReminderService    ReminderService    `inject:""`
    Logger             *logrus.Logger     `inject:""`
}

func (h *AppointmentScheduledEventHandler) Handle(ctx context.Context, e AppointmentScheduledEvent) error {
    // Log access for HIPAA compliance
    h.AuditService.LogAccess(ctx, AuditLog{
        UserID:     getUserID(ctx),
        Action:     "appointment_scheduled",
        ResourceID: fmt.Sprintf("appointment:%d", e.AppointmentID),
        Timestamp:  time.Now(),
        IPAddress:  getIPAddress(ctx),
    })
    
    // Send notifications (with privacy controls)
    patientNotification := PatientNotification{
        PatientID: e.PatientID,
        Type:      "appointment_confirmation",
        Message:   fmt.Sprintf("Your appointment is scheduled for %s", e.AppointmentTime.Format("January 2, 2006 at 3:04 PM")),
        Sensitive: true, // Requires secure delivery
    }
    
    if err := h.NotificationService.SendSecureNotification(ctx, patientNotification); err != nil {
        h.Logger.WithError(err).Error("Failed to send patient notification")
    }
    
    // Add to doctor's calendar
    calendarEvent := CalendarEvent{
        DoctorID:    e.DoctorID,
        StartTime:   e.AppointmentTime,
        Duration:    e.Duration,
        Title:       fmt.Sprintf("Patient Appointment - %s", e.AppointmentType),
        PatientInfo: redactPatientInfo(e.PatientID), // Minimal info for calendar
    }
    
    if err := h.CalendarService.AddEvent(ctx, calendarEvent); err != nil {
        h.Logger.WithError(err).Error("Failed to add calendar event")
    }
    
    // Schedule reminders
    reminderTime := e.AppointmentTime.Add(-24 * time.Hour) // 24 hours before
    if err := h.ReminderService.ScheduleReminder(ctx, ReminderRequest{
        PatientID:   e.PatientID,
        ScheduleFor: reminderTime,
        Type:        "appointment_reminder",
        Message:     "You have an appointment tomorrow",
    }); err != nil {
        h.Logger.WithError(err).Error("Failed to schedule reminder")
    }
    
    return nil
}
```

---

These examples demonstrate real-world usage patterns including:

- **Complex business logic** with multi-step validation
- **Error handling and compensation** strategies
- **Event-driven workflows** and saga patterns
- **Domain-specific requirements** (HIPAA, financial regulations)
- **Performance optimizations** (caching, async processing)
- **Production patterns** (auditing, monitoring, security)

For more specific implementation details, see the [API Reference](./api-reference.md) and [Patterns Guide](./patterns.md). 