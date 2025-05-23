# API Design (Hybrid GraphQL/REST)

**Related Documents:**
- [Database Schema](./database-schema.md) - Database structure that backs this API
- [Architecture](./architecture.md) - System architecture context
- [Authentication](./authentication.md) - Auth requirements for API access

## API Strategy

**Primary API**: GraphQL for complex queries, real-time subscriptions, and most operations
**Secondary API**: REST for simple operations, file uploads, and authentication

## GraphQL Schema

### User and Authentication Types

```graphql
type User {
  id: ID!
  email: String!
  name: String!
  displayName: String!
  avatarUrl: String
  timezone: String! # User's timezone preference (e.g., "America/New_York")
  dietaryPreferences: [String!]!
  tribes: [Tribe!]!
  personalLists: [List!]!
  activityHistory: [ActivityEntry!]!
  createdAt: DateTime!
}
```

### Tribe Management Types

```graphql
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
  tribeDisplayName: String
  invitedAt: DateTime!
  invitedBy: User!
  joinedAt: DateTime!
  lastLoginAt: DateTime
  isActive: Boolean!
  isCreator: Boolean! # Computed: user.id == invitedBy.id
  isSenior: Boolean! # Computed: earliest invited_at among active members
}

enum TribeMemberRole {
  CREATOR # Informational only - same permissions as MEMBER
  MEMBER
}
```

### List Management Types

```graphql
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
  type: String # "restaurant", "entertainment", "activity"
  phone: String
  website: String
  priceRange: String
  regularHours: RegularHours
  timezone: String
  notes: String
}

type RegularHours {
  monday: DayHours
  tuesday: DayHours
  wednesday: DayHours
  thursday: DayHours
  friday: DayHours
  saturday: DayHours
  sunday: DayHours
}

type DayHours {
  open: String # "HH:MM" format
  close: String # "HH:MM" format
  closed: Boolean!
}

type DietaryInfo {
  vegetarian: Boolean!
  vegan: Boolean!
  glutenFree: Boolean!
  customTags: [String!]!
}
```

### Activity Tracking Types

```graphql
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
```

### Decision Making Types

```graphql
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
  timeBasedFilter: TimeBasedFilter
  priceRange: PriceRange
  tags: [String!]!
  excludeTags: [String!]!
}

type TimeBasedFilter {
  mustBeOpenFor: Int # minutes from now
  mustBeOpenUntil: String # "HH:MM" in user's timezone
  checkDate: String # ISO date string, optional
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
  skipType: SkipType!
  skippedAt: DateTime!
}

enum SkipType {
  QUICK_SKIP
  TIMEOUT_SKIP
  FORFEITED
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
  canQuickSkip: Boolean!
  quickSkipsUsed: Int!
  quickSkipsLimit: Int!
  isCatchUpPhase: Boolean!
}
```

### GraphQL Mutations

```graphql
type Mutation {
  # User Management
  updateUserProfile(input: UpdateUserProfileInput!): User!
  deleteAccount: Boolean!
  
  # Tribe Management
  createTribe(input: CreateTribeInput!): Tribe!
  inviteToTribe(tribeId: ID!, email: String!, suggestedDisplayName: String): Boolean!
  acceptInvitation(invitationId: ID!): Tribe!
  voteOnInvitation(invitationId: ID!, approve: Boolean!): Boolean!
  petitionMemberRemoval(tribeId: ID!, targetUserId: ID!, reason: String!): MemberRemovalPetition!
  voteOnMemberRemoval(petitionId: ID!, approve: Boolean!): Boolean!
  leaveTribe(tribeId: ID!): Boolean!
  petitionTribeDeletion(tribeId: ID!, reason: String!): TribeDeletionPetition!
  voteOnTribeDeletion(petitionId: ID!, approve: Boolean!): Boolean!
  
  # List Management
  createList(input: CreateListInput!): List!
  updateList(id: ID!, input: UpdateListInput!): List!
  deleteList(id: ID!): Boolean!
  petitionListDeletion(listId: ID!, reason: String!): ListDeletionPetition!
  resolveListDeletion(petitionId: ID!, confirm: Boolean!): Boolean!
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
  
  # Decision Making with Quick-Skip
  createDecisionSession(input: CreateDecisionSessionInput!): DecisionSession!
  addListsToSession(sessionId: ID!, listIds: [ID!]!): DecisionSession!
  applyFilters(sessionId: ID!, filters: FilterCriteriaInput!): DecisionSession!
  startElimination(sessionId: ID!): DecisionSession!
  eliminateItem(sessionId: ID!, itemId: ID!): DecisionSession!
  quickSkipTurn(sessionId: ID!): DecisionSession!
  rejoinElimination(sessionId: ID!): DecisionSession!
  pinSession(sessionId: ID!): DecisionSession!
  completeDecision(sessionId: ID!): DecisionSession!
  cancelDecision(sessionId: ID!): DecisionSession!
}
```

### GraphQL Queries

```graphql
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
```

### GraphQL Subscriptions

```graphql
type Subscription {
  decisionSessionUpdated(sessionId: ID!): DecisionSession!
  tribeUpdated(tribeId: ID!): Tribe!
}
```

### Input Types (Key Examples)

```graphql
input CreateTribeInput {
  name: String!
  description: String
  maxMembers: Int
}

input CreateListInput {
  name: String!
  description: String
  category: String
  ownerType: String! # "user" or "tribe"
  ownerId: ID!
}

input AddListItemInput {
  name: String!
  description: String
  category: String
  tags: [String!]
  location: LocationInput
  businessInfo: BusinessInfoInput
  dietaryInfo: DietaryInfoInput
}

input FilterCriteriaInput {
  categories: [String!]
  excludeCategories: [String!]
  dietaryRequirements: [String!]
  maxDistance: Float
  centerLocation: LocationInput
  excludeRecentlyVisited: Boolean
  recentlyVisitedDays: Int
  timeBasedFilter: TimeBasedFilterInput
  tags: [String!]
  excludeTags: [String!]
}

input LogActivityInput {
  listItemId: ID!
  userId: ID!
  tribeId: ID
  activityType: ActivityType!
  activityStatus: ActivityStatus
  completedAt: DateTime!
  durationMinutes: Int
  participants: [ID!]
  notes: String
}
```

## REST Endpoints

### Authentication Endpoints
```
POST   /api/v1/auth/google/login     # OAuth login
POST   /api/v1/auth/refresh          # Refresh JWT token
POST   /api/v1/auth/logout           # Logout
```

### Health and Meta Endpoints
```
GET    /api/v1/health               # Health check
GET    /api/v1/version              # API version
```

### File Upload Endpoints
```
POST   /api/v1/uploads/avatar       # Upload user avatar
POST   /api/v1/uploads/list-image   # Upload list item image
```

## API Authentication

### JWT Token Structure
All GraphQL requests require a valid JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

### Token Claims
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "provider": "google",
  "exp": 1234567890,
  "iat": 1234567890
}
```

## Error Handling

### GraphQL Errors
```json
{
  "errors": [
    {
      "message": "User not found",
      "code": "USER_NOT_FOUND",
      "path": ["user"],
      "extensions": {
        "code": "USER_NOT_FOUND",
        "timestamp": "2024-01-01T00:00:00Z"
      }
    }
  ],
  "data": null
}
```

### REST Errors
```json
{
  "error": {
    "code": "INVALID_TOKEN",
    "message": "JWT token is invalid or expired",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

## Rate Limiting

### General Limits
- 1000 requests per hour per user for GraphQL
- 100 requests per hour per user for file uploads
- 10 requests per minute for authentication endpoints

### Decision Session Limits
- 1 active decision session per tribe at a time
- Maximum 10 eliminations per user per session

## Real-time Features

### WebSocket Connection
GraphQL subscriptions are delivered over WebSocket connections using the graphql-ws protocol.

### Subscription Usage
```typescript
// Subscribe to decision session updates
const subscription = useSubscription(DECISION_SESSION_UPDATED, {
  variables: { sessionId: "uuid" }
});

// Subscribe to tribe updates
const tribeSubscription = useSubscription(TRIBE_UPDATED, {
  variables: { tribeId: "uuid" }
});
```

## Performance Considerations

### Query Complexity Analysis
- Maximum query depth: 10 levels
- Maximum query complexity score: 1000
- Automatic query analysis and rejection

### Caching Strategy
- List data: 5 minutes
- User profile data: 15 minutes
- Decision session data: No cache (real-time)
- Activity history: 1 hour

### Pagination
All list queries support cursor-based pagination:
```graphql
query GetLists($first: Int!, $after: String) {
  lists(first: $first, after: $after) {
    edges {
      node {
        id
        name
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

## API Development Guidelines

### Naming Conventions
- Types: PascalCase (User, Tribe, ListItem)
- Fields: camelCase (displayName, createdAt)
- Enums: UPPER_SNAKE_CASE (DECISION_STATUS)

### Null Handling
- Required fields: Use ! to indicate non-null
- Optional fields: Allow null returns
- Lists: Always return empty array instead of null

### Versioning Strategy
- GraphQL schema evolution (additive changes only)
- REST API versioning (/api/v1/, /api/v2/)
- Deprecation notices in schema descriptions

---

*For database mapping details, see [Database Schema](./database-schema.md)*
*For authentication implementation, see [Authentication](./authentication.md)* 