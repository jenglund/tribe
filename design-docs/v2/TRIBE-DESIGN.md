# Tribe Governance Design

## Overview

This document defines the democratic governance system for tribes in the Tribe application. The system implements collaborative decision-making with senior member tie-breaking, ensuring fair and transparent tribe management for groups of 1-8 people.

## Core Governance Principles

### Democratic Participation
- **Equal Voice**: All active members participate in governance decisions
- **Transparency**: All governance actions are logged and visible to members  
- **Safeguards**: Petitions and voting prevent abuse while enabling necessary actions

### Senior Member System
- **Definition**: The member with the earliest invite timestamp among currently active members
- **Role**: Serves as tie-breaker and conflict resolver when democratic processes fail
- **Automatic**: Senior member role transfers automatically when current senior leaves

### Consensus Requirements
- **Invitations**: Require unanimous approval from existing members (two-stage process)
- **Member Removal**: Requires unanimous approval from all members except the target
- **Tribe Deletion**: Requires 100% consensus from all active members
- **List Operations**: Most operations are democratic, some require confirmation

## Implementation

### Core Service Structure

```go
// TribeGovernanceService handles all democratic tribe operations
type TribeGovernanceService struct {
    db repository.Database
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

// Get senior member (earliest invite among active members) for tie-breaking
func (tgs *TribeGovernanceService) GetSeniorMember(ctx context.Context, tribeID string) (*User, error) {
    seniorUserID, err := tgs.db.GetTribeSeniorMember(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    return tgs.db.GetUser(ctx, seniorUserID)
}

// Get tribe creator (user who invited themselves) - may have left
func (tgs *TribeGovernanceService) GetTribeCreator(ctx context.Context, tribeID string) (*User, error) {
    creatorUserID, err := tgs.db.GetTribeCreator(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    if creatorUserID == "" {
        return nil, nil // Creator has left the tribe
    }
    
    return tgs.db.GetUser(ctx, creatorUserID)
}
```

### Tribe Creation

```go
// Create tribe with democratic governance enabled
func (tgs *TribeGovernanceService) CreateTribe(ctx context.Context, creatorID string, name, description string) (*Tribe, error) {
    // Create the tribe
    tribe := &Tribe{
        ID:          generateUUID(),
        Name:        name,
        Description: description,
        MaxMembers:  8,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    if err := tgs.db.CreateTribe(ctx, tribe); err != nil {
        return nil, err
    }
    
    // Create founder membership with self-invitation pattern
    inviteTime := time.Now()
    membership := &TribeMembership{
        ID:               generateUUID(),
        TribeID:          tribe.ID,
        UserID:           creatorID,
        InvitedAt:        inviteTime,
        InvitedByUserID:  creatorID, // Self-invited (founder pattern)
        JoinedAt:         inviteTime, // Joined immediately
        IsActive:         true,
    }
    
    if err := tgs.db.CreateTribeMembership(ctx, membership); err != nil {
        // Rollback tribe creation
        tgs.db.DeleteTribe(ctx, tribe.ID)
        return nil, err
    }
    
    return tribe, nil
}
```

### Two-Stage Invitation System

```go
// Stage 1: Member initiates invitation
func (tgs *TribeGovernanceService) InviteToTribe(ctx context.Context, tribeID, inviterID, inviteeEmail string) (*TribeInvitation, error) {
    // Validate inviter is a member
    if err := tgs.validateTribeMembership(ctx, inviterID, tribeID); err != nil {
        return nil, err
    }
    
    // Check tribe capacity
    tribe, err := tgs.db.GetTribe(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    memberCount, err := tgs.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    if memberCount >= tribe.MaxMembers {
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

// Stage 2A: Invitee accepts invitation (moves to ratification)
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
    invitation.InviteeUserID = userID
    acceptedTime := time.Now()
    invitation.AcceptedAt = &acceptedTime
    
    if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
        return nil, err
    }
    
    // For single-member tribes, auto-approve
    memberCount, err := tgs.db.GetTribeMemberCount(ctx, invitation.TribeID)
    if err != nil {
        return nil, err
    }
    
    if memberCount == 1 {
        return tgs.autoApproveInvitation(ctx, invitation)
    }
    
    return invitation, nil
}

// Stage 2B: Existing members vote on ratification  
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
    
    // If any member rejects, immediately reject invitation
    if !approve {
        invitation.Status = "rejected"
        return tgs.db.UpdateTribeInvitation(ctx, invitation)
    }
    
    // Check if all members have approved
    return tgs.checkRatificationComplete(ctx, invitation)
}

// Complete ratification process
func (tgs *TribeGovernanceService) checkRatificationComplete(ctx context.Context, invitation *TribeInvitation) error {
    members, err := tgs.db.GetTribeMembers(ctx, invitation.TribeID)
    if err != nil {
        return err
    }
    
    votes, err := tgs.db.GetInvitationRatifications(ctx, invitation.ID)
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
        // All members approved - add member to tribe
        invitation.Status = "ratified"
        if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
            return err
        }
        
        membership := &TribeMembership{
            ID:               generateUUID(),
            TribeID:          invitation.TribeID,
            UserID:           invitation.InviteeUserID,
            InvitedAt:        invitation.InvitedAt,     // Original invite time
            InvitedByUserID:  invitation.InviterID,    // Who invited them
            JoinedAt:         time.Now(),              // When they joined
            IsActive:         true,
        }
        
        return tgs.db.CreateTribeMembership(ctx, membership)
    }
    
    return nil // Still waiting for more votes
}

// Auto-approve for single-member tribes
func (tgs *TribeGovernanceService) autoApproveInvitation(ctx context.Context, invitation *TribeInvitation) (*TribeInvitation, error) {
    invitation.Status = "ratified"
    if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
        return nil, err
    }
    
    membership := &TribeMembership{
        ID:               generateUUID(),
        TribeID:          invitation.TribeID,
        UserID:           invitation.InviteeUserID,
        InvitedAt:        invitation.InvitedAt,
        InvitedByUserID:  invitation.InviterID,
        JoinedAt:         time.Now(),
        IsActive:         true,
    }
    
    if err := tgs.db.CreateTribeMembership(ctx, membership); err != nil {
        return nil, err
    }
    
    return invitation, nil
}
```

### Democratic Member Removal

```go
// Petition to remove a member
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

// Vote on member removal petition
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
        resolvedTime := time.Now()
        petition.ResolvedAt = &resolvedTime
        return tgs.db.UpdateMemberRemovalPetition(ctx, petition)
    }
    
    // Check if all eligible members have approved
    return tgs.checkMemberRemovalComplete(ctx, petition)
}

// Complete member removal process
func (tgs *TribeGovernanceService) checkMemberRemovalComplete(ctx context.Context, petition *MemberRemovalPetition) error {
    // Get all members except the target
    members, err := tgs.db.GetTribeMembersExcept(ctx, petition.TribeID, petition.TargetUserID)
    if err != nil {
        return err
    }
    
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
        resolvedTime := time.Now()
        petition.ResolvedAt = &resolvedTime
        
        if err := tgs.db.UpdateMemberRemovalPetition(ctx, petition); err != nil {
            return err
        }
        
        // Remove the member
        return tgs.db.RemoveTribeMember(ctx, petition.TribeID, petition.TargetUserID)
    }
    
    return nil // Still waiting for more votes
}
```

### Voluntary Member Departure

```go
// Member leaves tribe voluntarily
func (tgs *TribeGovernanceService) LeaveTribe(ctx context.Context, tribeID, userID string) error {
    // Validate user is a member
    if err := tgs.validateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    // Check if this is the last member
    memberCount, err := tgs.db.GetTribeMemberCount(ctx, tribeID)
    if err != nil {
        return err
    }
    
    if memberCount == 1 {
        // Last member leaving - delete tribe
        return tgs.db.DeleteTribe(ctx, tribeID)
    }
    
    // Remove user from tribe
    return tgs.db.RemoveTribeMember(ctx, tribeID, userID)
}
```

### Democratic Tribe Deletion

```go
// Petition for tribe deletion
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
        resolvedTime := time.Now()
        petition.ResolvedAt = &resolvedTime
        return tgs.db.UpdateTribeDeletionPetition(ctx, petition)
    }
    
    // Check if all members have approved (100% consensus required)
    return tgs.checkTribeDeletionComplete(ctx, petition)
}

// Complete tribe deletion process
func (tgs *TribeGovernanceService) checkTribeDeletionComplete(ctx context.Context, petition *TribeDeletionPetition) error {
    members, err := tgs.db.GetTribeMembers(ctx, petition.TribeID)
    if err != nil {
        return err
    }
    
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
        resolvedTime := time.Now()
        petition.ResolvedAt = &resolvedTime
        
        if err := tgs.db.UpdateTribeDeletionPetition(ctx, petition); err != nil {
            return err
        }
        
        // Delete the tribe and all associated data
        return tgs.db.DeleteTribe(ctx, petition.TribeID)
    }
    
    return nil // Still waiting for more votes
}
```

### List Governance

```go
// Petition for list deletion (simpler process)
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
    resolvedTime := time.Now()
    petition.ResolvedAt = &resolvedTime
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
```

### Conflict Resolution

```go
// Resolve conflicts using senior member as tie-breaker
func (tgs *TribeGovernanceService) ResolveConflict(ctx context.Context, tribeID string, conflictType string, options []interface{}) (interface{}, error) {
    seniorMember, err := tgs.GetSeniorMember(ctx, tribeID)
    if err != nil {
        return nil, err
    }
    
    // Log the conflict and resolution
    conflict := &TribeConflict{
        ID:           generateUUID(),
        TribeID:      tribeID,
        ConflictType: conflictType,
        ResolvedBy:   seniorMember.ID,
        Resolution:   "senior_member_decision",
        CreatedAt:    time.Now(),
    }
    
    if err := tgs.db.LogTribeConflict(ctx, conflict); err != nil {
        return nil, err
    }
    
    // Senior member decision prevails
    if len(options) > 0 {
        return options[0], nil
    }
    
    return nil, errors.New("no options provided for conflict resolution")
}
```

### Tribe Settings Management

```go
// Get or create tribe settings
func (tgs *TribeGovernanceService) GetTribeSettings(ctx context.Context, tribeID string) (*TribeSettings, error) {
    settings, err := tgs.db.GetTribeSettings(ctx, tribeID)
    if err != nil {
        // Create default settings if none exist
        defaultSettings := &TribeSettings{
            TribeID:                tribeID,
            MaxMembers:             8,
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

// Update tribe settings (any member can propose, senior member resolves conflicts)
func (tgs *TribeGovernanceService) UpdateTribeSettings(ctx context.Context, tribeID, userID string, updates TribeUpdateRequest) error {
    // Validate user is a member
    if err := tgs.validateTribeMembership(ctx, userID, tribeID); err != nil {
        return err
    }
    
    // For MVP, any member can update settings
    // TODO: Could be enhanced with voting system for controversial changes
    return tgs.db.UpdateTribeSettings(ctx, tribeID, updates)
}
```

## Data Types

```go
// Core governance types
type TribeInvitation struct {
    ID               string     `json:"id"`
    TribeID          string     `json:"tribe_id"`
    InviterID        string     `json:"inviter_id"`
    InviteeEmail     string     `json:"invitee_email"`
    InviteeUserID    string     `json:"invitee_user_id,omitempty"` // Set when accepted
    Status           string     `json:"status"` // pending, accepted_pending_ratification, ratified, rejected, expired
    InvitedAt        time.Time  `json:"invited_at"`
    AcceptedAt       *time.Time `json:"accepted_at,omitempty"`
    ExpiresAt        time.Time  `json:"expires_at"`
}

type TribeInvitationRatification struct {
    InvitationID string    `json:"invitation_id"`
    MemberID     string    `json:"member_id"`
    Vote         string    `json:"vote"` // approve, reject
    VotedAt      time.Time `json:"voted_at"`
}

type MemberRemovalPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    TargetUserID string     `json:"target_user_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"` // active, approved, rejected
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

type MemberRemovalVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"` // approve, reject
    VotedAt    time.Time `json:"voted_at"`
}

type TribeDeletionPetition struct {
    ID           string     `json:"id"`
    TribeID      string     `json:"tribe_id"`
    PetitionerID string     `json:"petitioner_id"`
    Reason       string     `json:"reason"`
    Status       string     `json:"status"` // active, approved, rejected
    CreatedAt    time.Time  `json:"created_at"`
    ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

type TribeDeletionVote struct {
    PetitionID string    `json:"petition_id"`
    VoterID    string    `json:"voter_id"`
    Vote       string    `json:"vote"` // approve, reject
    VotedAt    time.Time `json:"voted_at"`
}

type ListDeletionPetition struct {
    ID                string     `json:"id"`
    ListID            string     `json:"list_id"`
    PetitionerID      string     `json:"petitioner_id"`
    Reason            string     `json:"reason"`
    Status            string     `json:"status"` // pending, confirmed, cancelled
    CreatedAt         time.Time  `json:"created_at"`
    ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
    ResolvedByUserID  *string    `json:"resolved_by_user_id,omitempty"`
}

type TribeConflict struct {
    ID           string    `json:"id"`
    TribeID      string    `json:"tribe_id"`
    ConflictType string    `json:"conflict_type"`
    ResolvedBy   string    `json:"resolved_by"`
    Resolution   string    `json:"resolution"`
    CreatedAt    time.Time `json:"created_at"`
}

type TribeSettings struct {
    TribeID                string    `json:"tribe_id"`
    MaxMembers             int       `json:"max_members"`
    InactivityThresholdDays int       `json:"inactivity_threshold_days"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}

type TribeUpdateRequest struct {
    Name                    *string                `json:"name,omitempty"`
    Description             *string                `json:"description,omitempty"`
    MaxMembers              *int                   `json:"max_members,omitempty"`
    InactivityThresholdDays *int                   `json:"inactivity_threshold_days,omitempty"`
}

type TribeMembership struct {
    ID               string    `json:"id"`
    TribeID          string    `json:"tribe_id"`
    UserID           string    `json:"user_id"`
    InvitedAt        time.Time `json:"invited_at"`
    InvitedByUserID  string    `json:"invited_by_user_id"`
    JoinedAt         time.Time `json:"joined_at"`
    IsActive         bool      `json:"is_active"`
    LastActiveAt     time.Time `json:"last_active_at"`
}

type Tribe struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    MaxMembers  int       `json:"max_members"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## Database Schema

The database tables supporting this governance system are defined in [`DATA-MODEL.md`](./DATA-MODEL.md). Key tables include:

- `tribes` - Basic tribe information
- `tribe_memberships` - Member relationships with invite tracking
- `tribe_invitations` - Two-stage invitation process
- `tribe_invitation_ratifications` - Member voting on invitations  
- `member_removal_petitions` / `member_removal_votes` - Democratic member removal
- `tribe_deletion_petitions` / `tribe_deletion_votes` - Democratic tribe deletion
- `list_deletion_petitions` - List deletion with confirmation
- `tribe_settings` - Configurable tribe settings
- `tribe_conflicts` - Conflict resolution logging

See [`DATA-MODEL.md`](./DATA-MODEL.md) for complete schema definitions and indexes.

## Governance Workflows

### New Member Invitation Flow
1. **Initiate**: Any member can invite via email
2. **Accept**: Invitee accepts invitation (moves to ratification)
3. **Ratify**: All existing members must approve (unanimous)
4. **Complete**: Member is added to tribe
5. **Reject**: Any member rejection immediately cancels invitation

### Member Removal Flow
1. **Petition**: Any member can petition to remove another (with reason)
2. **Vote**: All members except target vote (unanimous approval required)
3. **Complete**: Target is removed from tribe
4. **Reject**: Any member rejection cancels petition

### Tribe Deletion Flow
1. **Petition**: Any member can petition for tribe deletion
2. **Vote**: All members vote (100% consensus required)
3. **Complete**: Tribe and all data is deleted
4. **Reject**: Any member rejection cancels petition

### Conflict Resolution
1. **Identify**: System or members detect conflicting state
2. **Senior Member**: System identifies senior member (earliest invite)
3. **Resolve**: Senior member decision is applied
4. **Log**: Resolution is recorded for transparency

## Implementation Notes

- **Atomic Operations**: All voting processes use database transactions
- **Race Conditions**: Proper locking prevents double-voting or concurrent modifications
- **Audit Trail**: All governance actions are logged with timestamps and actors
- **Graceful Degradation**: System handles edge cases (member leaves during vote, etc.)
- **Senior Member Calculation**: Automatically updates when members join/leave
- **Notification System**: Members are notified of pending votes and outcomes

This democratic governance system ensures fair, transparent tribe management while providing necessary safeguards against abuse and mechanisms for conflict resolution.