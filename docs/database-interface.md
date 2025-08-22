# Database Interface Documentation

## Database Lookup Functions Interface Guide

This document provides comprehensive documentation for writing functions that involve database lookups in the kjernekraft system.

## Core Interface Patterns

### 1. Database Connection Pattern

```go
type Database struct {
    Conn *sql.DB
}

// Standard connection setup
func Connect() (*sql.DB, error)
func Migrate(db *sql.DB) error
```

**Best Practices:**
- Always use the `DB_PATH` environment variable for configurable database paths
- Enable foreign key constraints with `PRAGMA foreign_keys = ON`
- Configure connection pooling for production use

### 2. Error Handling Patterns

#### Standard Error Handling
```go
func (db *Database) GetUserByID(userID int64) (*models.User, error) {
    var user models.User
    err := db.Conn.QueryRow(
        "SELECT id, name, email FROM users WHERE id = ?", 
        userID,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // Not found - return nil, nil
        }
        return nil, fmt.Errorf("failed to get user %d: %v", userID, err)
    }
    
    return &user, nil
}
```

#### Error Handling Best Practices:
- **Not Found**: Return `(nil, nil)` for single entity lookups when no rows found
- **Multiple Results**: Return `([]Type{}, nil)` for empty lists, not nil slices
- **Database Errors**: Wrap errors with context using `fmt.Errorf`
- **Validation Errors**: Return descriptive error messages for business logic violations

### 3. Query Patterns

#### Single Entity Lookup
```go
func (db *Database) GetUserByEmail(email string) (*models.User, error) {
    var user models.User
    query := `SELECT id, name, email, phone, address, created_at 
              FROM users WHERE email = ?`
    
    err := db.Conn.QueryRow(query, email).Scan(
        &user.ID, &user.Name, &user.Email, 
        &user.Phone, &user.Address, &user.CreatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get user by email %s: %v", email, err)
    }
    
    return &user, nil
}
```

#### Multiple Entity Lookup
```go
func (db *Database) GetUsersByRole(roleName string) ([]models.User, error) {
    query := `SELECT u.id, u.name, u.email 
              FROM users u
              JOIN user_roles ur ON u.id = ur.user_id
              JOIN roles r ON ur.role_id = r.id
              WHERE r.name = ?`
    
    rows, err := db.Conn.Query(query, roleName)
    if err != nil {
        return nil, fmt.Errorf("failed to query users by role %s: %v", roleName, err)
    }
    defer rows.Close()
    
    var users []models.User
    for rows.Next() {
        var user models.User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            return nil, fmt.Errorf("failed to scan user: %v", err)
        }
        users = append(users, user)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error during row iteration: %v", err)
    }
    
    return users, nil
}
```

#### Filtered Lookup with Dynamic Conditions
```go
func (db *Database) GetFilteredEvents(startDate, endDate, location string) ([]models.Event, error) {
    query := "SELECT id, title, start_time, location FROM events WHERE 1=1"
    var args []interface{}
    
    if startDate != "" {
        query += " AND start_time >= ?"
        args = append(args, startDate)
    }
    if endDate != "" {
        query += " AND end_time <= ?"
        args = append(args, endDate)
    }
    if location != "" {
        query += " AND location = ?"
        args = append(args, location)
    }
    
    rows, err := db.Conn.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query filtered events: %v", err)
    }
    defer rows.Close()
    
    var events []models.Event
    for rows.Next() {
        var event models.Event
        if err := rows.Scan(&event.ID, &event.Title, &event.StartTime, &event.Location); err != nil {
            return nil, fmt.Errorf("failed to scan event: %v", err)
        }
        events = append(events, event)
    }
    
    return events, nil
}
```

### 4. Transaction Patterns

#### Simple Transaction
```go
func (db *Database) CreateUserWithMembership(user models.User, membershipID int64) error {
    tx, err := db.Conn.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    // Create user
    result, err := tx.Exec(
        "INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
        user.Name, user.Email, user.Password,
    )
    if err != nil {
        return fmt.Errorf("failed to create user: %v", err)
    }
    
    userID, err := result.LastInsertId()
    if err != nil {
        return fmt.Errorf("failed to get user ID: %v", err)
    }
    
    // Add membership
    _, err = tx.Exec(
        "INSERT INTO user_memberships (user_id, membership_id, start_date) VALUES (?, ?, ?)",
        userID, membershipID, time.Now(),
    )
    if err != nil {
        return fmt.Errorf("failed to create membership: %v", err)
    }
    
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }
    
    return nil
}
```

### 5. Validation Patterns

#### Input Validation
```go
func (db *Database) CreateUser(user models.User) (int64, error) {
    // Validate required fields
    if user.Name == "" {
        return 0, fmt.Errorf("user name is required")
    }
    if user.Email == "" {
        return 0, fmt.Errorf("user email is required")
    }
    
    // Check for existing email
    var existingID int
    err := db.Conn.QueryRow("SELECT id FROM users WHERE email = ?", user.Email).Scan(&existingID)
    if err == nil {
        return 0, fmt.Errorf("email already exists")
    }
    if err != sql.ErrNoRows {
        return 0, fmt.Errorf("failed to check email uniqueness: %v", err)
    }
    
    // Proceed with creation...
}
```

### 6. Join Patterns for Related Data

#### One-to-Many with Details
```go
func (db *Database) GetUserMembership(userID int64) (*models.MembershipWithDetails, error) {
    query := `
        SELECT um.id, um.user_id, um.status, um.start_date,
               m.name, m.price, m.description
        FROM user_memberships um
        JOIN memberships m ON um.membership_id = m.id
        WHERE um.user_id = ? AND um.status = 'active'
        LIMIT 1`
    
    var membership models.MembershipWithDetails
    err := db.Conn.QueryRow(query, userID).Scan(
        &membership.UserMembership.ID,
        &membership.UserMembership.UserID,
        &membership.UserMembership.Status,
        &membership.UserMembership.StartDate,
        &membership.Membership.Name,
        &membership.Membership.Price,
        &membership.Membership.Description,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get user membership: %v", err)
    }
    
    return &membership, nil
}
```

### 7. Bulk Operations

#### Batch Insert Pattern
```go
func (db *Database) CreateMultipleEvents(events []models.Event) error {
    if len(events) == 0 {
        return nil
    }
    
    tx, err := db.Conn.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    stmt, err := tx.Prepare("INSERT INTO events (title, start_time, location) VALUES (?, ?, ?)")
    if err != nil {
        return fmt.Errorf("failed to prepare statement: %v", err)
    }
    defer stmt.Close()
    
    for _, event := range events {
        if _, err := stmt.Exec(event.Title, event.StartTime, event.Location); err != nil {
            return fmt.Errorf("failed to insert event %s: %v", event.Title, err)
        }
    }
    
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %v", err)
    }
    
    return nil
}
```

## Function Naming Conventions

### CRUD Operations
- **Create**: `CreateUser`, `CreateEvent`
- **Read Single**: `GetUserByID`, `GetEventByID`
- **Read Multiple**: `GetAllUsers`, `GetUsersByRole`
- **Read Filtered**: `GetFilteredEvents`, `GetActiveMembers`
- **Update**: `UpdateUser`, `UpdateEventTime`
- **Delete**: `DeleteUser`, `DeleteEvent`

### Business Logic Operations
- **Status Changes**: `UpdateMembershipStatus`, `FreezeUserMembership`
- **Relationships**: `AssignRoleToUser`, `SignupUserForEvent`
- **Validation**: `AuthenticateUser`, `ValidateUserAccess`

## Testing Patterns

### Setup Test Database
```go
func setupTestDB() (*Database, func()) {
    tmpfile, err := os.CreateTemp("", "test_*.db")
    if err != nil {
        log.Fatal(err)
    }
    tmpfile.Close()
    
    os.Setenv("DB_PATH", tmpfile.Name())
    
    dbConn, err := Connect()
    if err != nil {
        log.Fatal(err)
    }
    
    if err := Migrate(dbConn); err != nil {
        log.Fatal(err)
    }
    
    db := &Database{Conn: dbConn}
    
    cleanup := func() {
        dbConn.Close()
        os.Remove(tmpfile.Name())
    }
    
    return db, cleanup
}
```

### Test Database Operations
```go
func TestGetUserByEmail(t *testing.T) {
    db, cleanup := setupTestDB()
    defer cleanup()
    
    // Create test user
    user := models.User{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "password",
    }
    userID, err := db.CreateUser(user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    // Test retrieval
    retrieved, err := db.GetUserByEmail("test@example.com")
    if err != nil {
        t.Fatalf("Failed to get user: %v", err)
    }
    
    if retrieved == nil {
        t.Fatal("Expected user, got nil")
    }
    
    if retrieved.ID != int(userID) {
        t.Errorf("Expected ID %d, got %d", userID, retrieved.ID)
    }
}
```

## Performance Considerations

### Use Appropriate Indexes
Always ensure frequently queried columns have indexes:
```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_events_start_time ON events(start_time);
```

### Limit Results for Large Datasets
```go
func (db *Database) GetRecentEvents(limit int) ([]models.Event, error) {
    query := `SELECT id, title, start_time FROM events 
              ORDER BY start_time DESC LIMIT ?`
    // ... implementation
}
```

### Use EXISTS for Existence Checks
```go
func (db *Database) UserExists(email string) (bool, error) {
    var exists bool
    err := db.Conn.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", 
        email,
    ).Scan(&exists)
    return exists, err
}
```