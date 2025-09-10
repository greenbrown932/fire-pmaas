# Fire PMAAS Testing Guide

This document provides comprehensive information about the testing strategy, structure, and execution for the Fire PMAAS application.

## Table of Contents

- [Testing Strategy](#testing-strategy)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Categories](#test-categories)
- [Test Coverage](#test-coverage)
- [Writing Tests](#writing-tests)
- [Mocking and Fixtures](#mocking-and-fixtures)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Testing Strategy

Our testing strategy follows the testing pyramid with:

1. **Unit Tests (70%)**: Fast, isolated tests for individual functions and methods
2. **Integration Tests (20%)**: Tests for API endpoints and component interactions
3. **End-to-End Tests (10%)**: Full application workflow tests

### Test Principles

- **Fast**: Tests should run quickly to enable rapid feedback
- **Reliable**: Tests should be deterministic and not flaky
- **Isolated**: Tests should not depend on external services or other tests
- **Maintainable**: Tests should be easy to read, write, and maintain
- **Comprehensive**: Critical paths should have high test coverage

## Test Structure

```
fire-pmaas/
├── pkg/
│   ├── api/
│   │   ├── api_test.go           # API route tests
│   │   └── users_test.go         # User API endpoint tests
│   ├── db/
│   │   └── db_test.go            # Database interaction tests
│   ├── middleware/
│   │   └── auth_test.go          # Authentication/authorization tests
│   ├── models/
│   │   └── model_test.go         # Model unit tests
│   └── testutils/
│       ├── testutils.go          # Test utilities and helpers
│       └── fixtures.go           # Test data fixtures
├── tests/
│   └── integration/
│       └── rbac_test.go          # Role-based access control tests
├── Makefile                      # Test runner commands
└── TESTING.md                    # This documentation
```

## Running Tests

### Prerequisites

1. Go 1.24.6 or later
2. Docker (for database tests)
3. Make (optional, for convenience commands)

### Basic Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run specific test package
go test ./pkg/models/

# Run specific test
go test -run TestCreateUser ./pkg/models/
```

### Using Makefile Commands

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run tests with coverage
make test-coverage

# Run database tests
make test-db

# Run RBAC tests
make test-rbac

# Clean test cache
make test-clean
```

## Test Categories

### 1. Unit Tests

Test individual functions and methods in isolation.

**Location**: `pkg/*/`
**Naming**: `*_test.go`
**Examples**:
- `pkg/models/model_test.go` - Model CRUD operations
- `pkg/middleware/auth_test.go` - Authentication middleware

### 2. Integration Tests

Test component interactions and API endpoints.

**Location**: `tests/integration/`, `pkg/api/`
**Naming**: `*_test.go`
**Examples**:
- `pkg/api/users_test.go` - User API endpoints
- `tests/integration/rbac_test.go` - Role-based access control

### 3. Database Tests

Test database interactions with mocked SQL connections.

**Location**: `pkg/db/`, `pkg/models/`
**Features**:
- SQL mock using `go-sqlmock`
- Transaction testing
- Constraint violation testing

## Test Coverage

### Coverage Goals

- **Overall**: 80% minimum
- **Critical paths**: 95% minimum
- **Models**: 90% minimum
- **API endpoints**: 85% minimum

### Generating Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html

# Generate coverage badge
make coverage-badge
```

### Coverage Commands

```bash
# Basic coverage
go test -cover ./...

# Detailed coverage with HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Coverage by function
go tool cover -func=coverage.out
```

## Writing Tests

### Test File Structure

```go
package models

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/greenbrown932/fire-pmaas/pkg/testutils"
)

func TestFunctionName(t *testing.T) {
    // Arrange
    testData := setupTestData()
    
    // Act
    result, err := FunctionToTest(testData)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedValue, result)
}
```

### Test Naming Conventions

- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Table-driven tests: `TestFunctionName` with subtests
- Benchmark tests: `BenchmarkFunctionName`

### Example Test Patterns

#### Simple Unit Test

```go
func TestCreateUser(t *testing.T) {
    mock, cleanup := setupTestDB(t)
    defer cleanup()

    user := &User{Username: "test", Email: "test@example.com"}
    
    mock.ExpectQuery(`INSERT INTO users`).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

    err := CreateUser(user)
    assert.NoError(t, err)
    assert.Equal(t, 1, user.ID)
}
```

#### Table-Driven Test

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name     string
        user     User
        wantErr  bool
        errMsg   string
    }{
        {"valid user", User{Username: "test", Email: "test@example.com"}, false, ""},
        {"missing username", User{Email: "test@example.com"}, true, "username required"},
        {"invalid email", User{Username: "test", Email: "invalid"}, true, "invalid email"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.user.Validate()
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### API Integration Test

```go
func TestUserRegistration(t *testing.T) {
    router, testDB := setupUserAPITest(t)
    defer testDB.TeardownTestDB()

    payload := models.UserRegistration{
        Username: "newuser",
        Email:    "new@example.com",
        Password: "password123",
    }

    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/users/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)
}
```

## Mocking and Fixtures

### Database Mocking

Using `go-sqlmock` for database operation testing:

```go
func setupTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
    mockDB, mock, err := sqlmock.New()
    require.NoError(t, err)

    originalDB := db.DB
    db.DB = mockDB

    return mock, func() {
        db.DB = originalDB
        mockDB.Close()
    }
}
```

### Test Fixtures

Pre-defined test data in `pkg/testutils/fixtures.go`:

```go
func GetUserFixtures() *UserFixtures {
    return &UserFixtures{
        AdminUser: &models.User{
            ID:       1,
            Username: "admin",
            Email:    "admin@test.com",
            Roles:    []models.Role{{Name: "admin"}},
        },
        // ... more fixtures
    }
}
```

### Test Utilities

Helper functions in `pkg/testutils/testutils.go`:

```go
func CreateTestUser(t *testing.T, username, email, role string) *models.User
func SetupTestDB(t *testing.T) *TestDB
func MakeTestRequest(t *testing.T, method, url string, body interface{}) *http.Request
```

## Test Environment Setup

### Environment Variables

Required for testing:

```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=test_user
export POSTGRES_PASSWORD=test_pass
export POSTGRES_DB=test_db
export KEYCLOAK_ISSUER=http://localhost:8080/realms/test
```

### Docker Test Environment

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run tests
make test

# Clean up
docker-compose -f docker-compose.test.yml down
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.24.6
    - run: make ci-test
```

### Test Commands for CI

```bash
# Complete CI test suite
make ci-test

# Individual CI steps
make deps
make lint
make vet
make test-coverage
```

## Test Performance

### Benchmark Tests

```go
func BenchmarkUserCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        user := &User{Username: fmt.Sprintf("user%d", i)}
        CreateUser(user)
    }
}
```

### Performance Monitoring

```bash
# Run benchmarks
make test-benchmark

# CPU profiling
make profile-cpu

# Memory profiling
make profile-mem
```

## Troubleshooting

### Common Issues

1. **Test Database Connection**
   ```bash
   # Check database is running
   docker ps | grep postgres
   
   # Verify connection
   psql -h localhost -U test_user -d test_db
   ```

2. **Mock Expectations Not Met**
   ```go
   // Always check mock expectations
   require.NoError(t, testDB.Mock.ExpectationsWereMet())
   ```

3. **Race Conditions**
   ```bash
   # Run with race detection
   go test -race ./...
   ```

### Debug Commands

```bash
# Verbose test output
go test -v ./...

# Run specific test with debugging
go test -v -run TestSpecificFunction ./pkg/models/

# Test with timeout
go test -timeout 30s ./...
```

### Test Data Cleanup

```bash
# Clean test cache
go clean -testcache

# Remove coverage files
rm -f coverage.out coverage.html
```

## Best Practices

1. **Test Organization**
   - Group related tests in the same file
   - Use descriptive test names
   - Keep tests focused and small

2. **Assertions**
   - Use `require` for critical assertions that should stop the test
   - Use `assert` for non-critical checks
   - Provide meaningful assertion messages

3. **Test Data**
   - Use fixtures for complex test data
   - Keep test data minimal and focused
   - Clean up after tests when needed

4. **Mocking**
   - Mock external dependencies
   - Verify mock expectations
   - Keep mocks simple and focused

5. **Coverage**
   - Aim for high coverage but focus on quality
   - Test error paths and edge cases
   - Don't test trivial getters/setters

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)
- [Testing Best Practices](https://github.com/golang/go/wiki/TestComments)