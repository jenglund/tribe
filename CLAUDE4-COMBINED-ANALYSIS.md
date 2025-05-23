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
- **US-TRIBE-002**: As a tribe member, I want to invite others via email to join our tribe
- **US-TRIBE-003**: As a user, I want to accept/decline tribe invitations
- **US-TRIBE-004**: As a tribe member, I want to see all members and understand who the senior member is
- **US-TRIBE-005**: As a tribe member, I want to remove other members when necessary (with appropriate safeguards)
- **US-TRIBE-006**: As a tribe member, I want to leave a tribe I no longer want to participate in
- **US-TRIBE-007**: As a user, I want to be part of multiple tribes for different social contexts
- **US-TRIBE-008**: As a tribe member, I want to update tribe settings (name, description, preferences) collectively
- **US-TRIBE-009**: As a tribe member, I want conflicts resolved fairly using the senior member as tie-breaker

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
- **Table Names**: Plural form (users, tribes, lists, etc.)
- **Soft Deletion**: Deferred for MVP - may use hard deletes with confirmation prompts
- **Timestamps**: All tables include created_at, updated_at with timezone support

### Schema Definition

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL, -- Global default display name
    avatar_url VARCHAR(500),
    oauth_provider VARCHAR(50) NOT NULL, -- 'google', 'dev' (for development)
    oauth_id VARCHAR(255) NOT NULL,
    dietary_preferences JSONB DEFAULT '[]'::jsonb, -- ['vegetarian', 'vegan', 'gluten_free']
    location_preferences JSONB, -- Default location, max distance, etc.
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(oauth_provider, oauth_id)
);
```

#### Tribes Table
```sql
CREATE TABLE tribes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id), -- Informational only, not functional
    max_members INTEGER DEFAULT 8,
    decision_preferences JSONB DEFAULT '{"k": 2, "m": 3}'::jsonb, -- Default K=2, M=3
    show_elimination_details BOOLEAN DEFAULT TRUE, -- Configurable elimination visibility
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Tribe Memberships Table
```sql
CREATE TABLE tribe_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- 'creator', 'member' (all have equal rights)
    tribe_display_name VARCHAR(255), -- User's display name within this tribe
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    is_inactive BOOLEAN DEFAULT FALSE,
    UNIQUE(tribe_id, user_id)
);

-- Function to get senior member (longest-standing) for tie-breaking
CREATE OR REPLACE FUNCTION get_tribe_senior_member(tribe_uuid UUID)
RETURNS UUID AS $$
BEGIN
    RETURN (
        SELECT user_id 
        FROM tribe_memberships 
        WHERE tribe_id = tribe_uuid 
        ORDER BY joined_at ASC 
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql;
```

#### Lists Table (Using owner_type/owner_id approach)
```sql
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_type VARCHAR(50) NOT NULL, -- 'user', 'tribe'
    owner_id UUID NOT NULL, -- References users.id or tribes.id based on owner_type
    category VARCHAR(100), -- 'restaurants', 'movies', 'activities', etc.
    metadata JSONB DEFAULT '{}'::jsonb, -- Flexible metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
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
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### List Shares Table (Read-only sharing only)
```sql
CREATE TABLE list_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    shared_with_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    shared_with_tribe_id UUID REFERENCES tribes(id) ON DELETE CASCADE,
    permission_level VARCHAR(50) DEFAULT 'read', -- Only 'read' for now
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
    activity_status VARCHAR(50) DEFAULT 'confirmed', -- 'confirmed', 'tentative', 'cancelled'
    completed_at TIMESTAMPTZ NOT NULL, -- When activity happened/will happen
    duration_minutes INTEGER, -- Optional duration
    participants JSONB DEFAULT '[]'::jsonb, -- Array of user IDs who participated
    notes TEXT,
    recorded_by_user_id UUID NOT NULL REFERENCES users(id), -- Who logged this entry
    decision_session_id UUID REFERENCES decision_sessions(id), -- If from decision result
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Decision Sessions Table
```sql
CREATE TABLE decision_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id),
    name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'configuring', -- 'configuring', 'eliminating', 'completed', 'expired', 'cancelled'
    filters JSONB DEFAULT '{}'::jsonb, -- Applied filter criteria
    algorithm_params JSONB NOT NULL, -- {k: 2, n: 2, m: 3, initial_count: 7}
    elimination_order JSONB DEFAULT '[]'::jsonb, -- Randomized user order for turns
    current_turn_index INTEGER DEFAULT 0, -- Index in elimination_order array
    current_round INTEGER DEFAULT 1, -- Which elimination round (1 to K)
    turn_started_at TIMESTAMPTZ, -- When current turn started (for timeout)
    turn_timeout_minutes INTEGER DEFAULT 5, -- Timeout per turn
    session_timeout_minutes INTEGER DEFAULT 30, -- Overall session inactivity timeout
    last_activity_at TIMESTAMPTZ DEFAULT NOW(), -- Track overall session activity
    skipped_users JSONB DEFAULT '[]'::jsonb, -- Users who were skipped and their missed turns
    initial_candidates JSONB DEFAULT '[]'::jsonb, -- Array of list_item_ids
    current_candidates JSONB DEFAULT '[]'::jsonb, -- Remaining after eliminations
    final_selection_id UUID REFERENCES list_items(id),
    runners_up JSONB DEFAULT '[]'::jsonb, -- The M set (other final candidates)
    elimination_history JSONB DEFAULT '[]'::jsonb, -- Complete elimination timeline
    is_pinned BOOLEAN DEFAULT FALSE, -- Prevent automatic cleanup
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ -- For automatic cleanup (1 month after completion)
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
CREATE INDEX idx_lists_owner ON lists(owner_type, owner_id);
CREATE INDEX idx_lists_category ON lists(category);
CREATE INDEX idx_list_items_list ON list_items(list_id);
CREATE INDEX idx_list_items_category ON list_items(category);
CREATE INDEX idx_list_items_tags ON list_items USING GIN(tags);
CREATE INDEX idx_activity_history_user ON activity_history(user_id);
CREATE INDEX idx_activity_history_item ON activity_history(list_item_id);
CREATE INDEX idx_activity_history_date ON activity_history(completed_at);
CREATE INDEX idx_decision_sessions_tribe ON decision_sessions(tribe_id);
CREATE INDEX idx_decision_sessions_status ON decision_sessions(status);
CREATE INDEX idx_list_shares_list ON list_shares(list_id);
CREATE INDEX idx_list_shares_user ON list_shares(shared_with_user_id);
CREATE INDEX idx_list_shares_tribe ON list_shares(shared_with_tribe_id);

-- Filter configuration indexes
CREATE INDEX idx_filter_configurations_user ON filter_configurations(user_id);
CREATE INDEX idx_filter_configurations_default ON filter_configurations(user_id, is_default) WHERE is_default = true;

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
  creator: User! # Informational only - all members have equal rights
  seniorMember: User! # Longest-standing member for tie-breaking
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
  isCreator: Boolean! # Informational flag
  isSenior: Boolean! # True for longest-standing member
}

enum TribeMemberRole {
  CREATOR # Informational only - same permissions as MEMBER
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
  activityStatus: ActivityStatus!
  completedAt: DateTime!
  durationMinutes: Int
  participants: [User!]!
  notes: String
  recordedBy: User!
  decisionSession: DecisionSession
  createdAt: DateTime!
  updatedAt: DateTime!
}

enum ActivityType {
  VISITED
  WATCHED
  COMPLETED
}

enum ActivityStatus {
  CONFIRMED
  TENTATIVE
  CANCELLED
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
  eliminationOrder: [User!]!
  currentTurnIndex: Int!
  currentRound: Int!
  turnStartedAt: DateTime
  turnTimeoutMinutes: Int!
  skippedUsers: [SkippedTurn!]!
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
  CATCH_UP_PHASE
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

type SkippedTurn {
  user: User!
  round: Int!
  turnInRound: Int!
}

type EliminationStatus {
  sessionId: ID!
  currentCandidates: [ListItem!]!
  currentUserTurn: User
  isYourTurn: Boolean!
  currentRound: Int!
  turnTimeRemaining: Int! # seconds
  eliminationOrder: [User!]!
  skippedUsers: [SkippedTurn!]!
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
  updateTentativeActivity(id: ID!, input: UpdateActivityInput!): ActivityEntry!
  cancelTentativeActivity(id: ID!): ActivityEntry!
  confirmTentativeActivity(id: ID!, input: ConfirmActivityInput!): ActivityEntry!
  deleteActivity(id: ID!): Boolean!
  logDecisionResult(sessionId: ID!, scheduledFor: DateTime): ActivityEntry!
  
  # Decision Making
  createDecisionSession(input: CreateDecisionSessionInput!): DecisionSession!
  addListsToSession(sessionId: ID!, listIds: [ID!]!): DecisionSession!
  applyFilters(sessionId: ID!, filters: FilterCriteriaInput!): DecisionSession!
  startElimination(sessionId: ID!): DecisionSession!
  eliminateItem(sessionId: ID!, itemId: ID!): DecisionSession!
  skipTurn(sessionId: ID!): DecisionSession!
  rejoinElimination(sessionId: ID!): DecisionSession!
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
  eliminationStatus(sessionId: ID!): EliminationStatus!
  activityEntry(id: ID!): ActivityEntry
  
  # Search and filtering
  searchLists(query: String!, type: ListType): [List!]!
  searchListItems(query: String!, filters: FilterCriteriaInput): [ListItem!]!
  suggestKMValues(availableCount: Int!, tribeSize: Int!): [AlgorithmParams!]!
  
  # Activity queries
  listItemActivities(listItemId: ID!, tribeId: ID): [ActivityEntry!]!
  userActivities(userId: ID!, tribeId: ID): [ActivityEntry!]!
  tentativeActivities(tribeId: ID!): [ActivityEntry!]!
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

## Authentication & Authorization

### OAuth Strategy
- **Primary Provider**: Google OAuth
- **Development Mode**: Dev/Test login allowing email input
- **Extensible Design**: Interface to support multiple providers
- **Key Requirements**: Extract verified email address

### Session Management (Simple JWT for MVP)
- **Access Tokens**: JWT with 7-day expiry (no refresh tokens initially)
- **Storage**: Frontend localStorage with fallback to sessionStorage
- **Concurrent Sessions**: Supported across multiple devices
- **Token Claims**: User ID, email, expiry timestamp
- **Migration Path**: Adding refresh tokens later requires minimal changes

```go
// JWT Token Structure
type JWTClaims struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    Provider  string `json:"provider"`
    ExpiresAt int64  `json:"exp"`
    IssuedAt  int64  `json:"iat"`
    jwt.StandardClaims
}

// JWT Configuration
type JWTConfig struct {
    SecretKey    string        // From environment
    ExpiryTime   time.Duration // 7 days for MVP
    Issuer       string        // "tribe-app"
}
```

### Authorization Model
- **List Ownership**: Users own personal lists, any tribe member can modify tribe lists
- **Sharing**: Read-only sharing only
- **Revocation**: List owners can revoke shares at any time
- **Tribe Departure**: Complex rules for maintaining/losing access to shared content

## Decision-Making Algorithm (Enhanced)

### Advanced Filter Engine with Priority System

```go
// Filter Configuration
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

// Filter Results with Violation Tracking
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

// Enhanced Filter Engine
type FilterEngine struct {
    db repository.Database
}

func (fe *FilterEngine) ApplyFiltersWithPriority(ctx context.Context, items []ListItem, config FilterConfiguration) ([]FilterResult, error) {
    results := make([]FilterResult, 0, len(items))
    
    // Sort filters by priority (lowest number = highest priority)
    sort.Slice(config.Items, func(i, j int) bool {
        return config.Items[i].Priority < config.Items[j].Priority
    })
    
    for _, item := range items {
        result := FilterResult{
            Item:              item,
            PassedHardFilters: true,
            SoftFilterResults: make([]SoftFilterResult, 0),
            ViolationCount:    0,
        }
        
        // Apply filters in priority order
        for _, filter := range config.Items {
            passed := fe.evaluateFilter(ctx, item, filter)
            
            if filter.IsHard {
                if !passed {
                    result.PassedHardFilters = false
                    break // Skip this item entirely
                }
            } else {
                // Soft filter - track result but don't exclude
                softResult := SoftFilterResult{
                    FilterID:    filter.ID,
                    FilterType:  filter.Type,
                    Passed:      passed,
                    Priority:    filter.Priority,
                    Description: filter.Description,
                }
                result.SoftFilterResults = append(result.SoftFilterResults, softResult)
                
                if !passed {
                    result.ViolationCount++
                }
            }
        }
        
        // Only include items that passed all hard filters
        if result.PassedHardFilters {
            result.PriorityScore = fe.calculatePriorityScore(result.SoftFilterResults)
            results = append(results, result)
        }
    }
    
    // Sort results: higher priority score = better match
    sort.Slice(results, func(i, j int) bool {
        if results[i].PriorityScore != results[j].PriorityScore {
            return results[i].PriorityScore > results[j].PriorityScore
        }
        // Tie-breaker: fewer violations is better
        return results[i].ViolationCount < results[j].ViolationCount
    })
    
    return results, nil
}

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

func (fe *FilterEngine) evaluateFilter(ctx context.Context, item ListItem, filter FilterItem) bool {
    switch filter.Type {
    case "category":
        criteria := filter.Criteria.(CategoryFilterCriteria)
        return fe.evaluateCategoryFilter(item, criteria)
    case "dietary":
        criteria := filter.Criteria.(DietaryFilterCriteria)
        return fe.evaluateDietaryFilter(item, criteria)
    case "location":
        criteria := filter.Criteria.(LocationFilterCriteria)
        return fe.evaluateLocationFilter(item, criteria)
    case "recent_activity":
        criteria := filter.Criteria.(RecentActivityFilterCriteria)
        return fe.evaluateRecentActivityFilter(ctx, item, criteria)
    case "opening_hours":
        criteria := filter.Criteria.(OpeningHoursFilterCriteria)
        return fe.evaluateOpeningHoursFilter(item, criteria)
    case "tags":
        criteria := filter.Criteria.(TagFilterCriteria)
        return fe.evaluateTagFilter(item, criteria)
    default:
        return true // Unknown filter types pass by default
    }
}

// Specific filter criteria types
type CategoryFilterCriteria struct {
    IncludeCategories []string `json:"include_categories"`
    ExcludeCategories []string `json:"exclude_categories"`
}

type DietaryFilterCriteria struct {
    RequiredOptions []string `json:"required_options"` // ["vegetarian", "vegan", "gluten_free"]
}

type LocationFilterCriteria struct {
    CenterLat    float64 `json:"center_lat"`
    CenterLng    float64 `json:"center_lng"`
    MaxDistance  float64 `json:"max_distance"` // in miles
}

type RecentActivityFilterCriteria struct {
    ExcludeDays int      `json:"exclude_days"`
    UserID      string   `json:"user_id"`
    TribeID     *string  `json:"tribe_id"`
}

type OpeningHoursFilterCriteria struct {
    MustBeOpenFor int `json:"must_be_open_for"` // minutes from now
}

type TagFilterCriteria struct {
    RequiredTags []string `json:"required_tags"`
    ExcludedTags []string `json:"excluded_tags"`
}
```

### Filter Configuration Storage

```sql
-- Add to schema for storing user filter configurations
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

### Frontend Filter Management

```typescript
// Filter Management Interface
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

// Filter Builder Component Props
interface FilterBuilderProps {
  configuration: FilterConfiguration;
  onConfigurationChange: (config: FilterConfiguration) => void;
  availableFilters: FilterType[];
}

// Filter Result Display
interface FilterResultDisplayProps {
  results: FilterResult[];
  showViolations: boolean;
  onItemSelect: (item: ListItem) => void;
}

// Enhanced GraphQL types for filtering
type FilterResult {
  item: ListItem!
  passedHardFilters: Boolean!
  softFilterResults: [SoftFilterResult!]!
  violationCount: Int!
  priorityScore: Float!
}

type SoftFilterResult {
  filterId: String!
  filterType: String!
  passed: Boolean!
  priority: Int!
  description: String!
}
```

### Default Parameters
```go
type DecisionDefaults struct {
    K int `json:"k"` // Default: 2 eliminations per person
    M int `json:"m"` // Default: 3 final options for random selection
    N int `json:"n"` // Always equals tribe size (1-8)
}

type TribeDecisionPreferences struct {
    DefaultK int `json:"default_k"` // Configurable per tribe
    DefaultM int `json:"default_m"` // Configurable per tribe
    MaxK     int `json:"max_k"`     // Maximum eliminations allowed
    MaxM     int `json:"max_m"`     // Maximum final options allowed
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

### Turn-Based Elimination Algorithm

```go
// Turn-based elimination system
type EliminationSession struct {
    SessionID           string    `json:"session_id"`
    TribeID             string    `json:"tribe_id"`
    EliminationOrder    []string  `json:"elimination_order"`    // Randomized user IDs
    CurrentTurnIndex    int       `json:"current_turn_index"`   // Index in elimination_order
    CurrentRound        int       `json:"current_round"`        // Which round (1 to K)
    TurnStartedAt       time.Time `json:"turn_started_at"`
    TurnTimeoutMinutes  int       `json:"turn_timeout_minutes"`
    SkippedUsers        []SkippedTurn `json:"skipped_users"`
    CurrentCandidates   []string  `json:"current_candidates"`   // Remaining item IDs
}

type SkippedTurn struct {
    UserID string `json:"user_id"`
    Round  int    `json:"round"`
    TurnInRound int `json:"turn_in_round"`
}

type EliminationAlgorithm struct {
    db repository.Database
}

func (ea *EliminationAlgorithm) StartEliminationPhase(ctx context.Context, sessionID string) (*EliminationSession, error) {
    session, err := ea.getDecisionSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Get tribe members
    members, err := ea.getTribeMembers(ctx, session.TribeID)
    if err != nil {
        return nil, err
    }
    
    // Randomize elimination order
    rand.Shuffle(len(members), func(i, j int) {
        members[i], members[j] = members[j], members[i]
    })
    
    eliminationOrder := make([]string, len(members))
    for i, member := range members {
        eliminationOrder[i] = member.UserID
    }
    
    elimSession := &EliminationSession{
        SessionID:           sessionID,
        TribeID:            session.TribeID,
        EliminationOrder:   eliminationOrder,
        CurrentTurnIndex:   0,
        CurrentRound:       1,
        TurnStartedAt:      time.Now(),
        TurnTimeoutMinutes: 5,
        SkippedUsers:       []SkippedTurn{},
        CurrentCandidates:  session.InitialCandidates,
    }
    
    // Update session status and save state
    session.Status = "eliminating"
    if err := ea.saveEliminationState(ctx, elimSession); err != nil {
        return nil, err
    }
    
    return elimSession, nil
}

func (ea *EliminationAlgorithm) ProcessElimination(ctx context.Context, sessionID, userID, eliminatedItemID string) (*EliminationSession, error) {
    session, err := ea.getEliminationSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Verify it's this user's turn
    currentUserID := session.EliminationOrder[session.CurrentTurnIndex]
    if currentUserID != userID {
        return nil, errors.New("not your turn")
    }
    
    // Verify item is still available
    if !contains(session.CurrentCandidates, eliminatedItemID) {
        return nil, errors.New("item not available for elimination")
    }
    
    // Remove item from candidates
    session.CurrentCandidates = removeItem(session.CurrentCandidates, eliminatedItemID)
    
    // Record elimination
    if err := ea.recordElimination(ctx, sessionID, userID, eliminatedItemID, session.CurrentRound); err != nil {
        return nil, err
    }
    
    // Advance to next turn
    return ea.advanceToNextTurn(ctx, session)
}

func (ea *EliminationAlgorithm) advanceToNextTurn(ctx context.Context, session *EliminationSession) (*EliminationSession, error) {
    params, err := ea.getAlgorithmParams(ctx, session.SessionID)
    if err != nil {
        return nil, err
    }
    
    // Move to next player in order
    session.CurrentTurnIndex = (session.CurrentTurnIndex + 1) % len(session.EliminationOrder)
    
    // If we've completed a full cycle, move to next round
    if session.CurrentTurnIndex == 0 {
        session.CurrentRound++
    }
    
    session.TurnStartedAt = time.Now()
    
    // Check if elimination phase is complete
    if session.CurrentRound > params.K {
        return ea.finalizeCatchUpPhase(ctx, session)
    }
    
    // Save updated state
    if err := ea.saveEliminationState(ctx, session); err != nil {
        return nil, err
    }
    
    return session, nil
}

func (ea *EliminationAlgorithm) HandleTimeout(ctx context.Context, sessionID string) (*EliminationSession, error) {
    session, err := ea.getEliminationSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check if current turn has timed out
    timeout := time.Duration(session.TurnTimeoutMinutes) * time.Minute
    if time.Since(session.TurnStartedAt) < timeout {
        return session, nil // Not timed out yet
    }
    
    // Record skipped turn
    currentUserID := session.EliminationOrder[session.CurrentTurnIndex]
    skippedTurn := SkippedTurn{
        UserID:      currentUserID,
        Round:       session.CurrentRound,
        TurnInRound: session.CurrentTurnIndex,
    }
    session.SkippedUsers = append(session.SkippedUsers, skippedTurn)
    
    // Advance to next turn
    return ea.advanceToNextTurn(ctx, session)
}

func (ea *EliminationAlgorithm) finalizeCatchUpPhase(ctx context.Context, session *EliminationSession) (*EliminationSession, error) {
    // Allow skipped users to make their eliminations
    if len(session.SkippedUsers) > 0 {
        // Process catch-up eliminations in order
        return ea.startCatchUpPhase(ctx, session)
    }
    
    // Complete the session
    return ea.completeEliminationPhase(ctx, session)
}

func (ea *EliminationAlgorithm) completeEliminationPhase(ctx context.Context, session *EliminationSession) (*EliminationSession, error) {
    params, err := ea.getAlgorithmParams(ctx, session.SessionID)
    if err != nil {
        return nil, err
    }
    
    // Determine final selection
    var finalSelection *string
    remainingCount := len(session.CurrentCandidates)
    
    if remainingCount == 1 {
        finalSelection = &session.CurrentCandidates[0]
    } else if remainingCount > 1 {
        // Adjust M if needed due to skipped eliminations
        effectiveM := params.M
        if remainingCount > effectiveM {
            effectiveM = remainingCount
        }
        
        // Random selection from remaining candidates
        if effectiveM == 1 {
            idx := rand.Intn(len(session.CurrentCandidates))
            finalSelection = &session.CurrentCandidates[idx]
        }
        // If M > 1, present multiple options for final selection
    }
    
    // Update decision session
    return ea.finalizeDecisionSession(ctx, session.SessionID, finalSelection)
}

// Polling endpoint for frontend
func (ea *EliminationAlgorithm) GetEliminationStatus(ctx context.Context, sessionID, userID string) (*EliminationStatus, error) {
    session, err := ea.getEliminationSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check for timeout
    session, _ = ea.HandleTimeout(ctx, sessionID)
    
    currentUserID := ""
    if len(session.EliminationOrder) > 0 {
        currentUserID = session.EliminationOrder[session.CurrentTurnIndex]
    }
    
    return &EliminationStatus{
        SessionID:         sessionID,
        CurrentCandidates: session.CurrentCandidates,
        CurrentUserTurn:   currentUserID,
        IsYourTurn:        currentUserID == userID,
        CurrentRound:      session.CurrentRound,
        TurnTimeRemaining: ea.calculateTimeRemaining(session),
        EliminationOrder:  session.EliminationOrder,
        SkippedUsers:      session.SkippedUsers,
    }, nil
}

type EliminationStatus struct {
    SessionID         string        `json:"session_id"`
    CurrentCandidates []string      `json:"current_candidates"`
    CurrentUserTurn   string        `json:"current_user_turn"`
    IsYourTurn        bool          `json:"is_your_turn"`
    CurrentRound      int           `json:"current_round"`
    TurnTimeRemaining time.Duration `json:"turn_time_remaining"`
    EliminationOrder  []string      `json:"elimination_order"`
    SkippedUsers      []SkippedTurn `json:"skipped_users"`
}
```

### Activity Tracking System

```go
// Enhanced activity tracking with tentative entries
type ActivityEntry struct {
    ID                string    `json:"id"`
    ListItemID        string    `json:"list_item_id"`
    UserID            string    `json:"user_id"`           // Who the activity is for
    TribeID           *string   `json:"tribe_id"`
    ActivityType      string    `json:"activity_type"`     // 'visited', 'watched', 'completed'
    ActivityStatus    string    `json:"activity_status"`   // 'confirmed', 'tentative', 'cancelled'
    CompletedAt       time.Time `json:"completed_at"`      // When it happened/will happen
    DurationMinutes   *int      `json:"duration_minutes"`
    Participants      []string  `json:"participants"`      // User IDs who participated
    Notes             string    `json:"notes"`
    RecordedByUserID  string    `json:"recorded_by_user_id"` // Who logged this
    DecisionSessionID *string   `json:"decision_session_id"` // If from decision result
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}

type ActivityService struct {
    db repository.Database
}

func (as *ActivityService) LogActivity(ctx context.Context, req LogActivityRequest) (*ActivityEntry, error) {
    entry := &ActivityEntry{
        ID:               generateUUID(),
        ListItemID:       req.ListItemID,
        UserID:           req.UserID,
        TribeID:          req.TribeID,
        ActivityType:     req.ActivityType,
        ActivityStatus:   req.ActivityStatus,
        CompletedAt:      req.CompletedAt,
        DurationMinutes:  req.DurationMinutes,
        Participants:     req.Participants,
        Notes:            req.Notes,
        RecordedByUserID: req.RecordedByUserID,
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }
    
    // Auto-determine status based on completion time
    if entry.ActivityStatus == "" {
        if entry.CompletedAt.After(time.Now()) {
            entry.ActivityStatus = "tentative"
        } else {
            entry.ActivityStatus = "confirmed"
        }
    }
    
    // Validate tribe membership
    if req.TribeID != nil {
        if err := as.validateTribeMembership(ctx, req.RecordedByUserID, *req.TribeID); err != nil {
            return nil, err
        }
    }
    
    if err := as.db.CreateActivityEntry(ctx, entry); err != nil {
        return nil, err
    }
    
    return entry, nil
}

func (as *ActivityService) UpdateTentativeActivity(ctx context.Context, entryID, userID string, req UpdateActivityRequest) (*ActivityEntry, error) {
    entry, err := as.db.GetActivityEntry(ctx, entryID)
    if err != nil {
        return nil, err
    }
    
    // Only allow updates to tentative entries
    if entry.ActivityStatus != "tentative" {
        return nil, errors.New("can only update tentative activities")
    }
    
    // Verify user is in the tribe
    if entry.TribeID != nil {
        if err := as.validateTribeMembership(ctx, userID, *entry.TribeID); err != nil {
            return nil, err
        }
    }
    
    // Update fields
    if req.ActivityStatus != nil {
        entry.ActivityStatus = *req.ActivityStatus
    }
    if req.CompletedAt != nil {
        entry.CompletedAt = *req.CompletedAt
    }
    if req.Participants != nil {
        entry.Participants = req.Participants
    }
    if req.Notes != nil {
        entry.Notes = *req.Notes
    }
    
    entry.UpdatedAt = time.Now()
    
    if err := as.db.UpdateActivityEntry(ctx, entry); err != nil {
        return nil, err
    }
    
    return entry, nil
}

func (as *ActivityService) LogDecisionResult(ctx context.Context, sessionID, userID string, scheduledFor *time.Time) (*ActivityEntry, error) {
    session, err := as.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    if session.FinalSelection == nil {
        return nil, errors.New("no final selection available")
    }
    
    // Get tribe members as default participants
    members, err := as.db.GetTribeMembers(ctx, session.TribeID)
    if err != nil {
        return nil, err
    }
    
    participants := make([]string, len(members))
    for i, member := range members {
        participants[i] = member.UserID
    }
    
    completedAt := time.Now()
    status := "confirmed"
    
    if scheduledFor != nil {
        completedAt = *scheduledFor
        if completedAt.After(time.Now()) {
            status = "tentative"
        }
    }
    
    req := LogActivityRequest{
        ListItemID:        *session.FinalSelection,
        UserID:            userID,
        TribeID:           &session.TribeID,
        ActivityType:      "visited", // Default, can be changed
        ActivityStatus:    status,
        CompletedAt:       completedAt,
        Participants:      participants,
        RecordedByUserID:  userID,
        DecisionSessionID: &sessionID,
    }
    
    return as.LogActivity(ctx, req)
}

type LogActivityRequest struct {
    ListItemID        string     `json:"list_item_id"`
    UserID            string     `json:"user_id"`
    TribeID           *string    `json:"tribe_id"`
    ActivityType      string     `json:"activity_type"`
    ActivityStatus    string     `json:"activity_status"`
    CompletedAt       time.Time  `json:"completed_at"`
    DurationMinutes   *int       `json:"duration_minutes"`
    Participants      []string   `json:"participants"`
    Notes             string     `json:"notes"`
    RecordedByUserID  string     `json:"recorded_by_user_id"`
    DecisionSessionID *string    `json:"decision_session_id"`
}

type UpdateActivityRequest struct {
    ActivityStatus *string    `json:"activity_status"`
    CompletedAt    *time.Time `json:"completed_at"`
    Participants   []string   `json:"participants"`
    Notes          *string    `json:"notes"`
}
```

### Common Ownership Tribe Management

```go
// Tribe service implementing common ownership model
type TribeService struct {
    db repository.Database
}

// Common ownership validation - any member can perform tribe operations
func (ts *TribeService) ValidateTribeMembership(ctx context.Context, userID, tribeID string) error {
    isMember, err := ts.db.IsUserTribeMember(ctx, userID, tribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    return nil
}

// Get senior member (longest-standing) for tie-breaking scenarios
func (ts *TribeService) GetSeniorMember(ctx context.Context, tribeID string) (*User, error) {
    return ts.db.GetTribeSeniorMember(ctx, tribeID)
}

// Any member can invite new members
func (ts *TribeService) InviteMember(ctx context.Context, tribeID, inviterID, inviteeEmail string) error {
    // Validate inviter is a member
    if err := ts.ValidateTribeMembership(ctx, inviterID, tribeID); err != nil {
        return err
    }
    
    // Check tribe capacity
    tribe, err := ts.db.GetTribe(ctx, tribeID)
    if err != nil {
        return err
    }
    
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    if memberCount >= tribe.MaxMembers {
        return errors.New("tribe is at maximum capacity")
    }
    
    // Create invitation
    invitation := &TribeInvitation{
        TribeID:      tribeID,
        InviterID:    inviterID,
        InviteeEmail: inviteeEmail,
        Status:       "pending",
        CreatedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
    }
    
    return ts.db.CreateTribeInvitation(ctx, invitation)
}

// Any member can remove other members (with safeguards)
func (ts *TribeService) RemoveMember(ctx context.Context, tribeID, removerID, targetUserID string) error {
    // Validate remover is a member
    if err := ts.ValidateTribeMembership(ctx, removerID, tribeID); err != nil {
        return err
    }
    
    // Validate target is a member
    if err := ts.ValidateTribeMembership(ctx, targetUserID, tribeID); err != nil {
        return err
    }
    
    // Prevent self-removal (use LeaveTribe instead)
    if removerID == targetUserID {
        return errors.New("use leave tribe function to remove yourself")
    }
    
    // Get current member count
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    // Prevent removing last member
    if memberCount <= 1 {
        return errors.New("cannot remove the last member of a tribe")
    }
    
    // **SAFETY CHECK**: Prevent mass removal in short timeframe
    recentRemovals, err := ts.db.GetRecentRemovals(ctx, tribeID, time.Hour)
    if err != nil {
        return err
    }
    
    // If more than half the tribe has been removed in the last hour, require senior member approval
    if len(recentRemovals) >= memberCount/2 {
        seniorMember, err := ts.GetSeniorMember(ctx, tribeID)
        if err != nil {
            return err
        }
        
        if removerID != seniorMember.ID {
            return errors.New("mass removal detected - senior member approval required")
        }
    }
    
    // Remove member
    if err := ts.db.RemoveTribeMember(ctx, tribeID, targetUserID); err != nil {
        return err
    }
    
    // Log the removal for safety tracking
    return ts.db.LogTribeMemberRemoval(ctx, tribeID, removerID, targetUserID)
}

// Members can leave tribes themselves
func (ts *TribeService) LeaveTribe(ctx context.Context, tribeID, userID string) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    // Check if this is the last member
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    if memberCount == 1 {
        // Last member leaving - offer to delete tribe or transfer to someone else
        return ts.handleLastMemberDeparture(ctx, tribeID, userID)
    }
    
    // Remove user from tribe
    return ts.db.RemoveTribeMember(ctx, tribeID, userID)
}

// Handle last member leaving - clean up or transfer
func (ts *TribeService) handleLastMemberDeparture(ctx context.Context, tribeID, userID string) error {
    // For now, just delete the tribe when last member leaves
    // Could be enhanced to offer transfer to recent members or archive
    return ts.db.DeleteTribe(ctx, tribeID)
}

// Any member can update tribe settings
func (ts *TribeService) UpdateTribeSettings(ctx context.Context, tribeID, userID string, updates TribeUpdateRequest) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    return ts.db.UpdateTribe(ctx, tribeID, updates)
}

// Conflict resolution using senior member
func (ts *TribeService) ResolveConflict(ctx context.Context, tribeID string, conflictType string, options []interface{}) (interface{}, error) {
    seniorMember, err := ts.GetSeniorMember(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    // Log the conflict and resolution method
    conflict := &TribeConflict{
        TribeID:      tribeID,
        ConflictType: conflictType,
        ResolvedBy:   seniorMember.ID,
        Resolution:   "senior_member_decision",
        CreatedAt:    time.Now(),
    }
    
    if err := ts.db.LogTribeConflict(ctx, conflict); err != nil {
        return nil, err
    }
    
    // For now, senior member gets to decide
    // Could be enhanced with voting mechanisms, etc.
    return options[0], nil // Default to first option
}

// Tribe deletion with consensus mechanism
func (ts *TribeService) DeleteTribe(ctx context.Context, tribeID, requestingUserID string, consensusRequired bool) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, requestingUserID, tribeID); err != nil {
        return err
    }
    
    if consensusRequired {
        // Check if all members have agreed to deletion
        members, err := ts.db.GetTribeMembers(ctx, tribeID)
        if err != nil {
            return err
        }
        
        approvals, err := ts.db.GetDeletionApprovals(ctx, tribeID)
        if err != nil {
            return err
        }
        
        if len(approvals) < len(members) {
            // Not all members have approved - record this user's approval
            approval := &TribeDeletionApproval{
                TribeID:   tribeID,
                UserID:    requestingUserID,
                CreatedAt: time.Now(),
            }
            return ts.db.RecordDeletionApproval(ctx, approval)
        }
    }
    
    // All approvals received or consensus not required - delete tribe
    return ts.db.DeleteTribe(ctx, tribeID)
}

type TribeUpdateRequest struct {
    Name                 *string                `json:"name"`
    Description          *string                `json:"description"`
    MaxMembers          *int                   `json:"max_members"`
    DecisionPreferences *DecisionPreferences   `json:"decision_preferences"`
}

type TribeConflict struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    ConflictType string    `json:"conflict_type"`
    ResolvedBy   string    `json:"resolved_by"`
    Resolution   string    `json:"resolution"`
    CreatedAt    time.Time `json:"created_at"`
}

type TribeDeletionApproval struct {
    TribeID   string    `json:"tribe_id"`
    UserID    string    `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}

type TribeInvitation struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    InviterID    string    `json:"inviter_id"`
    InviteeEmail string    `json:"invitee_email"`
    Status       string    `json:"status"` // pending, accepted, declined, expired
    CreatedAt    time.Time `json:"created_at"`
    ExpiresAt    time.Time `json:"expires_at"`
}
```

#### Tribe Invitations Table (Enhanced Two-Stage System)
```sql
CREATE TABLE tribe_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id),
    invitee_email VARCHAR(255) NOT NULL,
    suggested_tribe_display_name VARCHAR(255), -- Inviter can suggest display name
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'accepted_pending_ratification', 'ratified', 'rejected', 'revoked', 'expired'
    invited_at TIMESTAMPTZ DEFAULT NOW(),
    accepted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '7 days',
    UNIQUE(tribe_id, invitee_email)
);
```

#### Tribe Invitation Ratifications Table
```sql
CREATE TABLE tribe_invitation_ratifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invitation_id UUID NOT NULL REFERENCES tribe_invitations(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES users(id),
    vote VARCHAR(50) NOT NULL, -- 'approve', 'reject'
    voted_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(invitation_id, member_id)
);
```

#### Member Removal Petitions Table
```sql
CREATE TABLE member_removal_petitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    petitioner_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID NOT NULL REFERENCES users(id),
    reason TEXT,
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'approved', 'rejected'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    UNIQUE(tribe_id, target_user_id) -- Only one active petition per user
);
```

#### Member Removal Votes Table
```sql
CREATE TABLE member_removal_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    petition_id UUID NOT NULL REFERENCES member_removal_petitions(id) ON DELETE CASCADE,
    voter_id UUID NOT NULL REFERENCES users(id),
    vote VARCHAR(50) NOT NULL, -- 'approve', 'reject'
    voted_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(petition_id, voter_id)
);
```

#### Tribe Deletion Petitions Table
```sql
CREATE TABLE tribe_deletion_petitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    petitioner_id UUID NOT NULL REFERENCES users(id),
    reason TEXT,
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'approved', 'rejected'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    UNIQUE(tribe_id) -- Only one active deletion petition per tribe
);
```

#### Tribe Deletion Votes Table
```sql
CREATE TABLE tribe_deletion_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    petition_id UUID NOT NULL REFERENCES tribe_deletion_petitions(id) ON DELETE CASCADE,
    voter_id UUID NOT NULL REFERENCES users(id),
    vote VARCHAR(50) NOT NULL, -- 'approve', 'reject'
    voted_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(petition_id, voter_id)
);
```

#### List Deletion Petitions Table
```sql
CREATE TABLE list_deletion_petitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    petitioner_id UUID NOT NULL REFERENCES users(id),
    reason TEXT,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'confirmed', 'cancelled'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id UUID REFERENCES users(id),
    UNIQUE(list_id) -- Only one active petition per list
);
```

#### Tribe Settings Table (for configurable inactivity thresholds)
```sql
CREATE TABLE tribe_settings (
    tribe_id UUID PRIMARY KEY REFERENCES tribes(id) ON DELETE CASCADE,
    inactivity_threshold_days INTEGER DEFAULT 30, -- 1 to 730 (2 years)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Enhanced Tribe Memberships Table
```sql
-- Add columns to existing tribe_memberships table
ALTER TABLE tribe_memberships 
ADD COLUMN last_login_at TIMESTAMPTZ,
ADD COLUMN is_inactive BOOLEAN DEFAULT FALSE;

-- Update existing constraint
ALTER TABLE tribe_memberships 
DROP CONSTRAINT IF EXISTS tribe_memberships_role_check;

ALTER TABLE tribe_memberships 
ADD CONSTRAINT tribe_memberships_role_check 
CHECK (role IN ('creator', 'member', 'pending'));
```

### Democratic Tribe Governance System

```go
// Enhanced tribe service with democratic governance
type TribeGovernanceService struct {
    db repository.Database
}

// Two-stage invitation system
func (tgs *TribeGovernanceService) InviteToTribe(ctx context.Context, tribeID, inviterID, inviteeEmail string) (*TribeInvitation, error) {
    // Validate inviter is a member
    if err := tgs.validateTribeMembership(ctx, inviterID, tribeID); err != nil {
        return nil, err
    }
    
    // Check tribe capacity
    settings, err := tgs.getTribeSettings(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    memberCount, err := tgs.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    if memberCount >= settings.MaxMembers {
        return nil, errors.New("tribe is at maximum capacity")
    }
    
    // Create invitation (stage 1)
    invitation := &TribeInvitation{
        ID:           generateUUID(),
        TribeID:      tribeID,
        InviterID:    inviterID,
        InviteeEmail: inviteeEmail,
        Status:       "pending",
        InvitedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
    }
    
    return invitation, tgs.db.CreateTribeInvitation(ctx, invitation)
}

// Invitee accepts invitation (moves to ratification stage)
func (tgs *TribeGovernanceService) AcceptInvitation(ctx context.Context, invitationID, userID string) (*TribeInvitation, error) {
    invitation, err := tgs.db.GetTribeInvitation(ctx, invitationID)
    if err != nil {
        return nil, err
    }
    
    if invitation.Status != "pending" {
        return nil, errors.New("invitation is not in pending state")
    }
    
    if time.Now().After(invitation.ExpiresAt) {
        invitation.Status = "expired"
        tgs.db.UpdateTribeInvitation(ctx, invitation)
        return nil, errors.New("invitation has expired")
    }
    
    // Move to ratification stage
    invitation.Status = "accepted_pending_ratification"
    invitation.AcceptedAt = &time.Now()
    
    if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
        return nil, err
    }
    
    // Start ratification process - notify all existing members
    return invitation, tgs.startRatificationProcess(ctx, invitation)
}

// Existing member votes on ratification
func (tgs *TribeGovernanceService) VoteOnInvitation(ctx context.Context, invitationID, voterID string, approve bool) error {
    invitation, err := tgs.db.GetTribeInvitation(ctx, invitationID)
    if err != nil {
        return err
    }
    
    if invitation.Status != "accepted_pending_ratification" {
        return errors.New("invitation is not pending ratification")
    }
    
    // Validate voter is a member
    if err := tgs.validateTribeMembership(ctx, voterID, invitation.TribeID); err != nil {
        return err
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    ratification := &TribeInvitationRatification{
        InvitationID: invitationID,
        MemberID:     voterID,
        Vote:         vote,
        VotedAt:      time.Now(),
    }
    
    if err := tgs.db.CreateInvitationRatification(ctx, ratification); err != nil {
        return err
    }
    
    // If any member rejects, immediately revoke invitation
    if !approve {
        invitation.Status = "rejected"
        return tgs.db.UpdateTribeInvitation(ctx, invitation)
    }
    
    // Check if all members have approved
    return tgs.checkRatificationComplete(ctx, invitation)
}

func (tgs *TribeGovernanceService) checkRatificationComplete(ctx context.Context, invitation *TribeInvitation) error {
    // Get all current members
    members, err := tgs.db.GetTribeMembers(ctx, invitation.TribeID)
    if err != nil {
        return err
    }
    
    // Get all ratification votes
    votes, err := tgs.db.GetInvitationRatifications(ctx, invitation.ID)
    if err != nil {
        return err
    }
    
    // Check if all members have approved
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // All members approved - complete ratification
        invitation.Status = "ratified"
        if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
            return err
        }
        
        // Add user to tribe
        membership := &TribeMembership{
            TribeID:  invitation.TribeID,
            UserID:   invitation.InviteeUserID, // Set when they accept
            Role:     "member",
            JoinedAt: time.Now(),
        }
        
        return tgs.db.CreateTribeMembership(ctx, membership)
    }
    
    return nil // Still waiting for more votes
}

// Member removal petition system
func (tgs *TribeGovernanceService) PetitionMemberRemoval(ctx context.Context, tribeID, petitionerID, targetUserID, reason string) (*MemberRemovalPetition, error) {
    // Validate petitioner is a member
    if err := tgs.validateTribeMembership(ctx, petitionerID, tribeID); err != nil {
        return nil, err
    }
    
    // Validate target is a member
    if err := tgs.validateTribeMembership(ctx, targetUserID, tribeID); err != nil {
        return nil, err
    }
    
    // Cannot petition to remove yourself
    if petitionerID == targetUserID {
        return nil, errors.New("cannot petition to remove yourself - use leave tribe instead")
    }
    
    // Check if petition already exists
    existing, err := tgs.db.GetActiveMemberRemovalPetition(ctx, tribeID, targetUserID)
    if err == nil && existing != nil {
        return nil, errors.New("active petition already exists for this member")
    }
    
    petition := &MemberRemovalPetition{
        ID:           generateUUID(),
        TribeID:      tribeID,
        PetitionerID: petitionerID,
        TargetUserID: targetUserID,
        Reason:       reason,
        Status:       "active",
        CreatedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateMemberRemovalPetition(ctx, petition); err != nil {
        return nil, err
    }
    
    return petition, nil
}

// Vote on member removal
func (tgs *TribeGovernanceService) VoteOnMemberRemoval(ctx context.Context, petitionID, voterID string, approve bool) error {
    petition, err := tgs.db.GetMemberRemovalPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "active" {
        return errors.New("petition is not active")
    }
    
    // Validate voter is a member (but not the target)
    if err := tgs.validateTribeMembership(ctx, voterID, petition.TribeID); err != nil {
        return err
    }
    
    if voterID == petition.TargetUserID {
        return errors.New("target user cannot vote on their own removal")
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    removalVote := &MemberRemovalVote{
        PetitionID: petitionID,
        VoterID:    voterID,
        Vote:       vote,
        VotedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateMemberRemovalVote(ctx, removalVote); err != nil {
        return err
    }
    
    // If any member rejects, petition fails
    if !approve {
        petition.Status = "rejected"
        petition.ResolvedAt = &time.Now()
        return tgs.db.UpdateMemberRemovalPetition(ctx, petition)
    }
    
    // Check if all eligible members have approved
    return tgs.checkMemberRemovalComplete(ctx, petition)
}

func (tgs *TribeGovernanceService) checkMemberRemovalComplete(ctx context.Context, petition *MemberRemovalPetition) error {
    // Get all members except the target
    members, err := tgs.db.GetTribeMembersExcept(ctx, petition.TribeID, petition.TargetUserID)
    if err != nil {
        return err
    }
    
    // Get all votes
    votes, err := tgs.db.GetMemberRemovalVotes(ctx, petition.ID)
    if err != nil {
        return err
    }
    
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // Unanimous approval - remove member
        petition.Status = "approved"
        petition.ResolvedAt = &time.Now()
        
        if err := tgs.db.UpdateMemberRemovalPetition(ctx, petition); err != nil {
            return err
        }
        
        // Remove the member
        return tgs.db.RemoveTribeMember(ctx, petition.TribeID, petition.TargetUserID)
    }
    
    return nil // Still waiting for more votes
}

// Tribe deletion with 100% consensus
func (tgs *TribeGovernanceService) PetitionTribeDeletion(ctx context.Context, tribeID, petitionerID, reason string) (*TribeDeletionPetition, error) {
    // Validate petitioner is a member
    if err := tgs.validateTribeMembership(ctx, petitionerID, tribeID); err != nil {
        return nil, err
    }
    
    // Check if petition already exists
    existing, err := tgs.db.GetActiveTribeDeletionPetition(ctx, tribeID)
    if err == nil && existing != nil {
        return nil, errors.New("active deletion petition already exists")
    }
    
    petition := &TribeDeletionPetition{
        ID:           generateUUID(),
        TribeID:      tribeID,
        PetitionerID: petitionerID,
        Reason:       reason,
        Status:       "active",
        CreatedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateTribeDeletionPetition(ctx, petition); err != nil {
        return nil, err
    }
    
    return petition, nil
}

// Vote on tribe deletion
func (tgs *TribeGovernanceService) VoteOnTribeDeletion(ctx context.Context, petitionID, voterID string, approve bool) error {
    petition, err := tgs.db.GetTribeDeletionPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "active" {
        return errors.New("petition is not active")
    }
    
    // Validate voter is a member
    if err := tgs.validateTribeMembership(ctx, voterID, petition.TribeID); err != nil {
        return err
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    deletionVote := &TribeDeletionVote{
        PetitionID: petitionID,
        VoterID:    voterID,
        Vote:       vote,
        VotedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateTribeDeletionVote(ctx, deletionVote); err != nil {
        return err
    }
    
    // If any member rejects, petition fails
    if !approve {
        petition.Status = "rejected"
        petition.ResolvedAt = &time.Now()
        return tgs.db.UpdateTribeDeletionPetition(ctx, petition)
    }
    
    // Check if all members have approved
    return tgs.checkTribeDeletionComplete(ctx, petition)
}

func (tgs *TribeGovernanceService) checkTribeDeletionComplete(ctx context.Context, petition *TribeDeletionPetition) error {
    // Get all members
    members, err := tgs.db.GetTribeMembers(ctx, petition.TribeID)
    if err != nil {
        return err
    }
    
    // Get all votes
    votes, err := tgs.db.GetTribeDeletionVotes(ctx, petition.ID)
    if err != nil {
        return err
    }
    
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // 100% consensus achieved - delete tribe
        petition.Status = "approved"
        petition.ResolvedAt = &time.Now()
        
        if err := tgs.db.UpdateTribeDeletionPetition(ctx, petition); err != nil {
            return err
        }
        
        // Delete the tribe
        return tgs.db.DeleteTribe(ctx, petition.TribeID)
    }
    
    return nil // Still waiting for more votes
}

// List deletion petition system (simpler - only needs one confirmation)
func (tgs *TribeGovernanceService) PetitionListDeletion(ctx context.Context, listID, petitionerID, reason string) (*ListDeletionPetition, error) {
    list, err := tgs.db.GetList(ctx, listID)
    if err != nil {
        return nil, err
    }
    
    // Only applies to tribe lists
    if list.OwnerType != "tribe" {
        return nil, errors.New("list deletion petitions only apply to tribe lists")
    }
    
    // Validate petitioner is a tribe member
    if err := tgs.validateTribeMembership(ctx, petitionerID, list.OwnerID); err != nil {
        return nil, err
    }
    
    petition := &ListDeletionPetition{
        ID:           generateUUID(),
        ListID:       listID,
        PetitionerID: petitionerID,
        Reason:       reason,
        Status:       "pending",
        CreatedAt:    time.Now(),
    }
    
    return petition, tgs.db.CreateListDeletionPetition(ctx, petition)
}

// Confirm or cancel list deletion
func (tgs *TribeGovernanceService) ResolveListDeletion(ctx context.Context, petitionID, resolverID string, confirm bool) error {
    petition, err := tgs.db.GetListDeletionPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "pending" {
        return errors.New("petition is not pending")
    }
    
    list, err := tgs.db.GetList(ctx, petition.ListID)
    if err != nil {
        return err
    }
    
    // Validate resolver is a tribe member
    if err := tgs.validateTribeMembership(ctx, resolverID, list.OwnerID); err != nil {
        return err
    }
    
    status := "cancelled"
    if confirm {
        status = "confirmed"
    }
    
    petition.Status = status
    petition.ResolvedAt = &time.Now()
    petition.ResolvedByUserID = &resolverID
    
    if err := tgs.db.UpdateListDeletionPetition(ctx, petition); err != nil {
        return err
    }
    
    // If confirmed, delete the list
    if confirm {
        return tgs.db.DeleteList(ctx, petition.ListID)
    }
    
    return nil
}

// Helper function to validate tribe membership
func (tgs *TribeGovernanceService) validateTribeMembership(ctx context.Context, userID, tribeID string) error {
    isMember, err := tgs.db.IsUserTribeMember(ctx, userID, tribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    return nil
}

// Get tribe settings including inactivity threshold
func (tgs *TribeGovernanceService) getTribeSettings(ctx context.Context, tribeID string) (*TribeSettings, error) {
    settings, err := tgs.db.GetTribeSettings(ctx, tribeID)
    if err != nil {
        // Create default settings if none exist
        defaultSettings := &TribeSettings{
            TribeID:                tribeID,
            InactivityThresholdDays: 30,
            CreatedAt:              time.Now(),
            UpdatedAt:              time.Now(),
        }
        if err := tgs.db.CreateTribeSettings(ctx, defaultSettings); err != nil {
            return nil, err
        }
        return defaultSettings, nil
    }
    return settings, nil
}

// Data structures for the new governance system
type TribeInvitation struct {
    ID             string     `json:"id"`
    TribeID        string     `json:"tribe_id"`
    InviterID      string     `json:"inviter_id"`
    InviteeEmail   string     `json:"invitee_email"`
    InviteeUserID  *string    `json:"invitee_user_id"` // Set when they accept
    Status         string     `json:"status"`
    InvitedAt      time.Time  `json:"invited_at"`
    AcceptedAt     *time.Time `json:"accepted_at"`
    ExpiresAt      time.Time  `json:"expires_at"`
}

type TribeInvitationRatification struct {
    InvitationID string    `json:"invitation_id"`
    MemberID     string    `json:"member_id"`
    Vote         string    `json:"vote"`
    VotedAt      time.Time `json:"voted_at"`
}

type MemberRemovalPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    TargetUserID string     `json:"target_user_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type MemberRemovalVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type TribeDeletionPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type TribeDeletionVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type ListDeletionPetition struct {
    ID                 string     `json:"id"`
    ListID             string     `json:"list_id"`
    PetitionerID       string     `json:"petitioner_id"`
    Reason             string     `json:"reason"`
    Status             string     `json:"status"`
    CreatedAt          time.Time  `json:"created_at"`
    ResolvedAt         *time.Time `json:"resolved_at"`
    ResolvedByUserID   *string    `json:"resolved_by_user_id"`
}

type TribeSettings struct {
    TribeID                string    `json:"tribe_id"`
    InactivityThresholdDays int       `json:"inactivity_threshold_days"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}
```

### Decision Session Management with Timeouts

```go
// Enhanced decision session service with timeout management
type DecisionSessionService struct {
    db repository.Database
}

// Update session activity (called on any user interaction)
func (dss *DecisionSessionService) UpdateSessionActivity(ctx context.Context, sessionID string) error {
    return dss.db.UpdateSessionActivity(ctx, sessionID, time.Now())
}

// Check for expired sessions (run periodically)
func (dss *DecisionSessionService) ProcessExpiredSessions(ctx context.Context) error {
    cutoffTime := time.Now().Add(-30 * time.Minute)
    expiredSessions, err := dss.db.GetInactiveSessionsSince(ctx, cutoffTime)
    if err != nil {
        return err
    }
    
    for _, session := range expiredSessions {
        if err := dss.expireSession(ctx, session.ID); err != nil {
            log.Printf("Failed to expire session %s: %v", session.ID, err)
        }
    }
    
    return nil
}

// Expire a session due to inactivity
func (dss *DecisionSessionService) expireSession(ctx context.Context, sessionID string) error {
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    // Only expire active sessions
    if session.Status != "configuring" && session.Status != "eliminating" {
        return nil
    }
    
    session.Status = "expired"
    session.CompletedAt = &time.Now()
    
    return dss.db.UpdateDecisionSession(ctx, session)
}

// Complete session and set up history retention
func (dss *DecisionSessionService) CompleteSession(ctx context.Context, sessionID string, finalResult *DecisionResult) error {
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    session.Status = "completed"
    session.CompletedAt = &time.Now()
    session.ExpiresAt = &time.Now().Add(30 * 24 * time.Hour) // 1 month retention
    session.FinalSelectionID = finalResult.WinnerID
    session.RunnersUp = finalResult.RunnersUpIDs
    session.EliminationHistory = finalResult.EliminationHistory
    
    return dss.db.UpdateDecisionSession(ctx, session)
}

// Pin session to prevent automatic cleanup
func (dss *DecisionSessionService) PinSession(ctx context.Context, sessionID, userID string) error {
    // Verify user has access to this session
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    // Validate user is tribe member
    isMember, err := dss.db.IsUserTribeMember(ctx, userID, session.TribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    
    session.IsPinned = true
    return dss.db.UpdateDecisionSession(ctx, session)
}

// Get formatted decision history for display
func (dss *DecisionSessionService) GetDecisionHistory(ctx context.Context, sessionID, userID string) (*DecisionHistory, error) {
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Verify user has access
    isMember, err := dss.db.IsUserTribeMember(ctx, userID, session.TribeID)
    if err != nil {
        return nil, err
    }
    if !isMember {
        return nil, errors.New("user is not a member of this tribe")
    }
    
    // Get tribe settings for elimination visibility
    tribe, err := dss.db.GetTribe(ctx, session.TribeID)
    if err != nil {
        return nil, err
    }
    
    history := &DecisionHistory{
        SessionID:     sessionID,
        SessionName:   session.Name,
        Status:        session.Status,
        CreatedAt:     session.CreatedAt,
        CompletedAt:   session.CompletedAt,
        IsPinned:      session.IsPinned,
        ShowDetails:   tribe.ShowEliminationDetails,
    }
    
    // Get list items for display
    if session.FinalSelectionID != nil {
        winner, err := dss.db.GetListItem(ctx, *session.FinalSelectionID)
        if err == nil {
            history.Winner = winner
        }
    }
    
    // Get runners-up
    if len(session.RunnersUp) > 0 {
        runnersUp, err := dss.db.GetListItems(ctx, session.RunnersUp)
        if err == nil {
            history.RunnersUp = runnersUp
        }
    }
    
    // Parse elimination history (stored as JSON)
    var eliminations []EliminationHistoryEntry
    if err := json.Unmarshal(session.EliminationHistory, &eliminations); err == nil {
        // Reverse order for display (most recent elimination first)
        for i := len(eliminations) - 1; i >= 0; i-- {
            elimination := eliminations[i]
            
            // Get list item details
            item, err := dss.db.GetListItem(ctx, elimination.ItemID)
            if err != nil {
                continue
            }
            
            historyEntry := DecisionHistoryEntry{
                Item:          item,
                EliminatedAt:  elimination.EliminatedAt,
                Round:         elimination.Round,
                EliminationOrder: len(eliminations) - i, // 1-based order
            }
            
            // Include eliminator info if visibility enabled
            if tribe.ShowEliminationDetails {
                eliminator, err := dss.db.GetUser(ctx, elimination.EliminatorID)
                if err == nil {
                    // Get tribe-specific display name
                    membership, err := dss.db.GetTribeMembership(ctx, session.TribeID, elimination.EliminatorID)
                    if err == nil && membership.TribeDisplayName != "" {
                        historyEntry.EliminatorName = membership.TribeDisplayName
                    } else {
                        historyEntry.EliminatorName = eliminator.DisplayName
                    }
                }
            }
            
            history.Eliminations = append(history.Eliminations, historyEntry)
        }
    }
    
    return history, nil
}

// Clean up expired session histories (run daily)
func (dss *DecisionSessionService) CleanupExpiredHistories(ctx context.Context) error {
    expiredSessions, err := dss.db.GetExpiredUnpinnedSessions(ctx)
    if err != nil {
        return err
    }
    
    for _, session := range expiredSessions {
        if err := dss.db.DeleteDecisionSession(ctx, session.ID); err != nil {
            log.Printf("Failed to delete expired session %s: %v", session.ID, err)
        }
    }
    
    return nil
}

// Mobile-optimized session status for polling
func (dss *DecisionSessionService) GetMobileSessionStatus(ctx context.Context, sessionID, userID string) (*MobileSessionStatus, error) {
    // Update activity when checking status
    if err := dss.UpdateSessionActivity(ctx, sessionID); err != nil {
        return nil, err
    }
    
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check if session has expired
    if session.Status == "configuring" || session.Status == "eliminating" {
        inactiveTime := time.Since(session.LastActivityAt)
        if inactiveTime > time.Duration(session.SessionTimeoutMinutes)*time.Minute {
            // Expire the session
            if err := dss.expireSession(ctx, sessionID); err != nil {
                return nil, err
            }
            session.Status = "expired"
        }
    }
    
    status := &MobileSessionStatus{
        SessionID:         sessionID,
        Status:            session.Status,
        CurrentCandidates: []ListItemSummary{}, // Simplified for mobile
        IsYourTurn:        false,
        TurnTimeRemaining: 0,
        SessionTimeRemaining: 0,
    }
    
    // Calculate remaining time
    if session.Status == "eliminating" {
        sessionTimeout := time.Duration(session.SessionTimeoutMinutes) * time.Minute
        elapsed := time.Since(session.LastActivityAt)
        if elapsed < sessionTimeout {
            status.SessionTimeRemaining = int((sessionTimeout - elapsed).Seconds())
        }
        
        // Check if it's user's turn
        if len(session.EliminationOrder) > 0 {
            currentUserID := session.EliminationOrder[session.CurrentTurnIndex]
            status.IsYourTurn = (currentUserID == userID)
            
            if status.IsYourTurn && session.TurnStartedAt != nil {
                turnTimeout := time.Duration(session.TurnTimeoutMinutes) * time.Minute
                turnElapsed := time.Since(*session.TurnStartedAt)
                if turnElapsed < turnTimeout {
                    status.TurnTimeRemaining = int((turnTimeout - turnElapsed).Seconds())
                }
            }
        }
    }
    
    // Get simplified candidate list for mobile
    if len(session.CurrentCandidates) > 0 {
        items, err := dss.db.GetListItems(ctx, session.CurrentCandidates)
        if err == nil {
            for _, item := range items {
                status.CurrentCandidates = append(status.CurrentCandidates, ListItemSummary{
                    ID:          item.ID,
                    Name:        item.Name,
                    Category:    item.Category,
                    Tags:        item.Tags[:min(3, len(item.Tags))], // Limit tags for mobile
                })
            }
        }
    }
    
    return status, nil
}

// Data structures for session management
type DecisionResult struct {
    WinnerID           *string                    `json:"winner_id"`
    RunnersUpIDs       []string                   `json:"runners_up_ids"`
    EliminationHistory json.RawMessage            `json:"elimination_history"`
}

type EliminationHistoryEntry struct {
    ItemID        string    `json:"item_id"`
    EliminatorID  string    `json:"eliminator_id"`
    Round         int       `json:"round"`
    EliminatedAt  time.Time `json:"eliminated_at"`
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

type DecisionHistoryEntry struct {
    Item             *ListItem `json:"item"`
    EliminatorName   string    `json:"eliminator_name"`
    EliminatedAt     time.Time `json:"eliminated_at"`
    Round            int       `json:"round"`
    EliminationOrder int       `json:"elimination_order"` // 1 = last eliminated, 2 = second-to-last, etc.
}

type MobileSessionStatus struct {
    SessionID            string             `json:"session_id"`
    Status               string             `json:"status"`
    CurrentCandidates    []ListItemSummary  `json:"current_candidates"`
    IsYourTurn           bool               `json:"is_your_turn"`
    TurnTimeRemaining    int                `json:"turn_time_remaining"`    // seconds
    SessionTimeRemaining int                `json:"session_time_remaining"` // seconds
}

type ListItemSummary struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Category string   `json:"category"`
    Tags     []string `json:"tags"`
}