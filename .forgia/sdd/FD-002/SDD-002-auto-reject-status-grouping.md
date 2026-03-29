---
id: "SDD-002"
fd: "FD-002"
title: "Auto-reject competitors + status grouping"
status: done
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [go, vault, status, competitive]
---

# SDD-002: Auto-reject competitors + status grouping

> Parent FD: [[FD-002]]

## Scope

### 1. RejectCompetitors() in vault package

Add `RejectCompetitors(ctx, approvedID, reason)` to FileVault:
- Find all FDs where `competes_with` includes the approved FD ID
- Set their status to `rejected`, `superseded_by` to the approved ID, `rejected_reason` to the provided reason
- Write updated frontmatter back to disk

### 2. forgia status grouping

Update `cmd/forgia/cmd/status.go` to group competing FDs:
- After listing FDs, detect groups sharing the same `upstream_issue`
- Display them grouped with a visual indicator (e.g., indent or prefix)
- Show rejected FDs dimmed or with [rejected] tag

## Acceptance Criteria

- [ ] `RejectCompetitors()` updates all competitor FDs to rejected status
- [ ] `superseded_by` and `rejected_reason` filled on rejected FDs
- [ ] `forgia status` groups competing FDs by upstream_issue
- [ ] Rejected FDs shown with [rejected] indicator
- [ ] RejectCompetitors is idempotent (safe to call twice)
- [ ] 4+ tests pass
