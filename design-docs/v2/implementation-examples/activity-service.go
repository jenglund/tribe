package services

import (
	"context"
	"errors"
	"time"

	"tribe/internal/repository"
)

// ActivityService handles activity tracking and logging
//
// For complete type definitions, see: ../DATA-MODEL.md#activity-tracking-types
type ActivityService struct {
	db repository.Database
}

// NewActivityService creates a new activity service
func NewActivityService(db repository.Database) *ActivityService {
	return &ActivityService{db: db}
}

// LogActivity creates a new activity entry for a list item
func (as *ActivityService) LogActivity(ctx context.Context, req LogActivityRequest) (*ActivityEntry, error) {
	entry := &ActivityEntry{
		ID:                generateUUID(),
		ListItemID:        req.ListItemID,
		UserID:            req.UserID,
		TribeID:           req.TribeID,
		ActivityType:      req.ActivityType,
		ActivityStatus:    req.ActivityStatus,
		CompletedAt:       req.CompletedAt,
		DurationMinutes:   req.DurationMinutes,
		Participants:      req.Participants,
		Notes:             req.Notes,
		RecordedByUserID:  req.RecordedByUserID,
		DecisionSessionID: req.DecisionSessionID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Auto-determine status based on completion time
	if entry.ActivityStatus == "" {
		if entry.CompletedAt.After(time.Now()) {
			entry.ActivityStatus = "tentative"
		} else {
			entry.ActivityStatus = "confirmed"
		}
	}

	// Validate tribe membership if this is a tribe activity
	if req.TribeID != nil {
		if err := as.validateTribeMembership(ctx, req.RecordedByUserID, *req.TribeID); err != nil {
			return nil, err
		}
	}

	if err := as.db.CreateActivityEntry(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// UpdateTentativeActivity allows updating tentative activity entries
func (as *ActivityService) UpdateTentativeActivity(ctx context.Context, entryID, userID string, req UpdateActivityRequest) (*ActivityEntry, error) {
	entry, err := as.db.GetActivityEntry(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Only allow updates to tentative entries
	if entry.ActivityStatus != "tentative" {
		return nil, errors.New("can only update tentative activities")
	}

	// Verify user is in the tribe if this is a tribe activity
	if entry.TribeID != nil {
		if err := as.validateTribeMembership(ctx, userID, *entry.TribeID); err != nil {
			return nil, err
		}
	}

	// Update fields if provided
	if req.ActivityStatus != nil {
		entry.ActivityStatus = *req.ActivityStatus
	}
	if req.CompletedAt != nil {
		entry.CompletedAt = *req.CompletedAt
	}
	if req.Participants != nil {
		entry.Participants = req.Participants
	}
	if req.Notes != nil {
		entry.Notes = req.Notes
	}

	entry.UpdatedAt = time.Now()

	if err := as.db.UpdateActivityEntry(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// ConfirmTentativeActivity confirms a tentative activity
func (as *ActivityService) ConfirmTentativeActivity(ctx context.Context, entryID, userID string) (*ActivityEntry, error) {
	confirmedStatus := "confirmed"
	req := UpdateActivityRequest{
		ActivityStatus: &confirmedStatus,
	}

	return as.UpdateTentativeActivity(ctx, entryID, userID, req)
}

// CancelTentativeActivity cancels a tentative activity
func (as *ActivityService) CancelTentativeActivity(ctx context.Context, entryID, userID string) (*ActivityEntry, error) {
	cancelledStatus := "cancelled"
	req := UpdateActivityRequest{
		ActivityStatus: &cancelledStatus,
	}

	return as.UpdateTentativeActivity(ctx, entryID, userID, req)
}

// LogDecisionResult creates an activity entry for a completed decision session
func (as *ActivityService) LogDecisionResult(ctx context.Context, sessionID, userID string, scheduledFor *time.Time) (*ActivityEntry, error) {
	session, err := as.db.GetDecisionSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.FinalSelectionID == nil {
		return nil, errors.New("no final selection available")
	}

	// Get tribe members as default participants
	members, err := as.db.GetTribeMembers(ctx, session.TribeID)
	if err != nil {
		return nil, err
	}

	participants := make([]string, len(members))
	for i, member := range members {
		participants[i] = member.UserID
	}

	completedAt := time.Now()
	status := "confirmed"

	if scheduledFor != nil {
		completedAt = *scheduledFor
		if completedAt.After(time.Now()) {
			status = "tentative"
		}
	}

	req := LogActivityRequest{
		ListItemID:        *session.FinalSelectionID,
		UserID:            userID,
		TribeID:           &session.TribeID,
		ActivityType:      "visited", // Default, can be changed
		ActivityStatus:    status,
		CompletedAt:       completedAt,
		Participants:      participants,
		RecordedByUserID:  userID,
		DecisionSessionID: &sessionID,
	}

	return as.LogActivity(ctx, req)
}

// GetUserActivities retrieves activity history for a user
func (as *ActivityService) GetUserActivities(ctx context.Context, userID string, tribeID *string) ([]ActivityEntry, error) {
	return as.db.GetUserActivities(ctx, userID, tribeID)
}

// GetListItemActivities retrieves activity history for a specific list item
func (as *ActivityService) GetListItemActivities(ctx context.Context, listItemID string, tribeID *string) ([]ActivityEntry, error) {
	return as.db.GetListItemActivities(ctx, listItemID, tribeID)
}

// GetTentativeActivities retrieves all tentative activities for a tribe
func (as *ActivityService) GetTentativeActivities(ctx context.Context, tribeID string) ([]ActivityEntry, error) {
	return as.db.GetTentativeActivities(ctx, tribeID)
}

// DeleteActivity removes an activity entry
func (as *ActivityService) DeleteActivity(ctx context.Context, entryID, userID string) error {
	entry, err := as.db.GetActivityEntry(ctx, entryID)
	if err != nil {
		return err
	}

	// Only the recorder or tribe members can delete
	if entry.RecordedByUserID != userID {
		if entry.TribeID != nil {
			if err := as.validateTribeMembership(ctx, userID, *entry.TribeID); err != nil {
				return errors.New("only the recorder or tribe members can delete activities")
			}
		} else {
			return errors.New("only the recorder can delete personal activities")
		}
	}

	return as.db.DeleteActivityEntry(ctx, entryID)
}

// GetRecentActivities filters out items visited recently by user/tribe
func (as *ActivityService) GetRecentActivities(ctx context.Context, userID string, tribeID *string, days int) ([]string, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return as.db.GetRecentlyVisitedItems(ctx, userID, tribeID, cutoffDate)
}

// Helper function to validate tribe membership
func (as *ActivityService) validateTribeMembership(ctx context.Context, userID, tribeID string) error {
	isMember, err := as.db.IsUserTribeMember(ctx, userID, tribeID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this tribe")
	}
	return nil
}

// generateUUID is a placeholder for UUID generation
func generateUUID() string {
	// Implementation would use actual UUID library
	return "generated-uuid"
}
