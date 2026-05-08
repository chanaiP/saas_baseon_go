#!/usr/bin/env bash
set -euo pipefail

artifact="${1:-dist/release/saas_baseon_go.tar.gz}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$repo_root"
./scripts/check-delivery-clean.sh
./scripts/check-source-clean.sh

mkdir -p "$(dirname "$artifact")"
manifest="$(mktemp)"
trap 'rm -f "$manifest"' EXIT

git -c core.quotePath=false ls-files | grep -Ev '(^|/)\.DS_Store$|(^|/)__MACOSX(/|$)|^frontend/node_modules(/|$)|^frontend/dist(/|$)|^dist/release(/|$)|(^|/)(tmp|cache|\.cache)(/|$)|(^|/)[^/]+\.log$|(^|/)\._[^/]*$' >"$manifest"

COPYFILE_DISABLE=1 tar --exclude='._*' --exclude='.DS_Store' -czf "$artifact" -T "$manifest"
echo "release package created: $artifact"
