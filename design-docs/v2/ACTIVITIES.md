# Tribe Activities

## Activity History and Tracking

Activity tracking in Tribe allows users and tribes to log and manage activities for list items (restaurants visited, movies watched, activities completed). The system supports both confirmed (past) and tentative (planned) activities.

## Core Concepts

### Activity Types
- **Visited** - For restaurants and physical locations
- **Watched** - For movies, shows, and entertainment
- **Completed** - For general activities and experiences

### Activity Status
- **Confirmed** - Activity has been completed (default for past dates)
- **Tentative** - Activity is planned for the future
- **Cancelled** - Tentative activity that was cancelled

### Activity Scope
- **Personal Activities** - Individual user activities (TribeID is null)
- **Tribe Activities** - Group activities shared within a tribe

## Key Features

### 1. Activity Logging
- **Automatic Status Detection** - Status determined by completion date vs. current time
- **Participant Tracking** - Multiple users can be associated with group activities
- **Duration Tracking** - Optional duration for time-based activities
- **Notes and Context** - Free-form notes for additional details
- **Decision Session Linking** - Activities can be linked to decision results

### 2. Tentative Activity Management
- **Future Planning** - Schedule activities for future dates
- **Update Flexibility** - Modify tentative activities before completion
- **Status Transitions** - Convert tentative to confirmed or cancelled
- **Tribe Coordination** - Tribe members can manage shared tentative activities

### 3. Activity History
- **Personal History** - View individual activity history across all tribes
- **List Item History** - See all activities for a specific restaurant/movie/activity
- **Tribe Activity Feed** - View all activities within a tribe
- **Filtering and Search** - Filter by type, status, date range, participants

### 4. Decision Integration
- **Automatic Logging** - Decision results can be automatically logged as activities
- **Participant Inheritance** - Decision participants become activity participants
- **Context Preservation** - Link between decision process and actual experience

## Filtering Integration

### Recent Activity Exclusion
Activities are integrated with the filtering system to exclude recently visited items:

- **User-Scoped Filtering** - Exclude items visited by the user recently
- **Tribe-Scoped Filtering** - Exclude items visited by any tribe member recently
- **Configurable Timeframe** - Customizable "recent" period (e.g., 30 days)
- **Activity Type Awareness** - Different filters for different activity types

### Filter Configuration Examples
```json
{
  "type": "recent_activity",
  "criteria": {
    "exclude_days": 30,
    "user_id": "user-123",
    "tribe_id": "tribe-456",
    "activity_types": ["visited", "watched"]
  }
}
```

## Data Model

### Type Definitions
All activity-related types are defined in [DATA-MODEL.md](./DATA-MODEL.md#activity-tracking-types):

- `ActivityEntry` - Core activity record
- `LogActivityRequest` - Request structure for logging activities  
- `UpdateActivityRequest` - Request structure for updating tentative activities

### Database Schema
Database tables are defined in [DATA-MODEL.md](./DATA-MODEL.md):

- `activity_history` - Main activity tracking table
- Relationships to `users`, `tribes`, `list_items`, and `decision_sessions`

## Implementation

### Service Layer
Complete implementation examples are available in [implementation-examples/activity-service.go](./implementation-examples/activity-service.go).

Key service methods:
- `LogActivity()` - Create new activity entries
- `UpdateTentativeActivity()` - Modify planned activities
- `LogDecisionResult()` - Auto-log from decision sessions
- `GetUserActivities()` - Retrieve user activity history
- `GetListItemActivities()` - Get activities for specific items
- `GetRecentActivities()` - Support filtering integration

### API Design
Activity APIs follow the hybrid GraphQL/REST pattern:

**GraphQL Queries** (complex data retrieval):
```graphql
query getUserActivities($userId: ID!, $tribeId: ID) {
  userActivities(userId: $userId, tribeId: $tribeId) {
    id
    listItem { name, category }
    activityType
    activityStatus
    completedAt
    participants { name }
    notes
  }
}
```

**REST Endpoints** (simple operations):
```
POST /api/activities/log
PUT /api/activities/{id}/confirm
DELETE /api/activities/{id}
```

## Testing Strategy

Activity tracking follows the comprehensive testing approach outlined in [TESTING.md](./TESTING.md):

### Unit Tests
- Activity service logic
- Status determination algorithms
- Validation rules
- Participant management

### Integration Tests
- Database interactions
- Tribe membership validation
- Decision session integration
- Filter system integration

### End-to-End Tests
- Complete activity logging flow
- Tentative activity management
- Multi-user activity coordination
- Decision-to-activity workflows

### Test Coverage Goals
- **70% line coverage** for all activity-related code
- **Complete user journey coverage** in E2E tests
- **Edge case testing** for status transitions and validation

## Future Enhancements

### Activity Recommendations
- Suggest similar items based on activity history
- Recommend activities based on tribe preferences
- Time-based activity suggestions (seasonal, trending)

### Enhanced Analytics
- Activity frequency analysis
- Preference pattern recognition
- Tribe activity insights and trends

### External Integrations
- Calendar integration for tentative activities
- Photo/review integration for completed activities
- Social sharing of completed activities

### Advanced Filtering
- Location-based activity exclusion
- Preference-based activity weighting
- Dynamic timeframe calculation (more recent = stronger exclusion)

---

**Related Documentation:**
- [DATA-MODEL.md](./DATA-MODEL.md) - Complete type definitions and database schema
- [DECISION-MAKING.md](./DECISION-MAKING.md) - Decision session integration
- [TESTING.md](./TESTING.md) - Testing strategies and coverage goals
- [implementation-examples/activity-service.go](./implementation-examples/activity-service.go) - Complete service implementation
