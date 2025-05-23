# Advanced Filtering System

## Overview

The Tribe app uses a sophisticated priority-based filtering system that allows users to apply multiple criteria when selecting items for decision sessions. The system distinguishes between hard filters (required) and soft filters (preferred), providing intelligent ranking and graceful degradation when constraints are too restrictive.

## Filter Engine Architecture

### Core Concepts

- **Hard Filters**: Required criteria that items must meet to be included
- **Soft Filters**: Preferred criteria that influence ranking but don't exclude items
- **Priority System**: User-defined ordering of filter importance
- **Violation Tracking**: Detailed reporting of which filters items fail
- **Score Calculation**: Weighted scoring based on filter priority and compliance

### Filter Configuration Structure

```go
type FilterItem struct {
    ID          string      `json:"id"`           // Unique filter identifier
    Type        string      `json:"type"`         // "category", "dietary", "location", etc.
    IsHard      bool        `json:"is_hard"`      // true = required, false = preferred
    Priority    int         `json:"priority"`     // User-defined order (0 = highest)
    Criteria    interface{} `json:"criteria"`     // Type-specific filter data
    Description string      `json:"description"`  // Human-readable description
}

type FilterConfiguration struct {
    Items  []FilterItem `json:"items"`
    UserID string       `json:"user_id"`
}
```

## Filter Types & Criteria

### Category Filtering
Allows inclusion/exclusion of specific cuisine or activity categories.

```go
type CategoryFilterCriteria struct {
    IncludeCategories []string `json:"include_categories"` // ["italian", "mexican"]
    ExcludeCategories []string `json:"exclude_categories"` // ["fast_food", "chain"]
}
```

**Use Cases:**
- Include only specific cuisines for a themed dinner
- Exclude fast food for a special occasion
- Focus on specific activity types (outdoor, cultural, etc.)

### Dietary Restriction Filtering
Ensures items meet dietary requirements and preferences.

```go
type DietaryFilterCriteria struct {
    RequiredOptions []string `json:"required_options"` // ["vegetarian", "vegan", "gluten_free"]
}
```

**Use Cases:**
- Hard filter for allergies and religious restrictions
- Soft filter for dietary preferences
- Accommodate group members with different dietary needs

### Location-Based Filtering
Filters items by geographic proximity and accessibility.

```go
type LocationFilterCriteria struct {
    CenterLat    float64 `json:"center_lat"`
    CenterLng    float64 `json:"center_lng"`
    MaxDistance  float64 `json:"max_distance"` // in miles
}
```

**Use Cases:**
- Limit options to walking distance from a meeting point
- Filter by proximity to public transportation
- Exclude items beyond driving distance

### Recent Activity Filtering
Excludes items visited recently to encourage variety.

```go
type RecentActivityFilterCriteria struct {
    ExcludeDays int      `json:"exclude_days"`    // Don't include if visited within N days
    UserID      string   `json:"user_id"`         // Which user's activity to check
    TribeID     *string  `json:"tribe_id"`        // Optional: check tribe activity
}
```

**Use Cases:**
- Avoid restaurants visited in the last week
- Different recency rules for different activity types
- Consider both individual and group visit history

### Opening Hours Filtering
Ensures venues will be open and available when needed.

```go
type OpeningHoursFilterCriteria struct {
    MustBeOpenFor    int    `json:"must_be_open_for"`    // minutes from now
    MustBeOpenUntil  string `json:"must_be_open_until"`  // time like "23:00"
    UserTimezone     string `json:"user_timezone"`       // user's timezone
    CheckDate        *int64 `json:"check_date"`          // unix timestamp, defaults to now
}
```

**Use Cases:**
- Ensure restaurants will be open for dinner duration
- Check availability for planned future activities
- Account for time zones when traveling

### Tag-Based Filtering
Filters based on custom tags and attributes.

```go
type TagFilterCriteria struct {
    RequiredTags []string `json:"required_tags"` // ["outdoor", "date_night"]
    ExcludedTags []string `json:"excluded_tags"` // ["loud", "crowded"]
}
```

**Use Cases:**
- Find venues suitable for specific occasions
- Filter by atmosphere preferences
- Include/exclude based on special features

## Priority-Based Processing

### Filter Evaluation Order
Filters are processed in priority order (lowest number = highest priority):

1. **Hard Filters First**: Any item failing a hard filter is immediately excluded
2. **Soft Filters Second**: Remaining items are scored based on soft filter compliance
3. **Priority Weighting**: Earlier filters carry more weight in scoring
4. **Tie Breaking**: Lower violation count breaks ties in priority scores

### Scoring Algorithm

```go
func (fe *FilterEngine) calculatePriorityScore(softResults []SoftFilterResult) float64 {
    score := 0.0
    totalWeight := 0.0
    
    for _, result := range softResults {
        // Earlier filters (lower priority number) have higher weight
        weight := 1.0 / float64(result.Priority + 1)
        totalWeight += weight
        
        if result.Passed {
            score += weight
        }
    }
    
    if totalWeight == 0 {
        return 1.0 // No soft filters = perfect score
    }
    
    return score / totalWeight
}
```

### Results Structure

```go
type FilterResult struct {
    Item                ListItem           `json:"item"`
    PassedHardFilters   bool              `json:"passed_hard_filters"`
    SoftFilterResults   []SoftFilterResult `json:"soft_filter_results"`
    ViolationCount      int               `json:"violation_count"`
    PriorityScore       float64           `json:"priority_score"`
}

type SoftFilterResult struct {
    FilterID    string `json:"filter_id"`
    FilterType  string `json:"filter_type"`
    Passed      bool   `json:"passed"`
    Priority    int    `json:"priority"`
    Description string `json:"description"`
}
```

## Advanced Time-Based Filtering

### Timezone-Aware Processing
The system handles complex timezone scenarios for business hours:

```go
func (fe *FilterEngine) evaluateOpeningHoursFilter(item ListItem, criteria OpeningHoursFilterCriteria) bool {
    businessInfo := item.BusinessInfo
    if businessInfo == nil {
        return true // No business info = assume always open
    }
    
    // Get business timezone
    businessTimezone, _ := businessInfo["timezone"].(string)
    if businessTimezone == "" {
        businessTimezone = "UTC"
    }
    
    // Convert to business timezone for accurate hour checking
    businessLocation, err := time.LoadLocation(businessTimezone)
    if err != nil {
        businessLocation = time.UTC
    }
    
    // Handle user's check time requirements
    checkTime := time.Now()
    if criteria.CheckDate != nil {
        checkTime = time.Unix(*criteria.CheckDate, 0)
    }
    
    businessTime := checkTime.In(businessLocation)
    
    // Evaluate business hours for the check date
    return fe.checkBusinessHours(businessTime, businessInfo, criteria)
}
```

### Special Time Handling
- **24:00 Hours**: Correctly handled as midnight next day
- **Cross-Midnight**: Businesses closing after midnight properly evaluated
- **Holiday Hours**: Can be extended to support special schedules
- **Seasonal Variations**: Framework supports varying schedules

## Filter Configuration Storage

### Saved Configurations
Users can save and reuse filter configurations:

```sql
CREATE TABLE filter_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- "Date Night Filters", "Quick Lunch", etc.
    is_default BOOLEAN DEFAULT FALSE,
    configuration JSONB NOT NULL, -- FilterConfiguration JSON
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Default Configurations
- **Quick Lunch**: Short travel distance, fast service, open now
- **Date Night**: Romantic atmosphere, good ratings, reservations available
- **Family Meal**: Kid-friendly, dietary accommodations, moderate price
- **Adventure**: New places, highly rated, unique experiences

## Integration with Decision Process

### Pre-Decision Filtering
Before starting elimination rounds:

1. **Apply Hard Filters**: Immediately exclude non-qualifying items
2. **Rank by Soft Filters**: Sort remaining items by priority score
3. **Provide Feedback**: Show users which filters eliminated items
4. **Suggest Adjustments**: Recommend relaxing filters if too few items remain

### Dynamic Filtering
During decision sessions:

1. **Real-Time Updates**: Business hours checked against current time
2. **Progressive Relaxation**: Automatically suggest filter modifications if needed
3. **Manual Override**: Allow users to temporarily bypass specific filters
4. **Session Persistence**: Maintain filter state throughout decision process

## Error Handling & Edge Cases

### No Results Handling
When filters eliminate all options:

```go
type FilterSuggestion struct {
    Message     string              `json:"message"`
    Suggestions []RelaxationOption  `json:"suggestions"`
}

type RelaxationOption struct {
    FilterID    string  `json:"filter_id"`
    Description string  `json:"description"`
    Impact      int     `json:"impact"`     // How many items this would add back
}
```

### Common Issues & Solutions

1. **Over-Filtering**: Suggest removing lowest-priority soft filters
2. **Conflicting Criteria**: Identify and highlight logical conflicts
3. **External API Failures**: Graceful degradation when location/hours data unavailable
4. **Performance Issues**: Optimize filter evaluation for large datasets

## Frontend Filter Management

### Filter Builder Interface
```typescript
interface FilterBuilderProps {
    configuration: FilterConfiguration;
    onConfigurationChange: (config: FilterConfiguration) => void;
    availableFilters: FilterType[];
    previewResults?: FilterPreview;
}

interface FilterPreview {
    totalItems: number;
    filteredItems: number;
    eliminatedByFilter: Record<string, number>; // filterID -> count eliminated
}
```

### Real-Time Preview
As users build filter configurations:

- **Live Counts**: Show how many items each filter affects
- **Visual Feedback**: Highlight filters that eliminate too many/few items
- **Suggestions**: Recommend commonly used filter combinations
- **Validation**: Prevent impossible filter combinations

## Performance Optimizations

### Caching Strategy
- **Filter Results**: Cache results for common filter combinations
- **Business Hours**: Cache parsed hours data with invalidation
- **Location Calculations**: Cache distance calculations between points
- **User Preferences**: Cache frequently-used filter configurations

### Database Optimization
```sql
-- Optimized indexes for common filter queries
CREATE INDEX idx_list_items_category ON list_items(category);
CREATE INDEX idx_list_items_tags ON list_items USING GIN(tags);
CREATE INDEX idx_list_items_dietary ON list_items USING GIN(dietary_info);

-- Geospatial index for location queries (if PostGIS available)
-- CREATE INDEX idx_list_items_location ON list_items USING GIST(
--     (location->>'lat')::float, (location->>'lng')::float
-- );
```

### Query Optimization
- **Filter Ordering**: Apply most selective filters first
- **Batch Processing**: Group similar filter evaluations
- **Lazy Evaluation**: Only compute expensive filters when necessary
- **Result Limiting**: Stop processing once enough results found

## Future Enhancements

### Advanced Filter Types
- **Price Range**: Filter by cost estimates and budgets
- **Ratings**: Filter by user or external ratings
- **Popularity**: Filter by how often items are selected
- **Weather**: Consider weather conditions for outdoor activities
- **Traffic**: Factor in real-time travel conditions

### Machine Learning Integration
- **Preference Learning**: Automatically adjust filters based on user behavior
- **Collaborative Filtering**: Suggest filters based on similar users
- **Outcome Prediction**: Predict which filter combinations lead to successful decisions
- **Dynamic Weighting**: Adjust filter importance based on historical success

### External Data Integration
- **Real-Time Hours**: Live business hours from Google/Yelp
- **Event Data**: Consider special events affecting availability
- **Seasonal Adjustments**: Automatic seasonal filter modifications
- **Social Signals**: Incorporate trending and popular locations

---

*This document defines the advanced filtering system architecture. See [Decision Making Algorithm](./decision-making.md) for integration with the elimination process and [Database Schema](../database-schema.md) for data structure.* 