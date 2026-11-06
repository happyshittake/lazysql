---
# lazysql-1fs1
title: Configure GoReleaser for GitHub fork
status: completed
type: task
priority: normal
created_at: 2026-11-06T00:19:53Z
updated_at: 2026-11-06T00:23:57Z
---

Approved scope: publish tagged binary releases and checksums to happyshittake/lazysql using the built-in GitHub token; remove upstream Homebrew publishing. Do not publish a release or change the in-app updater.

- [x] Confirm release configuration and document implementation plan
- [x] Update workflow token and remove Homebrew publishing
- [x] Validate configuration without publishing

## Summary of Changes

Changed release workflow to secrets.GITHUB_TOKEN with existing contents: write permission; removed upstream Homebrew publisher. actionlint, fork configuration assertions, and git diff --check passed. GoReleaser v2.20 reports valid configuration but exits 2 due to pre-existing deprecated archive format keys; baseline checked and confirmed. Read-only independent review found no introduced issues. No release, push, commit, or updater change performed.
