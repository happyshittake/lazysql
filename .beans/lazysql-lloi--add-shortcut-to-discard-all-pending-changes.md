---
# lazysql-lloi
title: Add shortcut to discard all pending changes
status: completed
type: feature
priority: normal
created_at: 2026-11-06T01:27:36Z
updated_at: 2026-11-06T02:24:58Z
---

Design a keyboard shortcut to cancel pending edits, inserts, and deletes safely; confirm scope and behavior before implementation.

Approved: Ctrl+R with confirmation discards all pending edits/inserts/deletes across current connection. Design and plan saved under ~/.work-docs/lazysql-discard-pending-{design,plan}.md. Inspection found SQL editor already uses Ctrl+R to execute queries; preserve that behavior in editor tabs.

Approved revised key: F8. User explicitly permits working here on main. Baseline go test ./... passed. Ruling: restore affected Records views from their unchanged cached records, not a network refresh; pending changes only affect displayed cells, so local rebuild is immediate, works offline, and avoids refresh failures. Preserve metadata surfaces and editor buffers.

Implementation: F8 Home command with searchable help text, Cancel-default confirmation, local restoration of affected record tabs, tree highlight reset. Verified RED tests failed due missing F8 behavior, then GREEN. Tests cover cancellation/escape, mixed operations across tabs and closed tabs, empty/edit/filter guards, metadata surfaces and SQL text/focus preservation. Full go test ./... passed after implementation; final review pending.

## Summary of Changes

Added F8 / DiscardChanges to Home keymap and README; searchable help inherits the binding. Confirmation defaults to Cancel. Confirm clears all connection-wide pending edits/inserts/deletes and restores affected cached Records views, including filtered duplicate tabs, without DB writes. Preserves editor text, metadata surfaces, and valid sidebar focus; keyboard routing recovers from background focus changes. Current cell edits/filters and empty queues are guarded.

Independent review identified filtered tab lookup, identity drift, sidebar focus and background focus issues; regression tests reproduced each failure before fixes. Review dispositions: local rebuild is intentional and avoids network failures; F8 consumption with an empty queue is an intentional no-op; existing UpdateRows selection reset is retained rather than adding selection preservation. Generic label-comparison/style suggestions are not correctness issues and were left alone. Pre-existing AddInsertedRows behavior and tab identity mutation are outside scope.

Verification: go test ./... PASS; go test -race ./components -run TestDiscardChanges -count=1 PASS; go vet ./... PASS; git diff --check PASS. Not manually exercised in a real terminal. Changes left uncommitted in current checkout as requested.
