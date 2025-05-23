package services

import (
	"context"
	"errors"
	"time"

	"tribe/internal/repository"
)

// TribeGovernanceService handles all democratic tribe operations
//
// For complete type definitions, see: ../DATA-MODEL.md#go-type-definitions
type TribeGovernanceService struct {
	db repository.Database
}

// NewTribeGovernanceService creates a new tribe governance service
func NewTribeGovernanceService(db repository.Database) *TribeGovernanceService {
	return &TribeGovernanceService{db: db}
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

// GetSeniorMember gets senior member (earliest invite among active members) for tie-breaking
func (tgs *TribeGovernanceService) GetSeniorMember(ctx context.Context, tribeID string) (*User, error) {
	seniorUserID, err := tgs.db.GetTribeSeniorMember(ctx, tribeID)
	if err != nil {
		return nil, err
	}
	return tgs.db.GetUser(ctx, seniorUserID)
}

// GetTribeCreator gets tribe creator (user who invited themselves) - may have left
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

// CreateTribe creates tribe with democratic governance enabled
func (tgs *TribeGovernanceService) CreateTribe(ctx context.Context, creatorID string, name, description string) (*Tribe, error) {
	// Create the tribe
	tribe := &Tribe{
		ID:          generateUUID(),
		Name:        name,
		Description: &description,
		CreatorID:   creatorID,
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
		ID:              generateUUID(),
		TribeID:         tribe.ID,
		UserID:          creatorID,
		InvitedAt:       inviteTime,
		InvitedByUserID: creatorID,  // Self-invited (founder pattern)
		JoinedAt:        inviteTime, // Joined immediately
		IsActive:        true,
	}

	if err := tgs.db.CreateTribeMembership(ctx, membership); err != nil {
		// Rollback tribe creation
		tgs.db.DeleteTribe(ctx, tribe.ID)
		return nil, err
	}

	return tribe, nil
}

// InviteToTribe initiates invitation (Stage 1 of two-stage process)
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

// AcceptInvitation moves invitation to ratification stage (Stage 2A)
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
	invitation.InviteeUserID = &userID
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

// VoteOnInvitation allows existing members to vote on ratification (Stage 2B)
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
		ID:           generateUUID(),
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

// LeaveTribe allows member to leave tribe voluntarily
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

// PetitionMemberRemoval initiates member removal process
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
		Reason:       &reason,
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	if err := tgs.db.CreateMemberRemovalPetition(ctx, petition); err != nil {
		return nil, err
	}

	return petition, nil
}

// VoteOnMemberRemoval allows members to vote on removal petition
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
		ID:         generateUUID(),
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

// PetitionTribeDeletion initiates tribe deletion process
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
		Reason:       &reason,
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	if err := tgs.db.CreateTribeDeletionPetition(ctx, petition); err != nil {
		return nil, err
	}

	return petition, nil
}

// VoteOnTribeDeletion allows members to vote on tribe deletion
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
		ID:         generateUUID(),
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

// Helper methods for completing voting processes

func (tgs *TribeGovernanceService) autoApproveInvitation(ctx context.Context, invitation *TribeInvitation) (*TribeInvitation, error) {
	invitation.Status = "ratified"
	if err := tgs.db.UpdateTribeInvitation(ctx, invitation); err != nil {
		return nil, err
	}

	membership := &TribeMembership{
		ID:              generateUUID(),
		TribeID:         invitation.TribeID,
		UserID:          *invitation.InviteeUserID,
		InvitedAt:       invitation.InvitedAt,
		InvitedByUserID: invitation.InviterID,
		JoinedAt:        time.Now(),
		IsActive:        true,
	}

	if err := tgs.db.CreateTribeMembership(ctx, membership); err != nil {
		return nil, err
	}

	return invitation, nil
}

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
			ID:              generateUUID(),
			TribeID:         invitation.TribeID,
			UserID:          *invitation.InviteeUserID,
			InvitedAt:       invitation.InvitedAt, // Original invite time
			InvitedByUserID: invitation.InviterID, // Who invited them
			JoinedAt:        time.Now(),           // When they joined
			IsActive:        true,
		}

		return tgs.db.CreateTribeMembership(ctx, membership)
	}

	return nil // Still waiting for more votes
}

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

// generateUUID is a placeholder for UUID generation
func generateUUID() string {
	// Implementation would use actual UUID library
	return "generated-uuid"
}
