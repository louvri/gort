#!/usr/bin/env bash
# Regression tests for next-version.sh.
#
# Each case builds a throwaway repository, arranges tags and commits, and
# asserts the version the script computes. A wrong tag published to the module
# proxy is immutable, so every behaviour this script relies on is pinned here.
set -euo pipefail

script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/next-version.sh"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

failures=0
total=0

# repo NAME [TAG...] - start a fresh repository, optionally tagged at its root.
repo() {
  local name="$1" tag
  shift
  rm -rf "${workdir:?}/$name"
  mkdir -p "$workdir/$name"
  cd "$workdir/$name"
  git init -q .
  git config user.email t@example.com
  git config user.name Test
  git commit -q --allow-empty -m "chore: root"
  for tag in "$@"; do git tag "$tag"; done
  return 0
}

# change [MODULE] - stage an edit under MODULE/ (default: mod).
change() {
  local module="${1:-mod}"
  mkdir -p "$module"
  echo "$RANDOM" >> "$module/file"
  git add "$module/file"
}

# commit MESSAGE [MODULE] - commit an edit under MODULE/ (default: mod).
commit() { change "${2:-mod}"; git commit -q -m "$1"; }

# expect DESCRIPTION WANT [MODULE] - run the script for MODULE (default: mod)
# and compare its stdout.
expect() {
  local description="$1" want="$2" module="${3-mod}" got
  total=$((total + 1))
  got="$("$script" "$module" 2>/dev/null)" || got="<script failed>"
  if [ "$got" = "$want" ]; then
    printf 'ok   %s\n' "$description"
  else
    printf 'FAIL %s\n       want %s\n       got  %s\n' "$description" "$want" "$got"
    failures=$((failures + 1))
  fi
}

# --- level from the conventional-commit subject -----------------------------
repo a mod/v0.0.5; commit "fix: a bug"
expect "fix below v1.0.0 -> patch" mod/v0.0.6

repo b mod/v0.0.5; commit "feat: a feature"
expect "feat below v1.0.0 -> patch" mod/v0.0.6

repo c mod/v0.0.5; commit "feat!: a breaking change"
expect "breaking below v1.0.0 -> minor" mod/v0.1.0

repo d mod/v0.0.5; commit "refactor(api)!: drop Foo"
expect "any type with ! -> breaking" mod/v0.1.0

repo e mod/v1.2.3; commit "feat: a feature"
expect "feat at v1.x -> minor" mod/v1.3.0

# From v2 on, Go needs the major version in the module path; a tag the path
# does not match would never be served as that version.
repo f mod/v1.2.3; mkdir mod; echo "module example.com/mod/v2" > mod/go.mod
commit "feat!: a breaking change"
expect "breaking at v1.x -> major" mod/v2.0.0

repo f2 mod/v1.2.3; mkdir mod; echo "module example.com/mod" > mod/go.mod
commit "feat!: a breaking change"
expect "v2 without a /v2 module path is refused" "<script failed>"

repo f3 mod/v1.2.3; commit "feat!: a breaking change"
expect "v2 without a go.mod is refused" "<script failed>"

repo g mod/v1.2.3; commit "docs: a doc change"
expect "other types -> patch" mod/v1.2.4

# --- BREAKING CHANGE footers, both spellings the spec allows ----------------
repo h mod/v0.0.5; commit "fix: x

BREAKING CHANGE: removed the thing"
expect "BREAKING CHANGE footer" mod/v0.1.0

repo i mod/v0.0.5; commit "fix: x

BREAKING-CHANGE: removed the thing"
expect "BREAKING-CHANGE footer (hyphenated)" mod/v0.1.0

# --- the level must not be selectable from prose ----------------------------
repo j mod/v0.0.5; commit "docs: describe the old scheme

The level used to be read from [major]/[minor] markers."
expect "prose mentioning [major] is inert" mod/v0.0.6

repo k mod/v0.0.5; commit "docs: mention a trailer

Put Release-As: major in the message to force one."
expect "Release-As inside a sentence is inert" mod/v0.0.6

repo l mod/v0.0.5
commit "docs: document the convention

Subjects we accept:
* feat!: a breaking change
* fix: a bugfix"
commit "chore: a second commit"
expect "prose bullets over a multi-commit range are inert" mod/v0.0.6

# --- explicit override ------------------------------------------------------
repo m mod/v0.0.5; commit "docs: x

Release-As: major"
expect "Release-As trailer wins over the subject" mod/v1.0.0

repo n mod/v0.0.5; commit "feat!: breaking

Release-As: patch"
expect "Release-As can lower the level too" mod/v0.0.6

repo o mod/v0.0.5
commit "docs: x

Release-As: minor"
commit "chore: y"
expect "Release-As in a non-tip commit still counts" mod/v0.1.0

repo p1 mod/v0.0.5; commit "docs: x

Release-As: major   "
expect "Release-As tolerates trailing whitespace" mod/v1.0.0

repo p2 mod/v0.0.5; commit "docs: x

Release-As:   minor  "
expect "Release-As tolerates whitespace around the level" mod/v0.1.0

# An indented Release-As: is how a commit body documents the convention. Git
# does not treat it as a trailer, and neither may we: matching it would let a
# docs commit cut a release.
repo p2b mod/v0.0.5; commit "docs: document the release convention

To force a release level, add the trailer on its own line:

    Release-As: major

Otherwise the level comes from the commit subjects."
expect "an indented Release-As is not a trailer" mod/v0.0.6

repo p2c mod/v0.0.5; commit "docs: x

	Release-As: major"
expect "a tab-indented Release-As is not a trailer" mod/v0.0.6

repo p3 mod/v0.0.5; commit "feat!: breaking

Release-As: enormous"
expect "unreadable Release-As level falls back to the subject" mod/v0.1.0

# A change that reaches no consumer - a workflow, a README - should not
# publish a version identical to the last one.
repo p4 mod/v0.1.0; commit "ci: reshuffle a workflow

Release-As: skip"
expect "Release-As: skip publishes nothing" skip

repo p5 mod/v0.1.0; commit "feat!: breaking

Release-As: skip"
expect "skip beats the subject it is attached to" skip

repo p6 mod/v0.1.0; commit "ci: x

Release-As: skip"; commit "feat!: a real change"
expect "a skip below the merged commits does not skip" mod/v0.2.0

# A skip covers only its own commit: a change carried from an earlier push
# still releases, at the level it asked for.
repo p6c mod/v0.1.0; commit "ci: x

Release-As: patch"; commit "ci: y

Release-As: skip"
expect "a carried change still releases alongside a skip" mod/v0.1.1

repo p6d mod/v0.1.0; commit "ci: y

Release-As: skip"; commit "ci: x

Release-As: patch"
expect "and in the other commit order too" mod/v0.1.1

# Skipping creates no tag, so a skip read from anywhere in the range would
# stay in the range and disable every future release. It must defer, not
# suppress.
repo p6b mod/v0.1.0
commit "ci: reshuffle a workflow

Release-As: skip"
expect "the CI push itself skips" skip
commit "fix: a genuine bug fix"
expect "the next push releases, carrying the skipped commit" mod/v0.1.1

repo p7 mod/v0.1.0; commit "docs: x

    Release-As: skip"
expect "an indented skip is not a trailer" mod/v0.1.1

# A merge commit leaves the trailer on the branch commits, not on HEAD, so
# reading only HEAD would miss it.
repo p8a mod/v0.1.0
git checkout -q -b feature
commit "ci: reshuffle a workflow

Release-As: skip"
git checkout -q -
git merge -q --no-ff -m "Merge pull request #1 from feature" feature
expect "skip survives a merge commit" skip

# A parentless HEAD has no HEAD^ to diff against; it must not die under set -e.
repo p8b
change
git commit -q --amend -m "ci: first commit

Release-As: skip"
expect "skip works on a parentless HEAD" skip

# A squash collapses the branch into one commit whose body carries the
# original messages, trailer included.
repo p8 mod/v0.1.0
git checkout -q -b feature
commit "ci: reshuffle a workflow

Release-As: skip"
git checkout -q -
git merge -q --squash feature
git commit -q -m "ci: reshuffle a workflow (#10)

* ci: reshuffle a workflow

Release-As: skip"
expect "skip survives a squash merge" skip

repo p mod/v0.0.5
printf 'docs: x\r\n\r\nRelease-As: major\r\n' > "$workdir/crlf.txt"
change
git commit -q --cleanup=verbatim -F "$workdir/crlf.txt"
expect "Release-As survives CRLF line endings" mod/v1.0.0

# --- merge strategies -------------------------------------------------------
repo q mod/v0.0.5
git checkout -q -b feature
commit "feat!: the breaking change"
commit "docs: follow-up"
git checkout -q -
git merge -q --no-ff -m "Merge pull request #1 from feature" feature
expect "merge commit reads the branch commits" mod/v0.1.0

repo r mod/v0.0.5
git checkout -q -b feature
commit "feat!: the breaking change"
commit "docs: follow-up"
git checkout -q -
git merge -q --squash feature
git commit -q -m "Generated title that is not conventional (#1)

* feat!: the breaking change

* docs: follow-up"
expect "squash with a non-conventional title reads its body bullets" mod/v0.1.0

repo s mod/v0.0.5
git checkout -q -b feature
commit "feat!: the breaking change"
git checkout -q -
git merge -q feature
expect "fast-forward reads the branch commits" mod/v0.1.0

# --- tags -------------------------------------------------------------------
repo t; commit "feat!: first"
expect "no tags at all starts from v0.0.0" mod/v0.1.0

repo t2; commit "fix: first"
expect "a first release is at least v0.1.0" mod/v0.1.0

repo t3; commit "feat: first

Release-As: major"
expect "a first release can still ask for more" mod/v1.0.0

repo u mod/v0.0.5; git tag mod/v0.0.6-rc1; commit "fix: x"
expect "prerelease tags are skipped" mod/v0.0.6

repo v mod/v0.0.5; git tag -a mod/v0.0.6 -m annotated; commit "fix: x"
expect "annotated tags are read" mod/v0.0.7

repo w mod/v0.0.5
expect "no commits since the tag -> skip" skip

# --- per-module versioning --------------------------------------------------
repo x1 mod/v0.1.0 other/v0.1.0; commit "fix: x" other
expect "a change to another module only -> skip" skip
expect "while that module releases" other/v0.1.1 other

repo x2 mod/v0.1.0 other/v0.1.0
commit "fix: other

Release-As: major" other
expect "a trailer applies to the module it touched" other/v1.0.0 other
commit "fix: mod"
expect "but not to another module" mod/v0.1.1

# A module is released only by a push that changes it. Commits carried from an
# earlier push - a skipped one, or one whose release run failed - wait for the
# next change to the module instead of publishing an identical version.
repo x2b mod/v0.1.0
commit "ci: tidy the module

Release-As: skip"
expect "a skipped push" skip
commit "docs: unrelated" other
expect "is not released by a push that leaves the module alone" skip
commit "fix: a real change"
expect "but by its next change, carrying the skipped commit" mod/v0.1.1

repo x3 v0.9.0 other/v3.0.0 mod/v0.0.5; commit "fix: x"
expect "legacy and other-module tags are ignored" mod/v0.0.6

repo x4 mod/v0.0.5 modx/v5.0.0; commit "fix: x" modx
expect "a module whose name is a prefix does not see its sibling" skip
commit "fix: y"
expect "nor its sibling's tags" mod/v0.0.6

# A squash that touches several modules releases each with the same level.
repo x5 mod/v0.0.5 other/v0.2.3
git checkout -q -b feature
commit "Raise the Go floor

Release-As: minor"
change other
git commit -q -m "Also the other module"
git checkout -q -
git merge -q --squash feature
git commit -q -m "A pull request title (#16)

* Raise the Go floor

Release-As: minor

* Also the other module"
expect "a squash across modules releases mod at the shared level" mod/v0.1.0
expect "and the other module too" other/v0.3.0 other

# A path-limited log drops a merge commit whose tree matches the branch for
# the module, so the trailer in the merge commit's own message must still be
# read.
repo x7 mod/v0.1.0
git checkout -q -b feature
commit "fix: x"
git checkout -q -
git merge -q --no-ff -m "Merge pull request #1 from feature

Release-As: major" feature
expect "a trailer in the merge commit's own message counts" mod/v1.0.0

repo x7b mod/v0.1.0
git checkout -q -b feature
commit "fix: x"
git checkout -q -
git merge -q --no-ff -m "Merge pull request #1 from feature

Release-As: skip" feature
expect "and so does a skip there" skip

repo x6 mod/v0.0.5; commit "fix: x"
expect "a module name with a slash is refused" "<script failed>" "mod/sub"
expect "a missing module name is refused" "<script failed>" ""

# --- the range, not the tip, decides ----------------------------------------
# A push of several commits, or a run superseded by a newer push, must not
# lose a change because the tip commit does not touch the module.
repo y1 mod/v0.0.5; commit "feat!: x"; commit "chore: y" other
expect "a module change below a tip that leaves it alone still releases" mod/v0.1.0

# Only a GitHub squash title "(#N)" marks body bullets as squashed subjects.
repo y2 mod/v1.0.0; commit "fix: typo

Notes:
* feat!: was considered but not done"
expect "bullets in an ordinary commit body are inert" mod/v1.0.1

# --- base tag for release notes ---------------------------------------------
expect_base() {
  local description="$1" want="$2" module="${3-mod}" got
  total=$((total + 1))
  got="$("$script" --base "$module" 2>/dev/null)" || got="<script failed>"
  if [ "$got" = "$want" ]; then
    printf 'ok   %s\n' "$description"
  else
    printf 'FAIL %s\n       want %s\n       got  %s\n' "$description" "$want" "$got"
    failures=$((failures + 1))
  fi
}

repo z1 mod/v0.0.5 other/v9.9.9; git tag mod/v0.1.0-rc1
expect_base "base is the newest stable tag of the module" mod/v0.0.5

repo z2 other/v1.0.0
expect_base "base is empty before the first release" ""

# ----------------------------------------------------------------------------
cd /
printf '\n%d/%d passed\n' "$((total - failures))" "$total"
[ "$failures" -eq 0 ]
