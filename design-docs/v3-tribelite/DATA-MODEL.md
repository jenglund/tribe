# TribeLite v3 - Data Model

## Database Schema

This document defines the simplified database schema for TribeLite v3, eliminating authentication, tribal governance, and complex permissions in favor of a trust-based local network model.

### Design Principles
- **Simplicity First**: Minimal tables, maximum functionality
- **Trust-Based**: No complex permissions or authentication
- **Local Network Optimized**: Designed for small groups with shared trust
- **Activity-Centric**: Track what people actually do, not just what they plan

## Core Tables

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL, -- How they appear in the app
    avatar_url VARCHAR(500), -- Local file path or URL
    preferences JSONB DEFAULT '{}'::jsonb, -- User preferences and settings
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_created_at ON users(created_at);
```

#### User Preferences JSON Structure
```json
{
  "dietary_restrictions": ["vegetarian", "gluten_free"],
  "default_location": {
    "lat": 40.7128,
    "lng": -74.0060,
    "address": "New York, NY",
    "max_distance_miles": 20
  },
  "decision_defaults": {
    "k_eliminations": 2,
    "m_final_options": 1
  },
  "timezone": "America/New_York",
  "ui_preferences": {
    "theme": "light",
    "show_tutorials": true,
    "compact_view": false
  }
}
```

### Lists Table
```sql
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    is_community BOOLEAN DEFAULT FALSE, -- Community lists vs personal lists
    category VARCHAR(100), -- 'restaurants', 'movies', 'activities', 'vacations'
    k_default INTEGER DEFAULT 2, -- Default eliminations per participant
    m_default INTEGER DEFAULT 1, -- Default final options
    metadata JSONB DEFAULT '{}'::jsonb, -- Category-specific settings
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_lists_created_by ON lists(created_by);
CREATE INDEX idx_lists_category ON lists(category);
CREATE INDEX idx_lists_is_community ON lists(is_community);
CREATE INDEX idx_lists_updated_at ON lists(updated_at);
```

#### List Metadata JSON Examples
```json
// Restaurant list metadata
{
  "price_range_filter": true,
  "cuisine_tags": ["italian", "mexican", "asian", "american"],
  "default_filters": {
    "max_distance_miles": 15,
    "exclude_recently_visited_days": 30
  }
}

// Movie list metadata  
{
  "streaming_services": ["netflix", "hulu", "disney+"],
  "genre_tags": ["comedy", "drama", "action", "horror"],
  "default_filters": {
    "exclude_recently_watched_days": 90
  }
}

// Activity list metadata
{
  "activity_types": ["indoor", "outdoor", "social", "solo"],
  "weather_dependent": true,
  "default_filters": {
    "exclude_recently_done_days": 60
  }
}
```

### List Items Table
```sql
CREATE TABLE list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100), -- Item-specific category (e.g., 'italian', 'comedy')
    tags TEXT[] DEFAULT '{}', -- Flexible tagging system
    location JSONB, -- Location information if applicable
    metadata JSONB DEFAULT '{}'::jsonb, -- Item-specific data
    added_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_list_items_list_id ON list_items(list_id);
CREATE INDEX idx_list_items_category ON list_items(category);
CREATE INDEX idx_list_items_tags ON list_items USING GIN(tags);
CREATE INDEX idx_list_items_added_by ON list_items(added_by);
```

#### Location JSON Structure
```json
{
  "address": "123 Main St, New York, NY 10001",
  "lat": 40.7128,
  "lng": -74.0060,
  "city": "New York",
  "state": "NY",
  "country": "USA",
  "postal_code": "10001"
}
```

#### Item Metadata JSON Examples
```json
// Restaurant item metadata
{
  "cuisine_type": "italian",
  "price_range": "$$$",
  "phone": "+1-555-123-4567",
  "website": "https://example-restaurant.com",
  "dietary_options": {
    "vegetarian": true,
    "vegan": false,
    "gluten_free": true
  },
  "business_hours": {
    "monday": {"open": "11:00", "close": "22:00"},
    "tuesday": {"open": "11:00", "close": "22:00"},
    "wednesday": {"open": "11:00", "close": "22:00"},
    "thursday": {"open": "11:00", "close": "22:00"},
    "friday": {"open": "11:00", "close": "23:00"},
    "saturday": {"open": "10:00", "close": "23:00"},
    "sunday": {"closed": true}
  }
}

// Movie item metadata
{
  "genre": "comedy",
  "year": 2023,
  "runtime_minutes": 120,
  "rating": "PG-13",
  "streaming_services": ["netflix", "hulu"],
  "imdb_rating": 7.5
}

// Activity item metadata
{
  "activity_type": "outdoor",
  "duration_hours": 2,
  "difficulty": "moderate",
  "weather_dependent": true,
  "equipment_needed": ["hiking boots", "water bottle"],
  "best_seasons": ["spring", "summer", "fall"]
}
```

### List Subscriptions Table
```sql
CREATE TABLE list_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    is_snoozed BOOLEAN DEFAULT FALSE, -- Hide from main view
    subscribed_at TIMESTAMPTZ DEFAULT NOW(),
    snoozed_until TIMESTAMPTZ, -- Temporary snooze with auto-restore
    UNIQUE(user_id, list_id)
);

-- Indexes
CREATE INDEX idx_list_subscriptions_user_id ON list_subscriptions(user_id);
CREATE INDEX idx_list_subscriptions_list_id ON list_subscriptions(list_id);
CREATE INDEX idx_list_subscriptions_snoozed ON list_subscriptions(is_snoozed, snoozed_until);
```

### Activities Table
```sql
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_item_id UUID NOT NULL REFERENCES list_items(id),
    planned_by UUID NOT NULL REFERENCES users(id), -- Who made the decision
    participants JSONB NOT NULL, -- Array of user IDs who participated/will participate
    status VARCHAR(50) DEFAULT 'planned', -- 'planned', 'confirmed', 'completed', 'cancelled'
    planned_date TIMESTAMPTZ, -- When it's planned to happen
    completed_date TIMESTAMPTZ, -- When it actually happened
    notes TEXT, -- User notes about the activity
    rating INTEGER CHECK (rating >= 1 AND rating <= 5), -- Simple 1-5 rating
    metadata JSONB DEFAULT '{}'::jsonb, -- Activity-specific data
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_activities_list_item_id ON activities(list_item_id);
CREATE INDEX idx_activities_planned_by ON activities(planned_by);
CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_planned_date ON activities(planned_date);
CREATE INDEX idx_activities_completed_date ON activities(completed_date);
CREATE INDEX idx_activities_participants ON activities USING GIN(participants);
```

#### Activity Metadata JSON Structure
```json
{
  "decision_context": {
    "eliminations_made": 3,
    "total_options": 8,
    "k_value": 2,
    "m_value": 1,
    "filters_applied": ["max_distance_10", "vegetarian_required"]
  },
  "completion_details": {
    "actual_duration_minutes": 90,
    "photos": ["/uploads/activity_123_photo1.jpg"],
    "weather": "sunny, 72°F",
    "cost_per_person": 35.50
  }
}
```

## Go Type Definitions

### Core Types
```go
type User struct {
    ID          string                 `json:"id" db:"id"`
    Name        string                 `json:"name" db:"name"`
    DisplayName string                 `json:"display_name" db:"display_name"`
    AvatarURL   *string                `json:"avatar_url" db:"avatar_url"`
    Preferences map[string]interface{} `json:"preferences" db:"preferences"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type List struct {
    ID          string                 `json:"id" db:"id"`
    Name        string                 `json:"name" db:"name"`
    Description *string                `json:"description" db:"description"`
    CreatedBy   string                 `json:"created_by" db:"created_by"`
    IsCommunity bool                   `json:"is_community" db:"is_community"`
    Category    *string                `json:"category" db:"category"`
    KDefault    int                    `json:"k_default" db:"k_default"`
    MDefault    int                    `json:"m_default" db:"m_default"`
    Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type ListItem struct {
    ID          string                 `json:"id" db:"id"`
    ListID      string                 `json:"list_id" db:"list_id"`
    Name        string                 `json:"name" db:"name"`
    Description *string                `json:"description" db:"description"`
    Category    *string                `json:"category" db:"category"`
    Tags        []string               `json:"tags" db:"tags"`
    Location    map[string]interface{} `json:"location" db:"location"`
    Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
    AddedBy     string                 `json:"added_by" db:"added_by"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type ListSubscription struct {
    ID           string     `json:"id" db:"id"`
    UserID       string     `json:"user_id" db:"user_id"`
    ListID       string     `json:"list_id" db:"list_id"`
    IsSnoozed    bool       `json:"is_snoozed" db:"is_snoozed"`
    SubscribedAt time.Time  `json:"subscribed_at" db:"subscribed_at"`
    SnoozedUntil *time.Time `json:"snoozed_until" db:"snoozed_until"`
}

type Activity struct {
    ID            string                 `json:"id" db:"id"`
    ListItemID    string                 `json:"list_item_id" db:"list_item_id"`
    PlannedBy     string                 `json:"planned_by" db:"planned_by"`
    Participants  []string               `json:"participants" db:"participants"`
    Status        string                 `json:"status" db:"status"`
    PlannedDate   *time.Time             `json:"planned_date" db:"planned_date"`
    CompletedDate *time.Time             `json:"completed_date" db:"completed_date"`
    Notes         *string                `json:"notes" db:"notes"`
    Rating        *int                   `json:"rating" db:"rating"`
    Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
    CreatedAt     time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}
```

### Decision Making Types
```go
type DecisionRequest struct {
    ListIDs      []string               `json:"list_ids"`
    Participants []string               `json:"participants"`
    KValue       *int                   `json:"k_value"` // Override list default
    MValue       *int                   `json:"m_value"` // Override list default
    Filters      map[string]interface{} `json:"filters"`
}

type DecisionCandidate struct {
    Item         ListItem `json:"item"`
    IsEliminated bool     `json:"is_eliminated"`
    EliminatedBy string   `json:"eliminated_by,omitempty"`
}

type DecisionResult struct {
    WinnerID       string              `json:"winner_id"`
    RunnersUp      []string            `json:"runners_up"`
    TotalOptions   int                 `json:"total_options"`
    Eliminations   int                 `json:"eliminations"`
    Participants   []string            `json:"participants"`
    FiltersApplied []string            `json:"filters_applied"`
    DecisionMadeBy string              `json:"decision_made_by"`
    DecisionMadeAt time.Time           `json:"decision_made_at"`
}
```

### API Request/Response Types
```go
type CreateUserRequest struct {
    Name        string                 `json:"name" binding:"required"`
    DisplayName string                 `json:"display_name" binding:"required"`
    AvatarURL   *string                `json:"avatar_url"`
    Preferences map[string]interface{} `json:"preferences"`
}

type CreateListRequest struct {
    Name        string                 `json:"name" binding:"required"`
    Description *string                `json:"description"`
    IsCommunity bool                   `json:"is_community"`
    Category    *string                `json:"category"`
    KDefault    *int                   `json:"k_default"`
    MDefault    *int                   `json:"m_default"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type CreateListItemRequest struct {
    Name        string                 `json:"name" binding:"required"`
    Description *string                `json:"description"`
    Category    *string                `json:"category"`
    Tags        []string               `json:"tags"`
    Location    map[string]interface{} `json:"location"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type CreateActivityRequest struct {
    ListItemID   string     `json:"list_item_id" binding:"required"`
    Participants []string   `json:"participants" binding:"required"`
    PlannedDate  *time.Time `json:"planned_date"`
    Notes        *string    `json:"notes"`
}
```

## Notable Simplifications from v2

### What We Removed
- **Authentication tables**: No users.password, oauth_providers, sessions, tokens
- **Tribe management**: No tribes, memberships, invitations, governance, voting
- **Complex permissions**: No permission levels, sharing restrictions, or access controls
- **Decision sessions**: No persistent elimination state, turn management, or timeouts
- **External sync**: No sync configurations, conflict resolution, or external API state

### What We Kept Simple
- **User profiles**: Basic info with flexible JSONB preferences
- **List ownership**: Creator tracking for organization, not security
- **Activity tracking**: What people actually do, not complex planning workflows
- **Flexible metadata**: JSONB for extensibility without schema complexity

### Benefits of Simplification
- **Fewer joins**: Most queries can be satisfied with single table lookups
- **No cascade complexity**: Simpler foreign key relationships
- **JSON flexibility**: Easy to extend without migrations
- **Trust-based model**: No complex authorization logic needed
- **Local network optimized**: No external service dependencies

This simplified data model supports all core functionality while eliminating the complexity that made v2 difficult to implement and potentially over-engineered for the target use case. 