# Testing Strategy

## Overview

This document outlines the comprehensive testing strategy for the Tribe application, covering unit tests, integration tests, end-to-end tests, and performance testing. The strategy emphasizes test-driven development with high coverage requirements and automated testing pipelines.

## Testing Philosophy

### Core Principles
- **Test-Driven Development**: Write tests before implementation
- **High Coverage**: Minimum 80% overall, 90%+ for critical business logic
- **Fast Feedback**: Quick test execution for rapid development cycles
- **Reliable Tests**: Tests should be deterministic and not flaky
- **Real-World Scenarios**: Tests reflect actual user behaviors and edge cases

### Testing Pyramid
1. **Unit Tests (70%)**: Fast, isolated tests for individual components
2. **Integration Tests (20%)**: Test component interactions and API contracts
3. **End-to-End Tests (10%)**: Full user journey validation

## Backend Testing

### Unit Testing Framework
Using Go's built-in testing package with testify for assertions:

```go
package services

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

func TestFilterEngine_ApplyFilters(t *testing.T) {
    testCases := []struct {
        name     string
        items    []ListItem
        criteria FilterCriteria
        expected []ListItem
    }{
        {
            name: "filter by category includes only specified categories",
            items: []ListItem{
                {ID: "1", Name: "Pizza Place", Category: "italian"},
                {ID: "2", Name: "Taco Bell", Category: "mexican"},
                {ID: "3", Name: "Olive Garden", Category: "italian"},
            },
            criteria: FilterCriteria{Categories: []string{"italian"}},
            expected: []ListItem{
                {ID: "1", Name: "Pizza Place", Category: "italian"},
                {ID: "3", Name: "Olive Garden", Category: "italian"},
            },
        },
        {
            name: "dietary filter excludes non-compliant items",
            items: []ListItem{
                {ID: "1", Name: "Vegan Cafe", DietaryInfo: map[string]bool{"vegan": true}},
                {ID: "2", Name: "Steakhouse", DietaryInfo: map[string]bool{"vegan": false}},
            },
            criteria: FilterCriteria{DietaryRequirements: []string{"vegan"}},
            expected: []ListItem{
                {ID: "1", Name: "Vegan Cafe", DietaryInfo: map[string]bool{"vegan": true}},
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mockDB := &MockDatabase{}
            engine := NewFilterEngine(mockDB)
            
            result, err := engine.ApplyFilters(context.Background(), tc.items, tc.criteria)
            
            require.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

### Integration Testing
Testing component interactions and database operations:

```go
func TestDecisionAPI_EndToEnd(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create test data
    tribe := createTestTribe(t, db, "Test Tribe")
    users := createTestUsers(t, db, 3)
    lists := createTestLists(t, db, tribe, users)
    
    // Test complete decision flow
    sessionReq := CreateDecisionSessionRequest{
        TribeID: tribe.ID,
        Name:    "Lunch Decision",
        ListIDs: []string{lists[0].ID, lists[1].ID},
    }
    
    session, err := decisionService.CreateSession(context.Background(), sessionReq)
    require.NoError(t, err)
    assert.Equal(t, "configuring", session.Status)
    
    // Apply filters
    filters := FilterCriteria{
        Categories: []string{"italian", "mexican"},
        MaxDistance: 10.0,
    }
    
    session, err = decisionService.ApplyFilters(context.Background(), session.ID, filters)
    require.NoError(t, err)
    assert.Greater(t, len(session.CurrentCandidates), 0)
    
    // Start elimination
    session, err = decisionService.StartElimination(context.Background(), session.ID)
    require.NoError(t, err)
    assert.Equal(t, "eliminating", session.Status)
    
    // Simulate eliminations from all users
    for round := 1; round <= 2; round++ { // K=2
        for _, user := range users {
            if !isUserTurn(session, user.ID) {
                continue
            }
            
            // Eliminate one item
            candidateID := session.CurrentCandidates[0]
            session, err = decisionService.EliminateItem(
                context.Background(), 
                session.ID, 
                user.ID, 
                candidateID,
            )
            require.NoError(t, err)
        }
    }
    
    // Verify final result
    assert.Equal(t, "completed", session.Status)
    assert.NotNil(t, session.FinalSelection)
    assert.Equal(t, 3, len(session.RunnersUp)) // M=3
}
```

### Mock Interfaces
Comprehensive mocking for external dependencies:

```go
type MockDatabase struct {
    mock.Mock
}

func (m *MockDatabase) CreateUser(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockDatabase) GetUser(ctx context.Context, userID string) (*User, error) {
    args := m.Called(ctx, userID)
    return args.Get(0).(*User), args.Error(1)
}

type MockEmailService struct {
    mock.Mock
    SentEmails []EmailMessage
}

func (m *MockEmailService) SendInvitation(ctx context.Context, email EmailMessage) error {
    m.SentEmails = append(m.SentEmails, email)
    args := m.Called(ctx, email)
    return args.Error(0)
}
```

## Frontend Testing

### Component Testing with React Testing Library
Focus on testing behavior rather than implementation:

```typescript
import { render, screen, userEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import { DecisionWizard } from './DecisionWizard';
import { createMockTribe, createMockLists } from '../test-utils';

describe('DecisionWizard', () => {
  const mockTribe = createMockTribe();
  const mockLists = createMockLists();
  const mockOnComplete = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should guide user through complete decision process', async () => {
    const user = userEvent.setup();
    
    render(
      <DecisionWizard 
        tribe={mockTribe} 
        availableLists={mockLists} 
        onComplete={mockOnComplete}
      />
    );
    
    // Step 1: Select lists
    expect(screen.getByText('Select Lists')).toBeInTheDocument();
    await user.click(screen.getByLabelText('Restaurant List'));
    await user.click(screen.getByLabelText('Activity List'));
    await user.click(screen.getByText('Next'));
    
    // Step 2: Apply filters
    expect(screen.getByText('Apply Filters')).toBeInTheDocument();
    await user.click(screen.getByLabelText('Vegetarian Options'));
    await user.type(screen.getByLabelText('Max Distance (miles)'), '10');
    await user.click(screen.getByText('Apply Filters'));
    
    // Step 3: Review and start
    await waitFor(() => {
      expect(screen.getByText('7 options found')).toBeInTheDocument();
    });
    
    expect(screen.getByText('Start 7-5-3 Decision')).toBeInTheDocument();
    await user.click(screen.getByText('Start Decision'));
    
    // Verify API calls
    expect(mockCreateDecisionSession).toHaveBeenCalledWith({
      tribeId: mockTribe.id,
      listIds: ['list-1', 'list-2'],
      filters: {
        dietaryRequirements: ['vegetarian'],
        maxDistance: 10,
      },
    });
  });

  it('should handle filter combinations that yield no results', async () => {
    const user = userEvent.setup();
    
    // Mock empty results
    vi.mocked(mockApplyFilters).mockResolvedValue({ items: [], suggestions: [] });
    
    render(<DecisionWizard tribe={mockTribe} availableLists={mockLists} onComplete={mockOnComplete} />);
    
    // Apply restrictive filters
    await user.click(screen.getByLabelText('Vegan'));
    await user.click(screen.getByLabelText('Gluten Free'));
    await user.type(screen.getByLabelText('Max Distance'), '1');
    await user.click(screen.getByText('Apply Filters'));
    
    // Should show helpful message
    await waitFor(() => {
      expect(screen.getByText('No options match your filters')).toBeInTheDocument();
      expect(screen.getByText('Try relaxing some filters')).toBeInTheDocument();
    });
    
    // Should show suggestions
    expect(screen.getByText('Remove gluten-free requirement (+3 options)')).toBeInTheDocument();
    expect(screen.getByText('Increase distance to 5 miles (+7 options)')).toBeInTheDocument();
  });
});
```

### Hook Testing
Testing custom React hooks in isolation:

```typescript
import { renderHook, act } from '@testing-library/react';
import { useDecisionSession } from './useDecisionSession';
import { createWrapper } from '../test-utils';

describe('useDecisionSession', () => {
  it('should manage decision session state correctly', async () => {
    const { result } = renderHook(
      () => useDecisionSession('session-123'),
      { wrapper: createWrapper() }
    );
    
    // Initial state
    expect(result.current.session).toBeNull();
    expect(result.current.loading).toBe(true);
    
    // Wait for session to load
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.session).toMatchObject({
        id: 'session-123',
        status: 'eliminating',
      });
    });
    
    // Test elimination action
    await act(async () => {
      await result.current.eliminateItem('item-456');
    });
    
    expect(result.current.session.currentCandidates).not.toContain('item-456');
  });
});
```

## End-to-End Testing

### Playwright E2E Tests
Complete user journey testing:

```typescript
import { test, expect, Page } from '@playwright/test';

test.describe('Decision Making Flow', () => {
  let page: Page;
  
  test.beforeEach(async ({ page: p }) => {
    page = p;
    await page.goto('/login');
    await loginTestUser(page, 'user1@example.com');
    await page.goto('/tribes/test-tribe-id');
  });

  test('complete decision making journey', async () => {
    // Start decision
    await page.click('[data-testid="start-decision-button"]');
    
    // Select lists
    await page.check('input[value="restaurant-list"]');
    await page.check('input[value="activity-list"]');
    await page.click('button:text("Next")');
    
    // Apply filters
    await page.fill('input[name="maxDistance"]', '15');
    await page.check('input[name="vegetarian"]');
    await page.click('button:text("Apply Filters")');
    
    // Verify filter results
    await expect(page.locator('[data-testid="filter-results"]')).toContainText('12 options found');
    
    // Start elimination
    await page.click('button:text("Start 12-8-3")');
    
    // Wait for elimination phase
    await expect(page.locator('[data-testid="elimination-phase"]')).toBeVisible();
    
    // Simulate elimination rounds (test user's turns only)
    const rounds = 2; // K=2
    for (let round = 1; round <= rounds; round++) {
      await waitForMyTurn(page);
      
      // Eliminate one item
      const firstCandidate = page.locator('[data-testid="candidate-item"]').first();
      await firstCandidate.click();
      await page.click('button:text("Eliminate")');
      
      // Wait for turn to advance
      await page.waitForTimeout(1000);
    }
    
    // Wait for session completion (other users simulated by backend)
    await expect(page.locator('[data-testid="decision-result"]')).toBeVisible({ timeout: 30000 });
    
    // Verify result
    await expect(page.locator('[data-testid="final-selection"]')).toBeVisible();
    await expect(page.locator('[data-testid="runners-up"]')).toHaveCount(3);
    
    // Test activity logging
    await page.click('button:text("Log This Activity")');
    await page.fill('input[name="scheduledFor"]', '2024-12-25 19:00');
    await page.click('button:text("Schedule Activity")');
    
    await expect(page.locator('[data-testid="activity-logged"]')).toBeVisible();
  });

  test('handles network interruption gracefully', async () => {
    // Start decision process
    await page.click('[data-testid="start-decision-button"]');
    await setupQuickDecision(page);
    
    // Simulate network failure during elimination
    await page.context().setOffline(true);
    
    // Try to eliminate item
    const firstCandidate = page.locator('[data-testid="candidate-item"]').first();
    await firstCandidate.click();
    await page.click('button:text("Eliminate")');
    
    // Should show offline indicator
    await expect(page.locator('[data-testid="offline-indicator"]')).toBeVisible();
    
    // Restore network
    await page.context().setOffline(false);
    
    // Should automatically sync and continue
    await expect(page.locator('[data-testid="sync-success"]')).toBeVisible();
    await expect(page.locator('[data-testid="turn-indicator"]')).toContainText('Your turn');
  });
});

// Helper functions
async function waitForMyTurn(page: Page) {
  await expect(page.locator('[data-testid="turn-indicator"]')).toContainText('Your turn');
  await expect(page.locator('[data-testid="turn-timer"]')).toBeVisible();
}

async function setupQuickDecision(page: Page) {
  await page.check('input[value="small-list"]'); // List with 5 items
  await page.click('button:text("Next")');
  await page.click('button:text("Skip Filters")');
  await page.click('button:text("Start 5-3-1")');
}
```

### Mobile E2E Testing
Specialized tests for mobile experience:

```typescript
test.describe('Mobile Decision Experience', () => {
  test.use({ 
    viewport: { width: 375, height: 667 }, // iPhone SE
    hasTouch: true 
  });

  test('mobile elimination interface works correctly', async ({ page }) => {
    await setupMobileDecision(page);
    
    // Test swipe-to-eliminate gesture
    const candidate = page.locator('[data-testid="candidate-card"]').first();
    await candidate.hover();
    await page.mouse.down();
    await page.mouse.move(100, 0); // Swipe right
    await page.mouse.up();
    
    await expect(page.locator('[data-testid="elimination-confirmation"]')).toBeVisible();
    await page.tap('button:text("Confirm")');
    
    // Verify item was eliminated
    await expect(candidate).not.toBeVisible();
  });

  test('mobile quick-skip works with touch', async ({ page }) => {
    await setupMobileDecision(page);
    
    // Long press for quick-skip
    const skipButton = page.locator('[data-testid="quick-skip-button"]');
    await skipButton.tap({ delay: 1000 }); // Long press
    
    await expect(page.locator('[data-testid="turn-skipped"]')).toBeVisible();
    await expect(page.locator('[data-testid="next-user-turn"]')).toBeVisible();
  });
});
```

## Performance Testing

### Load Testing with k6
Testing API performance under load:

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

export let errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '2m', target: 10 },   // Ramp up
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 100 },  // Peak load
    { duration: '5m', target: 100 },  // Stay at peak
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    errors: ['rate<0.1'],             // Error rate under 10%
  },
};

export default function() {
  // Test decision session creation
  let sessionResponse = http.post('http://localhost:8080/api/v1/decisions', {
    tribeId: 'load-test-tribe',
    listIds: ['list-1', 'list-2'],
  }, {
    headers: { 
      'Authorization': `Bearer ${__ENV.TEST_TOKEN}`,
      'Content-Type': 'application/json',
    },
  });
  
  check(sessionResponse, {
    'session created': (r) => r.status === 201,
    'session has ID': (r) => r.json('id') !== '',
  }) || errorRate.add(1);
  
  let sessionId = sessionResponse.json('id');
  
  // Test filter application
  let filterResponse = http.post(`http://localhost:8080/api/v1/decisions/${sessionId}/filters`, {
    categories: ['italian', 'mexican'],
    maxDistance: 10,
  });
  
  check(filterResponse, {
    'filters applied': (r) => r.status === 200,
    'has candidates': (r) => r.json('candidates').length > 0,
  }) || errorRate.add(1);
  
  sleep(1);
}
```

### Database Performance Testing
Testing query performance with realistic data:

```go
func BenchmarkFilterEngine_ComplexFilters(b *testing.B) {
    db := setupBenchmarkDB(b)
    engine := NewFilterEngine(db)
    
    // Create realistic dataset
    items := generateTestItems(10000) // 10k items
    criteria := FilterCriteria{
        Categories: []string{"italian", "mexican", "chinese"},
        DietaryRequirements: []string{"vegetarian"},
        MaxDistance: 15.0,
        ExcludeRecentDays: 7,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := engine.ApplyFilters(context.Background(), items, criteria)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkDecisionElimination_LargeSession(b *testing.B) {
    session := &DecisionSession{
        CurrentCandidates: generateCandidateIDs(500), // 500 items
        EliminationOrder: generateUserIDs(8),         // 8 users
        AlgorithmParams: AlgorithmParams{K: 3, M: 5},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := processEliminationRound(session, "user-1", "item-123")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Test Data Management

### Test Database Setup
Consistent test environment setup:

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", os.Getenv("TEST_DATABASE_URL"))
    require.NoError(t, err)
    
    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    // Cleanup on test completion
    t.Cleanup(func() {
        cleanupTestDB(t, db)
    })
    
    return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
    // Clean all test data
    tables := []string{
        "decision_eliminations",
        "decision_sessions",
        "activity_history",
        "list_items",
        "lists",
        "tribe_memberships",
        "tribes",
        "users",
    }
    
    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at > NOW() - INTERVAL '1 hour'", table))
        if err != nil {
            t.Logf("Failed to clean table %s: %v", table, err)
        }
    }
}
```

### Test Data Factories
Consistent test data generation:

```go
func CreateTestTribe(t *testing.T, db *sql.DB, name string) *Tribe {
    tribe := &Tribe{
        ID:          generateUUID(),
        Name:        name,
        Description: "Test tribe for " + name,
        MaxMembers:  8,
        CreatedAt:   time.Now(),
    }
    
    err := insertTribe(db, tribe)
    require.NoError(t, err)
    
    return tribe
}

func CreateTestUsers(t *testing.T, db *sql.DB, count int) []*User {
    users := make([]*User, count)
    for i := 0; i < count; i++ {
        users[i] = &User{
            ID:       generateUUID(),
            Email:    fmt.Sprintf("testuser%d@example.com", i+1),
            Name:     fmt.Sprintf("Test User %d", i+1),
            Provider: "test",
        }
        
        err := insertUser(db, users[i])
        require.NoError(t, err)
    }
    return users
}
```

## Continuous Integration

### GitHub Actions Workflow
Automated testing pipeline:

```yaml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: tribe_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run migrations
      run: go run cmd/migrate/main.go up
      env:
        DATABASE_URL: postgres://postgres:test@localhost/tribe_test?sslmode=disable
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
      env:
        DATABASE_URL: postgres://postgres:test@localhost/tribe_test?sslmode=disable
    
    - name: Check coverage
      run: |
        go tool cover -func coverage.out
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
    
    - name: Install dependencies
      run: npm ci
    
    - name: Run tests
      run: npm run test:coverage
    
    - name: Run E2E tests
      run: npm run test:e2e
      env:
        PLAYWRIGHT_BROWSERS_PATH: 0

  integration-tests:
    runs-on: ubuntu-latest
    needs: [backend-tests, frontend-tests]
    steps:
    - uses: actions/checkout@v3
    
    - name: Start application
      run: docker-compose up -d
    
    - name: Wait for services
      run: ./scripts/wait-for-services.sh
    
    - name: Run integration tests
      run: npm run test:integration
    
    - name: Cleanup
      run: docker-compose down
```

## Coverage Requirements

### Coverage Targets
- **Overall Backend Coverage**: Minimum 80%
- **Critical Business Logic**: Minimum 90%
- **Frontend Components**: Minimum 85%
- **Integration Tests**: Cover all API endpoints
- **E2E Tests**: Cover all primary user journeys

### Coverage Monitoring
```bash
# Generate coverage reports
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Check coverage thresholds
go tool cover -func coverage.out | grep total | awk '{print $3}' | sed 's/%//' | awk '{if($1<80) exit 1}'
```

### Excluded from Coverage
- Generated code (protobuf, swagger)
- Third-party vendor code
- Simple getters/setters
- Database migration files
- Configuration structs

## Test Environment Management

### Environment Configuration
```yaml
# test-config.yaml
test:
  database:
    url: "postgres://test:test@localhost/tribe_test"
    max_connections: 10
  
  email:
    provider: "mock"
    capture_emails: true
  
  auth:
    jwt_secret: "test-secret-key"
    token_expiry: "1h"
  
  external_apis:
    mock_mode: true
    response_delay: "100ms"
```

### Test Data Isolation
- Each test uses isolated database transactions
- Parallel test execution with separate schemas
- Cleanup strategies for shared resources
- Deterministic test data generation

---

*This document defines the comprehensive testing strategy for the Tribe application. See [Development Guidelines](./development-guidelines.md) for code quality standards and [Implementation Roadmap](./roadmap.md) for testing milestones.* 