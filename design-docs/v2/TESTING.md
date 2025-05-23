# Tribe Testing

## Testing Strategy (Comprehensive)

### Backend Testing
```go
// Unit Tests
func TestFilterEngine_ApplyFilters(t *testing.T) {
    testCases := []struct {
        name     string
        items    []ListItem
        criteria FilterCriteria
        expected []ListItem
    }{
        {
            name: "filter by category",
            items: []ListItem{
                {Category: "italian"},
                {Category: "mexican"},
                {Category: "italian"},
            },
            criteria: FilterCriteria{Categories: []string{"italian"}},
            expected: []ListItem{
                {Category: "italian"},
                {Category: "italian"},
            },
        },
        // More test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            engine := NewFilterEngine(mockDB)
            result, err := engine.ApplyFilters(context.Background(), tc.items, tc.criteria)
            require.NoError(t, err)
            assert.Equal(t, tc.expected, result)
        })
    }
}

// Integration Tests
func TestDecisionAPI_EndToEnd(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create test data
    tribe := createTestTribe(t, db)
    users := createTestUsers(t, db, 2)
    lists := createTestLists(t, db, tribe, users)
    
    // Test complete decision flow
    session := startDecisionSession(t, tribe.ID, lists)
    applyFilters(t, session.ID, testFilters)
    
    // Simulate eliminations from both users
    eliminateItems(t, session.ID, users[0].ID, []string{"item1", "item2"})
    eliminateItems(t, session.ID, users[1].ID, []string{"item3", "item4"})
    
    // Verify final result
    finalSession := getDecisionSession(t, session.ID)
    assert.Equal(t, "completed", finalSession.Status)
    assert.NotNil(t, finalSession.FinalSelection)
}
```

### Frontend Testing
```typescript
// Component Tests
describe('DecisionWizard', () => {
  it('should guide user through decision process', async () => {
    const mockTribe = createMockTribe();
    const mockLists = createMockLists();
    
    render(
      <DecisionWizard 
        tribe={mockTribe} 
        availableLists={mockLists} 
        onComplete={jest.fn()}
      />
    );
    
    // Step 1: Select lists
    await userEvent.click(screen.getByText('Restaurant List'));
    await userEvent.click(screen.getByText('Next'));
    
    // Step 2: Apply filters
    await userEvent.click(screen.getByLabelText('Vegetarian Options'));
    await userEvent.type(screen.getByLabelText('Max Distance'), '10');
    await userEvent.click(screen.getByText('Apply Filters'));
    
    // Step 3: Configure algorithm
    expect(screen.getByText('5-3-1 Selection')).toBeInTheDocument();
    await userEvent.click(screen.getByText('Start Decision'));
    
    // Verify API calls
    expect(mockAPI.createDecisionSession).toHaveBeenCalled();
  });
});

// E2E Tests (Playwright)
test('complete decision making flow', async ({ page }) => {
  await page.goto('/tribes/test-tribe');
  await page.click('[data-testid="start-decision"]');
  
  // Select lists
  await page.check('input[value="restaurant-list"]');
  await page.click('button:text("Next")');
  
  // Apply filters
  await page.fill('input[name="maxDistance"]', '15');
  await page.check('input[name="vegetarian"]');
  await page.click('button:text("Apply Filters")');
  
  // Start elimination
  await page.click('button:text("Start 5-3-1")');
  
  // Eliminate items (simulate both users)
  await eliminateItemsAsUser(page, 'user1', ['item1', 'item2']);
  await eliminateItemsAsUser(page, 'user2', ['item3', 'item4']);
  
  // Verify result
  await expect(page.locator('[data-testid="final-result"]')).toBeVisible();
});
```

### Coverage Goals
- **Backend**: 70% line coverage
- **Frontend**: 70% line coverage
- **E2E**: Cover all primary user journeys

