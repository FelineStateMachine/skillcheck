#!/bin/sh
set -eu

before=$(git diff -- internal/report/generated internal/report/assets)
go generate ./internal/report/...
after=$(git diff -- internal/report/generated internal/report/assets)
test "$before" = "$after" || { echo "generated report files are stale" >&2; exit 1; }
npm --prefix web/report run build
post_build=$(git diff -- internal/report/assets)
test "$after" = "$post_build" || { echo "embedded report assets are stale" >&2; exit 1; }
