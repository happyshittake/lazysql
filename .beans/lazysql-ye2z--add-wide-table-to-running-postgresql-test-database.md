---
# lazysql-ye2z
title: Add wide table to running PostgreSQL test database
status: completed
type: task
priority: normal
created_at: 2026-11-06T08:02:33Z
updated_at: 2026-11-06T08:03:10Z
---

Create an isolated sidebar_scroll_test table with many columns and mixed-length synthetic values in the existing local PostgreSQL fixture. Preserve existing tables; verify row and column counts.

## Summary of Changes

Created public.sidebar_scroll_test in the running local PostgreSQL lazysql_test database (127.0.0.1:5433), with 41 columns (id plus field_01 through field_40) and 10 synthetic rows. Values alternate short, medium, long (1377 characters), multiline, JSON text, NULL, empty strings, and Unicode. Used a transaction, refused to overwrite an existing table, and asserted row/column counts and sample content before commit. Verified existing customers/orders remain at 1200/3000 rows. No application or seed SQL changes; table survives container restarts but not fixture volume reset. Live UI was not checked.
