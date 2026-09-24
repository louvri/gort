#!/usr/bin/env bash
# Usage: next-version.sh MODULE
#        next-version.sh --base MODULE
#
# Prints the tag for MODULE's next release (MODULE/vX.Y.Z), or "skip" when
# there is nothing to release. With --base, prints the newest release tag
# reachable from HEAD instead (empty before the first release): the start of
# the range the version is computed from, and so of the release notes. The
# number itself may go past a newer tag off the mainline; see below.
# Diagnostics go to stderr so stdout stays machine-readable.
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
# Each change is ranked on its own - a commit, a merge commit, or a GitHub
# squash as a whole - and the release takes the highest rank, so a trailer
# on one change can neither mask nor lower another's. "Release-As: skip"
# marks its change - and, on a mainline merge commit, the branch that merge
# brought in - as not triggering a release; skipped changes ship with the
# module's next real change, at no lower a level than they need. A module
# whose tree is unchanged since its tag is not released, whatever its commits
# ask for. Revert lines are not trusted - any message can carry one - so a
# reverted change still counts once something else changes; a bump too large
# is the safe mistake. Needs git 2.38 or later (merge-tree --write-tree). See
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

# stable TAG - succeed when TAG is exactly MODULE/vMAJOR.MINOR.PATCH, setting
# BASH_REMATCH to its numbers; the tag glob still admits MODULE/v1.2.3-rc1.
stable() {
  [[ "$1" =~ ^${module}/v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]
}

# The major version this module's path is for: N for a path ending in /vN,
# otherwise 0 or 1, which share a path.
head_major=1
if [ -f "${path}go.mod" ]; then
  module_path=$(sed -nE 's/^module[[:space:]]+([^[:space:]]+).*/\1/p' "${path}go.mod")
  if [[ "$module_path" =~ /v([0-9]+)$ ]]; then
    head_major="${BASH_REMATCH[1]}"
  fi
fi

# Tag lists are read into variables rather than piped to head: with
# `pipefail`, git being SIGPIPE'd once a list outgrows the pipe buffer would
# fail the step.
#
# The base - where the range, the unchanged-tree check and the release notes
# start - is the newest tag reachable from HEAD. A tag that is not reachable
# is one of two things:
# - a tag of v2 or later for another major: a version of the /vN module path,
#   not of this one, so it does not count at all
# - on a mainline commit after HEAD: a later run already released past this
#   one, and this run - a re-run of an old, failed release, say - is stale;
#   publishing would put a higher version on older code, so it releases
#   nothing. The mainline is MAINLINE_REF, which the release workflow fetches
#   itself; without it, no run is taken for stale, and one that does not
#   resolve fails every run - a broken setup should not wait for a rare tag
#   layout to show
# - anywhere else: a tag pushed by hand off the mainline. It is no base, but
#   it still owns its version on the module proxy, so the number goes past
#   it - reusing its name would collide, and a lower version would never be
#   @latest
# ancestor A B - succeed when A is an ancestor of B; a git error fails the step.
ancestor() {
  local status=0
  git merge-base --is-ancestor "$1" "$2" || status=$?
  if [ "$status" -gt 1 ]; then
    echo "git merge-base ${1} ${2} failed" >&2
    exit 1
  fi
  [ "$status" -eq 0 ]
}

mainline_tip=""
if [ "$mode" = next ] && [ -n "${MAINLINE_REF:-}" ]; then
  if ! mainline_tip=$(git rev-parse -q --verify "${MAINLINE_REF}^{commit}"); then
    echo "MAINLINE_REF ${MAINLINE_REF} does not resolve; refusing to release." >&2
    exit 1
  fi
fi

glob="${module}/v[0-9]*.[0-9]*.[0-9]*"
reachable=$(git tag --sort=-v:refname --merged HEAD --list "$glob")
every=$(git tag --sort=-v:refname --list "$glob")
tag=""
while IFS= read -r candidate; do
  if stable "$candidate"; then
    tag="$candidate"
    break
  fi
done <<< "$reachable"
top="$tag"
while IFS= read -r candidate; do
  if ! stable "$candidate" || grep -qxF "$candidate" <<< "$reachable"; then
    continue
  fi
  candidate_major="${BASH_REMATCH[1]}"
  if [ "$candidate_major" -ge 2 ] && [ "$candidate_major" -ne "$head_major" ]; then
    echo "Ignoring ${candidate}: a v${candidate_major} tag belongs to the /v${candidate_major} module path." >&2
    continue
  fi
  if [ -n "$mainline_tip" ] && ancestor HEAD "$candidate" && ancestor "$candidate" "$mainline_tip"; then
    echo "${candidate} is on a later mainline commit than HEAD; this run is stale, nothing to release." >&2
    echo "skip"
    exit 0
  fi
  if [ -z "$top" ] || [ "$(printf '%s\n%s\n' "$top" "$candidate" | sort -V | tail -n 1)" = "$candidate" ]; then
    top="$candidate"
  fi
done <<< "$every"
major=0
minor=0
patch=0
if stable "$top"; then
  major="${BASH_REMATCH[1]}"
  minor="${BASH_REMATCH[2]}"
  patch="${BASH_REMATCH[3]}"
fi

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

# changed FROM TO - succeed when the module differs between FROM and TO.
# A git error fails the step rather than counting as a change: the version is
# about to be published, and a tag cannot be withdrawn.
changed() {
  local status=0
  git diff --quiet "$1" "$2" -- "$path" || status=$?
  if [ "$status" -gt 1 ]; then
    echo "git diff ${1} ${2} failed" >&2
    exit 1
  fi
  [ "$status" -eq 1 ]
}

# A change reverted before its release leaves the module as it was tagged;
# publishing it would repeat the last version under a new number.
if [ "$first_release" = false ] && ! changed "$tag" HEAD; then
  echo "${path} is unchanged since ${tag}; nothing to release." >&2
  echo "skip"
  exit 0
fi

# Every commit since the tag that changed the module.
commits=$(git rev-list "$range" -- "$path")
if [ -z "$commits" ]; then
  echo "No changes to ${path} since ${tag}; nothing to release." >&2
  echo "skip"
  exit 0
fi

# raise VAR RANK - set VAR to RANK when RANK is a higher release rank.
raise() {
  if [ "$2" != skip ] && [ "$2" -gt "${!1}" ]; then
    printf -v "$1" '%s' "$2"
  fi
}

# release_as_levels MESSAGE - set `levels` to the level of every Release-As
# trailer in MESSAGE, lower-cased, one per line.
#
# A trailer has to be its own line and start it, as git's own trailer parsing
# requires: an indented Release-As: is how a commit body *documents* the
# convention, and prose that merely mentions it must not cut a release. Keys
# and levels are matched case-insensitively, as git treats trailer keys;
# whitespace the merge UI adds around the level is tolerated.
release_as_levels() {
  local status=0
  # grep exits 1 when there is no trailer; anything above that is a failing
  # tool, and reading it as "no trailers" would publish a different version.
  levels=$(grep -iE '^Release-As:' <<< "$1" | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/^release-as:[[:space:]]*//; s/[[:space:]]*$//') || status=$?
  if [ "$status" -gt 1 ]; then
    echo "failed to read Release-As trailers" >&2
    exit 1
  fi
}

# rank_change SUBJECTS MESSAGE DEFAULT - set `ranked` to the release rank one
# change asks for: 3 (major), 2 (minor), 1 (patch), or "skip". Each change is
# ranked on its own and the release takes the highest, so a trailer on one
# change can neither mask nor lower another's.
#
# A Release-As level sets the rank of its own change. Without one,
# conventional-commit markers decide - in any of SUBJECTS, or a footer in
# MESSAGE - and a change with neither ranks DEFAULT. A skipped change also
# sets `carried` to what its markers ask for, so the release that eventually
# ships it is not too small.
#
# The rank helpers set variables rather than print: in a $( ) subshell
# `set -e` is off, so a failing git call would be ranked as an ordinary
# change instead of failing the step. Messages are fed to grep with <<<,
# never piped from git: with `pipefail`, grep -q exiting early would fail
# the pipeline and read as "no match".
rank_change() {
  local subjects="$1" message="$2" default="$3" markers=0 explicit=""
  release_as_levels "$message"
  # Below v1.0.0, semver keeps breaking changes in the minor position and
  # everything else in the patch position.
  if grep -qE '^BREAKING[ -]CHANGE: ' <<< "$message" \
    || grep -qE '^[a-z]+(\([^)]*\))?!:' <<< "$subjects"; then
    if [ "$major" -eq 0 ]; then markers=2; else markers=3; fi
  elif grep -qE '^feat(\([^)]*\))?:' <<< "$subjects"; then
    if [ "$major" -eq 0 ]; then markers=1; else markers=2; fi
  fi
  carried="$markers"
  if grep -qx major <<< "$levels"; then
    explicit=3
  elif grep -qx minor <<< "$levels"; then
    explicit=2
  elif grep -qx patch <<< "$levels"; then
    explicit=1
  elif grep -qx skip <<< "$levels"; then
    explicit=skip
  fi
  if [ -z "$explicit" ]; then
    if [ -n "$levels" ]; then
      # Present but unreadable: fall back to the markers, but say so - doing
      # it silently would hide why the version is not the one asked for.
      echo "Release-As: trailer on '${subjects%%$'\n'*}' is not major, minor, patch or skip; ignoring it." >&2
    fi
    ranked="$markers"
    if [ "$markers" -eq 0 ]; then
      ranked="$default"
    fi
  else
    ranked="$explicit"
    if [ "$explicit" != skip ] && [ "$explicit" -lt "$markers" ]; then
      # Allowed - an internal-only breaking change can ship as a patch - but
      # said where the release log shows it.
      echo "Release-As on '${subjects%%$'\n'*}' ranks below its own breaking or feature marker; releasing at the lower level it asks for." >&2
    fi
  fi
}

# rank_commit COMMIT - rank a non-merge commit, as rank_change does.
#
# A GitHub squash merge is titled "<pull request title> (#N)", and when it
# squashes several commits its body lists their subjects as "* subject".
# Those are read as subjects too, so a squash with a non-conventional title
# does not hide the breaking change its commits declared. A squash is one
# change: its text cannot tell a trailer meant for its last squashed commit
# from one meant for the whole pull request, so a Release-As anywhere in it
# applies to all of it. Bullets count only for that title shape: in any
# other commit an ordinary markdown bullet would otherwise choose the level.
squash_title=' \(#[0-9]+\)$'
rank_commit() {
  local message subjects
  message=$(git log -1 --pretty=%B "$1" | tr -d '\r')
  subjects=${message%%$'\n'*}
  if [[ "$subjects" =~ $squash_title ]]; then
    subjects+=$'\n'"$(sed -nE 's/^\* //p' <<< "$message")"
  fi
  rank_change "$subjects" "$message" 1
}

# rank_merge MERGE - rank a merge commit on its own subject and body, which a
# path-limited walk drops along with the merge. It asks for a patch only when
# it changed the module itself; otherwise its branch's commits speak for it.
rank_merge() {
  local message default=0
  if merge_changed_module "$1"; then
    default=1
  fi
  message=$(git log -1 --pretty=%B "$1" | tr -d '\r')
  rank_change "${message%%$'\n'*}" "$message" "$default"
}

# merge_changed_module MERGE - succeed when the merge commit itself changed
# the module: its tree there differs from what git merges on its own, as with
# a hand-resolved conflict in the module or an edit made in the merge. A
# conflict elsewhere does not count, and git refuses to octopus-merge a
# conflict, so a merge of more than two parents never changed anything
# itself.
merge_changed_module() {
  local auto status=0
  if git rev-parse -q --verify "${1}^3" >/dev/null; then
    return 1
  fi
  # Exit status 1 means conflicts; the tree is still written, with conflict
  # markers in the conflicted files.
  auto=$(git merge-tree --write-tree --no-messages "${1}^1" "${1}^2") || status=$?
  if [ "$status" -gt 1 ]; then
    echo "git merge-tree failed on ${1}" >&2
    exit 1
  fi
  changed "${auto%%$'\n'*}" "$1"
}

# Lists are read into variables before looping over them: a failure inside
# `done <<< "$(...)"` would not stop the script.
mainline=$(git rev-list --first-parent --merges "$range")
merges=$(git rev-list --merges "$range")
changes=$(git rev-list --no-merges "$range" -- "$path")

# A mainline merge marked Release-As: skip covers everything it merged in,
# from every parent after the first and merges inside those branches
# included. Only mainline merges: one inside a pull request's own branch
# must not cover commits it did not introduce.
skipped=""
while IFS= read -r merge; do
  [ -n "$merge" ] || continue
  message=$(git log -1 --pretty=%B "$merge" | tr -d '\r')
  release_as_levels "$message"
  if grep -qx skip <<< "$levels" && ! grep -qxE 'major|minor|patch' <<< "$levels"; then
    skipped+=$'\n'"$(git rev-list "${merge}^1..${merge}")"
  fi
done <<< "$mainline"

# `best` is what triggers a release; `floor` is the level skipped changes
# need, applied only once something else triggers one.
best=0
floor=0
ranked=0
carried=0
note() {
  if [ "$ranked" = skip ]; then
    raise floor "$carried"
  elif grep -qxF "$1" <<< "$skipped"; then
    raise floor "$ranked"
  else
    raise best "$ranked"
  fi
}

while IFS= read -r merge; do
  if [ -z "$merge" ] || ! changed "${merge}^1" "$merge"; then
    continue
  fi
  rank_merge "$merge"
  note "$merge"
done <<< "$merges"

while IFS= read -r commit; do
  [ -n "$commit" ] || continue
  rank_commit "$commit"
  note "$commit"
done <<< "$changes"

# Only a change that asks for a release can trigger one; skipped changes
# ship with the module's next real change.
if [ "$best" -eq 0 ]; then
  echo "Every change to ${path} since ${tag} is marked Release-As: skip; nothing to release." >&2
  echo "skip"
  exit 0
fi
raise best "$floor"

case "$best" in
  3) level="major" ;;
  2) level="minor" ;;
  *) level="patch" ;;
esac

case "$level" in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
esac

# A module's first release is v0.1.0 unless its commits ask for more.
if [ "$first_release" = true ] && [ "$major" -eq 0 ] && [ "$minor" -eq 0 ]; then
  minor=1
  patch=0
  level="minor"
fi

# From v2 on, Go requires the major version in the module path. A tag the
# path does not match is not served as that version, and a published tag
# cannot be withdrawn, so refuse rather than publish it.
if [ "$major" -ge 2 ] && [ "$major" -ne "$head_major" ]; then
  echo "${module}/v${major} needs ${path}go.mod to declare a module path ending in /v${major}." >&2
  echo "If v${major} was not intended, tag the current main commit by hand with the version you want (git tag ${module}/vX.Y.Z origin/main && git push origin ${module}/vX.Y.Z); later runs start from it. A tag on a commit that is not on main is ignored as a base." >&2
  exit 1
fi

next="${module}/v${major}.${minor}.${patch}"
if [ -n "$top" ] && [ "$top" != "$tag" ]; then
  echo "Bumping ${top} (newest ${module} tag; changes counted from ${tag}) -> ${next} (${level})" >&2
else
  echo "Bumping ${tag} -> ${next} (${level})" >&2
fi
echo "$next"
