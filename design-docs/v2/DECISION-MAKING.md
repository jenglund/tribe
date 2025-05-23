# Tribe Decision Making

## Overview

This document outlines the decision-making system for Tribe, including the advanced filtering engine and the K+M elimination algorithm that helps small groups collaboratively choose from lists of options.

## Core Algorithm: K+M Elimination

### Algorithm Parameters

```go
type DecisionDefaults struct {
    K int `json:"k"` // Eliminations per person (default: 2)
    M int `json:"m"` // Final options for random selection (default: 3)
    N int `json:"n"` // Tribe size (always 1-8)
}

type TribeDecisionPreferences struct {
    DefaultK int `json:"default_k"` // Configurable per tribe
    DefaultM int `json:"default_m"` // Configurable per tribe
    MaxK     int `json:"max_k"`     // Maximum eliminations allowed
    MaxM     int `json:"max_m"`     // Maximum final options allowed
}
```

### Process Flow

1. **Filtering Phase**: Apply user-configured filters to reduce the candidate pool
2. **Elimination Phase**: K rounds of elimination with N users each
3. **Selection Phase**: Random selection from final M candidates
4. **History Phase**: Record results and elimination history

## Advanced Filter Engine

### Filter Configuration

```go
type FilterItem struct {
    ID          string      `json:"id"`           // Unique filter identifier
    Type        string      `json:"type"`         // Filter category
    IsHard      bool        `json:"is_hard"`      // Required vs preferred
    Priority    int         `json:"priority"`     // Execution order (0 = highest)
    Criteria    interface{} `json:"criteria"`     // Type-specific data
    Description string      `json:"description"`  // Human-readable description
}

type FilterConfiguration struct {
    Items  []FilterItem `json:"items"`
    UserID string       `json:"user_id"`
}
```

### Filter Types and Criteria

```go
// Category filtering
type CategoryFilterCriteria struct {
    IncludeCategories []string `json:"include_categories"`
    ExcludeCategories []string `json:"exclude_categories"`
}

// Dietary restrictions
type DietaryFilterCriteria struct {
    RequiredOptions []string `json:"required_options"` // ["vegetarian", "vegan", "gluten_free"]
}

// Geographic filtering
type LocationFilterCriteria struct {
    CenterLat    float64 `json:"center_lat"`
    CenterLng    float64 `json:"center_lng"`
    MaxDistance  float64 `json:"max_distance"` // in miles
}

// Recent activity exclusion
type RecentActivityFilterCriteria struct {
    ExcludeDays int      `json:"exclude_days"`
    UserID      string   `json:"user_id"`
    TribeID     *string  `json:"tribe_id"`
}

// Business hours filtering
type OpeningHoursFilterCriteria struct {
    MustBeOpenFor    int    `json:"must_be_open_for"`    // minutes from now
    MustBeOpenUntil  string `json:"must_be_open_until"`  // time in user timezone
    UserTimezone     string `json:"user_timezone"`       // user's timezone
    CheckDate        *int64 `json:"check_date"`          // unix timestamp, optional
}

// Tag-based filtering
type TagFilterCriteria struct {
    RequiredTags []string `json:"required_tags"`
    ExcludedTags []string `json:"excluded_tags"`
}
```

### Filter Results and Scoring

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

### Filter Engine Implementation

The FilterEngine applies filters in priority order (lowest number = highest priority), enforces hard constraints, and scores soft preferences. Items failing hard filters are excluded entirely, while soft filter violations reduce the priority score but don't eliminate items.

**Key Features:**
- Priority-based filter execution
- Hard vs soft constraint handling
- Timezone-aware business hours filtering
- Recent activity exclusion to avoid repeats
- Geographic radius filtering

## Turn-Based Elimination System

### Session Management

```go
type EliminationSession struct {
    SessionID           string         `json:"session_id"`
    TribeID             string         `json:"tribe_id"`
    EliminationOrder    []string       `json:"elimination_order"`    // Randomized user IDs
    CurrentTurnIndex    int            `json:"current_turn_index"`   
    CurrentRound        int            `json:"current_round"`        
    TurnStartedAt       time.Time      `json:"turn_started_at"`
    TurnTimeoutMinutes  int            `json:"turn_timeout_minutes"`
    SkippedUsers        []SkippedTurn  `json:"skipped_users"`
    UserSkipCounts      map[string]int `json:"user_skip_counts"`     
    CurrentCandidates   []string       `json:"current_candidates"`   
}

type SkippedTurn struct {
    UserID     string    `json:"user_id"`
    Round      int       `json:"round"`
    TurnInRound int      `json:"turn_in_round"`
    SkipType   string    `json:"skip_type"`    // "quick_skip", "timeout_skip", "forfeited"
    SkippedAt  time.Time `json:"skipped_at"`
}
```

### Quick-Skip Functionality

Users can skip their turns up to K times per session (equal to the number of elimination rounds). Skipped turns are deferred to a "catch-up phase" after regular rounds complete. Users cannot quick-skip turns that were already deferred.

### Timeout Handling

- **Turn Timeouts**: Individual turn time limits (configurable per tribe)
- **Session Timeouts**: Overall session inactivity limits (30 minutes default)
- **Catch-up Timeouts**: Forfeits all remaining skipped turns for that user

### Elimination Status

```go
type EliminationStatus struct {
    SessionID         string        `json:"session_id"`
    CurrentCandidates []string      `json:"current_candidates"`
    CurrentUserTurn   string        `json:"current_user_turn"`
    IsYourTurn        bool          `json:"is_your_turn"`
    CurrentRound      int           `json:"current_round"`
    TurnTimeRemaining time.Duration `json:"turn_time_remaining"`
    EliminationOrder  []string      `json:"elimination_order"`
    SkippedUsers      []SkippedTurn `json:"skipped_users"`
    CanQuickSkip      bool          `json:"can_quick_skip"`
    QuickSkipsUsed    int           `json:"quick_skips_used"`
    QuickSkipsLimit   int           `json:"quick_skips_limit"`
    IsCatchUpPhase    bool          `json:"is_catch_up_phase"`
}
```

## Decision History and Results

### Session Lifecycle

1. **Configuring**: Setting up filters and parameters
2. **Eliminating**: Active elimination rounds
3. **Completed**: Final selection made
4. **Expired**: Timed out due to inactivity

### History Storage

```go
type DecisionResult struct {
    WinnerID           *string                    `json:"winner_id"`
    RunnersUpIDs       []string                   `json:"runners_up_ids"`
    EliminationHistory json.RawMessage            `json:"elimination_history"`
}

type DecisionHistory struct {
    SessionID     string                    `json:"session_id"`
    SessionName   string                    `json:"session_name"`
    Status        string                    `json:"status"`
    CreatedAt     time.Time                 `json:"created_at"`
    CompletedAt   *time.Time                `json:"completed_at"`
    IsPinned      bool                      `json:"is_pinned"`
    ShowDetails   bool                      `json:"show_details"`
    Winner        *ListItem                 `json:"winner"`
    RunnersUp     []ListItem                `json:"runners_up"`
    Eliminations  []DecisionHistoryEntry    `json:"eliminations"`
}
```

### History Retention

- **Default Retention**: 30 days for completed sessions
- **Pinned Sessions**: Retained indefinitely until manually unpinned
- **Expired Sessions**: Cleaned up automatically
- **Privacy Controls**: Elimination details shown based on tribe settings

## Frontend Integration

### TypeScript Interfaces

```typescript
interface FilterItem {
  id: string;
  type: 'category' | 'dietary' | 'location' | 'recent_activity' | 'opening_hours' | 'tags';
  isHard: boolean;
  priority: number;
  criteria: any;
  description: string;
}

interface FilterConfiguration {
  items: FilterItem[];
  userId: string;
}

interface EliminationStatus {
  sessionId: string;
  currentCandidates: string[];
  currentUserTurn: string;
  isYourTurn: boolean;
  currentRound: number;
  turnTimeRemaining: number;
  canQuickSkip: boolean;
  quickSkipsUsed: number;
  quickSkipsLimit: number;
}
```

### Mobile Optimization

Mobile clients use simplified status polling to reduce bandwidth and battery usage. The `MobileSessionStatus` provides essential information without full candidate details until needed.

## Configuration and Storage

### Filter Configuration Storage

User filter configurations are saved in the `filter_configurations` table, allowing reuse of common filter combinations with friendly names like "Date Night Filters" or "Quick Lunch".

### Tribe Preferences

Each tribe can configure default K and M values, timeout settings, and visibility preferences for elimination details. These settings balance customization with simplicity for small group decision-making.

## Implementation Notes

### Timezone Handling

The opening hours filter properly handles timezone conversions between user locations and business locations, supporting multi-timezone tribe decision-making.

### Performance Considerations

- Filter execution is optimized by priority ordering
- Hard filters short-circuit evaluation when possible
- Mobile endpoints provide simplified data structures
- Session cleanup runs automatically to prevent storage bloat

### Error Handling

The system gracefully handles missing data, timezone parsing errors, and user availability issues. Default assumptions favor inclusion (e.g., missing hours = always open) to avoid over-constraining results.