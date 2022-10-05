#!/usr/bin/env bash

# Calculates the next SemVer version based on GitHub tags.
# Uses the latest tag on origin/master to avoid calculating against pre-release tags on other branches.

set -o errexit -o nounset -o pipefail
CURDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function has_commits_with_message_since_tag () {
  local tag=$1
  local message=$2

  git --no-pager log "$tag"..HEAD --format=%B | grep -P "$message" > /dev/null
}

MAJOR_BUMP_REGEX="^\s*BREAKING(\s+CHANGE)?:"
MINOR_BUMP_REGEX="feat(\([\w\d]+\))?:"
PATCH_BUMP_REGEX="(fix|build|chore|ci|refactor|docs|style|perf|test)(\([\w\d]+\))?:"

# If there are no tags so far, take initial commit and check from there
if [ -z "$(git show-ref --tags)" ]; then
  PREV_SEMVER_TAG=$(git rev-list --max-parents=0 HEAD)
else
  PREV_SEMVER_TAG=$(git describe --tags --first-parent --abbrev=0)
fi

echo "PREV_SEMVER_TAG=$PREV_SEMVER_TAG"

if [ -z "$(git show-ref --tags)" ]; then
  NEXT_SEMVER_TAG="1.0.0"
  echo "No previous tags, will publish first version"
else
  if has_commits_with_message_since_tag "$PREV_SEMVER_TAG" "$MAJOR_BUMP_REGEX" ; then
    NEXT_SEMVER_TAG=$("$CURDIR/get-semver.sh" bump major "$PREV_SEMVER_TAG")
    echo "Found breaking change commits, bumping major version"
  elif has_commits_with_message_since_tag "$PREV_SEMVER_TAG" "$MINOR_BUMP_REGEX" ; then
    NEXT_SEMVER_TAG=$("$CURDIR/get-semver.sh" bump minor "$PREV_SEMVER_TAG")
    echo "Found non-breaking feature commits, bumping minor version"
  elif has_commits_with_message_since_tag "$PREV_SEMVER_TAG" "$PATCH_BUMP_REGEX" ; then
    NEXT_SEMVER_TAG=$("$CURDIR/get-semver.sh" bump patch "$PREV_SEMVER_TAG")
    echo "Found non-breaking fix/refactor commits, bumping patch version"
  else
    echo >&2 "Could not find conventional commits in branch, bailing..."
    echo >&2 "Please make sure that your commit messages are properly structured."
    echo >&2 "Regular expressions are:"
    echo >&2 "MAJOR_BUMP_REGEX=\"$MAJOR_BUMP_REGEX\""
    echo >&2 "MINOR_BUMP_REGEX=\"$MINOR_BUMP_REGEX\""
    echo >&2 "PATCH_BUMP_REGEX=\"$PATCH_BUMP_REGEX\""
    echo >&2 "You can rebase your branch to fix your commits and then force push again."
    exit 1
  fi
fi

FULL_TAG="v${NEXT_SEMVER_TAG}"

echo "Version change ${PREV_SEMVER_TAG} -> ${FULL_TAG}"

export FULL_TAG