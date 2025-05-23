# Tribe Management User Stories

## Overview

These user stories define the requirements for tribe creation, membership management, and collaborative features. Tribes are the core social unit in the application, supporting 1-8 members with democratic governance and common ownership principles.

## Core Tribe Management (US-TRIBE)

### Tribe Creation & Setup
- **US-TRIBE-001**: As a user, I want to create a new tribe (max 8 members) so I can collaborate with specific groups
- **US-TRIBE-008**: As a tribe member, I want to update tribe settings (name, description, preferences) collectively

### Member Invitation & Onboarding
- **US-TRIBE-002**: As a tribe member, I want to invite others via email to join our tribe
- **US-TRIBE-003**: As a user, I want to accept/decline tribe invitations
- **US-TRIBE-009**: As a tribe member, I want conflicts resolved fairly using the senior member as tie-breaker

### Membership Management
- **US-TRIBE-004**: As a tribe member, I want to see all members and understand who the senior member is
- **US-TRIBE-005**: As a tribe member, I want to remove other members when necessary (with appropriate safeguards)
- **US-TRIBE-006**: As a tribe member, I want to leave a tribe I no longer want to participate in
- **US-TRIBE-007**: As a user, I want to be part of multiple tribes for different social contexts

## Democratic Governance System

### Two-Stage Invitation Process
- **US-GOVERN-001**: As a tribe member, I want to invite new people by email with a suggested display name
- **US-GOVERN-002**: As an invitee, I want to accept an invitation and have it subject to existing member ratification
- **US-GOVERN-003**: As an existing member, I want to vote to approve or reject new member applications
- **US-GOVERN-004**: As an existing member, if I reject an invitation, it should be immediately denied
- **US-GOVERN-005**: As an invitee, I want my membership confirmed only when all existing members approve

### Member Removal Petitions
- **US-GOVERN-006**: As a tribe member, I want to petition for another member's removal with a stated reason
- **US-GOVERN-007**: As a tribe member, I want to vote on member removal petitions (unanimous approval required)
- **US-GOVERN-008**: As a target of removal, I cannot vote on my own removal petition
- **US-GOVERN-009**: As a tribe member, if any member votes against removal, the petition should fail

### Tribe Deletion Consensus
- **US-GOVERN-010**: As a tribe member, I want to petition for tribe deletion with a stated reason
- **US-GOVERN-011**: As a tribe member, I want to vote on tribe deletion (100% consensus required)
- **US-GOVERN-012**: As a tribe member, if any member votes against deletion, it should be prevented

### List Governance
- **US-GOVERN-013**: As a tribe member, I want to petition for tribe list deletion with confirmation required
- **US-GOVERN-014**: As a tribe member, I want to confirm or cancel list deletion petitions
- **US-GOVERN-015**: As a tribe member, I want any member to be able to resolve list deletion petitions

## Common Ownership Principles

### Equal Access & Rights
- **US-COMMON-001**: As a tribe member, I want equal access to all tribe lists and decision-making features
- **US-COMMON-002**: As a tribe member, I want any member to be able to invite new people
- **US-COMMON-003**: As a tribe member, I want any member to be able to update tribe settings
- **US-COMMON-004**: As a tribe member, I want any member to be able to create and manage tribe lists

### Conflict Resolution
- **US-COMMON-005**: As a tribe member, I want the senior member (longest-standing) to resolve ties in voting
- **US-COMMON-006**: As a tribe member, I want decisions to be democratic with senior member as final arbiter
- **US-COMMON-007**: As a tribe member, I want safety mechanisms to prevent abuse of member removal

### Safety Mechanisms
- **US-SAFETY-001**: As a tribe member, I want protection against mass member removal in short timeframes
- **US-SAFETY-002**: As a tribe member, I want senior member approval required for mass removals
- **US-SAFETY-003**: As a tribe member, I want all member actions logged for accountability
- **US-SAFETY-004**: As a tribe member, I want automatic tribe deletion when the last member leaves

## Tribe Settings & Configuration

### Customizable Settings
- **US-CONFIG-001**: As a tribe member, I want to configure default decision algorithm parameters (K, M values)
- **US-CONFIG-002**: As a tribe member, I want to set tribe inactivity thresholds (1-730 days)
- **US-CONFIG-003**: As a tribe member, I want to control elimination detail visibility in decision history
- **US-CONFIG-004**: As a tribe member, I want to set maximum tribe size (up to 8 members)

### Display & Identity
- **US-DISPLAY-001**: As a tribe member, I want to set a tribe-specific display name different from my global name
- **US-DISPLAY-002**: As a tribe member, I want to see other members' tribe-specific display names
- **US-DISPLAY-003**: As a tribe member, I want to update tribe name and description collectively

## Member Roles & Seniority

### Role Identification
- **US-ROLE-001**: As a tribe member, I want to know who the tribe creator is (informational only)
- **US-ROLE-002**: As a tribe member, I want to identify the senior member (earliest invited active member)
- **US-ROLE-003**: As a tribe member, I want to understand that all members have equal permissions

### Seniority Calculation
- **US-SENIOR-001**: As a tribe member, I want seniority determined by earliest invitation timestamp
- **US-SENIOR-002**: As a tribe member, I want only active members considered for seniority
- **US-SENIOR-003**: As a tribe member, I want the senior member role to transfer if the current senior leaves

## Activity Tracking & History

### Member Activity
- **US-ACTIVITY-001**: As a tribe member, I want to see when other members were last active
- **US-ACTIVITY-002**: As a tribe member, I want inactive members identified based on tribe settings
- **US-ACTIVITY-003**: As a tribe member, I want to track member participation in decisions

### Tribe History
- **US-HISTORY-001**: As a tribe member, I want to see tribe creation history and milestones
- **US-HISTORY-002**: As a tribe member, I want to see member join/leave history
- **US-HISTORY-003**: As a tribe member, I want to see governance action history (votes, petitions)

## Integration with Other Features

### List Management Integration
- **US-INTEGRATE-001**: As a tribe member, I want tribe lists automatically accessible to all members
- **US-INTEGRATE-002**: As a tribe member, I want to see who added items to tribe lists
- **US-INTEGRATE-003**: As a tribe member, I want to manage list sharing permissions collectively

### Decision Making Integration
- **US-INTEGRATE-004**: As a tribe member, I want decision sessions automatically configured for our tribe size
- **US-INTEGRATE-005**: As a tribe member, I want tribe-specific decision preferences applied by default
- **US-INTEGRATE-006**: As a tribe member, I want decision history accessible to all tribe members

## Error Scenarios & Edge Cases

### Boundary Conditions
- **US-EDGE-001**: As the last tribe member, I want clear options when leaving (delete tribe or transfer)
- **US-EDGE-002**: As a tribe member, I want protection against creating tribes over the 8-member limit
- **US-EDGE-003**: As a user, I want clear error messages when tribe operations fail

### Data Consistency
- **US-CONSIST-001**: As a tribe member, I want tribe data to remain consistent during member changes
- **US-CONSIST-002**: As a tribe member, I want list access to be immediately updated when membership changes
- **US-CONSIST-003**: As a tribe member, I want decision sessions to handle member changes gracefully

## Acceptance Criteria Patterns

For all tribe management user stories:

### Common Success Criteria
- User receives immediate feedback on action success/failure
- All affected tribe members are notified of changes
- Data consistency is maintained across all operations
- Appropriate permissions are enforced

### Common Validation Rules
- Tribe size cannot exceed 8 members
- Only active tribe members can perform tribe operations
- All democratic processes require appropriate consensus
- Safety mechanisms prevent abuse of administrative functions

### Common Error Handling
- Clear error messages for permission violations
- Graceful degradation when tribe operations fail
- Automatic cleanup of orphaned data
- Recovery mechanisms for partial operation failures

---

*This document defines the user experience requirements for tribe management. See [Database Schema](../database-schema.md) for data structure and [Governance](../governance.md) for detailed implementation of democratic processes.* 