---
id: "SDD-003"
fd: "FD-002"
title: "Integration Wiring — fd-review triggers auto-reject, E2E test"
status: planned
agent: ""
assigned_to: ""
created: "2026-03-28"
started: ""
completed: ""
tags: [integration, wiring, e2e, competitive]
---

# SDD-003: Integration Wiring — competitive FD E2E

> Parent FD: [[FD-002]]

## Scope

Wire all components together and verify the full competitive FD workflow end-to-end:

1. **fd-review integration**: update `/fd-review` to call `RejectCompetitors()` when approving an FD that has `competes_with` entries. The reviewer confirms: "Approvare FD-A rejecta automaticamente FD-B. Procedere?"
2. **Board sync**: verify that rejected FDs sync to board as cards in "Rejected" column (if column exists)
3. **E2E test**: create 2 competing FDs in test vault, review and approve one, verify the other is auto-rejected with correct `superseded_by` and `rejected_reason`
4. **fd-compare post-to-issue**: verify `--post-to-issue` creates a GitHub comment (mock test)

## Acceptance Criteria

- [ ] `/fd-review` on competing FD shows auto-reject warning before approving
- [ ] Approving FD-A auto-rejects FD-B via RejectCompetitors()
- [ ] Rejected FD has `superseded_by: FD-A` and `rejected_reason` filled
- [ ] `forgia status` shows grouped competing FDs with [rejected] for losers
- [ ] Board sync handles rejected status
- [ ] E2E test: create 2 FDs → approve one → other rejected → status correct
- [ ] Existing single-FD workflow unchanged (no regression)
