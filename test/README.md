# Property-Based Testing for User Actions

This directory contains automated stateful property-based testing for user actions in the kjernekraft application.

## Overview

Property-based testing generates random test inputs and verifies that certain properties always hold, regardless of the specific inputs. This approach helps uncover edge cases and invariant violations that traditional example-based tests might miss.

## Implementation

### User Action Testing (`user_actions_property_test.go`)

Tests core user operations:
- **User Registration**: Creates new users with random data
- **Profile Updates**: Modifies user profile information
- **Login/Logout**: Handles authentication state changes

#### Properties Tested:
1. **User registration creates valid user**: Verifies that user registration always results in a valid user with a positive ID
2. **Login requires existing user**: Ensures login only succeeds when the email matches an existing user
3. **Profile update preserves ID**: Confirms that profile updates never change the user ID

### Membership Action Testing (`membership_actions_property_test.go`)

Tests membership-related operations:
- **Membership Changes**: Switching between different membership types
- **Freeze/Unfreeze**: Managing membership suspension and reactivation

#### Properties Tested:
1. **Freeze/unfreeze reversible**: Verifies that freezing and unfreezing a membership returns to the original state
2. **Membership change preserves valid ID**: Ensures membership changes maintain valid membership IDs

## Running Tests

```bash
# Run all property-based tests
go test ./test -v -run "Property"

# Run benchmarks
go test ./test -bench=.

# Run specific test
go test ./test -v -run "TestUserActionsPropertyBased"
```

## Key Invariants

The tests verify several critical invariants:

### User State Invariants:
1. If a user is logged in, they must have a session
2. User IDs must always be positive
3. Users in state must exist in the database

### Membership State Invariants:
1. Frozen status must match the IsFrozen field
2. Active status means not frozen
3. Membership IDs must be positive after changes

## Architecture

### Action Pattern
Each user action implements the `UserAction` interface:
```go
type UserAction interface {
    Apply(state *UserState, db *database.Database) error
    String() string
}
```

This pattern allows:
- Composable sequences of actions
- Easy testing of action combinations
- Clear separation of concerns

### State Management
- `UserState`: Tracks user information and login status
- `MembershipState`: Tracks membership information and freeze status

### Test Database Setup
Each test uses an isolated temporary SQLite database:
- Fresh database per test
- Full schema migration
- Automatic cleanup

## Benefits

1. **Edge Case Discovery**: Automatically finds unusual input combinations
2. **Regression Prevention**: Catches invariant violations as code evolves
3. **Documentation**: Properties serve as executable specifications
4. **Confidence**: Provides high confidence in system correctness

## Dependencies

- `github.com/leanovate/gopter`: Property-based testing framework for Go
- Standard Go testing framework
- Application database and models

## Performance

The property-based tests are designed to be fast:
- Tests run in ~6-7ms on average
- Benchmarks available for performance monitoring
- Configurable test parameters for deeper testing when needed

## Future Enhancements

Potential areas for expansion:
1. Event signup/cancellation testing
2. Payment processing workflows
3. Role-based access control verification
4. Concurrent user action testing
5. Database transaction integrity testing