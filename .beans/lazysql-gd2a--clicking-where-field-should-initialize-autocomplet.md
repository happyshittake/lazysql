---
# lazysql-gd2a
title: Clicking WHERE field should initialize autocomplete like /
status: completed
type: bug
priority: normal
created_at: 2026-11-06T03:02:37Z
updated_at: 2026-11-06T03:04:44Z
---

- [x] Reproduce click versus keyboard autocomplete difference
- [x] Fix shared filter focus initialization and add regression test
- [x] Verify targeted and component tests

## Summary of Changes

Autocomplete callbacks were initialized only by the / search action. Register them when constructing the table filter and refresh suggestions on mouse entry as well as keyboard search. Added simulated mouse/keyboard regression coverage for empty input, column prefixes, operators, and metadata arriving after construction. Verified go test ./... and git diff --check.
