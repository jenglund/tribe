# Tribe App - Consolidated Design Analysis

**Version:** Consolidated v1.0  
**Date:** January 2025  
**Based on:** Analysis of Claude4, ChatGPT4.5, ChatGPTo3, and Gemini2.5 design documents

## Executive Summary

This document consolidates the design recommendations from four AI models, reconciling differences and synthesizing the best practices into a unified specification for the Tribe collaborative decision-making application.

## Project Vision & Core Values

**Unified Vision**: Tribe is a collaborative decision-making web application designed to help small groups (1-8 people) make choices about activities, restaurants, entertainment, and shared experiences through structured list management and algorithmic decision processes.

**Core Values** (synthesized from all designs):
- **Simplicity**: Easy onboarding and intuitive user experience
- **Collaboration**: Designed for group decision-making without friction
- **Flexibility**: Adaptable to various decision types (restaurants, activities, entertainment)
- **Privacy**: Self-hostable, open-source with user data control
- **Small Scale**: Optimized for personal/friend groups, not enterprise scale
- **Reliability**: Robust error handling and graceful degradation

## Technology Stack (Consolidated)

### Core Technologies
- **Backend**: Go with Gin web framework
- **Frontend**: React with TypeScript
- **Database**: PostgreSQL 15+
- **API**: GraphQL for complex queries, REST for simple operations (hybrid approach)
- **Authentication**: OAuth 2.0 (Google primary) + JWT tokens
- **Deployment**: Docker-based for easy self-hosting

### Supporting Technologies
- **State Management**: React Context + useReducer (avoiding Redux complexity)
- **UI Framework**: shadcn/ui + Tailwind CSS
- **Testing**: 
  - Backend: Go testing package + testify
  - Frontend: Vitest + React Testing Library
  - E2E: Playwright
- **Build Tools**: Standard Go toolchain, Vite for frontend

## Comprehensive User Stories

### Authentication & User Management (US-AUTH)
- **US-AUTH-001**: As a new user, I want to sign up using Google OAuth so I can quickly access the app without passwords
- **US-AUTH-002**: As a user, I want to update my profile (name, avatar, preferences) so others can identify me
- **US-AUTH-003**: As a user, I want to set dietary preferences that automatically apply to my filters
- **US-AUTH-004**: As a user, I want to delete my account and all associated data for privacy

### Tribe Management (US-TRIBE)
- **US-TRIBE-001**: As a user, I want to create a new tribe (max 8 members) so I can collaborate with specific groups
- **US-TRIBE-002**: As a user, I want to invite others via email to join my tribe
- **US-TRIBE-003**: As a user, I want to accept/decline tribe invitations
- **US-TRIBE-004**: As a tribe member, I want to see all members and their roles
- **US-TRIBE-005**: As a tribe creator, I want to remove members when necessary
- **US-TRIBE-006**: As a tribe member, I want to leave a tribe I no longer want to participate in
- **US-TRIBE-007**: As a user, I want to be part of multiple tribes for different social contexts

### List Management (US-LIST)
- **US-LIST-001**: As a user, I want to create personal lists with categories (restaurants, movies, activities)
- **US-LIST-002**: As a tribe member, I want to create tribe lists that all members can edit
- **US-LIST-003**: As a user, I want to add detailed items to lists (name, description, location, tags, dietary info)
- **US-LIST-004**: As a list owner/editor, I want to edit or remove list items
- **US-LIST-005**: As a user, I want to share personal lists with tribes (read-only or editable)
- **US-LIST-006**: As a user, I want to organize lists by categories and tags
- **US-LIST-007**: As a user, I want to soft-delete lists with recovery within 30 days

### Activity Tracking (US-ACTIVITY)
- **US-ACTIVITY-001**: As a user, I want to log visits/completions with date, companions, and notes
- **US-ACTIVITY-002**: As a user, I want to rate experiences (1-5 scale)
- **US-ACTIVITY-003**: As a user, I want to see my activity history for any list item
- **US-ACTIVITY-004**: As a user, I want to filter out recently visited places (configurable timeframe)

### Decision Making (US-DECISION)
- **US-DECISION-001**: As a tribe member, I want to start a decision session using multiple lists
- **US-DECISION-002**: As a user, I want to apply multiple filters (cuisine, location, dietary, recency)
- **US-DECISION-003**: As a user, I want to get a single random result from filtered options
- **US-DECISION-004**: As a tribe, I want to use KN+M elimination process with configurable parameters
- **US-DECISION-005**: As a participant, I want the system to suggest optimal K and M values
- **US-DECISION-006**: As a user, I want to see decision history and outcomes
- **US-DECISION-007**: As a user, I want graceful handling when filters yield no results

## System Architecture

### High-Level Architecture
```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   React Frontend    │    │   Go Backend        │    │   PostgreSQL        │
│   (TypeScript)      │◄──►│   (Gin + GraphQL)   │◄──►│   (Database)        │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
         │                           │                          │
         │                           │                          │
         ▼                           ▼                          ▼
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   CDN/Static Files  │    │   External APIs     │    │   Redis (Cache)     │
│                     │    │   (OAuth, Maps)     │    │   (Optional)        │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

### Backend Structure
```
cmd/
├── server/
│   └── main.go                 # Application entry point
├── migrate/
│   └── main.go                 # Database migration tool
internal/
├── api/
│   ├── graphql/               # GraphQL schema and resolvers
│   ├── rest/                  # REST endpoints (auth, simple operations)
│   └── middleware/            # Auth, CORS, logging, rate limiting
├── auth/
│   ├── oauth.go              # OAuth providers
│   └── jwt.go                # JWT token management
├── models/                   # Domain models and validation
├── services/                 # Business logic layer
├── repository/               # Data access layer
├── filters/                  # Decision filtering engine
└── utils/                    # Shared utilities
migrations/                   # SQL migration files
tests/                       # Test files
docker/                      # Docker configuration
```

### Frontend Structure
```
src/
├── components/
│   ├── auth/                # Authentication components
│   ├── tribes/              # Tribe management
│   ├── lists/               # List management
│   ├── decisions/           # Decision-making flow
│   └── common/              # Reusable UI components
├── graphql/
│   ├── queries/             # GraphQL queries
│   ├── mutations/           # GraphQL mutations
│   └── fragments/           # Reusable fragments
├── hooks/                   # Custom React hooks
├── services/                # API client and utilities
├── store/                   # Context-based state management
├── types/                   # TypeScript definitions
└── utils/                   # Helper functions
```

## Database Schema (Consolidated)

### Core Design Decisions
- **Primary Keys**: UUIDs for all entities (better for distributed systems and external sync)
- **Naming**: snake_case for database columns (PostgreSQL convention)
- **Soft Deletion**: deleted_at timestamp for all major entities
- **Timestamps**: All tables include created_at, updated_at with timezone support

### Schema Definition

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    oauth_provider VARCHAR(50) NOT NULL, -- 'google', 'email' (future)
    oauth_id VARCHAR(255) NOT NULL,
    dietary_preferences JSONB DEFAULT '[]'::jsonb, -- ['vegetarian', 'vegan', 'gluten_free']
    location_preferences JSONB, -- Default location, max distance, etc.
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(oauth_provider, oauth_id)
);
```

#### Tribes Table
```sql
CREATE TABLE tribes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id),
    max_members INTEGER DEFAULT 8,
    decision_preferences JSONB DEFAULT '{}'::jsonb, -- Default K, M values, etc.
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

#### Tribe Memberships Table
```sql
CREATE TABLE tribe_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- 'creator', 'admin', 'member'
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tribe_id, user_id)
);
```

#### Lists Table (Unified Design)
```sql
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    list_type VARCHAR(50) NOT NULL, -- 'personal', 'tribe'
    owner_user_id UUID REFERENCES users(id), -- NULL for tribe lists
    owner_tribe_id UUID REFERENCES tribes(id), -- NULL for personal lists
    category VARCHAR(100), -- 'restaurants', 'movies', 'activities', etc.
    metadata JSONB DEFAULT '{}'::jsonb, -- Flexible metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT check_list_ownership CHECK (
        (list_type = 'personal' AND owner_user_id IS NOT NULL AND owner_tribe_id IS NULL) OR
        (list_type = 'tribe' AND owner_user_id IS NULL AND owner_tribe_id IS NOT NULL)
    )
);
```

#### List Items Table
```sql
CREATE TABLE list_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100), -- 'italian', 'mexican', 'comedy', 'thriller', etc.
    tags TEXT[] DEFAULT '{}', -- ['vegetarian', 'date_night', 'casual']
    location JSONB, -- {address, lat, lng, city, state, country}
    business_info JSONB, -- {hours, phone, website, price_range}
    dietary_info JSONB, -- {vegetarian: true, vegan: false, gluten_free: true}
    external_id VARCHAR(255), -- For future external API sync
    added_by_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

#### List Shares Table
```sql
CREATE TABLE list_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    shared_with_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    shared_with_tribe_id UUID REFERENCES tribes(id) ON DELETE CASCADE,
    permission_level VARCHAR(50) DEFAULT 'read', -- 'read', 'write'
    shared_by_user_id UUID NOT NULL REFERENCES users(id),
    shared_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT check_share_target CHECK (
        (shared_with_user_id IS NOT NULL AND shared_with_tribe_id IS NULL) OR
        (shared_with_user_id IS NULL AND shared_with_tribe_id IS NOT NULL)
    ),
    UNIQUE(list_id, shared_with_user_id),
    UNIQUE(list_id, shared_with_tribe_id)
);
```

#### Activity History Table
```sql
CREATE TABLE activity_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_item_id UUID NOT NULL REFERENCES list_items(id),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tribe_id UUID REFERENCES tribes(id), -- NULL if individual activity
    activity_type VARCHAR(50) DEFAULT 'visited', -- 'visited', 'watched', 'completed'
    completed_at DATE NOT NULL, -- Date of activity (not timestamp)
    duration_minutes INTEGER, -- Optional duration
    companions JSONB DEFAULT '[]'::jsonb, -- Array of user IDs and/or names
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(list_item_id, user_id, completed_at) -- Prevent duplicate same-day entries
);
```

#### Decision Sessions Table
```sql
CREATE TABLE decision_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id),
    name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'configuring', -- 'configuring', 'eliminating', 'completed', 'cancelled'
    filters JSONB DEFAULT '{}'::jsonb, -- Applied filter criteria
    algorithm_params JSONB NOT NULL, -- {k: 2, n: 2, m: 1, initial_count: 5}
    initial_candidates JSONB DEFAULT '[]'::jsonb, -- Array of list_item_ids
    current_candidates JSONB DEFAULT '[]'::jsonb, -- Remaining after eliminations
    final_selection_id UUID REFERENCES list_items(id),
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
```

#### Decision Session Lists Table
```sql
CREATE TABLE decision_session_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES decision_sessions(id) ON DELETE CASCADE,
    list_id UUID NOT NULL REFERENCES lists(id),
    UNIQUE(session_id, list_id)
);
```

#### Decision Eliminations Table
```sql
CREATE TABLE decision_eliminations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES decision_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    list_item_id UUID NOT NULL REFERENCES list_items(id),
    round_number INTEGER NOT NULL,
    eliminated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(session_id, user_id, list_item_id) -- User can't eliminate same item twice
);
```

### Database Indexes
```sql
-- Primary performance indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id);
CREATE INDEX idx_tribe_memberships_user ON tribe_memberships(user_id);
CREATE INDEX idx_tribe_memberships_tribe ON tribe_memberships(tribe_id);
CREATE INDEX idx_lists_owner_user ON lists(owner_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_lists_owner_tribe ON lists(owner_tribe_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_lists_type ON lists(list_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_items_list ON list_items(list_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_items_category ON list_items(category);
CREATE INDEX idx_list_items_tags ON list_items USING GIN(tags);
CREATE INDEX idx_activity_history_user ON activity_history(user_id);
CREATE INDEX idx_activity_history_item ON activity_history(list_item_id);
CREATE INDEX idx_activity_history_date ON activity_history(completed_at);
CREATE INDEX idx_decision_sessions_tribe ON decision_sessions(tribe_id);
CREATE INDEX idx_decision_sessions_status ON decision_sessions(status);

-- Soft delete indexes
CREATE INDEX idx_users_active ON users(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tribes_active ON tribes(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_lists_active ON lists(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_items_active ON list_items(id) WHERE deleted_at IS NULL;

-- Geospatial index for location queries (if PostGIS is available)
-- CREATE INDEX idx_list_items_location ON list_items USING GIST((location->>'lat')::float, (location->>'lng')::float);
```

## API Design (Hybrid GraphQL/REST)

### GraphQL Schema (Primary API)
```graphql
# User and Authentication
type User {
  id: ID!
  email: String!
  name: String!
  avatarUrl: String
  dietaryPreferences: [String!]!
  tribes: [Tribe!]!
  personalLists: [List!]!
  activityHistory: [ActivityEntry!]!
  createdAt: DateTime!
}

# Tribe Management
type Tribe {
  id: ID!
  name: String!
  description: String
  creator: User!
  members: [TribeMember!]!
  lists: [List!]!
  decisionSessions: [DecisionSession!]!
  maxMembers: Int!
  memberCount: Int!
  createdAt: DateTime!
}

type TribeMember {
  user: User!
  role: TribeMemberRole!
  joinedAt: DateTime!
}

enum TribeMemberRole {
  CREATOR
  ADMIN
  MEMBER
}

# List Management
type List {
  id: ID!
  name: String!
  description: String
  type: ListType!
  category: String
  owner: ListOwner
  items: [ListItem!]!
  shares: [ListShare!]!
  itemCount: Int!
  createdAt: DateTime!
  updatedAt: DateTime!
}

union ListOwner = User | Tribe

enum ListType {
  PERSONAL
  TRIBE
}

type ListItem {
  id: ID!
  name: String!
  description: String
  category: String
  tags: [String!]!
  location: Location
  businessInfo: BusinessInfo
  dietaryInfo: DietaryInfo!
  activityHistory: [ActivityEntry!]!
  addedBy: User!
  createdAt: DateTime!
}

type Location {
  address: String
  latitude: Float
  longitude: Float
  city: String
  state: String
  country: String
}

type BusinessInfo {
  hours: OpeningHours
  phone: String
  website: String
  priceRange: PriceRange
}

type DietaryInfo {
  vegetarian: Boolean!
  vegan: Boolean!
  glutenFree: Boolean!
  customTags: [String!]!
}

# Activity Tracking
type ActivityEntry {
  id: ID!
  listItem: ListItem!
  user: User!
  tribe: Tribe
  activityType: ActivityType!
  completedAt: Date!
  durationMinutes: Int
  companions: [Companion!]!
  rating: Int
  notes: String
  createdAt: DateTime!
}

enum ActivityType {
  VISITED
  WATCHED
  COMPLETED
}

type Companion {
  user: User
  name: String!
}

# Decision Making
type DecisionSession {
  id: ID!
  tribe: Tribe!
  name: String
  status: DecisionStatus!
  filters: FilterCriteria!
  algorithmParams: AlgorithmParams!
  sourceLists: [List!]!
  initialCandidates: [ListItem!]!
  currentCandidates: [ListItem!]!
  eliminations: [Elimination!]!
  finalSelection: ListItem
  createdBy: User!
  createdAt: DateTime!
  completedAt: DateTime
}

enum DecisionStatus {
  CONFIGURING
  ELIMINATING
  COMPLETED
  CANCELLED
}

type FilterCriteria {
  categories: [String!]!
  excludeCategories: [String!]!
  dietaryRequirements: [String!]!
  maxDistance: Float
  centerLocation: Location
  excludeRecentlyVisited: Boolean!
  recentlyVisitedDays: Int!
  mustBeOpenFor: Int
  priceRange: PriceRange
  tags: [String!]!
  excludeTags: [String!]!
}

type AlgorithmParams {
  k: Int! # Eliminations per person
  n: Int! # Number of participants
  m: Int! # Final choices for random selection
  initialCount: Int! # Total initial candidates
}

type Elimination {
  user: User!
  listItem: ListItem!
  roundNumber: Int!
  eliminatedAt: DateTime!
}

# Mutations
type Mutation {
  # User Management
  updateUserProfile(input: UpdateUserProfileInput!): User!
  deleteAccount: Boolean!
  
  # Tribe Management
  createTribe(input: CreateTribeInput!): Tribe!
  inviteToTribe(tribeId: ID!, email: String!): Boolean!
  joinTribe(inviteToken: String!): Tribe!
  leaveTribe(tribeId: ID!): Boolean!
  removeTribeMember(tribeId: ID!, userId: ID!): Boolean!
  deleteTribe(tribeId: ID!): Boolean!
  
  # List Management
  createList(input: CreateListInput!): List!
  updateList(id: ID!, input: UpdateListInput!): List!
  deleteList(id: ID!): Boolean!
  addListItem(listId: ID!, input: AddListItemInput!): ListItem!
  updateListItem(id: ID!, input: UpdateListItemInput!): ListItem!
  deleteListItem(id: ID!): Boolean!
  shareList(listId: ID!, input: ShareListInput!): ListShare!
  unshareList(shareId: ID!): Boolean!
  
  # Activity Tracking
  logActivity(input: LogActivityInput!): ActivityEntry!
  updateActivity(id: ID!, input: UpdateActivityInput!): ActivityEntry!
  deleteActivity(id: ID!): Boolean!
  
  # Decision Making
  createDecisionSession(input: CreateDecisionSessionInput!): DecisionSession!
  addListsToSession(sessionId: ID!, listIds: [ID!]!): DecisionSession!
  applyFilters(sessionId: ID!, filters: FilterCriteriaInput!): DecisionSession!
  startElimination(sessionId: ID!): DecisionSession!
  eliminateItem(sessionId: ID!, itemId: ID!): DecisionSession!
  completeDecision(sessionId: ID!): DecisionSession!
  cancelDecision(sessionId: ID!): DecisionSession!
}

# Queries
type Query {
  me: User
  tribe(id: ID!): Tribe
  list(id: ID!): List
  listItem(id: ID!): ListItem
  decisionSession(id: ID!): DecisionSession
  
  # Search and filtering
  searchLists(query: String!, type: ListType): [List!]!
  searchListItems(query: String!, filters: FilterCriteriaInput): [ListItem!]!
  suggestKMValues(availableCount: Int!, tribeSize: Int!): [AlgorithmParams!]!
}

# Subscriptions (for real-time features)
type Subscription {
  decisionSessionUpdated(sessionId: ID!): DecisionSession!
  tribeUpdated(tribeId: ID!): Tribe!
}
```

### REST Endpoints (Simple Operations)
```
# Authentication
POST   /api/v1/auth/google/login     # OAuth login
POST   /api/v1/auth/refresh          # Refresh JWT token
POST   /api/v1/auth/logout           # Logout

# Health and meta
GET    /api/v1/health               # Health check
GET    /api/v1/version              # API version

# File uploads (if needed)
POST   /api/v1/uploads/avatar       # Upload user avatar
POST   /api/v1/uploads/list-image   # Upload list item image
```

## Decision-Making Algorithm (Enhanced)

### Filter Engine
```go
type FilterEngine struct {
    db repository.Database
}

type FilterCriteria struct {
    Categories          []string    `json:"categories"`
    ExcludeCategories   []string    `json:"exclude_categories"`
    DietaryRequirements []string    `json:"dietary_requirements"`
    MaxDistance         *float64    `json:"max_distance"`
    CenterLocation      *Location   `json:"center_location"`
    ExcludeRecentDays   *int        `json:"exclude_recent_days"`
    MustBeOpenFor       *int        `json:"must_be_open_for"`
    PriceRange          *PriceRange `json:"price_range"`
    Tags                []string    `json:"tags"`
    ExcludeTags         []string    `json:"exclude_tags"`
    UserID              string      `json:"user_id"`
    TribeID             *string     `json:"tribe_id"`
}

func (fe *FilterEngine) ApplyFilters(ctx context.Context, items []ListItem, criteria FilterCriteria) ([]ListItem, error) {
    // Apply filters in order of selectivity (most restrictive first)
    filtered := items
    
    // 1. Category filters (high selectivity)
    if len(criteria.Categories) > 0 {
        filtered = fe.filterByCategories(filtered, criteria.Categories)
    }
    if len(criteria.ExcludeCategories) > 0 {
        filtered = fe.excludeCategories(filtered, criteria.ExcludeCategories)
    }
    
    // 2. Dietary requirements (high selectivity)
    if len(criteria.DietaryRequirements) > 0 {
        filtered = fe.filterByDietary(filtered, criteria.DietaryRequirements)
    }
    
    // 3. Location constraints (medium selectivity)
    if criteria.MaxDistance != nil && criteria.CenterLocation != nil {
        filtered = fe.filterByDistance(filtered, *criteria.CenterLocation, *criteria.MaxDistance)
    }
    
    // 4. Recent activity filters (medium selectivity)
    if criteria.ExcludeRecentDays != nil {
        var err error
        filtered, err = fe.filterByRecentActivity(ctx, filtered, criteria.UserID, criteria.TribeID, *criteria.ExcludeRecentDays)
        if err != nil {
            return nil, err
        }
    }
    
    // 5. Opening hours (low selectivity, expensive to calculate)
    if criteria.MustBeOpenFor != nil {
        filtered = fe.filterByOpeningHours(filtered, *criteria.MustBeOpenFor)
    }
    
    // 6. Price range (low selectivity)
    if criteria.PriceRange != nil {
        filtered = fe.filterByPriceRange(filtered, *criteria.PriceRange)
    }
    
    // 7. Tag filters (variable selectivity)
    if len(criteria.Tags) > 0 {
        filtered = fe.filterByTags(filtered, criteria.Tags)
    }
    if len(criteria.ExcludeTags) > 0 {
        filtered = fe.excludeTags(filtered, criteria.ExcludeTags)
    }
    
    return filtered, nil
}
```

### KN+M Algorithm
```go
type DecisionAlgorithm struct {
    filterEngine *FilterEngine
}

type AlgorithmParams struct {
    K            int `json:"k"`              // Eliminations per person
    N            int `json:"n"`              // Number of participants
    M            int `json:"m"`              // Final random selections
    InitialCount int `json:"initial_count"`  // Total candidates to start with
}

func (da *DecisionAlgorithm) SuggestOptimalParams(availableCount, tribeSize int, preferences UserPreferences) []AlgorithmParams {
    suggestions := []AlgorithmParams{}
    
    maxK := preferences.MaxEliminationsPerPerson
    if maxK == 0 {
        maxK = 3 // Default maximum
    }
    
    maxM := preferences.MaxFinalOptions
    if maxM == 0 {
        maxM = 3 // Default maximum
    }
    
    // Generate valid combinations
    for k := 0; k <= maxK; k++ {
        for m := 1; m <= maxM; m++ {
            initialNeeded := k*tribeSize + m
            if initialNeeded <= availableCount {
                suggestions = append(suggestions, AlgorithmParams{
                    K:            k,
                    N:            tribeSize,
                    M:            m,
                    InitialCount: initialNeeded,
                })
            }
        }
    }
    
    // Sort by preference (prefer higher K, then lower M)
    sort.Slice(suggestions, func(i, j int) bool {
        if suggestions[i].K != suggestions[j].K {
            return suggestions[i].K > suggestions[j].K
        }
        return suggestions[i].M < suggestions[j].M
    })
    
    return suggestions
}

func (da *DecisionAlgorithm) ProcessElimination(ctx context.Context, sessionID string, userID string, eliminatedItemIDs []string) (*DecisionSession, error) {
    session, err := da.getSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Validate elimination count
    if len(eliminatedItemIDs) > session.AlgorithmParams.K {
        return nil, errors.New("too many eliminations")
    }
    
    // Record eliminations
    for _, itemID := range eliminatedItemIDs {
        if err := da.recordElimination(ctx, sessionID, userID, itemID); err != nil {
            return nil, err
        }
    }
    
    // Check if all participants have completed their eliminations
    remainingParticipants, err := da.getRemainingParticipants(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    if len(remainingParticipants) == 0 {
        // All eliminations complete, finalize session
        return da.finalizeSession(ctx, sessionID)
    }
    
    return session, nil
}

func (da *DecisionAlgorithm) finalizeSession(ctx context.Context, sessionID string) (*DecisionSession, error) {
    session, err := da.getSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Get remaining candidates after all eliminations
    remaining, err := da.getRemainingCandidates(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    var finalSelection *ListItem
    if len(remaining) == 1 {
        finalSelection = &remaining[0]
    } else if len(remaining) > 1 {
        // Random selection from remaining candidates
        idx := rand.Intn(len(remaining))
        finalSelection = &remaining[idx]
    }
    
    // Update session status and final selection
    session.Status = "completed"
    session.FinalSelection = finalSelection
    session.CompletedAt = time.Now()
    
    return da.updateSession(ctx, session)
}
```

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
- **Backend**: 90% line coverage, 95% for critical business logic
- **Frontend**: 85% line coverage, 90% for business components
- **E2E**: Cover all primary user journeys

## Implementation Roadmap (Detailed)

### Phase 1: Foundation (Weeks 1-4)
**Goal**: Core authentication, user management, and basic tribe functionality

#### Week 1: Project Setup
- [ ] Initialize Go backend with Gin framework
- [ ] Set up PostgreSQL with Docker
- [ ] Create initial database migrations
- [ ] Set up React frontend with TypeScript and Vite
- [ ] Configure testing infrastructure (Jest, Playwright)
- [ ] Set up CI/CD pipeline (GitHub Actions)

#### Week 2: Authentication System
- [ ] Implement Google OAuth integration
- [ ] Create JWT token management
- [ ] Build authentication middleware
- [ ] Create user registration/profile API
- [ ] Build login/signup frontend components
- [ ] Add authentication state management

#### Week 3: Basic User Management
- [ ] Complete user profile CRUD operations
- [ ] Implement soft deletion for users
- [ ] Create user preferences system
- [ ] Build profile management UI
- [ ] Add avatar upload functionality

#### Week 4: Tribe Foundation
- [ ] Implement tribe CRUD operations
- [ ] Create tribe membership system
- [ ] Build invitation system (email-based)
- [ ] Create tribe management UI
- [ ] Add member management features

### Phase 2: List Management (Weeks 5-8)
**Goal**: Complete list creation, sharing, and item management

#### Week 5: List Infrastructure
- [ ] Create list data models and migrations
- [ ] Implement list CRUD API endpoints
- [ ] Build basic GraphQL schema for lists
- [ ] Create list creation and editing UI
- [ ] Add list categories and metadata

#### Week 6: List Items
- [ ] Implement list item CRUD operations
- [ ] Add comprehensive item metadata support
- [ ] Build item creation/editing UI
- [ ] Implement tag and category systems
- [ ] Add location data support

#### Week 7: List Sharing
- [ ] Create list sharing system
- [ ] Implement permission levels (read/write)
- [ ] Build sharing UI components
- [ ] Add shared list discovery
- [ ] Implement share notification system

#### Week 8: Activity Tracking
- [ ] Create activity history system
- [ ] Implement visit logging
- [ ] Build activity tracking UI
- [ ] Add rating and companion tracking
- [ ] Create activity history views

### Phase 3: Decision Making (Weeks 9-12)
**Goal**: Core decision-making functionality with filtering and KN+M algorithm

#### Week 9: Decision Infrastructure
- [ ] Design decision session data model
- [ ] Implement basic decision API endpoints
- [ ] Create filtering engine
- [ ] Build decision session management
- [ ] Add GraphQL mutations for decisions

#### Week 10: Filtering System
- [ ] Implement comprehensive filter criteria
- [ ] Add location-based filtering
- [ ] Create dietary restriction filtering
- [ ] Build recent activity filtering
- [ ] Add opening hours filtering

#### Week 11: KN+M Algorithm
- [ ] Implement KN+M selection algorithm
- [ ] Create parameter suggestion system
- [ ] Build elimination tracking
- [ ] Add random selection logic
- [ ] Implement edge case handling

#### Week 12: Decision UI
- [ ] Create decision wizard component
- [ ] Build filter selection interface
- [ ] Implement elimination round UI
- [ ] Add result display and confirmation
- [ ] Create decision history views

### Phase 4: Polish & Production (Weeks 13-16)
**Goal**: Production-ready application with deployment and optimization

#### Week 13: Real-time Features
- [ ] Implement WebSocket support for live decisions
- [ ] Add real-time tribe member updates
- [ ] Create notification system
- [ ] Build collaborative features
- [ ] Add progress indicators

#### Week 14: Performance & Testing
- [ ] Comprehensive testing and bug fixes
- [ ] Performance optimization
- [ ] Add caching layer (Redis)
- [ ] Implement rate limiting
- [ ] Add monitoring and logging

#### Week 15: UI/UX Polish
- [ ] UI/UX improvements and responsive design
- [ ] Add loading states and error handling
- [ ] Implement accessibility features
- [ ] Create onboarding flow
- [ ] Add empty states and help text

#### Week 16: Deployment
- [ ] Create Docker deployment configuration
- [ ] Set up production database
- [ ] Implement backup and recovery
- [ ] Create deployment documentation
- [ ] Set up monitoring and alerting

## Known Issues & Future Work

### Known Technical Considerations
1. **Real-time Collaboration**: WebSocket implementation needed for smooth elimination rounds
2. **Location Services**: Consider PostGIS for advanced geographic queries
3. **Caching Strategy**: Redis for session data and frequently accessed lists
4. **Performance**: Database query optimization for complex filters
5. **Mobile Experience**: Progressive Web App features for mobile usage

### High Priority Future Work
- [ ] Advanced filtering options (price range, ratings, custom criteria)
- [ ] Email notification system for invites and decisions
- [ ] List import/export functionality
- [ ] Decision analytics and insights
- [ ] Mobile application development

### Medium Priority Future Work
- [ ] Integration with external restaurant/activity APIs
- [ ] Advanced geographic features with mapping
- [ ] Social features (comments, recommendations)
- [ ] Multi-language support
- [ ] Advanced admin panel for self-hosted instances

### Low Priority Future Work
- [ ] AI-powered recommendations
- [ ] Integration with calendar applications
- [ ] Advanced analytics and reporting
- [ ] Plugin architecture for extensibility

## Development Guidelines for AI Collaboration

### Code Quality Standards
1. **Test-Driven Development**: Write tests before implementation
2. **Clear Documentation**: Comment complex business logic thoroughly
3. **Type Safety**: Leverage TypeScript and Go's type system fully
4. **Error Handling**: Implement comprehensive error handling and logging
5. **Consistent Patterns**: Follow established architectural patterns

### AI Agent Instructions
1. **Incremental Changes**: Make small, verifiable changes
2. **Test Coverage**: Ensure new code has appropriate test coverage
3. **Schema Consistency**: Database changes must include proper migrations
4. **API Consistency**: Follow GraphQL schema and REST conventions
5. **Documentation Updates**: Update relevant documentation with changes

### Review Checklist
- [ ] All tests pass with adequate coverage
- [ ] GraphQL schema is valid and consistent
- [ ] Database migrations are reversible
- [ ] API endpoints have proper authentication/authorization
- [ ] Frontend components are properly typed
- [ ] Error handling is comprehensive
- [ ] Documentation is updated

This consolidated design provides a comprehensive foundation for implementing the Tribe application, incorporating the best ideas from all four design approaches while maintaining internal consistency and following best practices. 