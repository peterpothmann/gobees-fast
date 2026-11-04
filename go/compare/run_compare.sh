#!/usr/bin/env bash
# Builds the reference commit, the base commit and the current code as command line programs and runs the
# comparisons of go/compare: every scenario against the reference, gobees and NetLogo, and the speed. The commits
# and the phase are in go/compare/commits.env.
#
# Usage (anywhere in the repository, Linux or Git Bash):
#   go/compare/run_compare.sh [further arguments of go test]
#
# The commits are checked out into temporary git worktrees, which are removed at the end. The current code is
# the working tree, including uncommitted changes. The pull request comment is written to compare-comment.md in the
# repository root (COMPARE_COMMENT to change it). Set COMPARE_REPORT_DIR to a folder to get the NetLogo report and
# its plots (the CI uses go/validation), and COMPARE_THREADS to limit the runs at the same time.
set -euo pipefail

root=$(git rev-parse --show-toplevel)
# shellcheck source=commits.env
source "$root/go/compare/commits.env"

# native turns a path into one that Go understands, also under Git Bash on Windows.
native() {
    if command -v cygpath >/dev/null; then cygpath -m "$1"; else echo "$1"; fi
}

work=$(mktemp -d)
worktrees=()
cleanup() {
    for w in "${worktrees[@]}"; do
        git -C "$root" worktree remove --force "$w" || true
    done
    rm -rf "$work"
}
trap cleanup EXIT

# build <commit> <name>: builds the command line program of the commit into $work/<name>.exe
build() {
    local dir="$work/$2"
    git -C "$root" worktree add --detach --quiet "$dir" "$1"
    worktrees+=("$dir")
    go -C "$(native "$dir/go")" build -o "$(native "$work/$2.exe")" .
}

reference=$(git -C "$root" rev-parse --verify "$REFERENCE_COMMIT^{commit}")
base=$(git -C "$root" rev-parse --verify "$BASE_COMMIT^{commit}")
echo "reference: $reference"
echo "base:      $base"

build "$reference" reference
COMPARE_REFERENCE_BIN=$(native "$work/reference.exe")
# The same program for the same commit: the tests run each program only once.
if [ "$base" = "$reference" ]; then
    COMPARE_BASE_BIN=$COMPARE_REFERENCE_BIN
else
    build "$base" base
    COMPARE_BASE_BIN=$(native "$work/base.exe")
fi
go -C "$(native "$root/go")" build -o "$(native "$work/current.exe")" .
COMPARE_CURRENT_BIN=$(native "$work/current.exe")

COMPARE_REFERENCE_COMMIT=$REFERENCE_COMMIT
COMPARE_BASE_COMMIT=$BASE_COMMIT
COMPARE_PHASE=$PHASE
COMPARE_REMOVED_COLUMNS=$REMOVED_COLUMNS
COMPARE_REFERENCE_INPUT=${REFERENCE_INPUT:-netlogo}
COMPARE_COMMENT=$(native "${COMPARE_COMMENT:-$root/compare-comment.md}")
export COMPARE_REFERENCE_BIN COMPARE_BASE_BIN COMPARE_CURRENT_BIN COMPARE_REFERENCE_COMMIT COMPARE_BASE_COMMIT
export COMPARE_PHASE COMPARE_COMMENT COMPARE_REMOVED_COLUMNS COMPARE_REFERENCE_INPUT

status=0
go -C "$(native "$root/go")" test ./compare -count=1 -v -timeout 60m -run 'Scenarios|Speed' "$@" || status=$?
echo "comment written to $COMPARE_COMMENT"
exit $status
