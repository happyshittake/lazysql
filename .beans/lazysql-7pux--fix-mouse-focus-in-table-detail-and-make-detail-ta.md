---
# lazysql-7pux
title: Fix mouse focus in table detail and make detail tabs clickable
status: completed
type: bug
priority: normal
created_at: 2026-11-06T08:08:20Z
updated_at: 2026-11-06T08:29:05Z
---

- [x] Trace detail focus and menu mouse routing
- [x] Add regression tests and implement mouse behavior
- [x] Verify affected tests

Implementation isolated in /Users/ndjoe/.worktrees/lazysql-table-detail-mouse on fix/table-detail-mouse. Baseline and current go test ./... passed. Regression tests reproduced stale pane focus and non-clickable menu. Final review pending.

Fresh review found four important gaps. Each was reproduced by a failing regression and fixed: preserve sidebar draft while selecting a surface; register Home rather than embedded Flex for argument connections; focus sidebar on border/whitespace clicks; block background Home mouse handling beneath front-most root pages. Full go test ./... -count=1 passed after fixes.

## Summary of Changes

Implemented bidirectional click focus and synchronized highlighting across tree, detail, and sidebar; clickable surface labels share keyboard rendering. Preserved drafts, modal input ownership, filter autocomplete, and argument-connection routing. Added mouse routing regressions including inline/overlay sidebar, editor/filter focus, keyboard navigation, and clipped labels. Fresh full suite passed; focused regressions passed five consecutive runs. Changes remain uncommitted in /Users/ndjoe/.worktrees/lazysql-table-detail-mouse on fix/table-detail-mouse, pending integration choice. No live-terminal check.
