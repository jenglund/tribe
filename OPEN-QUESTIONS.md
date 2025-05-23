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
- [x] **Decision made**: Simple JWT tokens with 7-day expiry for MVP
  - No refresh tokens initially to reduce complexity
  - Adding refresh tokens later is NOT a major architectural change:
    - Core JWT validation logic stays the same
    - Would only require: new refresh endpoint, frontend refresh logic, refresh token storage
    - Authentication interface remains unchanged
    - Can be added incrementally without major refactoring
- [x] **Decision made**: Support concurrent sessions across devices
  - Users should be able to interact from desktop and mobile simultaneously
  - All valid sessions should be accepted

**JWT Implementation Details:**
- **Access Tokens**: JWT with 7-day expiry
- **Storage**: Frontend localStorage/sessionStorage
- **Validation**: Standard JWT signature verification
- **Expiry Handling**: Redirect to login on token expiry
- **Future Migration**: Can add refresh tokens without major changes

### Authorization Levels
- [x] **Decision made**: List sharing permissions model
  - **Read-only sharing**: Shared lists are view-only for recipients
  - **Owner-only modifications**: Only list owners can edit or re-share lists
  - **List ownership**: User owns personal lists, tribe collectively owns tribe lists
  - **Tribe member equality**: All tribe members have equal rights to tribe lists and tribe management
  - **No re-sharing**: Recipients cannot re-share lists (designed to support, not implemented initially)
  - **Revocable shares**: List owners can revoke sharing at any time
  - **Tribe departure rules**:
    - User loses access to tribe's internal lists
    - User loses access to lists shared with the tribe (unless separately shared with user)
    - User's lists shared with tribe default to unshared, with option to preserve sharing
    - Remaining tribe members retain full access and control over tribe resources

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
- [x] **Decision made**: User-configurable filter priority with hard/soft filter distinction
  - **Priority Order**: Filters applied in user-defined order (earliest = highest priority)
  - **Filter Types**: 
    - **Hard Filters**: Must be satisfied (exclude non-matching results)
    - **Soft Filters**: Preferred but can be relaxed (mark violations but include results)
  - **User Interface**: Drag-and-drop list to reorder filter priority
  - **Result Handling**: 
    - Include all results that pass hard filters
    - Mark soft filter violations on each result
    - Sort results by: early soft filter satisfaction > fewer violations > original order
  - **Violation Display**: Show which soft filters each result violates

**Filter System Requirements:**
- User can mark each filter as "hard" (required) or "soft" (preferred)
- User can reorder filters to set priority
- Results show soft filter violations clearly
- Smart sorting prioritizes better matches

### Real-time vs Turn-based Elimination
- [x] **Decision made**: Turn-based elimination with polling updates and quick-skip functionality
  - **One elimination per round**: Users eliminate ONE choice per round (K = number of rounds)
  - **Randomized order**: Elimination order shuffled at start and maintained (e.g., BCA → BCABCA for K=2)
  - **Shared visibility**: All participants see remaining options after each elimination
  - **Polling updates**: Client polls API every 5-10 seconds for updates
  - **Timeout handling**: 5-minute timeout per turn, then skip to next user
  - **Quick-skip option**: Users can voluntarily skip their turn to defer until later
  - **Skip limits**: Maximum K quick-skips per user (one per round maximum)
  - **No double-skipping**: Cannot quick-skip a turn that was already deferred (timeout or previous quick-skip)
  - **Catch-up mechanism**: Skipped users can rejoin and apply missed eliminations at end
  - **Catch-up timeout penalty**: If user times out during catch-up phase, they forfeit ALL remaining skipped rounds
  - **Adaptive M value**: If unresolved turns remain, increase M to accommodate extra options

**Skip Types:**
- **Quick-skip**: Voluntary skip by user choice (limited to K per user)
- **Timeout-skip**: Automatic skip due to 5-minute timeout
- **Forfeited**: Skipped rounds lost due to timeout during catch-up phase

**Implementation Details:**
- Session state tracks current turn, elimination order, skipped users with skip types
- Polling endpoint returns current candidates and whose turn it is
- Timeout system automatically advances turns
- Quick-skip validation prevents double-skipping and enforces limits
- End-of-round catch-up phase for missed eliminations with timeout penalties

## List Item Metadata

### Location Data
- [ ] **Decision needed**: How to store and validate location information
  - Structured address fields vs free text
  - Lat/lon coordinates storage and precision
  - Integration with mapping services (if any)

### Opening Hours
- [x] **Decision made**: Structured JSON format for business hours storage
  - Store regular weekly hours (Monday-Sunday) in JSON format
  - All times stored as UTC/epoch on backend
  - User timezone preferences applied in frontend for display
  - No holiday hours support in MVP
- [x] **Decision made**: Time-based filtering capabilities for MVP
  - Support "closes within X minutes" filtering
  - Support "open until X time" filtering (e.g., "open until 11PM")
  - No location/driving distance calculations in MVP (requires Google Maps API)

**Business Hours JSON Structure:**
```json
{
  "regular_hours": {
    "monday": {"open": "11:00", "close": "22:00", "closed": false},
    "tuesday": {"open": "11:00", "close": "22:00", "closed": false},
    "wednesday": {"open": "11:00", "close": "22:00", "closed": false},
    "thursday": {"open": "11:00", "close": "22:00", "closed": false},
    "friday": {"open": "11:00", "close": "23:00", "closed": false},
    "saturday": {"open": "10:00", "close": "23:00", "closed": false},
    "sunday": {"closed": true}
  },
  "timezone": "America/New_York"
}
```

### Item Categories and Tags
- [ ] **Decision needed**: Predefined enums vs free-form tags
  - Restaurant categories: predefined list or user-defined?
  - Dietary restrictions: standardized values or free text?
  - Activity types: how to categorize non-restaurant items?

### Visit History Tracking
- [x] **Decision made**: Flexible activity logging with tentative future entries
  - Any user can log activity for any list item, whether they participated or not
  - Auto-filled defaults: Time defaults to "now", participants default to full tribe (both editable)
  - Decision integration: Completing decision offers option to log "doing this now/at time X"
  - Tentative entries: Future-dated activities marked as "tentative" and highlighted in UI
  - Confirmation workflow: Any tribe member can confirm/cancel tentative entries
  - Flexible participation: Can specify any subset of tribe members as participants

**Activity Entry Types:**
- **Immediate**: Recorded as happening "now" (confirmed)
- **Past**: Recorded for previous date (confirmed) 
- **Tentative**: Planned for future date (requires confirmation)
- **Confirmed**: Tentative entry that's been verified as completed
- **Cancelled**: Tentative entry that didn't happen

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

### Platform Prioritization
- [x] **Decision made**: Mobile web experience is primary focus
  - Mobile web app optimized for on-the-go usage
  - Desktop web secondary (for list curation, detailed management)
  - No native mobile app development planned
  - Mobile-first design philosophy

### Display Name System
- [x] **Decision made**: Global + per-tribe display name system
  - Users have default display name (configurable in profile)
  - Users can set tribe-specific display names when becoming full member
  - Inviting user can suggest initial tribe display name for invitee
  - Supports both privacy-focused and real-name-focused tribes
  - Accommodates different tribe cultures and privacy preferences

### Onboarding Flow
- [x] **Decision made**: Invitation-driven growth model
  - Users don't seek out tribes - tribes invite new users
  - New users typically join because existing user wants to add them
  - No social media "discoverability" features
  - Focus on intimate, close-knit groups (max 8 people)
  - Platform philosophy: deliberately non-expandable, invitation-only growth

### Mobile Experience
- [x] **Decision made**: Mobile web app with responsive design
  - Progressive Web App (PWA) features for mobile usage
  - Optimized for touch interactions and smaller screens
  - Desktop web provides enhanced experience for list management
  - No offline capability requirements for MVP

### Error Handling
- [ ] **Specification needed**: User-facing error messages and recovery flows
- [ ] **Question**: How to handle network failures during decision sessions?

## External Integrations

### Google Maps Integration
- [x] **Decision made**: Limited integration for MVP
  - No location/driving distance calculations initially (requires paid API)
  - May integrate for list sync in future phases
  - Focus on manual entry and time-based filtering for MVP
- [ ] **Future consideration**: Alternative location services to reduce Google Maps dependency
- [ ] **Future consideration**: Data import format and conflict resolution

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

### Tribe Management (Common Ownership Model)
- [x] **Decision made**: Common ownership model for all tribe members
  - **Equal Rights**: All tribe members have equal control and permissions
  - **No Functional Roles**: No special permissions for creator or any other role
  - **Seniority Calculation**: Senior member determined by earliest invite timestamp among active members
  - **Creator Detection**: Creator is member where `user_id == invited_by_user_id` (self-invited at creation)
  - **Invite Tracking**: Each membership records `invited_at`, `invited_by_user_id` for full audit trail
  - **Collective Ownership**: All members can invite, remove others, modify tribe settings
  - **Democratic Operations**: Any member can start decision sessions, create tribe lists

**Edge Cases Requiring Resolution:**
- [x] **Tribe Deletion**: Requires 100% consensus of all active members
  - All active members must explicitly approve tribe deletion
  - Deletion remains pending until all members have voted to approve
- [x] **Malicious Member Removal**: Unanimous vote required (very high barrier)
  - Any member can petition for removal of another member
  - Every other active member (excluding the one being voted on) must agree
  - Designed to be exceptional, not routine - protects against mass removal
- [x] **Creator Departure**: System handles gracefully since creator role is non-functional
  - Senior member (earliest invite among active members) serves as tie-breaker
  - Creator status is purely historical and doesn't affect tribe operations
- [x] **Invitation Management**: Two-stage ratification system
  - Invitees appear as "pending" members visible to all tribe members
  - Invitees see basic invite prompt but cannot access tribe details
  - ANY existing member can reject invitee (immediate revocation)
  - New members need: (1) accept invite AND (2) unanimous approval from existing members
- [x] **Inactive Member Handling**: Configurable inactivity thresholds with petition system
  - Default: can petition for removal after 1 month of inactivity
  - Configurable per-tribe: 1 day to 2 years
  - After 2+ years: automatic cleanup under privacy laws (low priority)
  - Still requires unanimous vote from other active members to actually remove
  - Inactive members excluded from seniority calculations

**Seniority and Creator Rules:**
- **Seniority**: Calculated dynamically from `invited_at` timestamps of active members only
- **Creator Detection**: `user_id == invited_by_user_id` indicates original tribe creator
- **Tie-Breaking**: Senior member (earliest active invite) resolves conflicts when needed
- **No Static Roles**: All role/status information derived from membership data, not stored separately

**Tribe List Deletion:**
- [x] **Decision made**: Petition + single confirmation system
  - Any member can petition for tribe list deletion
  - List appears as "pending deletion" to all members
  - Any other member can confirm OR cancel the deletion
  - Petitioner can immediately re-petition if cancelled
  - Lower barrier than member removal since lists can be rebuilt

### List Management
- [x] **Updated**: Tribe list ownership aligns with common ownership
  - Any tribe member can create, edit, delete tribe lists
  - Any tribe member can share tribe lists with external users/tribes
  - No special permissions for list creator within tribe context

### Decision Session Management
- [x] **Decision made**: 30-minute inactivity timeout for incomplete sessions
  - Session expires after 30 minutes without activity from any participant
  - No pause/resume functionality - sessions are time-bound and immediate
  - Designed for real-time collaborative decision-making
- [x] **Decision made**: One-month history retention with pinning option
  - Completed session results persist for 1 month automatically
  - Users can pin/sticky sessions to prevent automatic cleanup
  - History display order: final result → runners-up (M set) → reverse chronological eliminations
  - Shows elimination timeline: "last eliminated" → "first eliminated"
- [x] **Decision made**: Configurable elimination visibility per tribe
  - Default: show who eliminated what (transparency)
  - Tribes can configure to hide elimination details if desired
  - Promotes accountability while allowing privacy when needed

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
- Simple JWT tokens (7-day expiry) for MVP - no refresh tokens initially

**Tribe Ownership:**
- Common ownership model: all members have equal rights and control
- No functional roles: creator status is purely historical and non-functional
- Seniority calculated dynamically from invite timestamps of active members
- Creator detected by self-invite pattern (user_id == invited_by_user_id)
- Democratic governance with petition and voting systems
- Very high barriers for member removal (unanimous vote of all other active members)
- 100% consensus required for tribe deletion
- Two-stage invitation system with ratification by existing members

**Member Management:**
- Invitation process: invite → accept → unanimous ratification by existing members
- ANY existing member can reject an invitee (immediate revocation)
- Member removal requires petition + unanimous vote from all other members
- Configurable inactivity thresholds (1 day to 2 years, default 30 days)
- Inactive members can be petitioned for removal but still need unanimous vote

**List Management:**
- Personal lists: owner can delete immediately
- Tribe lists: petition + single member confirmation system
- Any member can petition for tribe list deletion
- Any other member can confirm or cancel the deletion
- Lower barrier than member removal since lists can be rebuilt

**Permissions:**
- Read-only list sharing
- Owner-only modifications and re-sharing for personal lists
- Any tribe member can manage tribe lists collectively
- Revocable shares
- Complex tribe departure rules

**Decision Algorithm:**
- N = tribe size, K = number of elimination rounds, M = 3 (configurable)
- Turn-based elimination: one item per round per user
- Randomized elimination order maintained throughout
- 5-minute timeouts with skip/catch-up mechanism
- Quick-skip option: voluntary turn deferral (max K per user)
- No double-skipping rule prevents gaming the system
- Catch-up timeout penalty: forfeit all remaining skipped rounds
- Adaptive M value for unresolved eliminations
- Polling-based updates (5-10 second intervals)

**Activity Tracking:**
- Flexible logging by any user for any list item
- Tentative vs confirmed entry system
- Auto-populated defaults with user override capability
- Integration with decision results
- Tribe member confirmation workflow for tentative entries

**Filter System:**
- User-configurable priority order
- Hard vs soft filter distinction
- Violation tracking and smart sorting
- Drag-and-drop filter management UI

**Decision Sessions:**
- 30-minute inactivity timeout for incomplete sessions
- No pause/resume - designed for immediate, time-bound decisions
- One-month automatic history retention with optional pinning
- History shows: winner → runners-up → reverse chronological eliminations
- Configurable elimination visibility per tribe (default: transparent)

**Platform & UX:**
- Mobile web experience is primary focus (Progressive Web App)
- Desktop web secondary for detailed list management
- No native mobile app development
- Invitation-driven growth model - no social discoverability
- Maximum 8 members per tribe (deliberately intimate design)

**Time & Business Hours:**
- All backend times stored as UTC/epoch timestamps
- User timezone preferences applied in frontend for display
- Structured JSON format for business hours (Monday-Sunday)
- Time-based filtering: "open for X minutes" and "open until X time"
- No holiday hours or location-based filtering in MVP
- No Google Maps API integration for MVP (cost considerations)

**Display Names:**
- Global default display name configurable in user profile
- Per-tribe display names set when becoming full member
- Inviting user can suggest tribe-specific name for invitee
- Supports both privacy-focused and real-name tribes

These questions should be resolved through discussion and decision-making before proceeding with detailed implementation planning. 