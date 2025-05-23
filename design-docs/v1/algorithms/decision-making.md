# Decision-Making Algorithm

**Related Documents:**
- [Database Schema](../database-schema.md) - Decision session tables
- [API Design](../api-design.md) - Decision-related GraphQL types
- [Filtering System](./filtering.md) - How filtering integrates with decisions

## Overview

The Tribe decision-making system uses a **KN+M elimination algorithm** where:
- **K**: Number of elimination rounds per person
- **N**: Number of participants 
- **M**: Final set size for random selection
- **Quick-Skip**: Users can defer their turn up to K times

## Algorithm Flow

### Phase 1: Setup and Filtering
1. **Session Creation**: User creates decision session for their tribe
2. **List Selection**: Choose which lists to include
3. **Filter Application**: Apply filters to narrow down candidates
4. **Parameter Configuration**: Set K, M values (with suggestions)
5. **Participant Confirmation**: All tribe members join the session

### Phase 2: Elimination Rounds
1. **Turn Order**: Randomize user elimination order
2. **Round-Robin Elimination**: Each user eliminates K items
3. **Quick-Skip Support**: Users can defer turns (limited to K total)
4. **Timeout Handling**: Auto-skip after timeout
5. **Catch-Up Phase**: Complete any deferred eliminations

### Phase 3: Final Selection
1. **Random Selection**: Choose winner from final M candidates
2. **Result Logging**: Record decision and activity history
3. **Session Cleanup**: Archive or delete based on settings

## Detailed Algorithm Specification

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

### Parameter Suggestion Algorithm

```go
func SuggestKMValues(candidateCount, tribeSize int) []AlgorithmParams {
    suggestions := []AlgorithmParams{}
    
    // Target: Reduce to 3-8 final candidates for good randomness
    targetFinal := min(max(3, tribeSize), 8)
    
    for k := 1; k <= 5; k++ {
        totalEliminations := k * tribeSize
        remaining := candidateCount - totalEliminations
        
        if remaining >= targetFinal && remaining <= targetFinal*2 {
            suggestions = append(suggestions, AlgorithmParams{
                K: k,
                N: tribeSize,
                M: min(remaining, targetFinal),
                InitialCount: candidateCount,
                Description: fmt.Sprintf("%d-%d-%d: %d rounds, %d final", 
                    candidateCount, totalEliminations, remaining, k, remaining),
            })
        }
    }
    
    // Always include default if no good suggestions
    if len(suggestions) == 0 {
        k := min(2, candidateCount/(tribeSize*2))
        m := min(3, candidateCount-(k*tribeSize))
        suggestions = append(suggestions, AlgorithmParams{
            K: k, N: tribeSize, M: m, InitialCount: candidateCount,
            Description: fmt.Sprintf("Default %d-%d-%d", candidateCount, k*tribeSize, m),
        })
    }
    
    return suggestions
}
```

## Turn-Based Elimination with Quick-Skip

### Elimination Session State

```go
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
```

### Quick-Skip Implementation

```go
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
```

### Timeout Handling

```go
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
```

### Catch-Up Phase

The catch-up phase handles deferred eliminations after all regular rounds complete:

```go
func (ea *EliminationAlgorithm) isCatchUpPhase(session *EliminationSession) bool {
    params, err := ea.getAlgorithmParams(context.Background(), session.SessionID)
    if err != nil {
        return false
    }
    
    // We're in catch-up if we've completed all regular rounds and have pending skips
    return session.CurrentRound > params.K && len(session.SkippedUsers) > 0
}

func (ea *EliminationAlgorithm) continueCatchUpPhase(ctx context.Context, session *EliminationSession) (*EliminationSession, error) {
    // Find next user with pending skipped turns
    for _, skip := range session.SkippedUsers {
        if skip.SkipType == "quick_skip" || skip.SkipType == "timeout_skip" {
            // Set this user as current turn
            session.CurrentTurnIndex = ea.findUserIndexInOrder(session.EliminationOrder, skip.UserID)
            session.CurrentRound = skip.Round
            session.TurnStartedAt = time.Now()
            
            return session, nil
        }
    }
    
    // No more catch-up turns needed - proceed to final selection
    return ea.proceedToFinalSelection(ctx, session)
}
```

## Elimination Status API

```go
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
```

## Final Selection and Random Choice

```go
func (ea *EliminationAlgorithm) proceedToFinalSelection(ctx context.Context, session *EliminationSession) (*EliminationSession, error) {
    params, err := ea.getAlgorithmParams(ctx, session.SessionID)
    if err != nil {
        return nil, err
    }
    
    // Ensure we have the right number of candidates
    if len(session.CurrentCandidates) > params.M {
        // Continue elimination until we reach M candidates
        return ea.continueElimination(ctx, session)
    }
    
    if len(session.CurrentCandidates) == 0 {
        return nil, errors.New("no candidates remaining")
    }
    
    // Random selection from final candidates
    rand.Seed(time.Now().UnixNano())
    winnerIndex := rand.Intn(len(session.CurrentCandidates))
    winner := session.CurrentCandidates[winnerIndex]
    
    // Set runners-up (all other final candidates)
    runnersUp := make([]string, 0, len(session.CurrentCandidates)-1)
    for i, candidate := range session.CurrentCandidates {
        if i != winnerIndex {
            runnersUp = append(runnersUp, candidate)
        }
    }
    
    // Update session with final result
    session.Status = "completed"
    session.FinalSelection = &winner
    session.RunnersUp = runnersUp
    session.CompletedAt = time.Now()
    
    return ea.saveDecisionSession(ctx, session)
}
```

## Edge Cases and Error Handling

### Insufficient Candidates
- If filtering results in fewer candidates than M, proceed with available candidates
- Minimum 1 candidate required to complete decision
- Show warning to users about limited options

### User Disconnection
- Timeout system handles absent users automatically
- Deferred turns can be completed when user returns
- Forfeiture option for permanently absent users

### Session Timeout
- Overall session timeout (30 minutes default)
- Automatic expiration for inactive sessions
- Option to extend timeout if users are active

### Concurrent Modifications
- Database-level concurrency control
- Optimistic locking for session updates
- Conflict resolution for simultaneous actions

## Testing Strategy

### Unit Tests
- Parameter suggestion algorithm with various inputs
- Quick-skip logic with edge cases
- Timeout handling scenarios
- Random selection verification

### Integration Tests
- Complete elimination flow with multiple users
- Real-time updates via WebSocket
- Database consistency during failures

### Performance Tests
- Large candidate sets (100+ items)
- Maximum tribe size (8 users)
- Concurrent sessions

---

*For filter integration, see [Filtering System](./filtering.md)*
*For database implementation, see [Database Schema](../database-schema.md)* 