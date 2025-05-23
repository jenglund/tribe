# Democratic Tribe Governance System

## Overview

The Tribe app implements a democratic governance system that ensures fair, transparent, and collaborative decision-making for tribe management. The system is built on principles of common ownership, equal participation, and democratic processes with appropriate safeguards against abuse.

## Core Governance Principles

### Common Ownership Model
- **Equal Rights**: All tribe members have equal permissions and access to tribe resources
- **Collective Decision Making**: Important tribe changes require member participation
- **Transparent Processes**: All governance actions are logged and visible to members
- **Democratic Safeguards**: Protection against individual abuse of privileges

### Seniority & Tie-Breaking
- **Senior Member**: Longest-standing active member (by invitation timestamp)
- **Tie-Breaking Role**: Senior member resolves deadlocks and conflicts
- **No Special Privileges**: Senior status is only for conflict resolution
- **Automatic Transfer**: Seniority transfers when current senior member leaves

## Two-Stage Invitation System

### Stage 1: Invitation
Any existing tribe member can invite new people by email:

```typescript
interface InvitationRequest {
  tribeId: string;
  inviterUserId: string;
  inviteeEmail: string;
  suggestedDisplayName?: string;
}
```

**Process:**
1. Member sends invitation with optional suggested display name
2. Invitation expires after 7 days if not accepted
3. Invitee receives email with tribe information and invitation link
4. Multiple invitations to same email are not allowed

### Stage 2: Ratification
Once invitee accepts, existing members vote to approve:

```typescript
interface RatificationProcess {
  invitationId: string;
  requiredVotes: number;        // All existing members
  approvalThreshold: number;    // 100% consensus required
  anyRejectionCausesFailure: boolean; // true
}
```

**Rules:**
- **Unanimous Approval Required**: All existing members must approve
- **Single Rejection Fails**: Any member voting "reject" immediately denies invitation
- **Time Limit**: Ratification process has reasonable time limit
- **Automatic Membership**: Once all approve, member is immediately added

### Implementation Flow

```go
// Stage 1: Create invitation
func (tgs *TribeGovernanceService) InviteToTribe(ctx context.Context, tribeID, inviterID, inviteeEmail string) (*TribeInvitation, error) {
    // Validate inviter membership and tribe capacity
    // Create pending invitation with 7-day expiry
    // Send invitation email to invitee
}

// Stage 2: Accept invitation (moves to ratification)
func (tgs *TribeGovernanceService) AcceptInvitation(ctx context.Context, invitationID, userID string) (*TribeInvitation, error) {
    // Update invitation status to "accepted_pending_ratification"
    // Notify all existing members for ratification votes
}

// Ratification voting
func (tgs *TribeGovernanceService) VoteOnInvitation(ctx context.Context, invitationID, voterID string, approve bool) error {
    // Record vote
    // If any rejection: immediately fail invitation
    // If all approve: complete membership
}
```

## Member Removal Petitions

### Petition Process
Any member can petition for another member's removal:

```go
type MemberRemovalPetition struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    PetitionerID string    `json:"petitioner_id"`
    TargetUserID string    `json:"target_user_id"`
    Reason       string    `json:"reason"`
    Status       string    `json:"status"` // "active", "approved", "rejected"
    CreatedAt    time.Time `json:"created_at"`
}
```

**Rules:**
- **Cannot Self-Petition**: Members cannot petition for their own removal
- **One Active Petition**: Only one active petition per target member
- **Stated Reason Required**: Petitioner must provide explanation
- **Democratic Voting**: All other members vote (target cannot vote)

### Voting Requirements
- **Eligible Voters**: All members except the target
- **Unanimous Approval**: All eligible members must vote to approve
- **Single Rejection Fails**: Any "reject" vote fails the petition
- **Immediate Action**: Approved petitions result in immediate removal

### Safeguards
- **Mass Removal Protection**: System detects and prevents mass removals
- **Senior Override**: Senior member approval required for bulk removals
- **Audit Trail**: All removal actions are permanently logged
- **Appeal Window**: Brief period for clarification before final removal

## Tribe Deletion Consensus

### Deletion Petitions
Any member can petition for complete tribe deletion:

```go
type TribeDeletionPetition struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    PetitionerID string    `json:"petitioner_id"`
    Reason       string    `json:"reason"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
}
```

### 100% Consensus Requirement
- **All Members Must Approve**: No exceptions for tribe deletion
- **Single Rejection Prevents**: Any member can prevent tribe deletion
- **Permanent Action**: No recovery once tribe is deleted
- **Data Cleanup**: All associated data is permanently removed

### Alternative to Deletion
- **Member Departure**: Members can leave without deleting tribe
- **Automatic Cleanup**: Tribe auto-deletes when last member leaves
- **Data Preservation**: Individual data preserved when members leave

## List Governance

### List Deletion Petitions
Simplified process for tribe list management:

```go
type ListDeletionPetition struct {
    ID                 string    `json:"id"`
    ListID             string    `json:"list_id"`
    PetitionerID       string    `json:"petitioner_id"`
    Reason             string    `json:"reason"`
    Status             string    `json:"status"` // "pending", "confirmed", "cancelled"
    ResolvedByUserID   *string   `json:"resolved_by_user_id"`
}
```

**Process:**
1. **Petition Creation**: Any member can petition to delete tribe list
2. **Confirmation Required**: Any other member can confirm or cancel
3. **Single Confirmation**: Only one confirmation needed (not unanimous)
4. **Immediate Action**: Confirmed deletions execute immediately

### List Management Rights
- **Equal Access**: All members can view and edit tribe lists
- **Add Items**: Any member can add items to tribe lists
- **Modify Items**: Any member can edit existing list items
- **Democratic Deletion**: List deletion requires petition + confirmation

## Safety Mechanisms

### Abuse Prevention
Multiple layers of protection against malicious behavior:

```go
// Mass removal detection
func (tgs *TribeGovernanceService) checkMassRemovalProtection(ctx context.Context, tribeID string) error {
    recentRemovals, err := tgs.db.GetRecentRemovals(ctx, tribeID, time.Hour)
    if err != nil {
        return err
    }
    
    memberCount, err := tgs.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    // If more than half removed in last hour, require senior approval
    if len(recentRemovals) >= memberCount/2 {
        return errors.New("mass removal detected - senior approval required")
    }
    
    return nil
}
```

### Activity Monitoring
- **Action Logging**: All governance actions permanently recorded
- **Pattern Detection**: Unusual voting patterns flagged for review
- **Rate Limiting**: Prevents rapid-fire malicious actions
- **Audit Trail**: Complete history available for dispute resolution

### Conflict Resolution
When normal democratic processes fail:

1. **Senior Member Authority**: Senior member makes final decision
2. **Documented Resolution**: All conflict resolutions logged with reasoning
3. **Appeal Process**: Framework for challenging senior decisions
4. **Community Standards**: Clear guidelines for acceptable behavior

## Database Schema for Governance

### Invitation Tables
```sql
-- Two-stage invitation tracking
CREATE TABLE tribe_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id),
    invitee_email VARCHAR(255) NOT NULL,
    suggested_tribe_display_name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    invited_at TIMESTAMPTZ DEFAULT NOW(),
    accepted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '7 days',
    UNIQUE(tribe_id, invitee_email)
);

-- Ratification vote tracking
CREATE TABLE tribe_invitation_ratifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invitation_id UUID NOT NULL REFERENCES tribe_invitations(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES users(id),
    vote VARCHAR(50) NOT NULL, -- 'approve', 'reject'
    voted_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(invitation_id, member_id)
);
```

### Petition Tables
```sql
-- Member removal petitions
CREATE TABLE member_removal_petitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tribe_id UUID NOT NULL REFERENCES tribes(id) ON DELETE CASCADE,
    petitioner_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID NOT NULL REFERENCES users(id),
    reason TEXT,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    UNIQUE(tribe_id, target_user_id)
);

-- Removal petition votes
CREATE TABLE member_removal_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    petition_id UUID NOT NULL REFERENCES member_removal_petitions(id) ON DELETE CASCADE,
    voter_id UUID NOT NULL REFERENCES users(id),
    vote VARCHAR(50) NOT NULL, -- 'approve', 'reject'
    voted_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(petition_id, voter_id)
);
```

## Frontend Governance Interface

### Invitation Management
```typescript
interface InvitationManagementProps {
  tribe: Tribe;
  pendingInvitations: TribeInvitation[];
  onInviteMember: (email: string, displayName?: string) => void;
  onVoteOnInvitation: (invitationId: string, approve: boolean) => void;
}

interface InvitationStatusDisplay {
  invitation: TribeInvitation;
  ratificationVotes: RatificationVote[];
  canVote: boolean;
  status: 'pending' | 'ratifying' | 'approved' | 'rejected' | 'expired';
}
```

### Petition Management
```typescript
interface PetitionInterface {
  activePetitions: Array<MemberRemovalPetition | TribeDeletionPetition>;
  canCreatePetition: boolean;
  onCreatePetition: (type: PetitionType, targetId: string, reason: string) => void;
  onVoteOnPetition: (petitionId: string, approve: boolean) => void;
}
```

### Governance Dashboard
- **Pending Actions**: All items requiring user's vote
- **Recent Activity**: History of governance actions
- **Member Status**: Current members with roles and seniority
- **Petition Status**: Progress of active petitions

## Error Handling & Edge Cases

### Network Failures
- **Optimistic Updates**: UI updates immediately with rollback on failure
- **Retry Logic**: Automatic retry for transient network issues
- **State Synchronization**: Ensure all clients see consistent governance state
- **Conflict Resolution**: Handle simultaneous votes gracefully

### Edge Cases
- **Simultaneous Petitions**: Prevent conflicting petitions for same target
- **Member Departure During Process**: Handle gracefully when participants leave
- **Invitation Expiry**: Clean up expired invitations automatically
- **Senior Member Removal**: Handle seniority transfer correctly

### Data Integrity
- **Atomic Operations**: Governance actions are all-or-nothing
- **Consistency Checks**: Validate system state after each action
- **Audit Verification**: Ensure audit trail accuracy and completeness
- **Recovery Procedures**: Handle partial failures gracefully

## Security Considerations

### Authentication & Authorization
- **Session Validation**: Verify user session for all governance actions
- **Permission Checks**: Confirm user can perform requested action
- **Rate Limiting**: Prevent spam voting and petition creation
- **Input Validation**: Sanitize all user inputs and reasons

### Data Protection
- **Privacy Controls**: Respect user privacy in governance processes
- **Data Retention**: Clear policies for governance history retention
- **Audit Security**: Protect audit logs from tampering
- **External Exposure**: Limit exposure of internal governance details

## Future Enhancements

### Advanced Governance Features
- **Delegation**: Allow members to delegate voting to others
- **Weighted Voting**: Consider tenure or contribution in voting weight
- **Appeal Process**: Formal process for challenging governance decisions
- **Mediation**: Third-party mediation for complex disputes

### Integration Features
- **External Notifications**: Email/SMS notifications for governance actions
- **Calendar Integration**: Schedule governance meetings and deadlines
- **Documentation**: Link governance decisions to supporting documents
- **Metrics**: Track governance health and member engagement

### Automation
- **Smart Defaults**: AI-suggested governance parameters
- **Automated Enforcement**: Automatic enforcement of governance rules
- **Pattern Recognition**: Detect and flag unusual governance patterns
- **Workflow Optimization**: Streamline common governance processes

---

*This document defines the democratic governance system for tribe management. See [Database Schema](./database-schema.md) for data structure and [Tribe Management User Stories](./user-stories/tribes.md) for user experience requirements.* 