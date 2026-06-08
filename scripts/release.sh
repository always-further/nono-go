#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/release.sh [--dry-run] [--skip-checks] vX.Y.Z

Creates and pushes an annotated release tag. Pushing the tag triggers the
Release GitHub Action, which validates the tag and creates the GitHub release.

Options:
  --dry-run      Print the tag and push commands without running them.
  --skip-checks  Skip local Go checks before tagging.
EOF
}

dry_run=false
skip_checks=false
tag=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run)
      dry_run=true
      shift
      ;;
    --skip-checks)
      skip_checks=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      if [ -n "$tag" ]; then
        echo "unexpected extra argument: $1" >&2
        usage >&2
        exit 2
      fi
      tag="$1"
      shift
      ;;
  esac
done

if [ -z "$tag" ]; then
  usage >&2
  exit 2
fi

if ! echo "$tag" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  echo "invalid release tag: $tag" >&2
  echo "expected vX.Y.Z, optionally with a prerelease suffix" >&2
  exit 1
fi

branch="$(git branch --show-current)"
if [ "$branch" != "main" ]; then
  echo "release tags must be created from main; current branch is $branch" >&2
  exit 1
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "working tree is not clean" >&2
  exit 1
fi

git fetch origin main --tags

head="$(git rev-parse HEAD)"
origin_main="$(git rev-parse origin/main)"
if [ "$head" != "$origin_main" ]; then
  echo "HEAD is not equal to origin/main" >&2
  echo "HEAD:        $head" >&2
  echo "origin/main: $origin_main" >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
  echo "local tag already exists: $tag" >&2
  exit 1
fi

if git ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null 2>&1; then
  echo "remote tag already exists: $tag" >&2
  exit 1
fi

if [ "$skip_checks" = "false" ]; then
  make fmt-check
  make tidy-check
  make verify
  make clib-manifest
  make vet
  make test
fi

if [ "$dry_run" = "true" ]; then
  echo "Would create annotated tag: $tag"
  echo "Would push tag to origin:   $tag"
  exit 0
fi

git tag -a "$tag" -m "Release $tag"
git push origin "$tag"

echo "Pushed $tag. The Release GitHub Action will create the GitHub release after validation passes."
