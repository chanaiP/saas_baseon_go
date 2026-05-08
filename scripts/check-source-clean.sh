#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

blocked_regex='(^|/)\.DS_Store$|(^|/)__MACOSX(/|$)|^frontend/node_modules(/|$)|^frontend/dist(/|$)|(^|/)\.git(/|$)|(^|/)(tmp|cache|\.cache)(/|$)|(^|/)[^/]+\.log$'

tracked_dirty="$(git -c core.quotePath=false ls-files | grep -E "$blocked_regex" || true)"
untracked_dirty="$(git -c core.quotePath=false ls-files --others --exclude-standard | grep -E "$blocked_regex" || true)"

if [[ -d frontend/node_modules ]]; then
  untracked_dirty="${untracked_dirty}"$'\n'"frontend/node_modules"
fi
if [[ -d frontend/dist ]]; then
  untracked_dirty="${untracked_dirty}"$'\n'"frontend/dist"
fi
if [[ -d __MACOSX ]]; then
  untracked_dirty="${untracked_dirty}"$'\n'"__MACOSX"
fi

dirty="$(printf '%s\n%s\n' "$tracked_dirty" "$untracked_dirty" | sed '/^[[:space:]]*$/d' | sort -u)"
if [[ -n "$dirty" ]]; then
  echo "source tree contains non-delivery artifacts:" >&2
  echo "$dirty" >&2
  exit 1
fi

echo "source tree clean for release"
