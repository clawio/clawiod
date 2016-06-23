#!/usr/bin/env bash
#
# ClawIO build script. Add build variables to compiled binary.
#
# Usage:
#
#     $ ./build.bash [output_filename] [git_repo]
#
# Outputs compiled program in current directory.
# Default file name is 'clawiod'.
# Default git repo is current directory.
# Builds always take place from current directory.

set -euo pipefail

: ${output_filename:="${1:-}"}
: ${output_filename:="clawiod"}

: ${git_repo:="${2:-}"}
: ${git_repo:="."}

pkg=main
ldflags=()

# Timestamp of build
ts_name="${pkg}.buildDate"
ts_value=$(date -u +"%a %b %d %H:%M:%S %Z %Y")
ldflags+=("-X" "\"${ts_name}=${ts_value}\"")

# Current tag, if HEAD is on a tag
# This value is used to determine if the current build is a dev build or a release build
# If this value is empty means we are not on an tag, thus is a dev build
current_tag_name="${pkg}.gitTag"
set +e
current_tag_value="$(git -C "${git_repo}" describe --exact-match HEAD 2>/dev/null)"
set -e
ldflags+=("-X" "\"${current_tag_name}=${current_tag_value}\"")

# Nearest tag on branch
tag_name="${pkg}.gitNearestTag"
tag_value="$(git -C "${git_repo}" describe --abbrev=0 --tags HEAD)"
ldflags+=("-X" "\"${tag_name}=${tag_value}\"")

# Commit SHA
commit_name="${pkg}.gitCommit"
commit_value="$(git -C "${git_repo}" rev-parse --short HEAD)"
ldflags+=("-X" "\"${commit_name}=${commit_value}\"")


mkdir -p ${git_repo}/releases

if [[ -z "${current_tag_value}" ]]; then
	# dev build
	current_date=$(date +"%m_%d_%Y_%H_%M_%S")
	GOOS=linux   GOARCH=amd64 go build -ldflags "${ldflags[*]}" -o releases/"${output_filename}"-${tag_value}-linux_amd64-${current_date}-${commit_value}
	GOOS=darwin  GOARCH=amd64 go build -ldflags "${ldflags[*]}" -o releases/"${output_filename}"-${tag_value}-darwin-_md64-${current_date}-${commit_value}
	GOOS=windows GOARCH=amd64 go build -ldflags "${ldflags[*]}" -o releases/"${output_filename}"-${tag_value}-windows_amd64-${current_date}-${commit_value}
else
	# release build
	GOOS=linux   GOARCH=amd64 go build -ldflags "${ldflags[*]}" -o releases/"${output_filename}"-${tag_value}-linux-amd64
	GOOS=darwin  GOARCH=amd64 go build -ldflags "${ldflags[*]}" -o releases/"${output_filename}"-${tag_value}-darwin-amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "${ldflags[*]}" -o releases/"${output_filename}"-${tag_value}-windows-amd64
fi

