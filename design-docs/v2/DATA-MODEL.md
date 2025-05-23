# Tribe Data Model

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
    timezone VARCHAR(100) DEFAULT 'UTC', -- User's timezone preference (e.g., 'America/New_York')
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
    tribe_display_name VARCHAR(255), -- User's display name within this tribe
    invited_at TIMESTAMPTZ NOT NULL, -- When invite was sent (used for seniority calculation)
    invited_by_user_id UUID NOT NULL REFERENCES users(id), -- Who invited this user (self-reference for creator)
    joined_at TIMESTAMPTZ DEFAULT NOW(), -- When user actually joined tribe
    last_login_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE, -- For marking inactive members
    UNIQUE(tribe_id, user_id)
);

-- Function to get senior member (earliest invite among active members)
CREATE OR REPLACE FUNCTION get_tribe_senior_member(tribe_uuid UUID)
RETURNS UUID AS $$
BEGIN
    RETURN (
        SELECT user_id 
        FROM tribe_memberships 
        WHERE tribe_id = tribe_uuid 
          AND is_active = TRUE
        ORDER BY invited_at ASC 
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get tribe creator (self-invited member)
CREATE OR REPLACE FUNCTION get_tribe_creator(tribe_uuid UUID)
RETURNS UUID AS $$
BEGIN
    RETURN (
        SELECT user_id 
        FROM tribe_memberships 
        WHERE tribe_id = tribe_uuid 
          AND user_id = invited_by_user_id
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
    business_info JSONB, -- Structured business information (see examples below)
    dietary_info JSONB, -- {vegetarian: true, vegan: false, gluten_free: true}
    external_id VARCHAR(255), -- For future external API sync
    added_by_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Business Info JSON Examples
```sql
-- Restaurant with business hours
{
  "type": "restaurant",
  "phone": "+1-555-123-4567",
  "website": "https://example-restaurant.com",
  "price_range": "$$$",
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

-- Movie theater
{
  "type": "entertainment",
  "phone": "+1-555-987-6543",
  "website": "https://example-theater.com",
  "regular_hours": {
    "monday": {"open": "12:00", "close": "23:00", "closed": false},
    "tuesday": {"open": "12:00", "close": "23:00", "closed": false},
    "wednesday": {"open": "12:00", "close": "23:00", "closed": false},
    "thursday": {"open": "12:00", "close": "23:00", "closed": false},
    "friday": {"open": "12:00", "close": "24:00", "closed": false},
    "saturday": {"open": "10:00", "close": "24:00", "closed": false},
    "sunday": {"open": "12:00", "close": "22:00", "closed": false}
  },
  "timezone": "America/New_York"
}

-- Activity without specific hours
{
  "type": "activity",
  "website": "https://example-park.gov",
  "notes": "Open during daylight hours"
}
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
    skipped_users JSONB DEFAULT '[]'::jsonb, -- Users who were skipped with skip type and details
    user_skip_counts JSONB DEFAULT '{}'::jsonb, -- Track quick-skip usage per user: {"user_id": 2}
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

#### Tribe Settings Table
```sql
CREATE TABLE tribe_settings (
    tribe_id UUID PRIMARY KEY REFERENCES tribes(id) ON DELETE CASCADE,
    inactivity_threshold_days INTEGER DEFAULT 30, -- 1 to 730 (2 years)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Filter Configurations Table
```sql
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

-- Governance and invitation indexes
CREATE INDEX idx_tribe_invitations_tribe ON tribe_invitations(tribe_id);
CREATE INDEX idx_tribe_invitations_invitee ON tribe_invitations(invitee_email);
CREATE INDEX idx_tribe_invitations_status ON tribe_invitations(status);
CREATE INDEX idx_tribe_invitation_ratifications_invitation ON tribe_invitation_ratifications(invitation_id);
CREATE INDEX idx_member_removal_petitions_tribe ON member_removal_petitions(tribe_id);
CREATE INDEX idx_member_removal_petitions_target ON member_removal_petitions(target_user_id);
CREATE INDEX idx_member_removal_votes_petition ON member_removal_votes(petition_id);
CREATE INDEX idx_tribe_deletion_petitions_tribe ON tribe_deletion_petitions(tribe_id);
CREATE INDEX idx_tribe_deletion_votes_petition ON tribe_deletion_votes(petition_id);
CREATE INDEX idx_list_deletion_petitions_list ON list_deletion_petitions(list_id);

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
  displayName: String!
  avatarUrl: String
  timezone: String! # User's timezone preference (e.g., "America/New_York")
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

# Mutations
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

## Go Type Definitions

### Core Entity Types

```go
// User represents a user in the system
type User struct {
    ID                  string    `json:"id" db:"id"`
    Email               string    `json:"email" db:"email"`
    Name                string    `json:"name" db:"name"`
    DisplayName         string    `json:"display_name" db:"display_name"`
    AvatarURL           *string   `json:"avatar_url" db:"avatar_url"`
    OAuthProvider       string    `json:"oauth_provider" db:"oauth_provider"`
    OAuthID             string    `json:"oauth_id" db:"oauth_id"`
    Timezone            string    `json:"timezone" db:"timezone"`
    DietaryPreferences  []string  `json:"dietary_preferences" db:"dietary_preferences"`
    LocationPreferences *Location `json:"location_preferences" db:"location_preferences"`
    EmailVerified       bool      `json:"email_verified" db:"email_verified"`
    CreatedAt           time.Time `json:"created_at" db:"created_at"`
    UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// Tribe represents a group of users
type Tribe struct {
    ID                    string                     `json:"id" db:"id"`
    Name                  string                     `json:"name" db:"name"`
    Description           *string                    `json:"description" db:"description"`
    CreatorID             string                     `json:"creator_id" db:"creator_id"`
    MaxMembers            int                        `json:"max_members" db:"max_members"`
    DecisionPreferences   *TribeDecisionPreferences  `json:"decision_preferences" db:"decision_preferences"`
    ShowEliminationDetails bool                      `json:"show_elimination_details" db:"show_elimination_details"`
    CreatedAt             time.Time                  `json:"created_at" db:"created_at"`
    UpdatedAt             time.Time                  `json:"updated_at" db:"updated_at"`
}

// TribeMembership represents the relationship between users and tribes
type TribeMembership struct {
    ID               string     `json:"id" db:"id"`
    TribeID          string     `json:"tribe_id" db:"tribe_id"`
    UserID           string     `json:"user_id" db:"user_id"`
    TribeDisplayName *string    `json:"tribe_display_name" db:"tribe_display_name"`
    InvitedAt        time.Time  `json:"invited_at" db:"invited_at"`
    InvitedByUserID  string     `json:"invited_by_user_id" db:"invited_by_user_id"`
    JoinedAt         time.Time  `json:"joined_at" db:"joined_at"`
    LastLoginAt      *time.Time `json:"last_login_at" db:"last_login_at"`
    IsActive         bool       `json:"is_active" db:"is_active"`
}

// List represents a collection of items
type List struct {
    ID          string                 `json:"id" db:"id"`
    Name        string                 `json:"name" db:"name"`
    Description *string                `json:"description" db:"description"`
    OwnerType   string                 `json:"owner_type" db:"owner_type"` // 'user' or 'tribe'
    OwnerID     string                 `json:"owner_id" db:"owner_id"`
    Category    *string                `json:"category" db:"category"`
    Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ListItem represents an individual item within a list
type ListItem struct {
    ID             string                 `json:"id" db:"id"`
    ListID         string                 `json:"list_id" db:"list_id"`
    Name           string                 `json:"name" db:"name"`
    Description    *string                `json:"description" db:"description"`
    Category       *string                `json:"category" db:"category"`
    Tags           []string               `json:"tags" db:"tags"`
    Location       *Location              `json:"location" db:"location"`
    BusinessInfo   *BusinessInfo          `json:"business_info" db:"business_info"`
    DietaryInfo    *DietaryInfo           `json:"dietary_info" db:"dietary_info"`
    ExternalID     *string                `json:"external_id" db:"external_id"`
    AddedByUserID  string                 `json:"added_by_user_id" db:"added_by_user_id"`
    CreatedAt      time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// Location represents geographical information
type Location struct {
    Address   *string  `json:"address"`
    Latitude  *float64 `json:"latitude"`
    Longitude *float64 `json:"longitude"`
    City      *string  `json:"city"`
    State     *string  `json:"state"`
    Country   *string  `json:"country"`
}

// BusinessInfo represents business-specific information
type BusinessInfo struct {
    Type         *string       `json:"type"`
    Phone        *string       `json:"phone"`
    Website      *string       `json:"website"`
    PriceRange   *string       `json:"price_range"`
    RegularHours *RegularHours `json:"regular_hours"`
    Timezone     *string       `json:"timezone"`
    Notes        *string       `json:"notes"`
}

// RegularHours represents business operating hours
type RegularHours struct {
    Monday    *DayHours `json:"monday"`
    Tuesday   *DayHours `json:"tuesday"`
    Wednesday *DayHours `json:"wednesday"`
    Thursday  *DayHours `json:"thursday"`
    Friday    *DayHours `json:"friday"`
    Saturday  *DayHours `json:"saturday"`
    Sunday    *DayHours `json:"sunday"`
}

// DayHours represents operating hours for a specific day
type DayHours struct {
    Open   *string `json:"open"`   // "HH:MM" format
    Close  *string `json:"close"`  // "HH:MM" format
    Closed bool    `json:"closed"`
}

// DietaryInfo represents dietary restriction information
type DietaryInfo struct {
    Vegetarian  bool     `json:"vegetarian"`
    Vegan       bool     `json:"vegan"`
    GlutenFree  bool     `json:"gluten_free"`
    CustomTags  []string `json:"custom_tags"`
}
```

### Decision Making Types

```go
// DecisionSession represents a collaborative decision-making session
type DecisionSession struct {
    ID                     string                 `json:"id" db:"id"`
    TribeID                string                 `json:"tribe_id" db:"tribe_id"`
    Name                   *string                `json:"name" db:"name"`
    Status                 string                 `json:"status" db:"status"`
    Filters                map[string]interface{} `json:"filters" db:"filters"`
    AlgorithmParams        *AlgorithmParams       `json:"algorithm_params" db:"algorithm_params"`
    EliminationOrder       []string               `json:"elimination_order" db:"elimination_order"`
    CurrentTurnIndex       int                    `json:"current_turn_index" db:"current_turn_index"`
    CurrentRound           int                    `json:"current_round" db:"current_round"`
    TurnStartedAt          *time.Time             `json:"turn_started_at" db:"turn_started_at"`
    TurnTimeoutMinutes     int                    `json:"turn_timeout_minutes" db:"turn_timeout_minutes"`
    SessionTimeoutMinutes  int                    `json:"session_timeout_minutes" db:"session_timeout_minutes"`
    LastActivityAt         time.Time              `json:"last_activity_at" db:"last_activity_at"`
    SkippedUsers           []SkippedTurn          `json:"skipped_users" db:"skipped_users"`
    UserSkipCounts         map[string]int         `json:"user_skip_counts" db:"user_skip_counts"`
    InitialCandidates      []string               `json:"initial_candidates" db:"initial_candidates"`
    CurrentCandidates      []string               `json:"current_candidates" db:"current_candidates"`
    FinalSelectionID       *string                `json:"final_selection_id" db:"final_selection_id"`
    RunnersUp              []string               `json:"runners_up" db:"runners_up"`
    EliminationHistory     []map[string]interface{} `json:"elimination_history" db:"elimination_history"`
    IsPinned               bool                   `json:"is_pinned" db:"is_pinned"`
    CreatedByUserID        string                 `json:"created_by_user_id" db:"created_by_user_id"`
    CreatedAt              time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt              time.Time              `json:"updated_at" db:"updated_at"`
    CompletedAt            *time.Time             `json:"completed_at" db:"completed_at"`
    ExpiresAt              *time.Time             `json:"expires_at" db:"expires_at"`
}

// AlgorithmParams represents K+M elimination algorithm parameters
type AlgorithmParams struct {
    K            int `json:"k"`             // Eliminations per person
    N            int `json:"n"`             // Number of participants
    M            int `json:"m"`             // Final choices for random selection
    InitialCount int `json:"initial_count"` // Total initial candidates
}

// TribeDecisionPreferences represents tribe-specific decision settings
type TribeDecisionPreferences struct {
    DefaultK int `json:"default_k"`
    DefaultM int `json:"default_m"`
    MaxK     int `json:"max_k"`
    MaxM     int `json:"max_m"`
}

// SkippedTurn represents a turn that was skipped during elimination
type SkippedTurn struct {
    UserID      string    `json:"user_id"`
    Round       int       `json:"round"`
    TurnInRound int       `json:"turn_in_round"`
    SkipType    string    `json:"skip_type"` // "quick_skip", "timeout_skip", "forfeited"
    SkippedAt   time.Time `json:"skipped_at"`
}

// EliminationStatus represents the current state of an elimination session
type EliminationStatus struct {
    SessionID         string        `json:"session_id"`
    CurrentCandidates []string      `json:"current_candidates"`
    CurrentUserTurn   *string       `json:"current_user_turn"`
    IsYourTurn        bool          `json:"is_your_turn"`
    CurrentRound      int           `json:"current_round"`
    TurnTimeRemaining time.Duration `json:"turn_time_remaining"`
    EliminationOrder  []string      `json:"elimination_order"`
    SkippedUsers      []SkippedTurn `json:"skipped_users"`
    CanQuickSkip      bool          `json:"can_quick_skip"`
    QuickSkipsUsed    int           `json:"quick_skips_used"`
    QuickSkipsLimit   int           `json:"quick_skips_limit"`
    IsCatchUpPhase    bool          `json:"is_catch_up_phase"`
}

// DecisionElimination represents an eliminated item in a decision session
type DecisionElimination struct {
    ID           string    `json:"id" db:"id"`
    SessionID    string    `json:"session_id" db:"session_id"`
    UserID       string    `json:"user_id" db:"user_id"`
    ListItemID   string    `json:"list_item_id" db:"list_item_id"`
    RoundNumber  int       `json:"round_number" db:"round_number"`
    EliminatedAt time.Time `json:"eliminated_at" db:"eliminated_at"`
}
```

### Filtering System Types

```go
// FilterItem represents a single filter with its criteria
type FilterItem struct {
    ID          string      `json:"id"`
    Type        string      `json:"type"`        // Filter category
    IsHard      bool        `json:"is_hard"`     // Required vs preferred
    Priority    int         `json:"priority"`    // Execution order (0 = highest)
    Criteria    interface{} `json:"criteria"`    // Type-specific data
    Description string      `json:"description"` // Human-readable description
}

// FilterConfiguration represents a complete filter setup
type FilterConfiguration struct {
    Items  []FilterItem `json:"items"`
    UserID string       `json:"user_id"`
}

// CategoryFilterCriteria for filtering by item categories
type CategoryFilterCriteria struct {
    IncludeCategories []string `json:"include_categories"`
    ExcludeCategories []string `json:"exclude_categories"`
}

// DietaryFilterCriteria for filtering by dietary restrictions
type DietaryFilterCriteria struct {
    RequiredOptions []string `json:"required_options"` // ["vegetarian", "vegan", "gluten_free"]
}

// LocationFilterCriteria for geographic filtering
type LocationFilterCriteria struct {
    CenterLat   float64 `json:"center_lat"`
    CenterLng   float64 `json:"center_lng"`
    MaxDistance float64 `json:"max_distance"` // in miles
}

// RecentActivityFilterCriteria for excluding recently visited items
type RecentActivityFilterCriteria struct {
    ExcludeDays int     `json:"exclude_days"`
    UserID      string  `json:"user_id"`
    TribeID     *string `json:"tribe_id"`
}

// OpeningHoursFilterCriteria for business hours filtering
type OpeningHoursFilterCriteria struct {
    MustBeOpenFor   int     `json:"must_be_open_for"`   // minutes from now
    MustBeOpenUntil *string `json:"must_be_open_until"` // time in user timezone
    UserTimezone    string  `json:"user_timezone"`      // user's timezone
    CheckDate       *int64  `json:"check_date"`         // unix timestamp, optional
}

// TagFilterCriteria for tag-based filtering
type TagFilterCriteria struct {
    RequiredTags []string `json:"required_tags"`
    ExcludedTags []string `json:"excluded_tags"`
}

// FilterResult represents the result of applying filters to an item
type FilterResult struct {
    Item              ListItem           `json:"item"`
    PassedHardFilters bool               `json:"passed_hard_filters"`
    SoftFilterResults []SoftFilterResult `json:"soft_filter_results"`
    ViolationCount    int                `json:"violation_count"`
    PriorityScore     float64            `json:"priority_score"`
}

// SoftFilterResult represents the result of a single soft filter
type SoftFilterResult struct {
    FilterID    string `json:"filter_id"`
    FilterType  string `json:"filter_type"`
    Passed      bool   `json:"passed"`
    Priority    int    `json:"priority"`
    Description string `json:"description"`
}
```

### Activity Tracking Types

```go
// ActivityEntry represents a logged activity for a list item
type ActivityEntry struct {
    ID                string     `json:"id" db:"id"`
    ListItemID        string     `json:"list_item_id" db:"list_item_id"`
    UserID            string     `json:"user_id" db:"user_id"`
    TribeID           *string    `json:"tribe_id" db:"tribe_id"`
    ActivityType      string     `json:"activity_type" db:"activity_type"`         // 'visited', 'watched', 'completed'
    ActivityStatus    string     `json:"activity_status" db:"activity_status"`     // 'confirmed', 'tentative', 'cancelled'
    CompletedAt       time.Time  `json:"completed_at" db:"completed_at"`
    DurationMinutes   *int       `json:"duration_minutes" db:"duration_minutes"`
    Participants      []string   `json:"participants" db:"participants"`           // User IDs who participated
    Notes             *string    `json:"notes" db:"notes"`
    RecordedByUserID  string     `json:"recorded_by_user_id" db:"recorded_by_user_id"`
    DecisionSessionID *string    `json:"decision_session_id" db:"decision_session_id"`
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// LogActivityRequest represents a request to log an activity
type LogActivityRequest struct {
    ListItemID        string     `json:"list_item_id"`
    UserID            string     `json:"user_id"`
    TribeID           *string    `json:"tribe_id"`
    ActivityType      string     `json:"activity_type"`
    ActivityStatus    string     `json:"activity_status"`
    CompletedAt       time.Time  `json:"completed_at"`
    DurationMinutes   *int       `json:"duration_minutes"`
    Participants      []string   `json:"participants"`
    Notes             *string    `json:"notes"`
    RecordedByUserID  string     `json:"recorded_by_user_id"`
    DecisionSessionID *string    `json:"decision_session_id"`
}

// UpdateActivityRequest represents a request to update an activity
type UpdateActivityRequest struct {
    ActivityStatus *string    `json:"activity_status"`
    CompletedAt    *time.Time `json:"completed_at"`
    Participants   []string   `json:"participants"`
    Notes          *string    `json:"notes"`
}
```

### Governance Types

```go
// TribeInvitation represents an invitation to join a tribe
type TribeInvitation struct {
    ID                         string     `json:"id" db:"id"`
    TribeID                    string     `json:"tribe_id" db:"tribe_id"`
    InviterID                  string     `json:"inviter_id" db:"inviter_id"`
    InviteeEmail               string     `json:"invitee_email" db:"invitee_email"`
    InviteeUserID              *string    `json:"invitee_user_id" db:"invitee_user_id"`
    SuggestedTribeDisplayName  *string    `json:"suggested_tribe_display_name" db:"suggested_tribe_display_name"`
    Status                     string     `json:"status" db:"status"` // 'pending', 'accepted_pending_ratification', 'ratified', 'rejected', 'revoked', 'expired'
    InvitedAt                  time.Time  `json:"invited_at" db:"invited_at"`
    AcceptedAt                 *time.Time `json:"accepted_at" db:"accepted_at"`
    ExpiresAt                  time.Time  `json:"expires_at" db:"expires_at"`
}

// TribeInvitationRatification represents a member's vote on an invitation
type TribeInvitationRatification struct {
    ID           string    `json:"id" db:"id"`
    InvitationID string    `json:"invitation_id" db:"invitation_id"`
    MemberID     string    `json:"member_id" db:"member_id"`
    Vote         string    `json:"vote" db:"vote"` // 'approve', 'reject'
    VotedAt      time.Time `json:"voted_at" db:"voted_at"`
}

// MemberRemovalPetition represents a petition to remove a member
type MemberRemovalPetition struct {
    ID           string     `json:"id" db:"id"`
    TribeID      string     `json:"tribe_id" db:"tribe_id"`
    PetitionerID string     `json:"petitioner_id" db:"petitioner_id"`
    TargetUserID string     `json:"target_user_id" db:"target_user_id"`
    Reason       *string    `json:"reason" db:"reason"`
    Status       string     `json:"status" db:"status"` // 'active', 'approved', 'rejected'
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at" db:"resolved_at"`
}

// MemberRemovalVote represents a vote on a member removal petition
type MemberRemovalVote struct {
    ID         string    `json:"id" db:"id"`
    PetitionID string    `json:"petition_id" db:"petition_id"`
    VoterID    string    `json:"voter_id" db:"voter_id"`
    Vote       string    `json:"vote" db:"vote"` // 'approve', 'reject'
    VotedAt    time.Time `json:"voted_at" db:"voted_at"`
}

// TribeDeletionPetition represents a petition to delete a tribe
type TribeDeletionPetition struct {
    ID           string     `json:"id" db:"id"`
    TribeID      string     `json:"tribe_id" db:"tribe_id"`
    PetitionerID string     `json:"petitioner_id" db:"petitioner_id"`
    Reason       *string    `json:"reason" db:"reason"`
    Status       string     `json:"status" db:"status"` // 'active', 'approved', 'rejected'
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at" db:"resolved_at"`
}

// TribeDeletionVote represents a vote on a tribe deletion petition
type TribeDeletionVote struct {
    ID         string    `json:"id" db:"id"`
    PetitionID string    `json:"petition_id" db:"petition_id"`
    VoterID    string    `json:"voter_id" db:"voter_id"`
    Vote       string    `json:"vote" db:"vote"` // 'approve', 'reject'
    VotedAt    time.Time `json:"voted_at" db:"voted_at"`
}

// ListDeletionPetition represents a petition to delete a list
type ListDeletionPetition struct {
    ID                string     `json:"id" db:"id"`
    ListID            string     `json:"list_id" db:"list_id"`
    PetitionerID      string     `json:"petitioner_id" db:"petitioner_id"`
    Reason            *string    `json:"reason" db:"reason"`
    Status            string     `json:"status" db:"status"` // 'pending', 'confirmed', 'cancelled'
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
    ResolvedAt        *time.Time `json:"resolved_at" db:"resolved_at"`
    ResolvedByUserID  *string    `json:"resolved_by_user_id" db:"resolved_by_user_id"`
}

// TribeSettings represents configurable tribe settings
type TribeSettings struct {
    TribeID                 string    `json:"tribe_id" db:"tribe_id"`
    InactivityThresholdDays int       `json:"inactivity_threshold_days" db:"inactivity_threshold_days"`
    CreatedAt               time.Time `json:"created_at" db:"created_at"`
    UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}
```

### Authentication Types

```go
// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    Provider  string `json:"provider"`
    ExpiresAt int64  `json:"exp"`
    IssuedAt  int64  `json:"iat"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
    SecretKey  string        `json:"secret_key"`
    ExpiryTime time.Duration `json:"expiry_time"`
    Issuer     string        `json:"issuer"`
}

// OAuthConfig represents OAuth provider configuration
type OAuthConfig struct {
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    RedirectURL  string `json:"redirect_url"`
}
```

### Shared Types

```go
// ListShare represents a shared list relationship
type ListShare struct {
    ID               string     `json:"id" db:"id"`
    ListID           string     `json:"list_id" db:"list_id"`
    SharedWithUserID *string    `json:"shared_with_user_id" db:"shared_with_user_id"`
    SharedWithTribeID *string   `json:"shared_with_tribe_id" db:"shared_with_tribe_id"`
    PermissionLevel  string     `json:"permission_level" db:"permission_level"`
    SharedByUserID   string     `json:"shared_by_user_id" db:"shared_by_user_id"`
    SharedAt         time.Time  `json:"shared_at" db:"shared_at"`
}

// DecisionSessionList represents the relationship between decision sessions and lists
type DecisionSessionList struct {
    ID        string `json:"id" db:"id"`
    SessionID string `json:"session_id" db:"session_id"`
    ListID    string `json:"list_id" db:"list_id"`
}

// FilterConfigurationSaved represents a saved filter configuration
type FilterConfigurationSaved struct {
    ID            string                 `json:"id" db:"id"`
    UserID        string                 `json:"user_id" db:"user_id"`
    Name          string                 `json:"name" db:"name"`
    IsDefault     bool                   `json:"is_default" db:"is_default"`
    Configuration map[string]interface{} `json:"configuration" db:"configuration"`
    CreatedAt     time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}
```

---

**Note**: All type definitions in this section are authoritative. Other design documents should reference these types rather than redefining them. For implementation examples using these types, see the [implementation-examples](./implementation-examples/) directory.

