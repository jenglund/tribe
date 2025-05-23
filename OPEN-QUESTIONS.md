# Open Questions and Unresolved Issues

This document contains ambiguities, unclear specifications, unresolved questions, and conflicts between the design specifications that need to be resolved before implementation.

## Database Design Conflicts

### Naming Conventions
- [ ] **Decision needed**: Snake_case (chatgpto3, claude4) vs camelCase (chatgpt45, gemini25) for database columns
- [ ] **Decision needed**: Table naming conventions - plural vs singular (users vs user)

### Primary Key Strategy
- [ ] **Decision needed**: UUID vs SERIAL/AUTO_INCREMENT primary keys
  - Claude4 and ChatGPT45 prefer UUIDs
  - Gemini25 suggests "UUIDs for distributed systems, integers for simplicity"
  - ChatGPTo3 uses UUIDs

### List Ownership Model
- [ ] **Conflict**: Unified lists table vs separate personal/tribe lists
  - Claude4: Unified lists table with type field and owner_id/tribe_id
  - ChatGPT45: Unified with owner_user_id/owner_tribe_id (both nullable)
  - Gemini25: Suggests unified with list_visibility_type
  - ChatGPTo3: Unified with owner_type and owner_id

### Soft Deletion Strategy
- [ ] **Clarification needed**: Consistent soft deletion implementation across all entities
- [ ] **Decision needed**: Vacuum/cleanup strategy for soft-deleted records
- [ ] **Question**: Should soft-deleted entities cascade delete relationships?

## Authentication & Authorization

### OAuth Provider Support
- [ ] **Priority clarification**: Which OAuth providers beyond Google?
  - Claude4: Mentions "other providers" as future
  - ChatGPT45: "Apple/Microsoft OAuth?"
  - All agree Google is primary

### Session Management
- [ ] **Technical decision**: JWT tokens vs session cookies vs hybrid approach
- [ ] **Question**: Token refresh strategy and expiration times
- [ ] **Question**: How to handle concurrent sessions across devices?

### Authorization Levels
- [ ] **Clarification needed**: Exact permission model for shared lists
  - Read-only vs read-write sharing
  - Can shared list permissions be revoked?
  - What happens to shared lists when user leaves tribe?

## Decision-Making Algorithm

### KN+M Parameter Defaults
- [ ] **Question**: Default K and M values for different tribe sizes
- [ ] **Question**: Maximum and minimum values for K and M
- [ ] **Algorithm**: How to handle edge cases:
  - [ ] What if filtered results < K*N + M?
  - [ ] What if tribe member doesn't participate in elimination?
  - [ ] What if multiple people try to eliminate the same item?

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
- [ ] **Question**: Can personal lists be shared with individual users or only tribes?
- [ ] **Question**: Can tribe lists be shared outside the tribe?
- [ ] **Clarification**: What permissions do shared list recipients have?

### Data Privacy
- [ ] **Important**: When user leaves tribe or deletes account:
  - [ ] What happens to their contributed list items?
  - [ ] What happens to their visit history?
  - [ ] What happens to lists they created but shared?
  - [ ] GDPR-style deletion requirements?

## Technical Architecture

### API Design
- [ ] **Decision needed**: REST vs GraphQL vs hybrid approach
  - User specified GraphQL preference
  - Need to determine where GraphQL is appropriate vs REST

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

These questions should be resolved through discussion and decision-making before proceeding with detailed implementation planning. 