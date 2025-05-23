# Tribe Activities

## Activity History and Tracking

### Activity Tracking System

```go
// Enhanced activity tracking with tentative entries
type ActivityEntry struct {
    ID                string    `json:"id"`
    ListItemID        string    `json:"list_item_id"`
    UserID            string    `json:"user_id"`           // Who the activity is for
    TribeID           *string   `json:"tribe_id"`
    ActivityType      string    `json:"activity_type"`     // 'visited', 'watched', 'completed'
    ActivityStatus    string    `json:"activity_status"`   // 'confirmed', 'tentative', 'cancelled'
    CompletedAt       time.Time `json:"completed_at"`      // When it happened/will happen
    DurationMinutes   *int      `json:"duration_minutes"`
    Participants      []string  `json:"participants"`      // User IDs who participated
    Notes             string    `json:"notes"`
    RecordedByUserID  string    `json:"recorded_by_user_id"` // Who logged this
    DecisionSessionID *string   `json:"decision_session_id"` // If from decision result
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}

type ActivityService struct {
    db repository.Database
}

func (as *ActivityService) LogActivity(ctx context.Context, req LogActivityRequest) (*ActivityEntry, error) {
    entry := &ActivityEntry{
        ID:               generateUUID(),
        ListItemID:       req.ListItemID,
        UserID:           req.UserID,
        TribeID:          req.TribeID,
        ActivityType:     req.ActivityType,
        ActivityStatus:   req.ActivityStatus,
        CompletedAt:      req.CompletedAt,
        DurationMinutes:  req.DurationMinutes,
        Participants:     req.Participants,
        Notes:            req.Notes,
        RecordedByUserID: req.RecordedByUserID,
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }
    
    // Auto-determine status based on completion time
    if entry.ActivityStatus == "" {
        if entry.CompletedAt.After(time.Now()) {
            entry.ActivityStatus = "tentative"
        } else {
            entry.ActivityStatus = "confirmed"
        }
    }
    
    // Validate tribe membership
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

func (as *ActivityService) UpdateTentativeActivity(ctx context.Context, entryID, userID string, req UpdateActivityRequest) (*ActivityEntry, error) {
    entry, err := as.db.GetActivityEntry(ctx, entryID)
    if err != nil {
        return nil, err
    }
    
    // Only allow updates to tentative entries
    if entry.ActivityStatus != "tentative" {
        return nil, errors.New("can only update tentative activities")
    }
    
    // Verify user is in the tribe
    if entry.TribeID != nil {
        if err := as.validateTribeMembership(ctx, userID, *entry.TribeID); err != nil {
            return nil, err
        }
    }
    
    // Update fields
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
        entry.Notes = *req.Notes
    }
    
    entry.UpdatedAt = time.Now()
    
    if err := as.db.UpdateActivityEntry(ctx, entry); err != nil {
        return nil, err
    }
    
    return entry, nil
}

func (as *ActivityService) LogDecisionResult(ctx context.Context, sessionID, userID string, scheduledFor *time.Time) (*ActivityEntry, error) {
    session, err := as.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    if session.FinalSelection == nil {
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
        ListItemID:        *session.FinalSelection,
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

type LogActivityRequest struct {
    ListItemID        string     `json:"list_item_id"`
    UserID            string     `json:"user_id"`
    TribeID           *string    `json:"tribe_id"`
    ActivityType      string     `json:"activity_type"`
    ActivityStatus    string     `json:"activity_status"`
    CompletedAt       time.Time  `json:"completed_at"`
    DurationMinutes   *int       `json:"duration_minutes"`
    Participants      []string   `json:"participants"`
    Notes             string     `json:"notes"`
    RecordedByUserID  string     `json:"recorded_by_user_id"`
    DecisionSessionID *string    `json:"decision_session_id"`
}

type UpdateActivityRequest struct {
    ActivityStatus *string    `json:"activity_status"`
    CompletedAt    *time.Time `json:"completed_at"`
    Participants   []string   `json:"participants"`
    Notes          *string    `json:"notes"`
}
```

