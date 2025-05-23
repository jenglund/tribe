# Database Schema

**Related Documents:**
- [Architecture](./architecture.md) - System architecture context
- [API Design](./api-design.md) - How schema maps to API
- [Governance](./governance.md) - Governance tables usage

## Core Design Decisions

- **Primary Keys**: UUIDs for all entities (better for distributed systems and external sync)
- **Naming**: snake_case for database columns (PostgreSQL convention)
- **Table Names**: Plural form (users, tribes, lists, etc.)
- **Soft Deletion**: Deferred for MVP - may use hard deletes with confirmation prompts
- **Timestamps**: All tables include created_at, updated_at with timezone support

## Core Tables

### Users Table
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

### Tribes Table
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

### Tribe Memberships Table
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

## List Management Tables

### Lists Table
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

### List Items Table
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

### Business Info JSON Examples
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

### List Shares Table
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

## Activity Tracking Tables

### Activity History Table
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

## Decision Making Tables

### Decision Sessions Table
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

### Decision Session Lists Table
```sql
CREATE TABLE decision_session_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES decision_sessions(id) ON DELETE CASCADE,
    list_id UUID NOT NULL REFERENCES lists(id),
    UNIQUE(session_id, list_id)
);
```

### Decision Eliminations Table
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

## Governance Tables

*See [Governance](./governance.md) for detailed governance system documentation.*

### Tribe Invitations Table
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

### Tribe Invitation Ratifications Table
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

### Member Removal Petitions Table
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

### Other Governance Tables
- `member_removal_votes` - Votes on member removal petitions
- `tribe_deletion_petitions` - Petitions to delete entire tribes
- `tribe_deletion_votes` - Votes on tribe deletion
- `list_deletion_petitions` - Petitions to delete tribe lists
- `tribe_settings` - Configurable tribe settings

*See [Governance](./governance.md) for complete governance table definitions.*

## Performance Indexes

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

-- Geospatial index for location queries (if PostGIS is available)
-- CREATE INDEX idx_list_items_location ON list_items USING GIST((location->>'lat')::float, (location->>'lng')::float);
```

## Database Functions and Triggers

### Automatic Timestamp Updates
```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply to all tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tribes_updated_at BEFORE UPDATE ON tribes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- ... (apply to all relevant tables)
```

### Data Validation Examples
```sql
-- Ensure tribe max_members is within bounds
ALTER TABLE tribes ADD CONSTRAINT tribes_max_members_check CHECK (max_members >= 1 AND max_members <= 8);

-- Ensure activity completed_at is reasonable
ALTER TABLE activity_history ADD CONSTRAINT activity_completed_at_check 
CHECK (completed_at >= '2020-01-01'::timestamptz AND completed_at <= NOW() + INTERVAL '1 year');

-- Ensure decision algorithm parameters are valid
ALTER TABLE decision_sessions ADD CONSTRAINT decision_params_check 
CHECK ((algorithm_params->>'k')::int >= 1 AND (algorithm_params->>'m')::int >= 1);
```

## Migration Strategy

### Initial Migration (001_initial_schema.sql)
- Create all core tables (users, tribes, tribe_memberships)
- Create essential indexes
- Add basic constraints

### Feature Migrations
- 002_lists_and_items.sql - List management tables
- 003_activity_tracking.sql - Activity history
- 004_decision_making.sql - Decision session tables
- 005_governance.sql - Democratic governance tables
- 006_performance_indexes.sql - Additional performance indexes

### Data Migration Considerations
- UUID generation for existing data
- Timezone migration for timestamp columns
- JSON migration for flexible data fields

---

*For API usage of this schema, see [API Design](./api-design.md)*
*For governance table usage, see [Governance](./governance.md)* 