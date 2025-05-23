# Open Questions and Unresolved Issues

This document contains ambiguities, unclear specifications, unresolved questions, and conflicts between the design specifications that need to be resolved before implementation.

## Database Design Conflicts

### Naming Conventions
- [x] **Decision made**: Snake_case for database columns (e.g., `tribe_name`, `created_at`)
- [x] **Decision made**: Plural table names (e.g., `users`, `tribes`, `lists`)

### Primary Key Strategy
- [x] **Decision made**: UUID primary keys for all entities
  - Provides better support for distributed systems and external sync
  - Avoids sequential ID exposure in URLs

### List Ownership Model
- [x] **Decision made**: Use ChatGPTo3's approach - unified lists table with `owner_type` and `owner_id`
  - `owner_type`: 'user' | 'tribe'
  - `owner_id`: UUID referencing either users.id or tribes.id
  - Simpler than multiple nullable columns

### Soft Deletion Strategy
- [ ] **Deferred**: May not implement soft deletion for MVP
  - Hard deletes with confirmation prompts may be sufficient
  - Database backups can serve as recovery mechanism
  - Will revisit if user feedback indicates need for "undo" functionality

## Authentication & Authorization

### OAuth Provider Support
- [x] **Decision made**: Google OAuth only for initial implementation
  - Design interface to support multiple providers in future
  - Interface should extract email address as primary identifier
  - Include "Dev/Test User Login" mode for development environment
    - Active only in development mode
    - Allows typing in any email address to "login"
    - Implements same interface as OAuth providers

### Session Management
- [ ] **Analysis needed**: JWT tokens vs session cookies vs hybrid approach

#### Detailed Comparison

**JWT Tokens (Stateless)**
- **Pros**: 
  - Stateless - no server-side session storage needed
  - Works well with microservices (if we scale later)
  - Contains user claims directly
  - Works across multiple devices naturally
- **Cons**: 
  - Cannot revoke until expiration
  - Larger payload in every request
  - Need refresh token mechanism for security
- **Our Use Case Fit**: ⭐⭐⭐⭐ (Good - supports concurrent devices, simple for monolith)

**Session Cookies (Stateful)**
- **Pros**: 
  - Can revoke immediately 
  - Smaller request overhead
  - More secure for sensitive apps
  - Traditional web app approach
- **Cons**: 
  - Requires server-side session storage (Redis/DB)
  - More complex for multiple devices
  - Additional infrastructure dependency
- **Our Use Case Fit**: ⭐⭐⭐ (Okay - more complex for our scale)

**Hybrid Approach (JWT + Refresh Tokens)**
- **Pros**: 
  - Short-lived JWTs (15-30 min) with refresh tokens
  - Can revoke refresh tokens
  - Good security/usability balance
- **Cons**: 
  - Most complex implementation
  - Still need some server-side storage for refresh tokens
- **Our Use Case Fit**: ⭐⭐⭐⭐⭐ (Best - security + usability + supports our needs)

**Recommendation**: Hybrid approach with:
- Short-lived JWT access tokens (30 minutes)
- Longer-lived refresh tokens (7 days) stored server-side
- Automatic refresh in frontend
- Support for multiple concurrent sessions

- [x] **Decision made**: Support concurrent sessions across devices
  - Users should be able to interact from desktop and mobile simultaneously
  - All valid sessions should be accepted

### Authorization Levels
- [x] **Decision made**: List sharing permissions model
  - **Read-only sharing**: Shared lists are view-only for recipients
  - **Owner-only modifications**: Only list owners can edit or re-share lists
  - **List ownership**: User owns personal lists, any tribe member owns tribe lists
  - **No re-sharing**: Recipients cannot re-share lists (designed to support, not implemented initially)
  - **Revocable shares**: List owners can revoke sharing at any time
  - **Tribe departure rules**:
    - User loses access to tribe's internal lists
    - User loses access to lists shared with the tribe (unless separately shared with user)
    - User's lists shared with tribe default to unshared, with option to preserve sharing

## Decision-Making Algorithm

### KN+M Parameter Defaults
- [x] **Decision made**: Default values and configuration
  - **N**: Always equals tribe size (1-8 members)
  - **K**: Default 2 eliminations per person (configurable per tribe and per session)
  - **M**: Default 3 final options for random selection (configurable per tribe and per session)
  - **Configuration**: Easy to change defaults at tribe level and override per decision session

### Algorithm for Insufficient Results
- [x] **Decision made**: Strategy when filtered results (R) < K*N + M
  - If K*N + M > R, apply reduction algorithm:
    1. If K > 2: reduce K by 1, check again
    2. If M > 3: reduce M by 1, check again  
    3. If K > 1: reduce K by 1, check again
    4. If M > 1: reduce M by 1, check again
    5. Else: set K=0, M=R (final fallback)
  - Skip algorithm if R = 0 (no valid results)

### Filtering Priority
- [ ] **Specification needed**: Order of filter application and conflicts
- [ ] **Question**: How to handle empty filter results - suggest relaxing which filters?
- [ ] **Question**: Should filters be "hard" (must match) or "weighted" (prefer matches)?

### Real-time vs Turn-based Elimination
- [ ] **Major UX decision**: How should elimination rounds work?
  - Real-time with WebSockets (Claude4 suggests this as future)
  - Turn-based with polling
  - Asynchronous with notifications
  - Simple refresh-to-update model

## List Item Metadata

### Location Data
- [ ] **Decision needed**: How to store and validate location information
  - Structured address fields vs free text
  - Lat/lon coordinates storage and precision
  - Integration with mapping services (if any)

### Opening Hours
- [ ] **Specification needed**: Format for storing business hours
  - Structured JSON vs free text
  - Timezone handling
  - Holiday/special hours support

### Item Categories and Tags
- [ ] **Decision needed**: Predefined enums vs free-form tags
  - Restaurant categories: predefined list or user-defined?
  - Dietary restrictions: standardized values or free text?
  - Activity types: how to categorize non-restaurant items?

### Visit History Tracking
- [ ] **Clarification needed**: What constitutes a "visit"
- [ ] **Question**: How to handle group visits - individual or shared logging?
- [ ] **Question**: Can visit history be edited or deleted?

## Sharing and Permissions

### List Sharing Scope
- [x] **Decision made**: Personal lists can be shared with users or tribes
- [x] **Decision made**: Tribe lists cannot be shared outside the tribe
- [x] **Decision made**: Recipients have read-only permissions

### Data Privacy
- [ ] **Important**: When user leaves tribe or deletes account:
  - [x] **Resolved**: Tribe departure behavior defined above
  - [ ] What happens to their contributed list items in tribe lists?
  - [ ] What happens to their visit history?
  - [ ] What happens to lists they created but shared?
  - [ ] GDPR-style deletion requirements?

## Technical Architecture

### API Design
- [x] **Decision confirmed**: Hybrid GraphQL/REST approach
  - GraphQL for complex queries and mutations
  - REST for simple operations (auth, file uploads, health checks)

### Real-time Features
- [ ] **Scope question**: Which features need real-time updates?
  - Decision elimination rounds
  - List collaboration
  - Tribe member actions
  - Notifications

### Caching Strategy
- [ ] **Question**: What data should be cached and where?
- [ ] **Question**: Cache invalidation strategy for shared/collaborative data

### File Storage
- [ ] **Question**: How to handle user avatars and potential list item images?
- [ ] **Question**: CDN strategy for self-hosted instances?

## User Experience Questions

### Onboarding Flow
- [ ] **Question**: What's the first-run experience for new users?
- [ ] **Question**: Should there be sample/demo data?
- [ ] **Question**: How do users discover and join existing tribes?

### Mobile Experience
- [ ] **Scope question**: Mobile web app vs native app vs responsive web
- [ ] **Question**: Offline capability requirements

### Error Handling
- [ ] **Specification needed**: User-facing error messages and recovery flows
- [ ] **Question**: How to handle network failures during decision sessions?

## External Integrations

### Google Maps Integration
- [ ] **Timeline question**: When to implement external list sync?
- [ ] **Question**: Alternative location services to reduce Google Maps dependency?
- [ ] **Question**: Data import format and conflict resolution

### Notification System
- [ ] **Question**: Email notifications for invites, decisions, etc.?
- [ ] **Question**: In-app notification system requirements?

## Testing Strategy

### Test Data Management
- [ ] **Question**: Strategy for test data in different environments
- [ ] **Question**: How to test collaborative features with multiple simulated users?

### Performance Testing
- [ ] **Question**: Performance benchmarks and acceptable limits
- [ ] **Question**: How to test filtering performance with large datasets?

## Deployment and Operations

### Self-Hosting Requirements
- [ ] **Specification needed**: Minimum system requirements
- [ ] **Question**: Docker vs native installation options
- [ ] **Question**: Database backup and migration strategy

### Configuration Management
- [ ] **Question**: How should self-hosters configure OAuth keys, email settings, etc.?
- [ ] **Question**: Runtime configuration vs build-time configuration

## Data Model Edge Cases

### Tribe Management
- [ ] **Question**: What happens when tribe creator leaves or deletes account?
- [ ] **Question**: Can tribe ownership be transferred?
- [ ] **Question**: How to handle inactive/abandoned tribes?

### List Management
- [ ] **Question**: Bulk operations support (import, export, bulk edit)?
- [ ] **Question**: List versioning or change history?
- [ ] **Question**: Maximum limits (items per list, lists per user, etc.)?

### Decision Session Management
- [ ] **Question**: How long do decision sessions remain active?
- [ ] **Question**: Can decision sessions be paused and resumed?
- [ ] **Question**: History retention for completed decisions?

## Future Feature Considerations

### Scalability Boundaries
- [ ] **Clarification needed**: Define the expected scale limits
  - Maximum users per instance
  - Maximum tribes per user
  - Maximum items per list
  - Maximum lists per decision session

### Extension Points
- [ ] **Question**: Plugin architecture for custom filters or integrations?
- [ ] **Question**: API for third-party integrations?

## Decisions Made Summary

**Database Design:**
- snake_case column naming
- Plural table names  
- UUID primary keys
- Unified lists table with owner_type/owner_id
- Soft deletion deferred (may use hard deletes)

**Authentication:**
- Google OAuth with extensible interface
- Dev/test login mode for development
- Concurrent sessions supported
- Hybrid JWT approach recommended

**Permissions:**
- Read-only list sharing
- Owner-only modifications and re-sharing
- Revocable shares
- Complex tribe departure rules

**Decision Algorithm:**
- N = tribe size, K = 2, M = 3 (configurable)
- Reduction algorithm for insufficient results

These questions should be resolved through discussion and decision-making before proceeding with detailed implementation planning. 