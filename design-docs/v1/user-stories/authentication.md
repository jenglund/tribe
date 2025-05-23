# User Stories: Authentication & User Management

**Related Documents:**
- [Authentication](../authentication.md) - Technical implementation
- [Database Schema](../database-schema.md) - User table structure
- [API Design](../api-design.md) - Auth-related API endpoints

## Authentication & User Management (US-AUTH)

### Primary Authentication Flow

**US-AUTH-001**: As a new user, I want to sign up using Google OAuth so I can quickly access the app without passwords

**Acceptance Criteria:**
- User can click "Sign in with Google" button
- OAuth flow redirects to Google and back to app
- User profile is created automatically from Google data
- User is logged in and redirected to dashboard
- JWT token is stored securely for subsequent requests

**Technical Notes:**
- Implement Google OAuth 2.0 flow
- Extract email, name, and profile picture from Google
- Generate JWT with 7-day expiry
- Store token in localStorage with sessionStorage fallback

---

**US-AUTH-002**: As a user, I want to update my profile (name, avatar, preferences) so others can identify me

**Acceptance Criteria:**
- User can access profile settings page
- User can update display name (shown globally)
- User can upload/change avatar image
- User can set timezone preference
- User can set dietary preferences (vegetarian, vegan, gluten-free, etc.)
- User can set default location preferences
- Changes are saved immediately with confirmation
- Other users see updated information in shared contexts

**Technical Notes:**
- Profile update via GraphQL mutation
- Image upload via REST endpoint
- Optimistic UI updates
- Validation for required fields

---

**US-AUTH-003**: As a user, I want to set dietary preferences that automatically apply to my filters

**Acceptance Criteria:**
- User can select multiple dietary preferences from predefined list
- User can add custom dietary requirements
- Preferences automatically apply when filtering restaurant lists
- User can override dietary filters on a per-decision basis
- Preferences are saved to user profile permanently

**Technical Notes:**
- Store as JSONB array in user table
- Apply as default filter criteria in decision sessions
- Allow per-session overrides

---

**US-AUTH-004**: As a user, I want to delete my account and all associated data for privacy

**Acceptance Criteria:**
- User can request account deletion from settings
- System shows what data will be deleted
- User must confirm deletion with password/re-authentication
- All personal data is permanently removed
- User is logged out and redirected to homepage
- Shared lists/tribes handle user departure gracefully

**Technical Notes:**
- Implement cascading deletes for user data
- Handle tribe membership cleanup
- Preserve anonymized activity history if needed
- Send deletion confirmation email

### Development/Testing Authentication

**US-AUTH-005**: As a developer, I want to use test login during development so I can test features without OAuth setup

**Acceptance Criteria:**
- Development environment shows "Dev Login" option
- Developer can enter any email and name to create test user
- Test users function identically to OAuth users
- Dev login is disabled in production environment
- Test accounts are clearly marked in the database

**Technical Notes:**
- Conditional dev login endpoint
- Environment variable to enable/disable
- Different OAuth provider type in database

### Session Management

**US-AUTH-006**: As a user, I want my login session to persist for a reasonable time so I don't have to re-authenticate frequently

**Acceptance Criteria:**
- User stays logged in for 7 days by default
- User session works across multiple browser tabs
- User can log in from multiple devices simultaneously
- Token automatically refreshes if user is active
- User gets warning before session expires

**Technical Notes:**
- 7-day JWT expiry for MVP
- Future enhancement: refresh token rotation
- Client-side token expiry checking

---

**US-AUTH-007**: As a user, I want to log out securely so my account is protected on shared devices

**Acceptance Criteria:**
- User can log out from any page via account menu
- Logout clears all local session data
- User is redirected to login page
- Subsequent requests are rejected until re-authentication
- "Logout all devices" option for security

**Technical Notes:**
- Clear localStorage and sessionStorage
- Invalidate token server-side (future enhancement)
- Redirect to auth page

### Profile Display and Management

**US-AUTH-008**: As a user, I want different display names in different tribes so I can have appropriate identity in each context

**Acceptance Criteria:**
- User has a global default display name
- User can set tribe-specific display names during invitation acceptance
- Tribe members see tribe-specific names in that context
- User can update tribe-specific names from tribe settings
- Global name is used as fallback if no tribe-specific name

**Technical Notes:**
- Store tribe_display_name in tribe_memberships table
- API returns appropriate name based on context
- Default to global display_name if tribe name is null

---

**US-AUTH-009**: As a user, I want to see when I last logged in for security awareness

**Acceptance Criteria:**
- User profile shows last login timestamp
- Login history shows recent login locations/devices (future)
- User gets email notification for new device logins (future)
- User can see active sessions (future enhancement)

**Technical Notes:**
- Track last_login_at in user table
- Update on each successful authentication
- Display in user-friendly format with timezone

### Privacy and Security

**US-AUTH-010**: As a user, I want my email address to be private by default so I control who can contact me

**Acceptance Criteria:**
- Email is not visible to other users by default
- User can choose to share email with tribe members
- Email is only used for system notifications and invitations
- User can opt out of non-essential emails

**Technical Notes:**
- Email visibility settings in user preferences
- API excludes email from user queries unless permission

---

**US-AUTH-011**: As a user, I want to verify my email address so others know my account is legitimate

**Acceptance Criteria:**
- New users receive email verification link
- Unverified users see verification prompt
- Verified status is shown in user profiles
- System can require verification for certain actions
- User can resend verification email

**Technical Notes:**
- email_verified boolean in user table
- Verification token system
- Email template for verification

### Error Handling and Recovery

**US-AUTH-012**: As a user, I want clear error messages when authentication fails so I can resolve issues

**Acceptance Criteria:**
- OAuth failures show user-friendly error messages
- Network issues show "try again" options
- Invalid tokens redirect to login with explanation
- Rate limiting shows clear retry timeframes
- Support contact info for unresolved issues

**Technical Notes:**
- Consistent error response format
- Client-side error handling and display
- Logging for debugging without exposing sensitive data

---

**US-AUTH-013**: As a user, I want to recover my account if I lose access to my Google account

**Acceptance Criteria:**
- User can contact support with account details
- Admin can verify identity through alternative means
- Account can be migrated to new email/OAuth provider
- User retains all tribes, lists, and history
- Process is documented and secure

**Technical Notes:**
- Admin tools for account recovery
- Identity verification process
- Email/OAuth provider migration capability

## Implementation Priority

### Phase 1 (MVP)
- US-AUTH-001: Google OAuth login
- US-AUTH-002: Basic profile management
- US-AUTH-006: Session persistence
- US-AUTH-007: Secure logout

### Phase 2 (Core Features)
- US-AUTH-003: Dietary preferences
- US-AUTH-008: Tribe-specific display names
- US-AUTH-012: Error handling

### Phase 3 (Enhancements)
- US-AUTH-004: Account deletion
- US-AUTH-009: Login history
- US-AUTH-011: Email verification

### Phase 4 (Advanced)
- US-AUTH-010: Privacy controls
- US-AUTH-013: Account recovery
- Enhanced session management

---

*For technical implementation, see [Authentication](../authentication.md)* 