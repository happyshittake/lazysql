---
# lazysql-5iyd
title: Wire all button clicks and save before connect
status: completed
type: bug
priority: normal
created_at: 2026-11-06T00:38:14Z
updated_at: 2026-11-06T00:54:19Z
---

- [x] Test and wire start-screen actions
- [x] Test and wire form actions and save-before-connect
- [x] Run suite and review changes

## Summary of Changes

Wired five selection buttons and four form buttons through shared actions. Connect/F3 validates, saves, and connects; duplicate names are rejected; edit retains hidden fields. Focused Enter activates buttons. Fixed startup-command config serialization panic and failed-delete backing-slice mutation. Cached sessions are handled on the UI thread; changed open-session settings report restart required to preserve existing work. Test-connection validation runs on the UI thread and DB results are queued back. Added click, validation, persistence, ordering, failed-save/delete, edit/options, cached-session and serialization regression tests.

Verification: go test ./... -count=1; go vet ./...; focused go test -race ./app ./components -count=3; git diff --check all PASS. Independent review completed; fixes verified via failing/passing tests. Not checked: real-terminal mouse delivery, live database/startup commands, disk-full/crash durability. Deferred minor: broader confirmation outcomes and focused-Enter selection-button coverage. No commits or integration performed.
