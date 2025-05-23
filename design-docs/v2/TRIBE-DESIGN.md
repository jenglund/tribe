# Tribe Design in Tribe

## Tribe Structure and Functional Principles

### Common Ownership Tribe Management

```go
// Tribe service implementing common ownership model
type TribeService struct {
    db repository.Database
}

// Common ownership validation - any member can perform tribe operations
func (ts *TribeService) ValidateTribeMembership(ctx context.Context, userID, tribeID string) error {
    isMember, err := ts.db.IsUserTribeMember(ctx, userID, tribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    return nil
}

// Get senior member (earliest invite among active members) for tie-breaking scenarios
func (ts *TribeService) GetSeniorMember(ctx context.Context, tribeID string) (*User, error) {
    // Use database function to get senior member based on earliest invite timestamp
    seniorUserID, err := ts.db.GetTribeSeniorMember(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    return ts.db.GetUser(ctx, seniorUserID)
}

// Get tribe creator (user who invited themselves)
func (ts *TribeService) GetTribeCreator(ctx context.Context, tribeID string) (*User, error) {
    // Use database function to find self-invited member
    creatorUserID, err := ts.db.GetTribeCreator(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    // Creator might have left the tribe, so this could return nil
    if creatorUserID == "" {
        return nil, nil // No creator found (they left)
    }
    
    return ts.db.GetUser(ctx, creatorUserID)
}

// Create tribe with proper self-invitation
func (ts *TribeService) CreateTribe(ctx context.Context, creatorID string, name, description string) (*Tribe, error) {
    // Create the tribe first
    tribe := &Tribe{
        ID:          generateUUID(),
        Name:        name,
        Description: description,
        MaxMembers:  8,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    if err := ts.db.CreateTribe(ctx, tribe); err != nil {
        return nil, err
    }
    
    // Create membership with self-invitation pattern
    inviteTime := time.Now()
    membership := &TribeMembership{
        ID:               generateUUID(),
        TribeID:          tribe.ID,
        UserID:           creatorID,
        InvitedAt:        inviteTime,
        InvitedByUserID:  creatorID, // Self-invited (creator pattern)
        JoinedAt:         inviteTime, // Joined immediately
        IsActive:         true,
    }
    
    if err := ts.db.CreateTribeMembership(ctx, membership); err != nil {
        // Rollback tribe creation if membership fails
        ts.db.DeleteTribe(ctx, tribe.ID)
        return nil, err
    }
    
    return tribe, nil
}

// Any member can invite new members
func (ts *TribeService) InviteMember(ctx context.Context, tribeID, inviterID, inviteeEmail string) error {
    // Validate inviter is a member
    if err := ts.ValidateTribeMembership(ctx, inviterID, tribeID); err != nil {
        return err
    }
    
    // Check tribe capacity
    tribe, err := ts.db.GetTribe(ctx, tribeID)
    if err != nil {
        return err
    }
    
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    if memberCount >= tribe.MaxMembers {
        return errors.New("tribe is at maximum capacity")
    }
    
    // Create invitation
    invitation := &TribeInvitation{
        TribeID:      tribeID,
        InviterID:    inviterID,
        InviteeEmail: inviteeEmail,
        Status:       "pending",
        CreatedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
    }
    
    return ts.db.CreateTribeInvitation(ctx, invitation)
}

// Any member can remove other members (with safeguards)
func (ts *TribeService) RemoveMember(ctx context.Context, tribeID, removerID, targetUserID string) error {
    // Validate remover is a member
    if err := ts.ValidateTribeMembership(ctx, removerID, tribeID); err != nil {
        return err
    }
    
    // Validate target is a member
    if err := ts.ValidateTribeMembership(ctx, targetUserID, tribeID); err != nil {
        return err
    }
    
    // Prevent self-removal (use LeaveTribe instead)
    if removerID == targetUserID {
        return errors.New("use leave tribe function to remove yourself")
    }
    
    // Get current member count
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    // Prevent removing last member
    if memberCount <= 1 {
        return errors.New("cannot remove the last member of a tribe")
    }
    
    // **SAFETY CHECK**: Prevent mass removal in short timeframe
    recentRemovals, err := ts.db.GetRecentRemovals(ctx, tribeID, time.Hour)
    if err != nil {
        return err
    }
    
    // If more than half the tribe has been removed in the last hour, require senior member approval
    if len(recentRemovals) >= memberCount/2 {
        seniorMember, err := ts.GetSeniorMember(ctx, tribeID)
        if err != nil {
            return err
        }
        
        if removerID != seniorMember.ID {
            return errors.New("mass removal detected - senior member approval required")
        }
    }
    
    // Remove member
    if err := ts.db.RemoveTribeMember(ctx, tribeID, targetUserID); err != nil {
        return err
    }
    
    // Log the removal for safety tracking
    return ts.db.LogTribeMemberRemoval(ctx, tribeID, removerID, targetUserID)
}

// Members can leave tribes themselves
func (ts *TribeService) LeaveTribe(ctx context.Context, tribeID, userID string) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    // Check if this is the last member
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    if memberCount == 1 {
        // Last member leaving - offer to delete tribe or transfer to someone else
        return ts.handleLastMemberDeparture(ctx, tribeID, userID)
    }
    
    // Remove user from tribe
    return ts.db.RemoveTribeMember(ctx, tribeID, userID)
}

// Handle last member leaving - clean up or transfer
func (ts *TribeService) handleLastMemberDeparture(ctx context.Context, tribeID, userID string) error {
    // For now, just delete the tribe when last member leaves
    // Could be enhanced to offer transfer to recent members or archive
    return ts.db.DeleteTribe(ctx, tribeID)
}

// Any member can update tribe settings
func (ts *TribeService) UpdateTribeSettings(ctx context.Context, tribeID, userID string, updates TribeUpdateRequest) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    return ts.db.UpdateTribe(ctx, tribeID, updates)
}

// Conflict resolution using senior member
func (ts *TribeService) ResolveConflict(ctx context.Context, tribeID string, conflictType string, options []interface{}) (interface{}, error) {
    seniorMember, err := ts.GetSeniorMember(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    // Log the conflict and resolution method
    conflict := &TribeConflict{
        TribeID:      tribeID,
        ConflictType: conflictType,
        ResolvedBy:   seniorMember.ID,
        Resolution:   "senior_member_decision",
        CreatedAt:    time.Now(),
    }
    
    if err := ts.db.LogTribeConflict(ctx, conflict); err != nil {
        return nil, err
    }
    
    // For now, senior member gets to decide
    // Could be enhanced with voting mechanisms, etc.
    return options[0], nil // Default to first option
}

// Tribe deletion with consensus mechanism
func (ts *TribeService) DeleteTribe(ctx context.Context, tribeID, requestingUserID string, consensusRequired bool) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, requestingUserID, tribeID); err != nil {
        return err
    }
    
    if consensusRequired {
        // Check if all members have agreed to deletion
        members, err := ts.db.GetTribeMembers(ctx, tribeID)
        if err != nil {
            return err
        }
        
        approvals, err := ts.db.GetDeletionApprovals(ctx, tribeID)
        if err != nil {
            return err
        }
        
        if len(approvals) < len(members) {
            // Not all members have approved - record this user's approval
            approval := &TribeDeletionApproval{
                TribeID:   tribeID,
                UserID:    requestingUserID,
                CreatedAt: time.Now(),
            }
            return ts.db.RecordDeletionApproval(ctx, approval)
        }
    }
    
    // All approvals received or consensus not required - delete tribe
    return ts.db.DeleteTribe(ctx, tribeID)
}

type TribeUpdateRequest struct {
    Name                 *string                `json:"name"`
    Description          *string                `json:"description"`
    MaxMembers          *int                   `json:"max_members"`
    DecisionPreferences *DecisionPreferences   `json:"decision_preferences"`
}

type TribeConflict struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    ConflictType string    `json:"conflict_type"`
    ResolvedBy   string    `json:"resolved_by"`
    Resolution   string    `json:"resolution"`
    CreatedAt    time.Time `json:"created_at"`
}

type TribeDeletionApproval struct {
    TribeID   string    `json:"tribe_id"`
    UserID    string    `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}

type TribeInvitation struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    InviterID    string    `json:"inviter_id"`
    InviteeEmail string    `json:"invitee_email"`
    Status       string    `json:"status"` // pending, accepted, declined, expired
    CreatedAt    time.Time `json:"created_at"`
    ExpiresAt    time.Time `json:"expires_at"`
}

type TribeInvitationRatification struct {
    InvitationID string    `json:"invitation_id"`
    MemberID     string    `json:"member_id"`
    Vote         string    `json:"vote"`
    VotedAt      time.Time `json:"voted_at"`
}

type MemberRemovalPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    TargetUserID string     `json:"target_user_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type MemberRemovalVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type TribeDeletionPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type TribeDeletionVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type ListDeletionPetition struct {
    ID                 string     `json:"id"`
    ListID             string     `json:"list_id"`
    PetitionerID       string     `json:"petitioner_id"`
    Reason             string     `json:"reason"`
    Status             string     `json:"status"`
    CreatedAt          time.Time  `json:"created_at"`
    ResolvedAt         *time.Time `json:"resolved_at"`
    ResolvedByUserID   *string    `json:"resolved_by_user_id"`
}

type TribeSettings struct {
    TribeID                string    `json:"tribe_id"`
    InactivityThresholdDays int       `json:"inactivity_threshold_days"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}

type InvitationStatus string

const (
    PENDING                      InvitationStatus = "PENDING"
    ACCEPTED_PENDING_RATIFICATION InvitationStatus = "ACCEPTED_PENDING_RATIFICATION"
    RATIFIED                       InvitationStatus = "RATIFIED"
    REJECTED                       InvitationStatus = "REJECTED"
    REVOKED                        InvitationStatus = "REVOKED"
    EXPIRED                        InvitationStatus = "EXPIRED"
)

type InvitationRatification struct {
    Member User!
    Vote   VoteType!
    VotedAt DateTime!
}

type VoteType string

const (
    APPROVE VoteType = "APPROVE"
    REJECT  VoteType = "REJECT"
)

type PetitionStatus string

const (
    ACTIVE PetitionStatus = "ACTIVE"
    APPROVED PetitionStatus = "APPROVED"
    REJECTED PetitionStatus = "REJECTED"
)

type ListPetitionStatus string

const (
    PENDING    ListPetitionStatus = "PENDING"
    CONFIRMED  ListPetitionStatus = "CONFIRMED"
    CANCELLED  ListPetitionStatus = "CANCELLED"
)
```

### Tribe Invitations Table (Enhanced Two-Stage System)
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

#### Tribe Settings Table (for configurable inactivity thresholds)
```sql
CREATE TABLE tribe_settings (
    tribe_id UUID PRIMARY KEY REFERENCES tribes(id) ON DELETE CASCADE,
    inactivity_threshold_days INTEGER DEFAULT 30, -- 1 to 730 (2 years)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Enhanced Tribe Memberships Table
```sql
-- Add columns to existing tribe_memberships table
ALTER TABLE tribe_memberships 
ADD COLUMN last_login_at TIMESTAMPTZ,
ADD COLUMN is_inactive BOOLEAN DEFAULT FALSE;

-- Update existing constraint
ALTER TABLE tribe_memberships 
DROP CONSTRAINT IF EXISTS tribe_memberships_role_check;

ALTER TABLE tribe_memberships 
ADD CONSTRAINT tribe_memberships_role_check 
CHECK (role IN ('creator', 'member', 'pending'));
```

### Democratic Tribe Governance System

```go
// Enhanced tribe service with democratic governance
type TribeGovernanceService struct {
    db repository.Database
}

// Two-stage invitation system
func (tgs *TribeGovernanceService) InviteToTribe(ctx context.Context, tribeID, inviterID, inviteeEmail string) (*TribeInvitation, error) {
    // Validate inviter is a member
    if err := tgs.validateTribeMembership(ctx, inviterID, tribeID); err != nil {
        return nil, err
    }
    
    // Check tribe capacity
    settings, err := tgs.getTribeSettings(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    memberCount, err := tgs.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    if memberCount >= settings.MaxMembers {
        return nil, errors.New("tribe is at maximum capacity")
    }
    
    // Create invitation (stage 1)
    invitation := &TribeInvitation{
        ID:           generateUUID(),
        TribeID:      tribeID,
        InviterID:    inviterID,
        InviteeEmail: inviteeEmail,
        Status:       "pending",
        InvitedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
    }
    
    return invitation, tgs.db.CreateTribeInvitation(ctx, invitation)
}

// Invitee accepts invitation (moves to ratification stage)
func (tgs *TribeGovernanceService) AcceptInvitation(ctx context.Context, invitationID, userID string) (*TribeInvitation, error) {
    invitation, err := tgs.db.GetTribeInvitation(ctx, invitationID)
    if err != nil {
        return nil, err
    }
    
    if invitation.Status != "pending" {
        return nil, errors.New("invitation is not in pending state")
    }
    
    if time.Now().After(invitation.ExpiresAt) {
        invitation.Status = "expired"
        tgs.db.UpdateTribeInvitation(ctx, invitation)
        return nil, errors.New("invitation has expired")
    }
    
    // Move to ratification stage
    invitation.Status = "accepted_pending_ratification"
    invitation.AcceptedAt = &time.Now()
    
    if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
        return nil, err
    }
    
    // Start ratification process - notify all existing members
    return invitation, tgs.startRatificationProcess(ctx, invitation)
}

// Existing member votes on ratification
func (tgs *TribeGovernanceService) VoteOnInvitation(ctx context.Context, invitationID, voterID string, approve bool) error {
    invitation, err := tgs.db.GetTribeInvitation(ctx, invitationID)
    if err != nil {
        return err
    }
    
    if invitation.Status != "accepted_pending_ratification" {
        return errors.New("invitation is not pending ratification")
    }
    
    // Validate voter is a member
    if err := tgs.validateTribeMembership(ctx, voterID, invitation.TribeID); err != nil {
        return err
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    ratification := &TribeInvitationRatification{
        InvitationID: invitationID,
        MemberID:     voterID,
        Vote:         vote,
        VotedAt:      time.Now(),
    }
    
    if err := tgs.db.CreateInvitationRatification(ctx, ratification); err != nil {
        return err
    }
    
    // If any member rejects, immediately revoke invitation
    if !approve {
        invitation.Status = "rejected"
        return tgs.db.UpdateTribeInvitation(ctx, invitation)
    }
    
    // Check if all members have approved
    return tgs.checkRatificationComplete(ctx, invitation)
}

func (tgs *TribeGovernanceService) checkRatificationComplete(ctx context.Context, invitation *TribeInvitation) error {
    // Get all current members
    members, err := tgs.db.GetTribeMembers(ctx, invitation.TribeID)
    if err != nil {
        return err
    }
    
    // Get all ratification votes
    votes, err := tgs.db.GetInvitationRatifications(ctx, invitation.ID)
    if err != nil {
        return err
    }
    
    // Check if all members have approved
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // All members approved - complete ratification
        invitation.Status = "ratified"
        if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
            return err
        }
        
        // Add user to tribe
        membership := &TribeMembership{
            TribeID:         invitation.TribeID,
            UserID:          invitation.InviteeUserID, // Set when they accept
            InvitedAt:       invitation.InvitedAt,     // Use original invite timestamp
            InvitedByUserID: invitation.InviterID,    // Who invited them
            JoinedAt:        time.Now(),              // When they actually joined
            IsActive:        true,
        }
        
        return tgs.db.CreateTribeMembership(ctx, membership)
    }
    
    return nil // Still waiting for more votes
}

// Member removal petition system
func (tgs *TribeGovernanceService) PetitionMemberRemoval(ctx context.Context, tribeID, petitionerID, targetUserID, reason string) (*MemberRemovalPetition, error) {
    // Validate petitioner is a member
    if err := tgs.validateTribeMembership(ctx, petitionerID, tribeID); err != nil {
        return nil, err
    }
    
    // Validate target is a member
    if err := tgs.validateTribeMembership(ctx, targetUserID, tribeID); err != nil {
        return nil, err
    }
    
    // Cannot petition to remove yourself
    if petitionerID == targetUserID {
        return nil, errors.New("cannot petition to remove yourself - use leave tribe instead")
    }
    
    // Check if petition already exists
    existing, err := tgs.db.GetActiveMemberRemovalPetition(ctx, tribeID, targetUserID)
    if err == nil && existing != nil {
        return nil, errors.New("active petition already exists for this member")
    }
    
    petition := &MemberRemovalPetition{
        ID:           generateUUID(),
        TribeID:      tribeID,
        PetitionerID: petitionerID,
        TargetUserID: targetUserID,
        Reason:       reason,
        Status:       "active",
        CreatedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateMemberRemovalPetition(ctx, petition); err != nil {
        return nil, err
    }
    
    return petition, nil
}

// Vote on member removal
func (tgs *TribeGovernanceService) VoteOnMemberRemoval(ctx context.Context, petitionID, voterID string, approve bool) error {
    petition, err := tgs.db.GetMemberRemovalPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "active" {
        return errors.New("petition is not active")
    }
    
    // Validate voter is a member (but not the target)
    if err := tgs.validateTribeMembership(ctx, voterID, petition.TribeID); err != nil {
        return err
    }
    
    if voterID == petition.TargetUserID {
        return errors.New("target user cannot vote on their own removal")
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    removalVote := &MemberRemovalVote{
        PetitionID: petitionID,
        VoterID:    voterID,
        Vote:       vote,
        VotedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateMemberRemovalVote(ctx, removalVote); err != nil {
        return err
    }
    
    // If any member rejects, petition fails
    if !approve {
        petition.Status = "rejected"
        petition.ResolvedAt = &time.Now()
        return tgs.db.UpdateMemberRemovalPetition(ctx, petition)
    }
    
    // Check if all eligible members have approved
    return tgs.checkMemberRemovalComplete(ctx, petition)
}

func (tgs *TribeGovernanceService) checkMemberRemovalComplete(ctx context.Context, petition *MemberRemovalPetition) error {
    // Get all members except the target
    members, err := tgs.db.GetTribeMembersExcept(ctx, petition.TribeID, petition.TargetUserID)
    if err != nil {
        return err
    }
    
    // Get all votes
    votes, err := tgs.db.GetMemberRemovalVotes(ctx, petition.ID)
    if err != nil {
        return err
    }
    
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // Unanimous approval - remove member
        petition.Status = "approved"
        petition.ResolvedAt = &time.Now()
        
        if err := tgs.db.UpdateMemberRemovalPetition(ctx, petition); err != nil {
            return err
        }
        
        // Remove the member
        return tgs.db.RemoveTribeMember(ctx, petition.TribeID, petition.TargetUserID)
    }
    
    return nil // Still waiting for more votes
}

// Tribe deletion with 100% consensus
func (tgs *TribeGovernanceService) PetitionTribeDeletion(ctx context.Context, tribeID, petitionerID, reason string) (*TribeDeletionPetition, error) {
    // Validate petitioner is a member
    if err := tgs.validateTribeMembership(ctx, petitionerID, tribeID); err != nil {
        return nil, err
    }
    
    // Check if petition already exists
    existing, err := tgs.db.GetActiveTribeDeletionPetition(ctx, tribeID)
    if err == nil && existing != nil {
        return nil, errors.New("active deletion petition already exists")
    }
    
    petition := &TribeDeletionPetition{
        ID:           generateUUID(),
        TribeID:      tribeID,
        PetitionerID: petitionerID,
        Reason:       reason,
        Status:       "active",
        CreatedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateTribeDeletionPetition(ctx, petition); err != nil {
        return nil, err
    }
    
    return petition, nil
}

// Vote on tribe deletion
func (tgs *TribeGovernanceService) VoteOnTribeDeletion(ctx context.Context, petitionID, voterID string, approve bool) error {
    petition, err := tgs.db.GetTribeDeletionPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "active" {
        return errors.New("petition is not active")
    }
    
    // Validate voter is a member
    if err := tgs.validateTribeMembership(ctx, voterID, petition.TribeID); err != nil {
        return err
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    deletionVote := &TribeDeletionVote{
        PetitionID: petitionID,
        VoterID:    voterID,
        Vote:       vote,
        VotedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateTribeDeletionVote(ctx, deletionVote); err != nil {
        return err
    }
    
    // If any member rejects, petition fails
    if !approve {
        petition.Status = "rejected"
        petition.ResolvedAt = &time.Now()
        return tgs.db.UpdateTribeDeletionPetition(ctx, petition)
    }
    
    // Check if all members have approved
    return tgs.checkTribeDeletionComplete(ctx, petition)
}

func (tgs *TribeGovernanceService) checkTribeDeletionComplete(ctx context.Context, petition *TribeDeletionPetition) error {
    // Get all members
    members, err := tgs.db.GetTribeMembers(ctx, petition.TribeID)
    if err != nil {
        return err
    }
    
    // Get all votes
    votes, err := tgs.db.GetTribeDeletionVotes(ctx, petition.ID)
    if err != nil {
        return err
    }
    
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // 100% consensus achieved - delete tribe
        petition.Status = "approved"
        petition.ResolvedAt = &time.Now()
        
        if err := tgs.db.UpdateTribeDeletionPetition(ctx, petition); err != nil {
            return err
        }
        
        // Delete the tribe
        return tgs.db.DeleteTribe(ctx, petition.TribeID)
    }
    
    return nil // Still waiting for more votes
}

// List deletion petition system (simpler - only needs one confirmation)
func (tgs *TribeGovernanceService) PetitionListDeletion(ctx context.Context, listID, petitionerID, reason string) (*ListDeletionPetition, error) {
    list, err := tgs.db.GetList(ctx, listID)
    if err != nil {
        return nil, err
    }
    
    // Only applies to tribe lists
    if list.OwnerType != "tribe" {
        return nil, errors.New("list deletion petitions only apply to tribe lists")
    }
    
    // Validate petitioner is a tribe member
    if err := tgs.validateTribeMembership(ctx, petitionerID, list.OwnerID); err != nil {
        return nil, err
    }
    
    petition := &ListDeletionPetition{
        ID:           generateUUID(),
        ListID:       listID,
        PetitionerID: petitionerID,
        Reason:       reason,
        Status:       "pending",
        CreatedAt:    time.Now(),
    }
    
    return petition, tgs.db.CreateListDeletionPetition(ctx, petition)
}

// Confirm or cancel list deletion
func (tgs *TribeGovernanceService) ResolveListDeletion(ctx context.Context, petitionID, resolverID string, confirm bool) error {
    petition, err := tgs.db.GetListDeletionPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "pending" {
        return errors.New("petition is not pending")
    }
    
    list, err := tgs.db.GetList(ctx, petition.ListID)
    if err != nil {
        return err
    }
    
    // Validate resolver is a tribe member
    if err := tgs.validateTribeMembership(ctx, resolverID, list.OwnerID); err != nil {
        return err
    }
    
    status := "cancelled"
    if confirm {
        status = "confirmed"
    }
    
    petition.Status = status
    petition.ResolvedAt = &time.Now()
    petition.ResolvedByUserID = &resolverID
    
    if err := tgs.db.UpdateListDeletionPetition(ctx, petition); err != nil {
        return err
    }
    
    // If confirmed, delete the list
    if confirm {
        return tgs.db.DeleteList(ctx, petition.ListID)
    }
    
    return nil
}

// Helper function to validate tribe membership
func (tgs *TribeGovernanceService) validateTribeMembership(ctx context.Context, userID, tribeID string) error {
    isMember, err := tgs.db.IsUserTribeMember(ctx, userID, tribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    return nil
}

// Get tribe settings including inactivity threshold
func (tgs *TribeGovernanceService) getTribeSettings(ctx context.Context, tribeID string) (*TribeSettings, error) {
    settings, err := tgs.db.GetTribeSettings(ctx, tribeID)
    if err != nil {
        // Create default settings if none exist
        defaultSettings := &TribeSettings{
            TribeID:                tribeID,
            InactivityThresholdDays: 30,
            CreatedAt:              time.Now(),
            UpdatedAt:              time.Now(),
        }
        if err := tgs.db.CreateTribeSettings(ctx, defaultSettings); err != nil {
            return nil, err
        }
        return defaultSettings, nil
    }
    return settings, nil
}

// Data structures for the new governance system
type TribeInvitation struct {
    ID             string     `json:"id"`
    TribeID        string     `json:"tribe_id"`
    InviterID      string     `json:"inviter_id"`
    InviteeEmail   string     `json:"invitee_email"`
    InviteeUserID  *string    `json:"invitee_user_id"` // Set when they accept
    Status         string     `json:"status"`
    InvitedAt      time.Time  `json:"invited_at"`
    AcceptedAt     *time.Time `json:"accepted_at"`
    ExpiresAt      time.Time  `json:"expires_at"`
}

type TribeInvitationRatification struct {
    InvitationID string    `json:"invitation_id"`
    MemberID     string    `json:"member_id"`
    Vote         string    `json:"vote"`
    VotedAt      time.Time `json:"voted_at"`
}

type MemberRemovalPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    TargetUserID string     `json:"target_user_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type MemberRemovalVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type TribeDeletionPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type TribeDeletionVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type ListDeletionPetition struct {
    ID                 string     `json:"id"`
    ListID             string     `json:"list_id"`
    PetitionerID       string     `json:"petitioner_id"`
    Reason             string     `json:"reason"`
    Status             string     `json:"status"`
    CreatedAt          time.Time  `json:"created_at"`
    ResolvedAt         *time.Time `json:"resolved_at"`
    ResolvedByUserID   *string    `json:"resolved_by_user_id"`
}

type TribeSettings struct {
    TribeID                string    `json:"tribe_id"`
    InactivityThresholdDays int       `json:"inactivity_threshold_days"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}
```

### Tribe Governance Types

```go
// Tribe service implementing common ownership model
type TribeService struct {
    db repository.Database
}

// Common ownership validation - any member can perform tribe operations
func (ts *TribeService) ValidateTribeMembership(ctx context.Context, userID, tribeID string) error {
    isMember, err := ts.db.IsUserTribeMember(ctx, userID, tribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    return nil
}

// Get senior member (longest-standing) for tie-breaking scenarios
func (ts *TribeService) GetSeniorMember(ctx context.Context, tribeID string) (*User, error) {
    return ts.db.GetTribeSeniorMember(ctx, tribeID)
}

// Any member can invite new members
func (ts *TribeService) InviteMember(ctx context.Context, tribeID, inviterID, inviteeEmail string) error {
    // Validate inviter is a member
    if err := ts.ValidateTribeMembership(ctx, inviterID, tribeID); err != nil {
        return err
    }
    
    // Check tribe capacity
    tribe, err := ts.db.GetTribe(ctx, tribeID)
    if err != nil {
        return err
    }
    
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    if memberCount >= tribe.MaxMembers {
        return errors.New("tribe is at maximum capacity")
    }
    
    // Create invitation
    invitation := &TribeInvitation{
        TribeID:      tribeID,
        InviterID:    inviterID,
        InviteeEmail: inviteeEmail,
        Status:       "pending",
        CreatedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
    }
    
    return ts.db.CreateTribeInvitation(ctx, invitation)
}

// Any member can remove other members (with safeguards)
func (ts *TribeService) RemoveMember(ctx context.Context, tribeID, removerID, targetUserID string) error {
    // Validate remover is a member
    if err := ts.ValidateTribeMembership(ctx, removerID, tribeID); err != nil {
        return err
    }
    
    // Validate target is a member
    if err := ts.ValidateTribeMembership(ctx, targetUserID, tribeID); err != nil {
        return err
    }
    
    // Prevent self-removal (use LeaveTribe instead)
    if removerID == targetUserID {
        return errors.New("use leave tribe function to remove yourself")
    }
    
    // Get current member count
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    // Prevent removing last member
    if memberCount <= 1 {
        return errors.New("cannot remove the last member of a tribe")
    }
    
    // **SAFETY CHECK**: Prevent mass removal in short timeframe
    recentRemovals, err := ts.db.GetRecentRemovals(ctx, tribeID, time.Hour)
    if err != nil {
        return err
    }
    
    // If more than half the tribe has been removed in the last hour, require senior member approval
    if len(recentRemovals) >= memberCount/2 {
        seniorMember, err := ts.GetSeniorMember(ctx, tribeID)
        if err != nil {
            return err
        }
        
        if removerID != seniorMember.ID {
            return errors.New("mass removal detected - senior member approval required")
        }
    }
    
    // Remove member
    if err := ts.db.RemoveTribeMember(ctx, tribeID, targetUserID); err != nil {
        return err
    }
    
    // Log the removal for safety tracking
    return ts.db.LogTribeMemberRemoval(ctx, tribeID, removerID, targetUserID)
}

// Members can leave tribes themselves
func (ts *TribeService) LeaveTribe(ctx context.Context, tribeID, userID string) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    // Check if this is the last member
    memberCount, err := ts.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    if memberCount == 1 {
        // Last member leaving - offer to delete tribe or transfer to someone else
        return ts.handleLastMemberDeparture(ctx, tribeID, userID)
    }
    
    // Remove user from tribe
    return ts.db.RemoveTribeMember(ctx, tribeID, userID)
}

// Handle last member leaving - clean up or transfer
func (ts *TribeService) handleLastMemberDeparture(ctx context.Context, tribeID, userID string) error {
    // For now, just delete the tribe when last member leaves
    // Could be enhanced to offer transfer to recent members or archive
    return ts.db.DeleteTribe(ctx, tribeID)
}

// Any member can update tribe settings
func (ts *TribeService) UpdateTribeSettings(ctx context.Context, tribeID, userID string, updates TribeUpdateRequest) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    return ts.db.UpdateTribe(ctx, tribeID, updates)
}

// Conflict resolution using senior member
func (ts *TribeService) ResolveConflict(ctx context.Context, tribeID string, conflictType string, options []interface{}) (interface{}, error) {
    seniorMember, err := ts.GetSeniorMember(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    // Log the conflict and resolution method
    conflict := &TribeConflict{
        TribeID:      tribeID,
        ConflictType: conflictType,
        ResolvedBy:   seniorMember.ID,
        Resolution:   "senior_member_decision",
        CreatedAt:    time.Now(),
    }
    
    if err := ts.db.LogTribeConflict(ctx, conflict); err != nil {
        return nil, err
    }
    
    // For now, senior member gets to decide
    // Could be enhanced with voting mechanisms, etc.
    return options[0], nil // Default to first option
}

// Tribe deletion with consensus mechanism
func (ts *TribeService) DeleteTribe(ctx context.Context, tribeID, requestingUserID string, consensusRequired bool) error {
    // Validate user is a member
    if err := ts.ValidateTribeMembership(ctx, requestingUserID, tribeID); err != nil {
        return err
    }
    
    if consensusRequired {
        // Check if all members have agreed to deletion
        members, err := ts.db.GetTribeMembers(ctx, tribeID)
        if err != nil {
            return err
        }
        
        approvals, err := ts.db.GetDeletionApprovals(ctx, tribeID)
        if err != nil {
            return err
        }
        
        if len(approvals) < len(members) {
            // Not all members have approved - record this user's approval
            approval := &TribeDeletionApproval{
                TribeID:   tribeID,
                UserID:    requestingUserID,
                CreatedAt: time.Now(),
            }
            return ts.db.RecordDeletionApproval(ctx, approval)
        }
    }
    
    // All approvals received or consensus not required - delete tribe
    return ts.db.DeleteTribe(ctx, tribeID)
}

type TribeUpdateRequest struct {
    Name                 *string                `json:"name"`
    Description          *string                `json:"description"`
    MaxMembers          *int                   `json:"max_members"`
    DecisionPreferences *DecisionPreferences   `json:"decision_preferences"`
}

type TribeConflict struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    ConflictType string    `json:"conflict_type"`
    ResolvedBy   string    `json:"resolved_by"`
    Resolution   string    `json:"resolution"`
    CreatedAt    time.Time `json:"created_at"`
}

type TribeDeletionApproval struct {
    TribeID   string    `json:"tribe_id"`
    UserID    string    `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}

type TribeInvitation struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    InviterID    string    `json:"inviter_id"`
    InviteeEmail string    `json:"invitee_email"`
    Status       string    `json:"status"` // pending, accepted, declined, expired
    CreatedAt    time.Time `json:"created_at"`
    ExpiresAt    time.Time `json:"expires_at"`
}

type TribeInvitationRatification struct {
    InvitationID string    `json:"invitation_id"`
    MemberID     string    `json:"member_id"`
    Vote         string    `json:"vote"`
    VotedAt      time.Time `json:"voted_at"`
}

type MemberRemovalPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    TargetUserID string     `json:"target_user_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type MemberRemovalVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type TribeDeletionPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type TribeDeletionVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type ListDeletionPetition struct {
    ID                 string     `json:"id"`
    ListID             string     `json:"list_id"`
    PetitionerID       string     `json:"petitioner_id"`
    Reason             string     `json:"reason"`
    Status             string     `json:"status"`
    CreatedAt          time.Time  `json:"created_at"`
    ResolvedAt         *time.Time `json:"resolved_at"`
    ResolvedByUserID   *string    `json:"resolved_by_user_id"`
}

type TribeSettings struct {
    TribeID                string    `json:"tribe_id"`
    InactivityThresholdDays int       `json:"inactivity_threshold_days"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}

type InvitationStatus string

const (
    PENDING                      InvitationStatus = "PENDING"
    ACCEPTED_PENDING_RATIFICATION InvitationStatus = "ACCEPTED_PENDING_RATIFICATION"
    RATIFIED                       InvitationStatus = "RATIFIED"
    REJECTED                       InvitationStatus = "REJECTED"
    REVOKED                        InvitationStatus = "REVOKED"
    EXPIRED                        InvitationStatus = "EXPIRED"
)

type InvitationRatification struct {
    Member User!
    Vote   VoteType!
    VotedAt DateTime!
}

type VoteType string

const (
    APPROVE VoteType = "APPROVE"
    REJECT  VoteType = "REJECT"
)

type PetitionStatus string

const (
    ACTIVE PetitionStatus = "ACTIVE"
    APPROVED PetitionStatus = "APPROVED"
    REJECTED PetitionStatus = "REJECTED"
)

type ListPetitionStatus string

const (
    PENDING    ListPetitionStatus = "PENDING"
    CONFIRMED  ListPetitionStatus = "CONFIRMED"
    CANCELLED  ListPetitionStatus = "CANCELLED"
)
```

### Tribe Invitations Table (Enhanced Two-Stage System)
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

#### Tribe Settings Table (for configurable inactivity thresholds)
```sql
CREATE TABLE tribe_settings (
    tribe_id UUID PRIMARY KEY REFERENCES tribes(id) ON DELETE CASCADE,
    inactivity_threshold_days INTEGER DEFAULT 30, -- 1 to 730 (2 years)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Enhanced Tribe Memberships Table
```sql
-- Add columns to existing tribe_memberships table
ALTER TABLE tribe_memberships 
ADD COLUMN last_login_at TIMESTAMPTZ,
ADD COLUMN is_inactive BOOLEAN DEFAULT FALSE;

-- Update existing constraint
ALTER TABLE tribe_memberships 
DROP CONSTRAINT IF EXISTS tribe_memberships_role_check;

ALTER TABLE tribe_memberships 
ADD CONSTRAINT tribe_memberships_role_check 
CHECK (role IN ('creator', 'member', 'pending'));
```

### Democratic Tribe Governance System

```go
// Enhanced tribe service with democratic governance
type TribeGovernanceService struct {
    db repository.Database
}

// Two-stage invitation system
func (tgs *TribeGovernanceService) InviteToTribe(ctx context.Context, tribeID, inviterID, inviteeEmail string) (*TribeInvitation, error) {
    // Validate inviter is a member
    if err := tgs.validateTribeMembership(ctx, inviterID, tribeID); err != nil {
        return nil, err
    }
    
    // Check tribe capacity
    settings, err := tgs.getTribeSettings(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    memberCount, err := tgs.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    if memberCount >= settings.MaxMembers {
        return nil, errors.New("tribe is at maximum capacity")
    }
    
    // Create invitation (stage 1)
    invitation := &TribeInvitation{
        ID:           generateUUID(),
        TribeID:      tribeID,
        InviterID:    inviterID,
        InviteeEmail: inviteeEmail,
        Status:       "pending",
        InvitedAt:    time.Now(),
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
    }
    
    return invitation, tgs.db.CreateTribeInvitation(ctx, invitation)
}

// Invitee accepts invitation (moves to ratification stage)
func (tgs *TribeGovernanceService) AcceptInvitation(ctx context.Context, invitationID, userID string) (*TribeInvitation, error) {
    invitation, err := tgs.db.GetTribeInvitation(ctx, invitationID)
    if err != nil {
        return nil, err
    }
    
    if invitation.Status != "pending" {
        return nil, errors.New("invitation is not in pending state")
    }
    
    if time.Now().After(invitation.ExpiresAt) {
        invitation.Status = "expired"
        tgs.db.UpdateTribeInvitation(ctx, invitation)
        return nil, errors.New("invitation has expired")
    }
    
    // Move to ratification stage
    invitation.Status = "accepted_pending_ratification"
    invitation.AcceptedAt = &time.Now()
    
    if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
        return nil, err
    }
    
    // Start ratification process - notify all existing members
    return invitation, tgs.startRatificationProcess(ctx, invitation)
}

// Existing member votes on ratification
func (tgs *TribeGovernanceService) VoteOnInvitation(ctx context.Context, invitationID, voterID string, approve bool) error {
    invitation, err := tgs.db.GetTribeInvitation(ctx, invitationID)
    if err != nil {
        return err
    }
    
    if invitation.Status != "accepted_pending_ratification" {
        return errors.New("invitation is not pending ratification")
    }
    
    // Validate voter is a member
    if err := tgs.validateTribeMembership(ctx, voterID, invitation.TribeID); err != nil {
        return err
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    ratification := &TribeInvitationRatification{
        InvitationID: invitationID,
        MemberID:     voterID,
        Vote:         vote,
        VotedAt:      time.Now(),
    }
    
    if err := tgs.db.CreateInvitationRatification(ctx, ratification); err != nil {
        return err
    }
    
    // If any member rejects, immediately revoke invitation
    if !approve {
        invitation.Status = "rejected"
        return tgs.db.UpdateTribeInvitation(ctx, invitation)
    }
    
    // Check if all members have approved
    return tgs.checkRatificationComplete(ctx, invitation)
}

func (tgs *TribeGovernanceService) checkRatificationComplete(ctx context.Context, invitation *TribeInvitation) error {
    // Get all current members
    members, err := tgs.db.GetTribeMembers(ctx, invitation.TribeID)
    if err != nil {
        return err
    }
    
    // Get all ratification votes
    votes, err := tgs.db.GetInvitationRatifications(ctx, invitation.ID)
    if err != nil {
        return err
    }
    
    // Check if all members have approved
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // All members approved - complete ratification
        invitation.Status = "ratified"
        if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
            return err
        }
        
        // Add user to tribe
        membership := &TribeMembership{
            TribeID:         invitation.TribeID,
            UserID:          invitation.InviteeUserID, // Set when they accept
            InvitedAt:       invitation.InvitedAt,     // Use original invite timestamp
            InvitedByUserID: invitation.InviterID,    // Who invited them
            JoinedAt:        time.Now(),              // When they actually joined
            IsActive:        true,
        }
        
        return tgs.db.CreateTribeMembership(ctx, membership)
    }
    
    return nil // Still waiting for more votes
}

// Member removal petition system
func (tgs *TribeGovernanceService) PetitionMemberRemoval(ctx context.Context, tribeID, petitionerID, targetUserID, reason string) (*MemberRemovalPetition, error) {
    // Validate petitioner is a member
    if err := tgs.validateTribeMembership(ctx, petitionerID, tribeID); err != nil {
        return nil, err
    }
    
    // Validate target is a member
    if err := tgs.validateTribeMembership(ctx, targetUserID, tribeID); err != nil {
        return nil, err
    }
    
    // Cannot petition to remove yourself
    if petitionerID == targetUserID {
        return nil, errors.New("cannot petition to remove yourself - use leave tribe instead")
    }
    
    // Check if petition already exists
    existing, err := tgs.db.GetActiveMemberRemovalPetition(ctx, tribeID, targetUserID)
    if err == nil && existing != nil {
        return nil, errors.New("active petition already exists for this member")
    }
    
    petition := &MemberRemovalPetition{
        ID:           generateUUID(),
        TribeID:      tribeID,
        PetitionerID: petitionerID,
        TargetUserID: targetUserID,
        Reason:       reason,
        Status:       "active",
        CreatedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateMemberRemovalPetition(ctx, petition); err != nil {
        return nil, err
    }
    
    return petition, nil
}

// Vote on member removal
func (tgs *TribeGovernanceService) VoteOnMemberRemoval(ctx context.Context, petitionID, voterID string, approve bool) error {
    petition, err := tgs.db.GetMemberRemovalPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "active" {
        return errors.New("petition is not active")
    }
    
    // Validate voter is a member (but not the target)
    if err := tgs.validateTribeMembership(ctx, voterID, petition.TribeID); err != nil {
        return err
    }
    
    if voterID == petition.TargetUserID {
        return errors.New("target user cannot vote on their own removal")
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    removalVote := &MemberRemovalVote{
        PetitionID: petitionID,
        VoterID:    voterID,
        Vote:       vote,
        VotedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateMemberRemovalVote(ctx, removalVote); err != nil {
        return err
    }
    
    // If any member rejects, petition fails
    if !approve {
        petition.Status = "rejected"
        petition.ResolvedAt = &time.Now()
        return tgs.db.UpdateMemberRemovalPetition(ctx, petition)
    }
    
    // Check if all eligible members have approved
    return tgs.checkMemberRemovalComplete(ctx, petition)
}

func (tgs *TribeGovernanceService) checkMemberRemovalComplete(ctx context.Context, petition *MemberRemovalPetition) error {
    // Get all members except the target
    members, err := tgs.db.GetTribeMembersExcept(ctx, petition.TribeID, petition.TargetUserID)
    if err != nil {
        return err
    }
    
    // Get all votes
    votes, err := tgs.db.GetMemberRemovalVotes(ctx, petition.ID)
    if err != nil {
        return err
    }
    
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // Unanimous approval - remove member
        petition.Status = "approved"
        petition.ResolvedAt = &time.Now()
        
        if err := tgs.db.UpdateMemberRemovalPetition(ctx, petition); err != nil {
            return err
        }
        
        // Remove the member
        return tgs.db.RemoveTribeMember(ctx, petition.TribeID, petition.TargetUserID)
    }
    
    return nil // Still waiting for more votes
}

// Tribe deletion with 100% consensus
func (tgs *TribeGovernanceService) PetitionTribeDeletion(ctx context.Context, tribeID, petitionerID, reason string) (*TribeDeletionPetition, error) {
    // Validate petitioner is a member
    if err := tgs.validateTribeMembership(ctx, petitionerID, tribeID); err != nil {
        return nil, err
    }
    
    // Check if petition already exists
    existing, err := tgs.db.GetActiveTribeDeletionPetition(ctx, tribeID)
    if err == nil && existing != nil {
        return nil, errors.New("active deletion petition already exists")
    }
    
    petition := &TribeDeletionPetition{
        ID:           generateUUID(),
        TribeID:      tribeID,
        PetitionerID: petitionerID,
        Reason:       reason,
        Status:       "active",
        CreatedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateTribeDeletionPetition(ctx, petition); err != nil {
        return nil, err
    }
    
    return petition, nil
}

// Vote on tribe deletion
func (tgs *TribeGovernanceService) VoteOnTribeDeletion(ctx context.Context, petitionID, voterID string, approve bool) error {
    petition, err := tgs.db.GetTribeDeletionPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "active" {
        return errors.New("petition is not active")
    }
    
    // Validate voter is a member
    if err := tgs.validateTribeMembership(ctx, voterID, petition.TribeID); err != nil {
        return err
    }
    
    vote := "approve"
    if !approve {
        vote = "reject"
    }
    
    // Record vote
    deletionVote := &TribeDeletionVote{
        PetitionID: petitionID,
        VoterID:    voterID,
        Vote:       vote,
        VotedAt:    time.Now(),
    }
    
    if err := tgs.db.CreateTribeDeletionVote(ctx, deletionVote); err != nil {
        return err
    }
    
    // If any member rejects, petition fails
    if !approve {
        petition.Status = "rejected"
        petition.ResolvedAt = &time.Now()
        return tgs.db.UpdateTribeDeletionPetition(ctx, petition)
    }
    
    // Check if all members have approved
    return tgs.checkTribeDeletionComplete(ctx, petition)
}

func (tgs *TribeGovernanceService) checkTribeDeletionComplete(ctx context.Context, petition *TribeDeletionPetition) error {
    // Get all members
    members, err := tgs.db.GetTribeMembers(ctx, petition.TribeID)
    if err != nil {
        return err
    }
    
    // Get all votes
    votes, err := tgs.db.GetTribeDeletionVotes(ctx, petition.ID)
    if err != nil {
        return err
    }
    
    approvals := 0
    for _, vote := range votes {
        if vote.Vote == "approve" {
            approvals++
        }
    }
    
    if approvals >= len(members) {
        // 100% consensus achieved - delete tribe
        petition.Status = "approved"
        petition.ResolvedAt = &time.Now()
        
        if err := tgs.db.UpdateTribeDeletionPetition(ctx, petition); err != nil {
            return err
        }
        
        // Delete the tribe
        return tgs.db.DeleteTribe(ctx, petition.TribeID)
    }
    
    return nil // Still waiting for more votes
}

// List deletion petition system (simpler - only needs one confirmation)
func (tgs *TribeGovernanceService) PetitionListDeletion(ctx context.Context, listID, petitionerID, reason string) (*ListDeletionPetition, error) {
    list, err := tgs.db.GetList(ctx, listID)
    if err != nil {
        return nil, err
    }
    
    // Only applies to tribe lists
    if list.OwnerType != "tribe" {
        return nil, errors.New("list deletion petitions only apply to tribe lists")
    }
    
    // Validate petitioner is a tribe member
    if err := tgs.validateTribeMembership(ctx, petitionerID, list.OwnerID); err != nil {
        return nil, err
    }
    
    petition := &ListDeletionPetition{
        ID:           generateUUID(),
        ListID:       listID,
        PetitionerID: petitionerID,
        Reason:       reason,
        Status:       "pending",
        CreatedAt:    time.Now(),
    }
    
    return petition, tgs.db.CreateListDeletionPetition(ctx, petition)
}

// Confirm or cancel list deletion
func (tgs *TribeGovernanceService) ResolveListDeletion(ctx context.Context, petitionID, resolverID string, confirm bool) error {
    petition, err := tgs.db.GetListDeletionPetition(ctx, petitionID)
    if err != nil {
        return err
    }
    
    if petition.Status != "pending" {
        return errors.New("petition is not pending")
    }
    
    list, err := tgs.db.GetList(ctx, petition.ListID)
    if err != nil {
        return err
    }
    
    // Validate resolver is a tribe member
    if err := tgs.validateTribeMembership(ctx, resolverID, list.OwnerID); err != nil {
        return err
    }
    
    status := "cancelled"
    if confirm {
        status = "confirmed"
    }
    
    petition.Status = status
    petition.ResolvedAt = &time.Now()
    petition.ResolvedByUserID = &resolverID
    
    if err := tgs.db.UpdateListDeletionPetition(ctx, petition); err != nil {
        return err
    }
    
    // If confirmed, delete the list
    if confirm {
        return tgs.db.DeleteList(ctx, petition.ListID)
    }
    
    return nil
}

// Helper function to validate tribe membership
func (tgs *TribeGovernanceService) validateTribeMembership(ctx context.Context, userID, tribeID string) error {
    isMember, err := tgs.db.IsUserTribeMember(ctx, userID, tribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    return nil
}

// Get tribe settings including inactivity threshold
func (tgs *TribeGovernanceService) getTribeSettings(ctx context.Context, tribeID string) (*TribeSettings, error) {
    settings, err := tgs.db.GetTribeSettings(ctx, tribeID)
    if err != nil {
        // Create default settings if none exist
        defaultSettings := &TribeSettings{
            TribeID:                tribeID,
            InactivityThresholdDays: 30,
            CreatedAt:              time.Now(),
            UpdatedAt:              time.Now(),
        }
        if err := tgs.db.CreateTribeSettings(ctx, defaultSettings); err != nil {
            return nil, err
        }
        return defaultSettings, nil
    }
    return settings, nil
}

// Data structures for the new governance system
type TribeInvitation struct {
    ID             string     `json:"id"`
    TribeID        string     `json:"tribe_id"`
    InviterID      string     `json:"inviter_id"`
    InviteeEmail   string     `json:"invitee_email"`
    InviteeUserID  *string    `json:"invitee_user_id"` // Set when they accept
    Status         string     `json:"status"`
    InvitedAt      time.Time  `json:"invited_at"`
    AcceptedAt     *time.Time `json:"accepted_at"`
    ExpiresAt      time.Time  `json:"expires_at"`
}

type TribeInvitationRatification struct {
    InvitationID string    `json:"invitation_id"`
    MemberID     string    `json:"member_id"`
    Vote         string    `json:"vote"`
    VotedAt      time.Time `json:"voted_at"`
}

type MemberRemovalPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    TargetUserID string     `json:"target_user_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type MemberRemovalVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type TribeDeletionPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at"`
}

type TribeDeletionVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"`
    VotedAt    time.Time `json:"voted_at"`
}

type ListDeletionPetition struct {
    ID                 string     `json:"id"`
    ListID             string     `json:"list_id"`
    PetitionerID       string     `json:"petitioner_id"`
    Reason             string     `json:"reason"`
    Status             string     `json:"status"`
    CreatedAt          time.Time  `json:"created_at"`
    ResolvedAt         *time.Time `json:"resolved_at"`
    ResolvedByUserID   *string    `json:"resolved_by_user_id"`
}

type TribeSettings struct {
    TribeID                string    `json:"tribe_id"`
    InactivityThresholdDays int       `json:"inactivity_threshold_days"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}
```

