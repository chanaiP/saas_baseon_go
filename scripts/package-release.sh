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

blocked_regex='(^|/)\._[^/]+$|(^|/)\.DS_Store$|(^|/)__MACOSX(/|$)|^frontend/node_modules(/|$)|^frontend/dist(/|$)|^dist/release(/|$)|(^|/)(tmp|\.cache)(/|$)|(^|/)[^/]+\.log$'
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git -c core.quotePath=false ls-files | grep -Ev "$blocked_regex" >"$manifest"
else
  find . \
    -path './.git' -prune -o \
    -path './dist/release' -prune -o \
    -type f -print |
    sed 's#^\./##' |
    grep -Ev "$blocked_regex" |
    sort >"$manifest"
fi

tar_args=(--exclude='._*' --exclude='.DS_Store')
if tar --help 2>&1 | grep -q -- '--no-xattrs'; then
  tar_args+=(--no-xattrs)
fi

COPYFILE_DISABLE=1 tar "${tar_args[@]}" -czf "$artifact" -T "$manifest"
echo "release package created: $artifact"
