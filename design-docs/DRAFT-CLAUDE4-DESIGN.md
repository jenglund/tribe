# Tribe App - Comprehensive Design Document

**Version:** Draft 1.0  
**Date:** January 2025  
**Author:** Claude-4 (based on requirements from project owner)

## Table of Contents

1. [Project Overview](#project-overview)
2. [User Stories](#user-stories)
3. [System Architecture](#system-architecture)
4. [Data Model](#data-model)
5. [API Design](#api-design)
6. [Frontend Architecture](#frontend-architecture)
7. [Authentication & Authorization](#authentication--authorization)
8. [Decision-Making Algorithm](#decision-making-algorithm)
9. [Testing Strategy](#testing-strategy)
10. [Implementation Roadmap](#implementation-roadmap)
11. [Known Issues & Future Work](#known-issues--future-work)

## Project Overview

Tribe is a collaborative decision-making web application designed to help small groups (1-8 people) make choices about activities, restaurants, entertainment, and other shared experiences. The core functionality revolves around creating and managing lists of options, applying filters and criteria, and using structured decision-making processes (like the KN+M selection system) to arrive at mutually agreeable choices.

### Core Values
- **Simplicity**: Easy to use for non-technical users
- **Collaboration**: Designed for group decision-making
- **Flexibility**: Adaptable to various types of decisions
- **Privacy**: Self-hostable, open-source solution
- **Small Scale**: Optimized for personal/friend group usage (not enterprise scale)

### Technology Stack
- **Backend**: Go (Gin web framework)
- **Frontend**: React with TypeScript
- **Database**: PostgreSQL
- **Authentication**: OAuth (Google) + OTP via email
- **Deployment**: Docker-based for easy self-hosting

## User Stories

### Authentication & User Management
- **US-001**: As a new user, I want to sign up using my Google account so I can quickly join without creating another password
- **US-002**: As a user without Google, I want to sign up using email OTP so I can access the app
- **US-003**: As a user, I want to update my profile information (name, preferences) so others can identify me
- **US-004**: As a user, I want to delete my account and all associated data so I can maintain privacy

### Tribe Management
- **US-005**: As a user, I want to create a new tribe so I can collaborate with my partner/friends
- **US-006**: As a user, I want to invite others to my tribe using their email addresses
- **US-007**: As a user, I want to join a tribe when invited so I can participate in group decisions
- **US-008**: As a tribe member, I want to see all other members so I know who's part of the group
- **US-009**: As a tribe creator, I want to remove members from my tribe when needed
- **US-010**: As a tribe member, I want to leave a tribe I no longer want to participate in
- **US-011**: As a user, I want to be part of multiple tribes for different social contexts

### List Management
- **US-012**: As a user, I want to create personal lists so I can track my own interests
- **US-013**: As a tribe member, I want to create tribe lists so we can collaboratively build options
- **US-014**: As a user, I want to add items to my personal lists with details (name, location, type, notes)
- **US-015**: As a tribe member, I want to add items to tribe lists
- **US-016**: As a list owner, I want to edit or remove items from my lists
- **US-017**: As a user, I want to share my personal lists with specific tribes
- **US-018**: As a user, I want to see all lists I have access to (personal, shared, tribe)
- **US-019**: As a user, I want to mark items as "visited/done" with timestamps and companions
- **US-020**: As a user, I want to add tags/categories to list items for better organization

### Decision Making
- **US-021**: As a tribe member, I want to start a decision-making session using available lists
- **US-022**: As a user, I want to apply filters (cuisine type, location, recency, dietary restrictions) to narrow options
- **US-023**: As a tribe member, I want to participate in KN+M selection process (e.g., 5-3-1)
- **US-024**: As a user, I want the system to automatically adjust K and M based on available options and tribe size
- **US-025**: As a tribe member, I want to see the final decision and mark it as completed
- **US-026**: As a user, I want to see the history of decisions made by my tribes

### Advanced Features
- **US-027**: As a user, I want to set dietary preferences that automatically filter options
- **US-028**: As a user, I want to avoid places I've visited recently (configurable timeframe)
- **US-029**: As a user, I want to avoid places other tribe members have visited recently
- **US-030**: As a user, I want to set location preferences and maximum distance filters
- **US-031**: As a user, I want to check if places are currently open and will remain open
- **US-032**: As a user, I want to import lists from external sources (future: Google Maps)

## System Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React App     │    │   Go Backend    │    │   PostgreSQL    │
│   (Frontend)    │◄──►│   (API Server)  │◄──►│   (Database)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                        │
        │                        │
        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐
│   Static Files  │    │ External APIs   │
│   (CDN/Nginx)   │    │ (OAuth, Maps)   │
└─────────────────┘    └─────────────────┘
```

### Backend Architecture (Go)

```
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP request handlers
│   │   ├── middleware/            # Authentication, logging, CORS
│   │   └── routes/                # Route definitions
│   ├── auth/
│   │   ├── oauth.go              # OAuth providers (Google)
│   │   └── otp.go                # Email OTP system
│   ├── models/                   # Data models and validation
│   ├── services/                 # Business logic
│   ├── repository/               # Database access layer
│   └── utils/                    # Shared utilities
├── migrations/                   # Database migrations
├── tests/                       # Test files
└── docker/                      # Docker configuration
```

### Frontend Architecture (React + TypeScript)

```
├── src/
│   ├── components/
│   │   ├── auth/                # Authentication components
│   │   ├── lists/               # List management components
│   │   ├── tribes/              # Tribe management components
│   │   ├── decisions/           # Decision-making flow
│   │   └── common/              # Reusable UI components
│   ├── hooks/                   # Custom React hooks
│   ├── services/                # API client functions
│   ├── store/                   # State management (Context/Redux)
│   ├── types/                   # TypeScript type definitions
│   ├── utils/                   # Helper functions
│   └── pages/                   # Page components
```

## Data Model

### Database Schema

#### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    oauth_provider VARCHAR(50), -- 'google', 'email'
    oauth_id VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    dietary_preferences JSONB, -- ['vegetarian', 'vegan', 'gluten-free', etc.]
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

#### Tribe Memberships Table
```sql
CREATE TABLE tribe_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- 'creator', 'admin', 'member'
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tribe_id, user_id)
);
```

#### Lists Table
```sql
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 'personal', 'tribe'
    owner_id UUID REFERENCES users(id), -- NULL for tribe lists
    tribe_id UUID REFERENCES tribes(id), -- NULL for personal lists
    category VARCHAR(100), -- 'restaurants', 'movies', 'activities', etc.
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_owner CHECK (
        (type = 'personal' AND owner_id IS NOT NULL AND tribe_id IS NULL) OR
        (type = 'tribe' AND owner_id IS NULL AND tribe_id IS NOT NULL)
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
    category VARCHAR(100),
    location JSONB, -- {address, lat, lng, city, state}
    metadata JSONB, -- Flexible field for cuisine_type, price_range, etc.
    external_id VARCHAR(255), -- For syncing with external APIs
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

#### List Shares Table
```sql
CREATE TABLE list_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    shared_with_tribe UUID REFERENCES tribes(id) ON DELETE CASCADE,
    shared_with_user UUID REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(50) DEFAULT 'read', -- 'read', 'write'
    shared_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_share_target CHECK (
        (shared_with_tribe IS NOT NULL AND shared_with_user IS NULL) OR
        (shared_with_tribe IS NULL AND shared_with_user IS NOT NULL)
    )
);
```

#### Activity History Table
```sql
CREATE TABLE activity_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    list_item_id UUID NOT NULL REFERENCES list_items(id),
    tribe_id UUID REFERENCES tribes(id), -- NULL if done individually
    completed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    companions JSONB, -- Array of user IDs who participated
    notes TEXT,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Decision Sessions Table
```sql
CREATE TABLE decision_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'completed', 'cancelled'
    filters JSONB, -- Applied filters (cuisine, location, etc.)
    k_value INTEGER, -- Number of eliminations per person
    m_value INTEGER, -- Final number of options
    final_selection UUID REFERENCES list_items(id),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
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
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    list_item_id UUID NOT NULL REFERENCES list_items(id) ON DELETE CASCADE,
    round INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(session_id, user_id, list_item_id)
);
```

### Database Indexes

```sql
-- Performance indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_oauth_provider_id ON users(oauth_provider, oauth_id);
CREATE INDEX idx_tribe_memberships_user_id ON tribe_memberships(user_id);
CREATE INDEX idx_tribe_memberships_tribe_id ON tribe_memberships(tribe_id);
CREATE INDEX idx_lists_owner_id ON lists(owner_id);
CREATE INDEX idx_lists_tribe_id ON lists(tribe_id);
CREATE INDEX idx_lists_type ON lists(type);
CREATE INDEX idx_list_items_list_id ON list_items(list_id);
CREATE INDEX idx_list_shares_list_id ON list_shares(list_id);
CREATE INDEX idx_activity_history_user_id ON activity_history(user_id);
CREATE INDEX idx_activity_history_completed_at ON activity_history(completed_at);
CREATE INDEX idx_decision_sessions_tribe_id ON decision_sessions(tribe_id);

-- Soft delete support
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_tribes_deleted_at ON tribes(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_lists_deleted_at ON lists(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_list_items_deleted_at ON list_items(deleted_at) WHERE deleted_at IS NULL;
```

## API Design

### Authentication Endpoints
```
POST /api/v1/auth/google/login     # OAuth login with Google
POST /api/v1/auth/email/request    # Request OTP via email
POST /api/v1/auth/email/verify     # Verify OTP and login
POST /api/v1/auth/refresh          # Refresh JWT token
POST /api/v1/auth/logout           # Logout and invalidate token
```

### User Endpoints
```
GET    /api/v1/users/me            # Get current user profile
PUT    /api/v1/users/me            # Update current user profile
DELETE /api/v1/users/me            # Delete current user account
```

### Tribe Endpoints
```
GET    /api/v1/tribes              # List user's tribes
POST   /api/v1/tribes              # Create new tribe
GET    /api/v1/tribes/{id}         # Get tribe details
PUT    /api/v1/tribes/{id}         # Update tribe (creator only)
DELETE /api/v1/tribes/{id}         # Delete tribe (creator only)
POST   /api/v1/tribes/{id}/invite  # Invite user to tribe
GET    /api/v1/tribes/{id}/members # List tribe members
DELETE /api/v1/tribes/{id}/members/{user_id} # Remove member
POST   /api/v1/tribes/{id}/leave   # Leave tribe
```

### List Endpoints
```
GET    /api/v1/lists               # List accessible lists
POST   /api/v1/lists               # Create new list
GET    /api/v1/lists/{id}          # Get list details
PUT    /api/v1/lists/{id}          # Update list
DELETE /api/v1/lists/{id}          # Delete list
POST   /api/v1/lists/{id}/share    # Share list with tribe/user
DELETE /api/v1/lists/{id}/share/{share_id} # Remove share
```

### List Item Endpoints
```
GET    /api/v1/lists/{id}/items    # Get list items
POST   /api/v1/lists/{id}/items    # Add item to list
PUT    /api/v1/lists/{id}/items/{item_id} # Update list item
DELETE /api/v1/lists/{id}/items/{item_id} # Delete list item
POST   /api/v1/lists/{id}/items/{item_id}/visited # Mark as visited
```

### Decision Endpoints
```
POST   /api/v1/decisions           # Start decision session
GET    /api/v1/decisions/{id}      # Get decision session
POST   /api/v1/decisions/{id}/eliminate # Eliminate options
POST   /api/v1/decisions/{id}/complete  # Complete decision
GET    /api/v1/decisions/{id}/history   # Get decision history
```

## Frontend Architecture

### Component Structure

#### Authentication Flow
```typescript
// components/auth/LoginPage.tsx
// components/auth/GoogleLoginButton.tsx
// components/auth/EmailLoginForm.tsx
// components/auth/OTPVerificationForm.tsx
```

#### Main Application
```typescript
// components/layout/AppLayout.tsx
// components/layout/Navigation.tsx
// components/layout/Sidebar.tsx
```

#### Tribe Management
```typescript
// components/tribes/TribeList.tsx
// components/tribes/TribeCard.tsx
// components/tribes/CreateTribeForm.tsx
// components/tribes/InviteMemberForm.tsx
// components/tribes/MemberList.tsx
```

#### List Management
```typescript
// components/lists/ListGrid.tsx
// components/lists/ListCard.tsx
// components/lists/CreateListForm.tsx
// components/lists/ListItemForm.tsx
// components/lists/ShareListModal.tsx
```

#### Decision Making
```typescript
// components/decisions/DecisionWizard.tsx
// components/decisions/FilterForm.tsx
// components/decisions/OptionCard.tsx
// components/decisions/EliminationRound.tsx
// components/decisions/ResultDisplay.tsx
```

### State Management

Using React Context + useReducer for global state:

```typescript
// store/AppContext.tsx
interface AppState {
  user: User | null;
  tribes: Tribe[];
  lists: List[];
  currentDecision: DecisionSession | null;
}

// store/actions.ts
type AppAction = 
  | { type: 'SET_USER'; payload: User }
  | { type: 'ADD_TRIBE'; payload: Tribe }
  | { type: 'UPDATE_LIST'; payload: List }
  | { type: 'START_DECISION'; payload: DecisionSession }
  // ... other actions
```

### API Client

```typescript
// services/api.ts
class ApiClient {
  private baseURL: string;
  private token: string | null;
  
  async auth(): Promise<AuthService> { /* ... */ }
  async users(): Promise<UserService> { /* ... */ }
  async tribes(): Promise<TribeService> { /* ... */ }
  async lists(): Promise<ListService> { /* ... */ }
  async decisions(): Promise<DecisionService> { /* ... */ }
}
```

## Authentication & Authorization

### JWT Token Structure
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "name": "User Name",
  "exp": 1640995200,
  "iat": 1640908800,
  "iss": "tribe-app"
}
```

### Authorization Levels
1. **Public**: Unauthenticated access (login pages only)
2. **User**: Authenticated user access to own resources
3. **Tribe Member**: Access to tribe resources for members only
4. **Tribe Creator**: Full control over tribe (invite, remove members)
5. **List Owner**: Full control over personal lists
6. **List Collaborator**: Read/write access to shared lists

### Middleware Chain
```go
router.Use(
    middleware.CORS(),
    middleware.Logger(),
    middleware.RateLimit(),
    middleware.Authentication(), // Extract user from JWT
    middleware.Authorization(), // Check permissions for specific routes
)
```

## Decision-Making Algorithm

### KN+M Selection Process

```go
type DecisionParams struct {
    K int // Eliminations per person
    M int // Final options count
    N int // Number of people
}

func CalculateOptimalKM(availableOptions int, tribeSize int, preferences UserPreferences) DecisionParams {
    // Algorithm to determine optimal K and M values
    // Constraints:
    // - K * N < availableOptions (must have options left)
    // - M >= 1 (must have at least one final option)
    // - K >= 0 (cannot have negative eliminations)
    // - K <= preferences.MaxEliminationsPerPerson
    // - M <= preferences.MaxFinalOptions
}
```

### Filtering System

```go
type FilterCriteria struct {
    Categories []string          // e.g., ["italian", "mexican"]
    ExcludeCategories []string   // e.g., ["fast-food"]
    MaxDistance float64          // in miles
    Location *Location           // center point for distance calculation
    DietaryRequirements []string // e.g., ["vegetarian", "vegan"]
    ExcludeRecentlyVisited bool  // exclude if visited in last N days
    RecentlyVisitedDays int      // default 60
    MustBeOpenFor int            // minutes from now
    PriceRange *PriceRange       // min/max price range
}

func ApplyFilters(items []ListItem, criteria FilterCriteria, user User) []ListItem {
    // Apply each filter in sequence
    // Return filtered list
}
```

## Testing Strategy

### Backend Testing (Go)

#### Unit Tests
- **Models**: Test validation, serialization, business logic
- **Services**: Test business logic, filtering algorithms
- **Repositories**: Test database operations (with test DB)
- **Handlers**: Test HTTP request/response handling

#### Integration Tests
- **API Endpoints**: Test complete request flows
- **Database**: Test migrations, constraints, transactions
- **Authentication**: Test OAuth and OTP flows

#### Test Structure
```go
// tests/unit/services/decision_test.go
func TestDecisionService_CalculateKM(t *testing.T) {
    testCases := []struct {
        name string
        availableOptions int
        tribeSize int
        expected DecisionParams
    }{
        {"Couple with 5 options", 5, 2, DecisionParams{K: 2, M: 1, N: 2}},
        {"Single user", 10, 1, DecisionParams{K: 0, M: 1, N: 1}},
        // ... more test cases
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Frontend Testing (React + TypeScript)

#### Unit Tests (Jest + React Testing Library)
- **Components**: Test rendering, user interactions, prop handling
- **Hooks**: Test custom hooks with different scenarios
- **Services**: Test API client functions (mocked)
- **Utils**: Test helper functions and utilities

#### Integration Tests
- **User Flows**: Test complete user journeys
- **API Integration**: Test with mock backend
- **State Management**: Test context and reducer functions

#### E2E Tests (Playwright)
- **Authentication Flow**: Complete login/logout process
- **Tribe Management**: Create, invite, join tribes
- **List Management**: CRUD operations on lists
- **Decision Making**: Complete decision-making flow

### Testing Infrastructure

#### Test Database
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  test-db:
    image: postgres:15
    environment:
      POSTGRES_DB: tribe_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5433:5432"
```

#### CI/CD Pipeline
```yaml
# .github/workflows/test.yml
name: Test Suite
on: [push, pull_request]
jobs:
  backend-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - run: go test -v -coverage ./...
      
  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npm test -- --coverage
      - run: npm run test:e2e
```

### Coverage Goals
- **Backend**: Minimum 85% line coverage, 90% for critical paths
- **Frontend**: Minimum 80% line coverage, 85% for business logic
- **Integration**: Cover all happy paths and major error scenarios

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-3)
**Goal**: Basic user authentication and tribe management

#### Week 1: Project Setup
- [ ] Initialize Go backend with Gin framework
- [ ] Set up PostgreSQL database with Docker
- [ ] Create database migrations for users and tribes
- [ ] Implement basic JWT authentication
- [ ] Set up React frontend with TypeScript
- [ ] Configure testing infrastructure

#### Week 2: User Management
- [ ] Implement Google OAuth integration
- [ ] Create user registration/login API endpoints
- [ ] Build authentication middleware
- [ ] Create user profile management
- [ ] Implement email OTP system (basic)
- [ ] Build login/signup frontend components

#### Week 3: Tribe Management
- [ ] Implement tribe CRUD operations
- [ ] Create tribe membership system
- [ ] Build invite/join functionality
- [ ] Create tribe management UI components
- [ ] Add member management features
- [ ] Implement basic authorization middleware

### Phase 2: List Management (Weeks 4-6)
**Goal**: Complete list creation, sharing, and item management

#### Week 4: List Foundation
- [ ] Create list data models and migrations
- [ ] Implement list CRUD API endpoints
- [ ] Build list creation and editing UI
- [ ] Add list item management
- [ ] Implement basic list viewing

#### Week 5: List Sharing
- [ ] Create list sharing system
- [ ] Implement share permissions (read/write)
- [ ] Build sharing UI components
- [ ] Add list discovery features
- [ ] Implement list categories

#### Week 6: Enhanced List Features
- [ ] Add list item metadata (location, tags)
- [ ] Implement search and filtering
- [ ] Create activity history tracking
- [ ] Build list import/export features
- [ ] Add bulk operations for list items

### Phase 3: Decision Making (Weeks 7-10)
**Goal**: Core decision-making functionality with KN+M system

#### Week 7: Decision Framework
- [ ] Design decision session data model
- [ ] Implement basic decision API endpoints
- [ ] Create decision session management
- [ ] Build filter system for list items
- [ ] Implement KM calculation algorithm

#### Week 8: Decision UI
- [ ] Create decision wizard component
- [ ] Build filter selection interface
- [ ] Implement option display and selection
- [ ] Create elimination round interface
- [ ] Add result display and confirmation

#### Week 9: Advanced Decision Features
- [ ] Implement location-based filtering
- [ ] Add recently visited filtering
- [ ] Create dietary preference filtering
- [ ] Build decision history tracking
- [ ] Add decision analytics

#### Week 10: Decision Polish
- [ ] Optimize decision algorithms
- [ ] Add real-time collaboration features
- [ ] Implement decision notifications
- [ ] Create decision sharing features
- [ ] Add mobile responsiveness

### Phase 4: Polish & Deployment (Weeks 11-12)
**Goal**: Production-ready application with deployment

#### Week 11: Polish
- [ ] Comprehensive testing and bug fixes
- [ ] Performance optimization
- [ ] UI/UX improvements
- [ ] Add loading states and error handling
- [ ] Implement proper error logging

#### Week 12: Deployment
- [ ] Create Docker deployment configuration
- [ ] Set up production database
- [ ] Implement proper logging and monitoring
- [ ] Create deployment documentation
- [ ] Set up backup and recovery procedures

### Future Phases (Post-MVP)

#### Phase 5: Enhanced Features
- [ ] Google Maps integration for location data
- [ ] Real-time collaboration with WebSockets
- [ ] Mobile app development
- [ ] Advanced analytics and insights
- [ ] External list synchronization

#### Phase 6: Scalability
- [ ] Performance optimization for larger datasets
- [ ] Caching layer implementation
- [ ] Database optimization and indexing
- [ ] API rate limiting and throttling
- [ ] Advanced monitoring and alerting

## Known Issues & Future Work

### Known Technical Considerations

1. **Email OTP System**: Currently planned as basic implementation. May need rate limiting and better security measures.

2. **Real-time Collaboration**: Decision sessions may benefit from WebSocket implementation for real-time updates during elimination rounds.

3. **Location Services**: Google Maps integration is desired but deprioritized due to API costs. Consider using OpenStreetMap or other free alternatives.

4. **Mobile Responsiveness**: Design should be mobile-first, but native mobile apps are out of scope for MVP.

5. **Performance**: With small user base, performance isn't critical, but filtering algorithms should be optimized for larger lists.

### Future Work

#### High Priority
- [ ] Advanced filtering options (price range, ratings, hours)
- [ ] Better decision history and analytics
- [ ] Email notifications for tribe invites and decisions
- [ ] List templates and categories
- [ ] Export/import functionality

#### Medium Priority
- [ ] Integration with external restaurant/activity APIs
- [ ] Social features (comments, recommendations)
- [ ] Advanced user preferences and learning
- [ ] Mobile application development
- [ ] Advanced admin panel for self-hosted instances

#### Low Priority
- [ ] Multi-language support
- [ ] Advanced analytics and reporting
- [ ] Integration with calendar applications
- [ ] AI-powered recommendations
- [ ] Advanced geographic features

### Development Guidelines for AI Collaboration

#### Code Quality Standards
1. **Test-Driven Development**: Write tests before implementation
2. **Clear Naming**: Use descriptive variable and function names
3. **Documentation**: Comment complex business logic thoroughly
4. **Error Handling**: Implement comprehensive error handling and logging
5. **Type Safety**: Leverage TypeScript and Go's type system fully

#### AI Agent Instructions
1. **Incremental Changes**: Make small, verifiable changes
2. **Test Coverage**: Ensure new code has appropriate test coverage
3. **Consistent Style**: Follow established patterns in the codebase
4. **Documentation Updates**: Update relevant documentation with changes
5. **Database Migrations**: Always create reversible database migrations

#### Review Checklist
- [ ] All tests pass
- [ ] Code coverage meets minimum requirements
- [ ] API endpoints have proper authentication/authorization
- [ ] Database changes include proper migrations
- [ ] Frontend components are properly typed
- [ ] Error handling is comprehensive
- [ ] Documentation is updated

---

**Next Steps**: This design document should be reviewed and refined based on feedback. Priority should be given to Phase 1 implementation, starting with project setup and basic authentication. 