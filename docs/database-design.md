# Database Design Documentation

## Current Shortcomings and Improvement Plan

### Identified Issues

#### 1. Schema Management
- **Problem**: Basic migration system without versioning or rollback capabilities
- **Impact**: Difficult to manage schema changes across environments
- **Solution**: Implement versioned migrations with up/down scripts

#### 2. Performance Issues
- **Problem**: No database indexes for performance optimization
- **Impact**: Slow queries on large datasets, especially for lookups by email, foreign keys
- **Solution**: Add strategic indexes for frequently queried columns

#### 3. Connection Management
- **Problem**: Basic connection handling without pooling configuration
- **Impact**: Potential connection exhaustion under load
- **Solution**: Configure proper connection pool settings

#### 4. Transaction Management
- **Problem**: Most operations lack proper transaction boundaries
- **Impact**: Data inconsistency risk during complex operations
- **Solution**: Implement transaction wrappers for multi-step operations

#### 5. Error Handling
- **Problem**: Inconsistent error handling and mixed language error messages
- **Impact**: Difficult debugging and poor user experience
- **Solution**: Standardize error handling patterns and messages

## Database Schema Overview

### Core Tables

#### Users (`users`)
Primary entity for user management with full profile information.

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    birthdate TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT NOT NULL UNIQUE,
    address TEXT,
    postal_code TEXT,
    city TEXT,
    country TEXT,
    password TEXT NOT NULL,
    newsletter_subscription BOOLEAN DEFAULT FALSE,
    terms_accepted BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### Events (`events`)
Manages classes, workshops, and other events.

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    role_requirements TEXT,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    location TEXT,
    organizer TEXT,
    attendees TEXT,
    class_type TEXT DEFAULT '',
    teacher_name TEXT DEFAULT '',
    capacity INTEGER DEFAULT 0,
    current_enrolment INTEGER DEFAULT 0,
    color TEXT DEFAULT ''
);
```

#### Memberships (`memberships`)
Defines available membership types and pricing.

#### User Memberships (`user_memberships`)
Links users to their active memberships with status tracking.

### Relationship Tables

- `user_roles`: Many-to-many relationship between users and roles
- `event_signups`: Tracks user event registrations
- `user_klippekort`: Manages user's class credits
- `payment_methods`: Stores user payment information

## Recommended Improvements

### 1. Add Database Indexes

```sql
-- User lookup indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);

-- Event lookup indexes
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_location ON events(location);
CREATE INDEX idx_events_teacher_name ON events(teacher_name);

-- Membership indexes
CREATE INDEX idx_user_memberships_user_id ON user_memberships(user_id);
CREATE INDEX idx_user_memberships_status ON user_memberships(status);

-- Event signup indexes
CREATE INDEX idx_event_signups_user_id ON event_signups(user_id);
CREATE INDEX idx_event_signups_event_id ON event_signups(event_id);
```

### 2. Connection Pool Configuration

```go
func Connect() (*sql.DB, error) {
    // ... existing code ...
    
    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db, nil
}
```

### 3. Transaction Management Pattern

```go
// Example transaction wrapper
func (db *Database) WithTransaction(fn func(*sql.Tx) error) error {
    tx, err := db.Conn.Begin()
    if err != nil {
        return err
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        } else if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()
    
    err = fn(tx)
    return err
}
```

### 4. Improved Migration System

Implement a versioned migration system with:
- Sequential version numbers
- Up and down migration scripts
- Migration status tracking table
- Rollback capabilities

### 5. Enhanced Error Handling

- Standardize error types and messages
- Use English consistently for error messages
- Add structured error information for debugging
- Implement retry logic for transient errors