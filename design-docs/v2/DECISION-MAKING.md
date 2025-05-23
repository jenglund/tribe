# Tribe Decision Making

## Decision-Making Algorithm (Enhanced)

### Advanced Filter Engine with Priority System

```go
// Filter Configuration
type FilterItem struct {
    ID          string      `json:"id"`           // Unique filter identifier
    Type        string      `json:"type"`         // "category", "dietary", "location", etc.
    IsHard      bool        `json:"is_hard"`      // true = required, false = preferred
    Priority    int         `json:"priority"`     // User-defined order (0 = highest)
    Criteria    interface{} `json:"criteria"`     // Type-specific filter data
    Description string      `json:"description"`  // Human-readable description
}

type FilterConfiguration struct {
    Items  []FilterItem `json:"items"`
    UserID string       `json:"user_id"`
}

// Filter Results with Violation Tracking
type FilterResult struct {
    Item                ListItem           `json:"item"`
    PassedHardFilters   bool              `json:"passed_hard_filters"`
    SoftFilterResults   []SoftFilterResult `json:"soft_filter_results"`
    ViolationCount      int               `json:"violation_count"`
    PriorityScore       float64           `json:"priority_score"`
}

type SoftFilterResult struct {
    FilterID    string `json:"filter_id"`
    FilterType  string `json:"filter_type"`
    Passed      bool   `json:"passed"`
    Priority    int    `json:"priority"`
    Description string `json:"description"`
}

// Enhanced Filter Engine
type FilterEngine struct {
    db repository.Database
}

func (fe *FilterEngine) ApplyFiltersWithPriority(ctx context.Context, items []ListItem, config FilterConfiguration) ([]FilterResult, error) {
    results := make([]FilterResult, 0, len(items))
    
    // Sort filters by priority (lowest number = highest priority)
    sort.Slice(config.Items, func(i, j int) bool {
        return config.Items[i].Priority < config.Items[j].Priority
    })
    
    for _, item := range items {
        result := FilterResult{
            Item:              item,
            PassedHardFilters: true,
            SoftFilterResults: make([]SoftFilterResult, 0),
            ViolationCount:    0,
        }
        
        // Apply filters in priority order
        for _, filter := range config.Items {
            passed := fe.evaluateFilter(ctx, item, filter)
            
            if filter.IsHard {
                if !passed {
                    result.PassedHardFilters = false
                    break // Skip this item entirely
                }
            } else {
                // Soft filter - track result but don't exclude
                softResult := SoftFilterResult{
                    FilterID:    filter.ID,
                    FilterType:  filter.Type,
                    Passed:      passed,
                    Priority:    filter.Priority,
                    Description: filter.Description,
                }
                result.SoftFilterResults = append(result.SoftFilterResults, softResult)
                
                if !passed {
                    result.ViolationCount++
                }
            }
        }
        
        // Only include items that passed all hard filters
        if result.PassedHardFilters {
            result.PriorityScore = fe.calculatePriorityScore(result.SoftFilterResults)
            results = append(results, result)
        }
    }
    
    // Sort results: higher priority score = better match
    sort.Slice(results, func(i, j int) bool {
        if results[i].PriorityScore != results[j].PriorityScore {
            return results[i].PriorityScore > results[j].PriorityScore
        }
        // Tie-breaker: fewer violations is better
        return results[i].ViolationCount < results[j].ViolationCount
    })
    
    return results, nil
}

func (fe *FilterEngine) calculatePriorityScore(softResults []SoftFilterResult) float64 {
    score := 0.0
    totalWeight := 0.0
    
    for _, result := range softResults {
        // Earlier filters (lower priority number) have higher weight
        weight := 1.0 / float64(result.Priority + 1)
        totalWeight += weight
        
        if result.Passed {
            score += weight
        }
    }
    
    if totalWeight == 0 {
        return 1.0 // No soft filters = perfect score
    }
    
    return score / totalWeight
}

func (fe *FilterEngine) evaluateFilter(ctx context.Context, item ListItem, filter FilterItem) bool {
    switch filter.Type {
    case "category":
        criteria := filter.Criteria.(CategoryFilterCriteria)
        return fe.evaluateCategoryFilter(item, criteria)
    case "dietary":
        criteria := filter.Criteria.(DietaryFilterCriteria)
        return fe.evaluateDietaryFilter(item, criteria)
    case "location":
        criteria := filter.Criteria.(LocationFilterCriteria)
        return fe.evaluateLocationFilter(item, criteria)
    case "recent_activity":
        criteria := filter.Criteria.(RecentActivityFilterCriteria)
        return fe.evaluateRecentActivityFilter(ctx, item, criteria)
    case "opening_hours":
        criteria := filter.Criteria.(OpeningHoursFilterCriteria)
        return fe.evaluateOpeningHoursFilter(item, criteria)
    case "tags":
        criteria := filter.Criteria.(TagFilterCriteria)
        return fe.evaluateTagFilter(item, criteria)
    default:
        return true // Unknown filter types pass by default
    }
}

// Specific filter criteria types
type CategoryFilterCriteria struct {
    IncludeCategories []string `json:"include_categories"`
    ExcludeCategories []string `json:"exclude_categories"`
}

type DietaryFilterCriteria struct {
    RequiredOptions []string `json:"required_options"` // ["vegetarian", "vegan", "gluten_free"]
}

type LocationFilterCriteria struct {
    CenterLat    float64 `json:"center_lat"`
    CenterLng    float64 `json:"center_lng"`
    MaxDistance  float64 `json:"max_distance"` // in miles
}

type RecentActivityFilterCriteria struct {
    ExcludeDays int      `json:"exclude_days"`
    UserID      string   `json:"user_id"`
    TribeID     *string  `json:"tribe_id"`
}

type OpeningHoursFilterCriteria struct {
    MustBeOpenFor    int    `json:"must_be_open_for"`    // minutes from now
    MustBeOpenUntil  string `json:"must_be_open_until"`  // time like "23:00" in user's timezone
    UserTimezone     string `json:"user_timezone"`       // user's timezone for calculations
    CheckDate        *int64 `json:"check_date"`          // unix timestamp, defaults to now
}

type TagFilterCriteria struct {
    RequiredTags []string `json:"required_tags"`
    ExcludedTags []string `json:"excluded_tags"`
}

// Enhanced opening hours filter with timezone support
func (fe *FilterEngine) evaluateOpeningHoursFilter(item ListItem, criteria OpeningHoursFilterCriteria) bool {
    businessInfo := item.BusinessInfo
    if businessInfo == nil {
        return true // No business info = assume always open
    }
    
    regularHours, exists := businessInfo["regular_hours"]
    if !exists {
        return true // No hours specified = assume always open
    }
    
    businessTimezone, _ := businessInfo["timezone"].(string)
    if businessTimezone == "" {
        businessTimezone = "UTC"
    }
    
    // Use provided check date or current time
    checkTime := time.Now()
    if criteria.CheckDate != nil {
        checkTime = time.Unix(*criteria.CheckDate, 0)
    }
    
    // Convert to business timezone
    businessLocation, err := time.LoadLocation(businessTimezone)
    if err != nil {
        businessLocation = time.UTC
    }
    businessTime := checkTime.In(businessLocation)
    
    // Get today's day of week
    dayName := strings.ToLower(businessTime.Weekday().String())
    
    hours, ok := regularHours.(map[string]interface{})[dayName]
    if !ok {
        return true // Day not specified = assume open
    }
    
    dayHours, ok := hours.(map[string]interface{})
    if !ok {
        return true
    }
    
    // Check if closed today
    if closed, ok := dayHours["closed"].(bool); ok && closed {
        return false
    }
    
    openTimeStr, ok1 := dayHours["open"].(string)
    closeTimeStr, ok2 := dayHours["close"].(string)
    if !ok1 || !ok2 {
        return true // No times specified = assume always open
    }
    
    // Parse business hours for today
    openTime, err1 := parseTimeInLocation(openTimeStr, businessTime, businessLocation)
    closeTime, err2 := parseTimeInLocation(closeTimeStr, businessTime, businessLocation)
    if err1 != nil || err2 != nil {
        return true // Parse error = assume open
    }
    
    // Handle closing time after midnight
    if closeTime.Before(openTime) {
        closeTime = closeTime.Add(24 * time.Hour)
    }
    
    currentTime := businessTime
    
    // Check if currently open
    if currentTime.Before(openTime) || currentTime.After(closeTime) {
        return false // Currently closed
    }
    
    // Check "must be open for X minutes"
    if criteria.MustBeOpenFor > 0 {
        requiredUntil := currentTime.Add(time.Duration(criteria.MustBeOpenFor) * time.Minute)
        if requiredUntil.After(closeTime) {
            return false // Won't be open long enough
        }
    }
    
    // Check "must be open until specific time"
    if criteria.MustBeOpenUntil != "" {
        // Convert user's desired time to business timezone
        userLocation, err := time.LoadLocation(criteria.UserTimezone)
        if err != nil {
            userLocation = time.UTC
        }
        
        userTime := checkTime.In(userLocation)
        requiredUntilTime, err := parseTimeInLocation(criteria.MustBeOpenUntil, userTime, userLocation)
        if err == nil {
            // Convert user's time requirement to business timezone
            requiredUntilInBusiness := requiredUntilTime.In(businessLocation)
            
            // Adjust to same day as current business time
            requiredUntilInBusiness = time.Date(
                businessTime.Year(), businessTime.Month(), businessTime.Day(),
                requiredUntilInBusiness.Hour(), requiredUntilInBusiness.Minute(), 0, 0,
                businessLocation,
            )
            
            if requiredUntilInBusiness.After(closeTime) {
                return false // Won't be open until required time
            }
        }
    }
    
    return true // Passes all time-based criteria
}

// Helper function to parse time string in specific location
func parseTimeInLocation(timeStr string, referenceTime time.Time, location *time.Location) (time.Time, error) {
    // Handle 24:00 as midnight next day
    if timeStr == "24:00" {
        return time.Date(
            referenceTime.Year(), referenceTime.Month(), referenceTime.Day(),
            0, 0, 0, 0, location,
        ).Add(24 * time.Hour), nil
    }
    
    parsedTime, err := time.Parse("15:04", timeStr)
    if err != nil {
        return time.Time{}, err
    }
    
    return time.Date(
        referenceTime.Year(), referenceTime.Month(), referenceTime.Day(),
        parsedTime.Hour(), parsedTime.Minute(), 0, 0, location,
    ), nil
}

// Time-based filter configuration for frontend
type TimeBasedFilterConfig struct {
    MustBeOpenFor    *int    `json:"must_be_open_for"`    // minutes
    MustBeOpenUntil  *string `json:"must_be_open_until"`  // "HH:MM" in user timezone
    UserTimezone     string  `json:"user_timezone"`       // user's timezone
    CheckDate        *string `json:"check_date"`          // ISO date string, optional
}

// Convert frontend config to backend criteria
func (config TimeBasedFilterConfig) ToCriteria() OpeningHoursFilterCriteria {
    criteria := OpeningHoursFilterCriteria{
        UserTimezone: config.UserTimezone,
    }
    
    if config.MustBeOpenFor != nil {
        criteria.MustBeOpenFor = *config.MustBeOpenFor
    }
    
    if config.MustBeOpenUntil != nil {
        criteria.MustBeOpenUntil = *config.MustBeOpenUntil
    }
    
    if config.CheckDate != nil {
        if checkTime, err := time.Parse("2006-01-02", *config.CheckDate); err == nil {
            timestamp := checkTime.Unix()
            criteria.CheckDate = &timestamp
        }
    }
    
    return criteria
}
```

### Filter Configuration Storage

User filter configurations are stored in the `filter_configurations` table defined in [`DATA-MODEL.md`](./DATA-MODEL.md). This allows users to save commonly used filter combinations with names like "Date Night Filters", "Quick Lunch", etc.

### Frontend Filter Management

```typescript
// Filter Management Interface
interface FilterItem {
  id: string;
  type: 'category' | 'dietary' | 'location' | 'recent_activity' | 'opening_hours' | 'tags';
  isHard: boolean;
  priority: number;
  criteria: any;
  description: string;
}

interface FilterConfiguration {
  items: FilterItem[];
  userId: string;
}

// Filter Builder Component Props
interface FilterBuilderProps {
  configuration: FilterConfiguration;
  onConfigurationChange: (config: FilterConfiguration) => void;
  availableFilters: FilterType[];
}

// Filter Result Display
interface FilterResultDisplayProps {
  results: FilterResult[];
  showViolations: boolean;
  onItemSelect: (item: ListItem) => void;
}

// Enhanced GraphQL types for filtering
type FilterResult {
  item: ListItem!
  passedHardFilters: Boolean!
  softFilterResults: [SoftFilterResult!]!
  violationCount: Int!
  priorityScore: Float!
}

type SoftFilterResult {
  filterId: String!
  filterType: String!
  passed: Boolean!
  priority: Int!
  description: String!
}
```

### Default Parameters
```go
type DecisionDefaults struct {
    K int `json:"k"` // Default: 2 eliminations per person
    M int `json:"m"` // Default: 3 final options for random selection
    N int `json:"n"` // Always equals tribe size (1-8)
}

type TribeDecisionPreferences struct {
    DefaultK int `json:"default_k"` // Configurable per tribe
    DefaultM int `json:"default_m"` // Configurable per tribe
    MaxK     int `json:"max_k"`     // Maximum eliminations allowed
    MaxM     int `json:"max_m"`     // Maximum final options allowed
}
```

### Turn-Based Elimination Algorithm

```go
// Turn-based elimination system with quick-skip
type EliminationSession struct {
    SessionID           string         `json:"session_id"`
    TribeID             string         `json:"tribe_id"`
    EliminationOrder    []string       `json:"elimination_order"`    // Randomized user IDs
    CurrentTurnIndex    int            `json:"current_turn_index"`   // Index in elimination_order
    CurrentRound        int            `json:"current_round"`        // Which round (1 to K)
    TurnStartedAt       time.Time      `json:"turn_started_at"`
    TurnTimeoutMinutes  int            `json:"turn_timeout_minutes"`
    SkippedUsers        []SkippedTurn  `json:"skipped_users"`
    UserSkipCounts      map[string]int `json:"user_skip_counts"`     // Track quick-skip usage
    CurrentCandidates   []string       `json:"current_candidates"`   // Remaining item IDs
}

type SkippedTurn struct {
    UserID     string    `json:"user_id"`
    Round      int       `json:"round"`
    TurnInRound int      `json:"turn_in_round"`
    SkipType   string    `json:"skip_type"`    // "quick_skip", "timeout_skip", "forfeited"
    SkippedAt  time.Time `json:"skipped_at"`
}

// Quick-skip functionality
func (ea *EliminationAlgorithm) ProcessQuickSkip(ctx context.Context, sessionID, userID string) (*EliminationSession, error) {
    session, err := ea.getEliminationSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Verify it's this user's turn
    currentUserID := session.EliminationOrder[session.CurrentTurnIndex]
    if currentUserID != userID {
        return nil, errors.New("not your turn")
    }
    
    // Get algorithm parameters
    params, err := ea.getAlgorithmParams(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check if user has exceeded quick-skip limit
    userSkipCount := session.UserSkipCounts[userID]
    if userSkipCount >= params.K {
        return nil, errors.New("maximum quick-skips exceeded")
    }
    
    // Check if this turn was already deferred
    currentTurn := SkippedTurn{
        UserID:      userID,
        Round:       session.CurrentRound,
        TurnInRound: session.CurrentTurnIndex,
    }
    
    if ea.wasTurnAlreadyDeferred(session.SkippedUsers, currentTurn) {
        return nil, errors.New("cannot quick-skip a turn that was already deferred")
    }
    
    // Record the quick-skip
    skippedTurn := SkippedTurn{
        UserID:      userID,
        Round:       session.CurrentRound,
        TurnInRound: session.CurrentTurnIndex,
        SkipType:    "quick_skip",
        SkippedAt:   time.Now(),
    }
    session.SkippedUsers = append(session.SkippedUsers, skippedTurn)
    
    // Increment user's quick-skip count
    if session.UserSkipCounts == nil {
        session.UserSkipCounts = make(map[string]int)
    }
    session.UserSkipCounts[userID]++
    
    // Advance to next turn
    return ea.advanceToNextTurn(ctx, session)
}

// Check if a turn was already deferred
func (ea *EliminationAlgorithm) wasTurnAlreadyDeferred(skippedUsers []SkippedTurn, currentTurn SkippedTurn) bool {
    for _, skip := range skippedUsers {
        if skip.UserID == currentTurn.UserID && 
           skip.Round == currentTurn.Round && 
           skip.TurnInRound == currentTurn.TurnInRound {
            return true
        }
    }
    return false
}

// Enhanced timeout handling
func (ea *EliminationAlgorithm) HandleTimeout(ctx context.Context, sessionID string) (*EliminationSession, error) {
    session, err := ea.getEliminationSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check if current turn has timed out
    timeout := time.Duration(session.TurnTimeoutMinutes) * time.Minute
    if time.Since(session.TurnStartedAt) < timeout {
        return session, nil // Not timed out yet
    }
    
    currentUserID := session.EliminationOrder[session.CurrentTurnIndex]
    
    // Determine if this is a catch-up phase timeout
    isCatchUpPhase := ea.isCatchUpPhase(session)
    
    if isCatchUpPhase {
        // Forfeit all remaining skipped rounds for this user
        return ea.forfeitRemainingSkips(ctx, session, currentUserID)
    } else {
        // Regular timeout-skip
        skippedTurn := SkippedTurn{
            UserID:      currentUserID,
            Round:       session.CurrentRound,
            TurnInRound: session.CurrentTurnIndex,
            SkipType:    "timeout_skip",
            SkippedAt:   time.Now(),
        }
        session.SkippedUsers = append(session.SkippedUsers, skippedTurn)
        
        // Advance to next turn
        return ea.advanceToNextTurn(ctx, session)
    }
}

// Forfeit all remaining skipped rounds for a user
func (ea *EliminationAlgorithm) forfeitRemainingSkips(ctx context.Context, session *EliminationSession, userID string) (*EliminationSession, error) {
    // Mark all remaining skipped turns for this user as forfeited
    for i := range session.SkippedUsers {
        if session.SkippedUsers[i].UserID == userID && 
           (session.SkippedUsers[i].SkipType == "quick_skip" || session.SkippedUsers[i].SkipType == "timeout_skip") {
            session.SkippedUsers[i].SkipType = "forfeited"
        }
    }
    
    // Continue with catch-up phase for other users
    return ea.continueCatchUpPhase(ctx, session)
}

// Check if we're in catch-up phase
func (ea *EliminationAlgorithm) isCatchUpPhase(session *EliminationSession) bool {
    params, err := ea.getAlgorithmParams(context.Background(), session.SessionID)
    if err != nil {
        return false
    }
    
    // We're in catch-up if we've completed all regular rounds and have pending skips
    return session.CurrentRound > params.K && len(session.SkippedUsers) > 0
}

// Enhanced elimination status with quick-skip info
func (ea *EliminationAlgorithm) GetEliminationStatus(ctx context.Context, sessionID, userID string) (*EliminationStatus, error) {
    session, err := ea.getEliminationSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check for timeout
    session, _ = ea.HandleTimeout(ctx, sessionID)
    
    currentUserID := ""
    if len(session.EliminationOrder) > 0 {
        currentUserID = session.EliminationOrder[session.CurrentTurnIndex]
    }
    
    // Get algorithm parameters for skip limits
    params, err := ea.getAlgorithmParams(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check if current turn was already deferred
    currentTurn := SkippedTurn{
        UserID:      userID,
        Round:       session.CurrentRound,
        TurnInRound: session.CurrentTurnIndex,
    }
    canQuickSkip := !ea.wasTurnAlreadyDeferred(session.SkippedUsers, currentTurn)
    
    // Check quick-skip limit
    userSkipCount := session.UserSkipCounts[userID]
    if userSkipCount >= params.K {
        canQuickSkip = false
    }
    
    return &EliminationStatus{
        SessionID:         sessionID,
        CurrentCandidates: session.CurrentCandidates,
        CurrentUserTurn:   currentUserID,
        IsYourTurn:        currentUserID == userID,
        CurrentRound:      session.CurrentRound,
        TurnTimeRemaining: ea.calculateTimeRemaining(session),
        EliminationOrder:  session.EliminationOrder,
        SkippedUsers:      session.SkippedUsers,
        CanQuickSkip:      canQuickSkip && currentUserID == userID,
        QuickSkipsUsed:    userSkipCount,
        QuickSkipsLimit:   params.K,
        IsCatchUpPhase:    ea.isCatchUpPhase(session),
    }, nil
}

type EliminationStatus struct {
    SessionID         string        `json:"session_id"`
    CurrentCandidates []string      `json:"current_candidates"`
    CurrentUserTurn   string        `json:"current_user_turn"`
    IsYourTurn        bool          `json:"is_your_turn"`
    CurrentRound      int           `json:"current_round"`
    TurnTimeRemaining time.Duration `json:"turn_time_remaining"`
    EliminationOrder  []string      `json:"elimination_order"`
    SkippedUsers      []SkippedTurn `json:"skipped_users"`
    CanQuickSkip      bool          `json:"can_quick_skip"`
    QuickSkipsUsed    int           `json:"quick_skips_used"`
    QuickSkipsLimit   int           `json:"quick_skips_limit"`
    IsCatchUpPhase    bool          `json:"is_catch_up_phase"`
}
```

### Decision Session Management with Timeouts

```go
// Enhanced decision session service with timeout management
type DecisionSessionService struct {
    db repository.Database
}

// Update session activity (called on any user interaction)
func (dss *DecisionSessionService) UpdateSessionActivity(ctx context.Context, sessionID string) error {
    return dss.db.UpdateSessionActivity(ctx, sessionID, time.Now())
}

// Check for expired sessions (run periodically)
func (dss *DecisionSessionService) ProcessExpiredSessions(ctx context.Context) error {
    cutoffTime := time.Now().Add(-30 * time.Minute)
    expiredSessions, err := dss.db.GetInactiveSessionsSince(ctx, cutoffTime)
    if err != nil {
        return err
    }
    
    for _, session := range expiredSessions {
        if err := dss.expireSession(ctx, session.ID); err != nil {
            log.Printf("Failed to expire session %s: %v", session.ID, err)
        }
    }
    
    return nil
}

// Expire a session due to inactivity
func (dss *DecisionSessionService) expireSession(ctx context.Context, sessionID string) error {
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    // Only expire active sessions
    if session.Status != "configuring" && session.Status != "eliminating" {
        return nil
    }
    
    session.Status = "expired"
    session.CompletedAt = &time.Now()
    
    return dss.db.UpdateDecisionSession(ctx, session)
}

// Complete session and set up history retention
func (dss *DecisionSessionService) CompleteSession(ctx context.Context, sessionID string, finalResult *DecisionResult) error {
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    session.Status = "completed"
    session.CompletedAt = &time.Now()
    session.ExpiresAt = &time.Now().Add(30 * 24 * time.Hour) // 1 month retention
    session.FinalSelectionID = finalResult.WinnerID
    session.RunnersUp = finalResult.RunnersUpIDs
    session.EliminationHistory = finalResult.EliminationHistory
    
    return dss.db.UpdateDecisionSession(ctx, session)
}

// Pin session to prevent automatic cleanup
func (dss *DecisionSessionService) PinSession(ctx context.Context, sessionID, userID string) error {
    // Verify user has access to this session
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    // Validate user is tribe member
    isMember, err := dss.db.IsUserTribeMember(ctx, userID, session.TribeID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("user is not a member of this tribe")
    }
    
    session.IsPinned = true
    return dss.db.UpdateDecisionSession(ctx, session)
}

// Get formatted decision history for display
func (dss *DecisionSessionService) GetDecisionHistory(ctx context.Context, sessionID, userID string) (*DecisionHistory, error) {
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Verify user has access
    isMember, err := dss.db.IsUserTribeMember(ctx, userID, session.TribeID)
    if err != nil {
        return nil, err
    }
    if !isMember {
        return nil, errors.New("user is not a member of this tribe")
    }
    
    // Get tribe settings for elimination visibility
    tribe, err := dss.db.GetTribe(ctx, session.TribeID)
    if err != nil {
        return nil, err
    }
    
    history := &DecisionHistory{
        SessionID:     sessionID,
        SessionName:   session.Name,
        Status:        session.Status,
        CreatedAt:     session.CreatedAt,
        CompletedAt:   session.CompletedAt,
        IsPinned:      session.IsPinned,
        ShowDetails:   tribe.ShowEliminationDetails,
    }
    
    // Get list items for display
    if session.FinalSelectionID != nil {
        winner, err := dss.db.GetListItem(ctx, *session.FinalSelectionID)
        if err == nil {
            history.Winner = winner
        }
    }
    
    // Get runners-up
    if len(session.RunnersUp) > 0 {
        runnersUp, err := dss.db.GetListItems(ctx, session.RunnersUp)
        if err == nil {
            history.RunnersUp = runnersUp
        }
    }
    
    // Parse elimination history (stored as JSON)
    var eliminations []EliminationHistoryEntry
    if err := json.Unmarshal(session.EliminationHistory, &eliminations); err == nil {
        // Reverse order for display (most recent elimination first)
        for i := len(eliminations) - 1; i >= 0; i-- {
            elimination := eliminations[i]
            
            // Get list item details
            item, err := dss.db.GetListItem(ctx, elimination.ItemID)
            if err != nil {
                continue
            }
            
            historyEntry := DecisionHistoryEntry{
                Item:          item,
                EliminatedAt:  elimination.EliminatedAt,
                Round:         elimination.Round,
                EliminationOrder: len(eliminations) - i, // 1-based order
            }
            
            // Include eliminator info if visibility enabled
            if tribe.ShowEliminationDetails {
                eliminator, err := dss.db.GetUser(ctx, elimination.EliminatorID)
                if err == nil {
                    // Get tribe-specific display name
                    membership, err := dss.db.GetTribeMembership(ctx, session.TribeID, elimination.EliminatorID)
                    if err == nil && membership.TribeDisplayName != "" {
                        historyEntry.EliminatorName = membership.TribeDisplayName
                    } else {
                        historyEntry.EliminatorName = eliminator.DisplayName
                    }
                }
            }
            
            history.Eliminations = append(history.Eliminations, historyEntry)
        }
    }
    
    return history, nil
}

// Clean up expired session histories (run daily)
func (dss *DecisionSessionService) CleanupExpiredHistories(ctx context.Context) error {
    expiredSessions, err := dss.db.GetExpiredUnpinnedSessions(ctx)
    if err != nil {
        return err
    }
    
    for _, session := range expiredSessions {
        if err := dss.db.DeleteDecisionSession(ctx, session.ID); err != nil {
            log.Printf("Failed to delete expired session %s: %v", session.ID, err)
        }
    }
    
    return nil
}

// Mobile-optimized session status for polling
func (dss *DecisionSessionService) GetMobileSessionStatus(ctx context.Context, sessionID, userID string) (*MobileSessionStatus, error) {
    // Update activity when checking status
    if err := dss.UpdateSessionActivity(ctx, sessionID); err != nil {
        return nil, err
    }
    
    session, err := dss.db.GetDecisionSession(ctx, sessionID)
    if err != nil {
        return nil, err
    }
    
    // Check if session has expired
    if session.Status == "configuring" || session.Status == "eliminating" {
        inactiveTime := time.Since(session.LastActivityAt)
        if inactiveTime > time.Duration(session.SessionTimeoutMinutes)*time.Minute {
            // Expire the session
            if err := dss.expireSession(ctx, sessionID); err != nil {
                return nil, err
            }
            session.Status = "expired"
        }
    }
    
    status := &MobileSessionStatus{
        SessionID:         sessionID,
        Status:            session.Status,
        CurrentCandidates: []ListItemSummary{}, // Simplified for mobile
        IsYourTurn:        false,
        TurnTimeRemaining: 0,
        SessionTimeRemaining: 0,
    }
    
    // Calculate remaining time
    if session.Status == "eliminating" {
        sessionTimeout := time.Duration(session.SessionTimeoutMinutes) * time.Minute
        elapsed := time.Since(session.LastActivityAt)
        if elapsed < sessionTimeout {
            status.SessionTimeRemaining = int((sessionTimeout - elapsed).Seconds())
        }
        
        // Check if it's user's turn
        if len(session.EliminationOrder) > 0 {
            currentUserID := session.EliminationOrder[session.CurrentTurnIndex]
            status.IsYourTurn = (currentUserID == userID)
            
            if status.IsYourTurn && session.TurnStartedAt != nil {
                turnTimeout := time.Duration(session.TurnTimeoutMinutes) * time.Minute
                turnElapsed := time.Since(*session.TurnStartedAt)
                if turnElapsed < turnTimeout {
                    status.TurnTimeRemaining = int((turnTimeout - turnElapsed).Seconds())
                }
            }
        }
    }
    
    // Get simplified candidate list for mobile
    if len(session.CurrentCandidates) > 0 {
        items, err := dss.db.GetListItems(ctx, session.CurrentCandidates)
        if err == nil {
            for _, item := range items {
                status.CurrentCandidates = append(status.CurrentCandidates, ListItemSummary{
                    ID:          item.ID,
                    Name:        item.Name,
                    Category:    item.Category,
                    Tags:        item.Tags[:min(3, len(item.Tags))], // Limit tags for mobile
                })
            }
        }
    }
    
    return status, nil
}

// Data structures for session management
type DecisionResult struct {
    WinnerID           *string                    `json:"winner_id"`
    RunnersUpIDs       []string                   `json:"runners_up_ids"`
    EliminationHistory json.RawMessage            `json:"elimination_history"`
}

type EliminationHistoryEntry struct {
    ItemID        string    `json:"item_id"`
    EliminatorID  string    `json:"eliminator_id"`
    Round         int       `json:"round"`
    EliminatedAt  time.Time `json:"eliminated_at"`
}

type DecisionHistory struct {
    SessionID     string                    `json:"session_id"`
    SessionName   string                    `json:"session_name"`
    Status        string                    `json:"status"`
    CreatedAt     time.Time                 `json:"created_at"`
    CompletedAt   *time.Time                `json:"completed_at"`
    IsPinned      bool                      `json:"is_pinned"`
    ShowDetails   bool                      `json:"show_details"`
    Winner        *ListItem                 `json:"winner"`
    RunnersUp     []ListItem                `json:"runners_up"`
    Eliminations  []DecisionHistoryEntry    `json:"eliminations"`
}

type DecisionHistoryEntry struct {
    Item             *ListItem `json:"item"`
    EliminatorName   string    `json:"eliminator_name"`
    EliminatedAt     time.Time `json:"eliminated_at"`
    Round            int       `json:"round"`
    EliminationOrder int       `json:"elimination_order"` // 1 = last eliminated, 2 = second-to-last, etc.
}

type MobileSessionStatus struct {
    SessionID            string             `json:"session_id"`
    Status               string             `json:"status"`
    CurrentCandidates    []ListItemSummary  `json:"current_candidates"`
    IsYourTurn           bool               `json:"is_your_turn"`
    TurnTimeRemaining    int                `json:"turn_time_remaining"`    // seconds
    SessionTimeRemaining int                `json:"session_time_remaining"` // seconds
}

type ListItemSummary struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Category string   `json:"category"`
    Tags     []string `json:"tags"`
}