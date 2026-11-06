---
# lazysql-texf
title: Publish mouse navigation fixes as v0.0.5
status: completed
type: task
priority: normal
created_at: 2026-11-06T11:40:08Z
updated_at: 2026-11-06T11:45:02Z
---

- [x] Verify and commit mouse tab fixes
- [x] Push main and v0.0.5 tag
- [x] Confirm release workflow and downloadable assets

## Summary of Changes

Committed upper tab click support as ec19467 and pushed main including f156389 pane-focus/detail-menu fixes. Published tag v0.0.5 through the existing GoReleaser workflow; run 42211997252 succeeded. Verified the public release contains the platform archives and checksum manifest; downloaded Darwin arm64 archive and verified SHA-256. Local go test ./... -count=1, go build ./..., and git diff --check passed. Release: https://github.com/happyshittake/lazysql/releases/tag/v0.0.5. Unrelated untracked files were left untouched.
