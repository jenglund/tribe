# TribeLite v3 - Open Questions & Design Issues

This document tracks unresolved questions, design conflicts, and ambiguities that need to be addressed before implementation begins. Each item should be resolved and checked off before proceeding with development.

## Core Functionality Questions

### User Profile Management
- [ ] **Profile Deletion**: What happens when someone deletes a user profile? Do we soft-delete and preserve activity history, or hard delete and lose all data?
- [ ] **Duplicate Profiles**: How do we prevent/handle accidentally creating duplicate users with the same name?
- [ ] **Profile Merging**: If someone creates duplicate profiles, how can they merge the data (lists, activities, preferences)?
- [ ] **Avatar Storage**: Where do we store avatar images? Local filesystem, database BLOB, or external service?
- [ ] **Profile Limits**: Should there be a limit on the number of user profiles to prevent abuse/clutter?

### Concurrent Usage & Data Consistency
- [ ] **Simultaneous Editing**: What happens when multiple people edit the same list item simultaneously?
- [ ] **List Creation Conflicts**: How do we handle multiple people creating lists with the same name at the same time?
- [ ] **Decision Conflicts**: What if two people try to make decisions from the same list simultaneously?
- [ ] **Real-time Updates**: Do we need WebSocket updates when someone else modifies a list you're viewing?
- [ ] **Optimistic Updates**: Should the UI update immediately and handle conflicts, or wait for server confirmation?

### Data Ownership & Permissions
- [ ] **List Deletion Rights**: Who can delete personal lists? Only the creator, or anyone?
- [ ] **Community List Management**: Who can delete community lists? First creator, anyone, or require consensus?
- [ ] **Activity Modification**: Can anyone edit/delete activity records, or only participants?
- [ ] **Bulk Operations**: Should there be safeguards against accidental bulk deletions?
- [ ] **Data Export Rights**: Can anyone export all data, or only data they created/participated in?

## Technical Implementation Questions

### Database Design
- [ ] **Primary Key Strategy**: UUIDs or auto-incrementing integers? UUIDs are better for distributed systems but we're local-only.
- [ ] **Indexing Strategy**: What indexes do we need for performance with ~10 users and ~100 lists?
- [ ] **JSON vs Relational**: Store preferences and metadata as JSONB or normalize into separate tables?
- [ ] **Migration Strategy**: How do we handle database schema changes in a self-hosted environment?
- [ ] **Foreign Key Cascades**: What should happen when a user is deleted? Cascade delete their data or preserve with null references?

### API Design
- [ ] **REST vs GraphQL**: Should we use REST for simplicity or GraphQL for efficiency?
- [ ] **Error Handling**: What error codes and messages should we return for various scenarios?
- [ ] **Rate Limiting**: Do we need rate limiting for a local network app?
- [ ] **API Versioning**: Do we need API versioning for a self-hosted app, or can we break changes?
- [ ] **Bulk Operations**: Should we support bulk create/update/delete operations for efficiency?

### Frontend Architecture
- [ ] **State Management**: React Context sufficient, or do we need Redux/Zustand for complexity?
- [ ] **Offline Capabilities**: What functionality should work without network connectivity?
- [ ] **Data Caching**: How long should we cache data locally? How do we invalidate stale cache?
- [ ] **Mobile Safari PWA**: Are there specific iOS Safari limitations we need to work around?
- [ ] **Touch Gestures**: What swipe/touch gestures should we implement for mobile UX?

## User Experience Questions

### Decision Making Flow
- [ ] **Elimination UI**: Should eliminations be swipe-based, tap-based, or checkbox-based?
- [ ] **Undo Functionality**: Can users undo eliminations during the decision process?
- [ ] **Decision History**: Should we show who made the decision and what eliminations they made?
- [ ] **Quick Decisions**: Should there be a "random pick" option that skips the elimination process?
- [ ] **Decision Sharing**: How do we communicate decision results to other users? In-app notifications, email, etc.?

### List Management UX
- [ ] **List Discovery**: How do users find community lists they might be interested in?
- [ ] **Import Workflows**: What formats should we support for importing lists? CSV, JSON, text?
- [ ] **Drag & Drop**: Should users be able to reorder lists and list items via drag and drop?
- [ ] **Search Functionality**: Do we need search across lists, items, or both?
- [ ] **List Templates**: Should we provide pre-made list templates for common use cases?

### Mobile Experience
- [ ] **App Icon**: Should we generate a PWA app icon that appears on home screens?
- [ ] **Push Notifications**: Can we send notifications about planned activities without a notification service?
- [ ] **Offline Indicators**: How do we show users when they're offline or have stale data?
- [ ] **Location Permissions**: How do we handle location permission requests for filtering?
- [ ] **Camera Integration**: Should users be able to take photos of activities directly in the app?

## Deployment & Operations Questions

### Docker Configuration
- [ ] **Environment Variables**: What configuration should be environment variables vs config files?
- [ ] **Port Configuration**: Should we default to port 3000, or make it configurable?
- [ ] **Volume Mounts**: What directories need to be mounted for persistence (database, uploads, logs)?
- [ ] **Health Checks**: What health check endpoints should we provide for Docker?
- [ ] **Resource Limits**: What are reasonable memory/CPU limits for typical home hardware?

### Network & Discovery
- [ ] **IP Address Detection**: How do we reliably detect and display the local network IP address?
- [ ] **mDNS Implementation**: Should we implement Bonjour/mDNS for "tribelite.local" access?
- [ ] **HTTPS Support**: Should we support HTTPS with self-signed certificates for security?
- [ ] **Multiple Network Interfaces**: How do we handle devices with multiple network interfaces (WiFi + Ethernet)?
- [ ] **VPN Compatibility**: What happens when users are connected to VPNs?

### Data Management
- [ ] **Backup Strategy**: Should we provide built-in backup functionality or rely on Docker volume backups?
- [ ] **Data Export**: What format should full data exports use? JSON, SQL dump, CSV?
- [ ] **Log Management**: What should we log, and how do we prevent log files from growing too large?
- [ ] **Database Maintenance**: Do we need automated vacuum/analyze jobs for PostgreSQL?
- [ ] **Upgrade Process**: How do users upgrade to new versions without losing data?

## Performance & Scalability Questions

### Hardware Requirements
- [ ] **Minimum Specs**: What are the minimum hardware requirements for smooth operation?
- [ ] **User Limits**: How many concurrent users can the system handle realistically?
- [ ] **Data Limits**: How many lists/items can the system handle before performance degrades?
- [ ] **Memory Usage**: What's the expected memory footprint for typical usage?
- [ ] **Storage Growth**: How quickly will storage requirements grow with normal usage?

### Optimization Strategies
- [ ] **Database Queries**: Which queries need optimization for larger datasets?
- [ ] **Frontend Bundling**: Should we code-split the frontend for faster initial loads?
- [ ] **Image Optimization**: How do we handle avatar images and activity photos efficiently?
- [ ] **Caching Strategy**: What should we cache and for how long?
- [ ] **Lazy Loading**: What data should be lazy-loaded vs loaded upfront?

## Security & Privacy Questions

### Data Protection
- [ ] **Local Network Security**: Are there security considerations for local network access?
- [ ] **XSS Protection**: What XSS protections do we need without authentication?
- [ ] **CSRF Protection**: Do we need CSRF tokens for state-changing operations?
- [ ] **Input Validation**: What server-side validation do we need to prevent data corruption?
- [ ] **File Upload Security**: If we allow file uploads, what validation/sanitization is needed?

### Privacy Considerations
- [ ] **Activity Tracking**: Should users be able to hide their activity from others?
- [ ] **Data Retention**: Should we automatically clean up old activity data?
- [ ] **Profile Privacy**: Should users be able to mark their profiles as private?
- [ ] **Anonymous Usage**: Should there be an option for anonymous/guest profiles?
- [ ] **Data Sharing**: What controls should users have over sharing their data?

## Integration & Extensibility Questions

### Future Expansion Paths
- [ ] **API Access**: Should we provide an API for third-party integrations?
- [ ] **Plugin System**: Do we want to support plugins or extensions?
- [ ] **External Services**: How would we add external API integrations (weather, maps, etc.) later?
- [ ] **Multi-Instance**: Could users run multiple TribeLite instances and sync between them?
- [ ] **Cloud Migration**: What would it take to migrate from local to cloud hosting later?

### Import/Export Features
- [ ] **Google Maps Integration**: How would we implement Google Maps "Want to Go" list imports?
- [ ] **Calendar Integration**: Should we be able to export planned activities to calendars?
- [ ] **Social Sharing**: Should users be able to share lists or activities externally?
- [ ] **Backup Integration**: Should we integrate with cloud storage for automated backups?
- [ ] **Migration Tools**: How do users migrate data between TribeLite instances?

## Testing & Quality Assurance Questions

### Testing Strategy
- [ ] **Test Environment**: How do we test multi-user scenarios without actual multiple users?
- [ ] **Database Testing**: Should we use an in-memory database for tests or full PostgreSQL?
- [ ] **Mobile Testing**: How do we test mobile experience during development?
- [ ] **Network Testing**: How do we test different network configurations and scenarios?
- [ ] **Performance Testing**: What performance benchmarks should we establish?

### Error Handling
- [ ] **Database Connectivity**: How do we handle database connection failures gracefully?
- [ ] **Network Issues**: What happens when mobile users lose network connectivity?
- [ ] **Concurrent Modifications**: How do we handle and display data conflicts to users?
- [ ] **Invalid Data**: How do we handle corrupted or invalid data in the database?
- [ ] **Resource Exhaustion**: What happens when the system runs out of disk space or memory?

---

## Resolution Process

Each question should be:
1. **Researched**: Investigate technical constraints and user experience implications
2. **Decided**: Make a clear decision with reasoning documented
3. **Validated**: Test the decision with prototypes or research if necessary
4. **Documented**: Update relevant design documents with the decision
5. **Checked Off**: Mark as resolved in this document

Questions marked as **[CRITICAL]** must be resolved before development begins.
Questions marked as **[PHASE2]** can be deferred to later phases.
Questions marked as **[NICE-TO-HAVE]** are optional enhancements. 