#!/usr/bin/env bash
set -euo pipefail

blocked=(
  ".DS_Store"
  "__MACOSX"
  "frontend/node_modules"
  "frontend/dist"
  "dist/release"
  "*.log"
)

if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  for path in "${blocked[@]}"; do
    if git ls-files --error-unmatch "$path" >/dev/null 2>&1; then
      echo "tracked non-delivery artifact: $path" >&2
      exit 1
    fi
  done
else
  echo "delivery tracked-file check skipped: .git not found"
fi

echo "delivery artifact check passed"
