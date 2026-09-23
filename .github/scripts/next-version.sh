#!/usr/bin/env bash
# Usage: next-version.sh MODULE
#        next-version.sh --base MODULE
#
# Prints the tag for MODULE's next release (MODULE/vX.Y.Z), or "skip" when
# there is nothing to release. With --base, prints the release tag that the
# next version is computed from instead (empty before the first release), so
# release notes start where the version bump does. Diagnostics go to stderr
# so stdout stays machine-readable.
#
# Each module is versioned on its own: its tags are MODULE/v*, and only
# commits that touch MODULE/ count towards its release, so a trailer on a
# commit that only changes another module does not move this one. A squash
# merge is a single commit, though, so a pull request's markers apply to
# every module it touches.
#
# The release is computed from every commit since the module's last release
# tag, not from what the latest push introduced, so a push whose release run
# was superseded or failed is carried by the next run instead of lost, and
# the result does not depend on whether a pull request was squashed, merged
# or rebased.
#
# "Release-As: skip" marks its commit - and, on a merge commit, the branch
# that merge brought in - as needing no release. Only commits without it can
# trigger one; skipped changes ship with the module's next real change. See
# next-version_test.sh for the behaviour this must keep.
set -euo pipefail

mode=next
if [ "${1:-}" = "--base" ]; then
  mode=base
  shift
fi

module="${1:-}"
# The name is used in a tag glob, a regex and a pathspec; keep it to
# characters that mean the same thing in all three.
if [[ ! "$module" =~ ^[a-z0-9_-]+$ ]]; then
  echo "usage: next-version.sh [--base] MODULE (a top-level module directory name)" >&2
  exit 2
fi
path="${module}/"

skip_trailer='^Release-As:[[:space:]]*skip[[:space:]]*$'

# Read the tag list into a variable rather than piping to head: with
# `pipefail`, git being SIGPIPE'd once the list outgrows the pipe
# buffer would fail the step.
tags=$(git tag --sort=-v:refname --list "${module}/v[0-9]*.[0-9]*.[0-9]*")

# Take the newest tag that is exactly MODULE/vMAJOR.MINOR.PATCH; the glob
# above still admits things like MODULE/v1.2.3-rc1.
tag=""
major=0
minor=0
patch=0
while IFS= read -r candidate; do
  if [[ "$candidate" =~ ^${module}/v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    tag="$candidate"
    major="${BASH_REMATCH[1]}"
    minor="${BASH_REMATCH[2]}"
    patch="${BASH_REMATCH[3]}"
    break
  fi
done <<< "$tags"

if [ "$mode" = base ]; then
  echo "$tag"
  exit 0
fi

first_release=false
if [ -n "$tag" ]; then
  range="${tag}..HEAD"
else
  tag="${module}/v0.0.0"
  range="HEAD"
  first_release=true
fi

# Every commit since the tag that changed the module.
commits=$(git rev-list "$range" -- "$path")
if [ -z "$commits" ]; then
  echo "No changes to ${path} since ${tag}; nothing to release." >&2
  echo "skip"
  exit 0
fi

# Commits marked Release-As: skip. The scan covers the whole range rather
# than only the module's commits, because a path-limited walk drops a merge
# commit whose tree matches the branch it merged - and a skip written on the
# merge covers everything that branch brought in. Messages are read into a
# variable before grep: with `pipefail`, grep -q exiting early would fail
# the pipeline and read as "no match".
skipped=""
merges=""
while IFS= read -r commit; do
  [ -n "$commit" ] || continue
  message=$(git log -1 --pretty=%B "$commit" | tr -d '\r')
  is_merge=false
  if git rev-parse -q --verify "${commit}^2" >/dev/null 2>&1; then
    is_merge=true
  fi
  if grep -qE "$skip_trailer" <<< "$message"; then
    skipped=$(printf '%s\n%s\n' "$skipped" "$commit")
    if [ "$is_merge" = true ]; then
      skipped=$(printf '%s\n%s\n' "$skipped" "$(git rev-list "${commit}^1..${commit}^2")")
    fi
  elif [ "$is_merge" = true ] && ! git diff --quiet "${commit}^1" "$commit" -- "$path"; then
    merges=$(printf '%s\n%s\n' "$merges" "$commit")
  fi
done <<< "$(git rev-list "$range")"

releasable=""
while IFS= read -r commit; do
  if ! grep -qxF "$commit" <<< "$skipped"; then
    releasable=$(printf '%s\n%s\n' "$releasable" "$commit")
  fi
done <<< "$commits"
releasable=$(sed '/^$/d' <<< "$releasable")
if [ -z "$releasable" ]; then
  echo "Every change to ${path} since ${tag} is marked Release-As: skip; nothing to release." >&2
  echo "skip"
  exit 0
fi

# Strip CR: the merge UI submits textarea content without git's
# message cleanup, and a trailing CR would defeat the whole-line
# match on the Release-As trailer below. A merge commit that changed the
# module carries its own message too, which a path-limited walk would drop.
subjects=$(git log --no-walk=unsorted --stdin --pretty=%s <<< "$releasable" | tr -d '\r')
messages=$(printf '%s\n%s\n' "$releasable" "$merges" | sed '/^$/d' \
  | git log --no-walk=unsorted --stdin --pretty=%B | tr -d '\r')

# A GitHub squash merge collapses the branch into one commit titled
# "<pull request title> (#N)" whose body lists the original subjects as
# "* subject", so read those as subjects too - otherwise a squash with a
# non-conventional title hides the breaking change that the commits
# themselves declared. Only for that shape: in any other commit, an
# ordinary markdown bullet would otherwise get to choose the release level.
if [ "$(wc -l <<< "$releasable")" -eq 1 ] && grep -qE ' \(#[0-9]+\)$' <<< "$subjects"; then
  status=0
  bullets=$(grep -E '^\* ' <<< "$messages") || status=$?
  if [ "$status" -gt 1 ]; then
    echo "grep failed while reading the squash body" >&2
    exit 1
  fi
  if [ -n "$bullets" ]; then
    subjects=$(printf '%s\n%s\n' "$subjects" "$(sed -E 's/^\* //' <<< "$bullets")")
  fi
fi

# An explicit override has to be its own trailer line. Matching a bare
# marker anywhere in the prose would let a commit that merely
# *mentions* it cut the wrong release.
# The trailer must start the line, as git's own trailer parsing requires:
# an indented Release-As: is how a commit body *documents* the convention, and
# matching that would let a docs commit cut a release. Whitespace after the
# colon and at the end of the line is noise the merge UI adds, so it is
# tolerated.
level=""
if grep -qE '^Release-As:[[:space:]]*major[[:space:]]*$' <<< "$messages"; then
  level="major"
elif grep -qE '^Release-As:[[:space:]]*minor[[:space:]]*$' <<< "$messages"; then
  level="minor"
elif grep -qE '^Release-As:[[:space:]]*patch[[:space:]]*$' <<< "$messages"; then
  level="patch"
elif grep -qE '^Release-As:' <<< "$messages"; then
  # Present but unreadable: say so rather than fall through to the commit
  # subject, which would silently produce a different version.
  echo "Release-As: trailer found but its level is not major, minor, patch or skip; ignoring it." >&2
fi

if [ -z "$level" ]; then
  if grep -qE '^BREAKING[ -]CHANGE: ' <<< "$messages" \
    || grep -qE '^[a-z]+(\([^)]*\))?!:' <<< "$subjects"; then
    intent="breaking"
  elif grep -qE '^feat(\([^)]*\))?:' <<< "$subjects"; then
    intent="feature"
  else
    intent="fix"
  fi

  # Below v1.0.0, semver keeps breaking changes in the minor position
  # and everything else in the patch position.
  if [ "$major" -eq 0 ]; then
    case "$intent" in
      breaking) level="minor" ;;
      *)        level="patch" ;;
    esac
  else
    case "$intent" in
      breaking) level="major" ;;
      feature)  level="minor" ;;
      *)        level="patch" ;;
    esac
  fi
fi

case "$level" in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
esac

# A module's first release is v0.1.0 unless its commits ask for more.
if [ "$first_release" = true ] && [ "$major" -eq 0 ] && [ "$minor" -eq 0 ]; then
  minor=1
  patch=0
fi

# From v2 on, Go requires the major version in the module path. A tag the
# path does not match is not served as that version, and a published tag
# cannot be withdrawn, so refuse rather than publish it.
if [ "$major" -ge 2 ]; then
  module_path=$(sed -nE 's/^module[[:space:]]+([^[:space:]]+).*/\1/p' "${path}go.mod" 2>/dev/null || true)
  if [[ "$module_path" != */v"$major" ]]; then
    echo "${module}/v${major} needs ${path}go.mod to declare a module path ending in /v${major} (found: '${module_path}')." >&2
    exit 1
  fi
fi

next="${module}/v${major}.${minor}.${patch}"
echo "Bumping ${tag} -> ${next} (${level})" >&2
echo "$next"
