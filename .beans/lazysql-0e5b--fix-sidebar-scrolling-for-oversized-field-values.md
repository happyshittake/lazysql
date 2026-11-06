---
# lazysql-0e5b
title: Fix sidebar scrolling for oversized field values
status: completed
type: bug
priority: normal
created_at: 2026-11-06T05:25:33Z
updated_at: 2026-11-06T05:54:11Z
---

- [x] Reproduce scrolling failure and identify root cause
- [x] Add regression coverage and minimal fix
- [x] Verify sidebar behavior and component tests

## Clarified scope

User confirmed scrolling between fields and sidebar-first mouse navigation. Click selects without editing; wheel navigates fields. Design and implementation plan saved under ~/.work-docs. Awaiting execution choice before implementation.

## Execution evidence

Baseline go test ./... passes. Regression tests fail: wheel unhandled, clicks focus first field without updating selection, and tall fields draw outside sidebar. Empty navigation panics. Native in-place execution approved by user (implement directly here); no commits requested. Existing unrelated untracked files remain untouched. Progress is tracked here rather than repository-local planning artifacts.

## Summary of Changes

Implemented sidebar-only mouse-wheel field navigation and click-to-select without entering edit mode. Shared draw-time visibility keeps keyboard/mouse selection within the sidebar, handles mixed field heights, and prevents drawing past the viewport. Empty navigation is safe. Editing retains native TextArea mouse handling and cannot switch fields accidentally.

Added simulation-screen regression tests for wheel boundaries, content/title clicks, keyboard navigation, tall fields, resizing, empty/short viewports, editing/discard after click selection, and press/redraw/release targeting. Independent review found two Important issues; both were reproduced RED then fixed GREEN: double hit-testing during click gestures and stale clickable rectangles after tiny resizing.

Verification: go test ./components -run TestSidebar -count=1 and go test ./... -count=1 pass; git diff --check clean; UI detector returned no findings. No live-terminal/database/OS clipboard integration checks. App-wide mouse support and scrolling inside long values remain outside approved scope. Changes left uncommitted as planned.
