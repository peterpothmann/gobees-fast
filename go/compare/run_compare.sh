#!/usr/bin/env bash
# Builds the reference commit, the base commit and the current code as command line programs and runs the
# comparisons of go/compare. The commits and the comparisons to run are in go/compare/commits.env.
#
# Usage (anywhere in the repository, Linux or Git Bash):
#   go/compare/run_compare.sh [further arguments of go test]
#
# The commits are checked out into temporary git worktrees, which are removed at the end. The current code is
# the working tree, including uncommitted changes. Set COMPARE_REPORT to a file to get the results as Markdown
# (the CI uses $GITHUB_STEP_SUMMARY), and COMPARE_THREADS to limit the runs at the same time.
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
if [ "$base" = "$reference" ]; then
    cp "$work/reference.exe" "$work/base.exe"
else
    build "$base" base
fi
go -C "$(native "$root/go")" build -o "$(native "$work/current.exe")" .

COMPARE_REFERENCE_BIN=$(native "$work/reference.exe")
COMPARE_BASE_BIN=$(native "$work/base.exe")
COMPARE_CURRENT_BIN=$(native "$work/current.exe")
export COMPARE_REFERENCE_BIN COMPARE_BASE_BIN COMPARE_CURRENT_BIN

go -C "$(native "$root/go")" test ./compare -count=1 -v -timeout 60m -run "$COMPARE_TESTS" "$@"
