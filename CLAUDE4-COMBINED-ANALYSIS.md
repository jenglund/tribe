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

For more information on the system's data model, see `DATA-MODEL.md` under `design-docs/v2`.

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

For additional information, you can refer to the following documents under `design-docs/v2` for additonal details on design and implementation:
- ACTIVITIES.md for activity-related functionality
- TRIBE-DESIGN.md for everything around Tribe structure, management, and invitations
- MVP-ROADMAP.md for a rundown of everything that's currently planned for this project.
- TESTING.md for anything related to designing, developing, or running our tests
- DECISION-MAKING.md for everything around decision making and the elimination process/system.

