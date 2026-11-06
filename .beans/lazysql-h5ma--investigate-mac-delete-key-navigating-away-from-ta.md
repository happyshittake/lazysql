---
# lazysql-h5ma
title: Investigate Mac Delete key navigating away from table
status: completed
type: bug
priority: normal
created_at: 2026-11-06T01:15:34Z
updated_at: 2026-11-06T01:19:30Z
---

Trace Backspace key routing versus text editing; reproduce reported navigation before applying a targeted fix.

## Summary of Changes

Reproduced mouse-focused WHERE input receiving Mac Backspace2 and incorrectly navigating to Connections. GetIsFiltering now recognizes actual filter input focus as well as keyboard-entered filtering state. Added simulation-screen regression coverage for both Backspace encodings, character deletion, focus exit, and navigation outside the filter. Verified go test ./... and git diff --check. No live-terminal manual test or installed binary rebuild.
