---
# lazysql-ue2e
title: Fix click switching in upper table tab bar
status: completed
type: bug
priority: normal
created_at: 2026-11-06T11:26:21Z
updated_at: 2026-11-06T11:40:24Z
---

- [x] Reproduce upper tab click routing
- [x] Wire tab clicks with focus and edit protection
- [x] Verify regression and full suite

## Summary of Changes

Upper tab headers now select the existing tab/page and focus its content on click, with right-pane highlighting for table tabs. Mouse-down no longer steals focus into inert header text. Active cell/sidebar edits block switching, while SQL editor text survives tab changes. Hidden header rectangles are cleared before grid layout to prevent stale hit targets after resize. Added four regression tests watched failing before implementation; full go test ./... -count=1 passed, focused tests passed five consecutive runs, and independent read-only review found no correctness issues. Included in the v0.0.5 release. Live-terminal behavior not tested.
