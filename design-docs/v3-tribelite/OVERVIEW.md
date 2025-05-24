# TribeLite v3 - Design Overview

## Vision Statement

TribeLite is a self-hosted, local network collaborative decision-making app designed for households and small groups who share physical space. It eliminates authentication complexity in favor of a trusted local environment where anyone on the network can create and manage lists, make decisions, and track activities together.

## Core Design Principles

### Trust-Based Local Network Model
- **No Authentication**: Users select/create profiles like streaming service accounts
- **Shared Network = Shared Trust**: If you can access the app, you can do anything
- **Self-Hosted Simplicity**: Single Docker Compose deployment, no external dependencies
- **Mobile-First, Desktop-Friendly**: Optimized for phones but great on all devices

### Collaborative Simplicity
- **Single Tribe Model**: Everyone on the instance is part of one community
- **Content Curation over Security**: Sharing focuses on organization, not permissions  
- **Ephemeral Decisions**: Decision-making is client-side, results are saved as activities
- **Snooze Don't Delete**: Hide content you don't care about rather than complex permissions

## Target Use Cases

### Primary: Household Decision Making
- **Couples deciding on dinner**: "What should we cook tonight?"
- **Family weekend planning**: "What movie should we watch?"
- **Roommate activity coordination**: "Where should we go hiking this weekend?"

### Secondary: Small Group Coordination  
- **Friend group vacation planning**: Shared lists of destinations and activities
- **Game night organization**: Lists of games, snacks, and participants
- **Book club selections**: Curated reading lists with member preferences

## System Architecture

### Deployment Model
```
Docker Host (Home Server/Laptop)
├── tribelite-app (Go backend + React frontend)
├── postgres (Database)
└── Optional: redis (Caching for performance)

Local Network Access: http://192.168.1.X:3000
```

### Core Components

#### User Profiles (No Authentication)
- **Profile Selection**: Choose from existing users or create new profile
- **Persistent Preferences**: Dietary restrictions, location defaults, decision preferences
- **Activity History**: Personal tracking of completed activities
- **List Subscriptions**: Which lists they care about seeing

#### Lists & Items
- **Personal Lists**: Created by individual users
- **Community Lists**: Visible to all users, editable by anyone
- **Sharing Model**: Users can "subscribe" to lists they care about
- **Snooze Function**: Hide lists from view without deleting

#### Simplified Decision Making
- **Client-Side Elimination**: One person makes all eliminations in browser
- **Configurable K+M Values**: Per-list defaults, global fallbacks
- **Participant Selection**: Choose who the decision is for (affects filtering)
- **Result Tracking**: Save decision outcomes as pending/completed activities

#### Activity Tracking
- **Visit History**: When, where, who participated
- **Activity Status**: Planned → Confirmed → Completed
- **Rating & Notes**: Simple feedback for future filtering

## Data Model Overview

### Core Tables
```sql
-- User profiles (no auth required)
users (id, name, avatar, preferences, created_at)

-- Lists with ownership and sharing
lists (id, name, created_by, is_community, category, k_default, m_default)

-- List items with rich metadata
list_items (id, list_id, name, description, location, metadata, added_by)

-- User relationships to lists
list_subscriptions (user_id, list_id, is_snoozed, subscribed_at)

-- Activity tracking
activities (id, item_id, participants, status, planned_date, completed_date, notes)
```

### Notable Simplifications
- **No authentication tables**: No users.password, oauth_providers, sessions, etc.
- **No tribe management**: No tribes, memberships, invitations, governance
- **No decision sessions**: Decisions are ephemeral, only results are stored
- **No complex permissions**: Everyone can read/write everything

## User Experience Flow

### First Time Setup
1. User navigates to local IP address in browser
2. App shows "Select Profile" with existing users + "Create New User" option
3. User creates profile with name, optional avatar, basic preferences
4. Lands on dashboard showing subscribed lists + community lists

### Daily Usage: Making a Decision
1. User selects "Make Decision" from any list
2. Chooses participants (for filtering purposes)
3. Applies filters (dietary, location, recent activity, etc.)
4. Performs K eliminations per participant (all done by one person)
5. System randomly selects from remaining M options
6. User confirms result and saves as planned activity

### List Management
1. Anyone can create personal or community lists
2. Users subscribe to lists they care about
3. Snooze lists to hide from main view
4. Add/edit items with location, tags, dietary info

## Technical Implementation

### Backend (Go + Gin)
- **REST API**: Simple CRUD operations, no complex auth middleware
- **Local Network Binding**: Listen on 0.0.0.0 to accept LAN connections
- **IP Discovery**: Log local network IP addresses on startup
- **Simple Validation**: Prevent data corruption, but trust users

### Frontend (React + TypeScript)
- **Mobile-First Design**: Touch-friendly interfaces, responsive layout
- **Profile Selection**: Persistent localStorage for "current user"
- **Offline-Capable**: Cache data for better mobile experience
- **Simple State Management**: React Context for current user, minimal global state

### Database (PostgreSQL)
- **Simple Schema**: Fewer tables, less complexity
- **JSON Preferences**: Store user preferences as JSONB
- **Activity Logging**: Track what people actually do
- **No Soft Deletes**: Trust users, provide clear confirmations

## Key Features

### Decision Making Engine
- **Flexible K+M Configuration**: Per-list defaults, global environment variables
- **Intelligent Scaling**: Automatically adjust K/M when insufficient options
- **Rich Filtering**: Location, dietary, recency, tags, custom criteria
- **Client-Side Processing**: No server-side session management

### Content Organization
- **List Categories**: Restaurants, Movies, Activities, Vacations, etc.
- **Smart Defaults**: Category-specific K/M values (Restaurants: K=2,M=1; Vacations: K=3,M=3)
- **Subscription Model**: See only lists you care about
- **Community Discovery**: Find lists other users have created

### Activity Tracking
- **Planning Integration**: Decision results become planned activities
- **Participation Tracking**: Who was involved in each activity
- **History-Based Filtering**: "Not visited in last 60 days"
- **Simple Rating**: Thumbs up/down for future reference

## Development Priorities

### Phase 1: Core Functionality (Weeks 1-4)
- [ ] User profile selection (no auth)
- [ ] Basic list and item CRUD
- [ ] Simple decision making (K+M elimination)
- [ ] Activity tracking and history

### Phase 2: Polish & Usability (Weeks 5-8) 
- [ ] Mobile-optimized UI
- [ ] Advanced filtering system
- [ ] List subscription and snoozing
- [ ] Docker deployment configuration

### Phase 3: Enhancement (Weeks 9-12)
- [ ] Rich metadata for locations and businesses
- [ ] Import/export functionality
- [ ] Performance optimization
- [ ] Documentation and setup guides

## Innovation Opportunities

### Local Network Discovery
- **mDNS/Bonjour**: Advertise service as "TribeLite.local"
- **QR Code Generation**: Display QR code with local URL for easy mobile access
- **Network Scanner**: Help users find the service on their network

### Mobile-First Features
- **PWA Capabilities**: Install as app on mobile devices
- **Offline Decision Making**: Cache lists for offline elimination
- **Location Services**: Auto-detect user location for filtering
- **Camera Integration**: Photo capture for activity documentation

### Smart Defaults
- **Time-Based Filtering**: Auto-filter by business hours
- **Weather Integration**: Filter outdoor activities by forecast
- **Learning Preferences**: Suggest K/M values based on usage patterns
- **Quick Actions**: One-tap common decisions ("Quick dinner choice")

---

## Questions Identified for OPEN-QUESTIONS.md

While designing this overview, several unresolved issues became apparent:

1. **Concurrent Usage**: What happens when multiple people use the app simultaneously?
2. **Data Consistency**: How do we prevent conflicts when multiple people edit the same list?
3. **Profile Management**: How do we handle profile deletion or merging?
4. **Activity Conflicts**: What if multiple people plan the same activity simultaneously?
5. **List Ownership**: Who can delete community lists? What about personal lists?
6. **Data Recovery**: How do we handle accidental deletions in a no-auth environment?
7. **Performance**: How many users/lists can the system handle on typical home hardware?
8. **Mobile Safari**: Are there any PWA limitations or iOS-specific considerations?
9. **Network Configuration**: How do we handle different router configurations, VPNs, etc.?
10. **Backup Strategy**: How should users backup their data in a self-hosted environment?

These questions need resolution before implementation begins. 