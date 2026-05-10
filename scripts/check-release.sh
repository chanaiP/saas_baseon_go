#!/usr/bin/env bash
set -euo pipefail

blocked_regex='(^|/)\._[^/]+$|(^|/)\.git(/|$)|(^|/)\.DS_Store$|(^|/)__MACOSX(/|$)|^frontend/node_modules(/|$)|^frontend/dist(/|$)|(^|/)(tmp|\.cache)(/|$)|(^|/)[^/]+\.log$'

check_artifact() {
  local artifact="$1"
  if [[ ! -f "$artifact" ]]; then
    echo "release artifact not found: $artifact" >&2
    return 1
  fi
  local dirty
  dirty="$(tar -tzf "$artifact" | grep -E "$blocked_regex" || true)"
  if [[ -n "$dirty" ]]; then
    echo "release artifact contains blocked paths:" >&2
    echo "$dirty" >&2
    return 1
  fi
  echo "release artifact content:"
  tar -tzf "$artifact" | awk 'NR <= 200 { print "  " $0 }'
  echo "release artifact check passed: $artifact"
}

self_test() {
  local dir artifact
  dir="$(mktemp -d)"
  trap 'rm -rf "$dir"' RETURN
  mkdir -p "$dir/frontend/node_modules"
  printf 'bad\n' >"$dir/frontend/node_modules/bad.txt"
  tar -czf "$dir/dirty.tar.gz" -C "$dir" frontend/node_modules/bad.txt
  if check_artifact "$dir/dirty.tar.gz" >/dev/null 2>&1; then
    echo "check-release self-test failed: dirty artifact was accepted" >&2
    return 1
  fi
  printf 'bad\n' >"$dir/._Makefile"
  COPYFILE_DISABLE=1 tar -czf "$dir/appledouble.tar.gz" -C "$dir" ._Makefile
  if check_artifact "$dir/appledouble.tar.gz" >/dev/null 2>&1; then
    echo "check-release self-test failed: AppleDouble artifact was accepted" >&2
    return 1
  fi
  mkdir -p "$dir/clean"
  printf 'ok\n' >"$dir/clean/README.md"
  artifact="$dir/clean.tar.gz"
  tar -czf "$artifact" -C "$dir" clean/README.md
  check_artifact "$artifact" >/dev/null
  echo "check-release self-test passed"
}

case "${1:-}" in
  --self-test)
    self_test
    ;;
  "")
    check_artifact "dist/release/saas_baseon_go.tar.gz"
    ;;
  *)
    check_artifact "$1"
    ;;
esac
