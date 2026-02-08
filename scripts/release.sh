#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/release.sh [--default VERSION] [--api VERSION] [--algo VERSION] [--pipelines VERSION] [--eval VERSION] [--suite VERSION] [--no-push]

Rules:
  - Default VERSION is used for all packages unless overridden.
  - Per-package VERSION '-' skips tagging/pushing that package.
  - Default VERSION cannot be '-'.
  - --no-push creates tags locally without pushing to origin.

Examples:
  scripts/release.sh --default v1.0.0
  scripts/release.sh --default v1.0.0 --api v1.0.1 --algo -
EOF
}

default_version="v1.0.0"
api_version=""
algo_version=""
pipelines_version=""
eval_version=""
suite_version=""
no_push=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --default)
      [[ $# -ge 2 ]] || { echo "Missing value for --default" >&2; exit 1; }
      default_version="$2"
      shift 2
      ;;
    --api)
      [[ $# -ge 2 ]] || { echo "Missing value for --api" >&2; exit 1; }
      api_version="$2"
      shift 2
      ;;
    --algo)
      [[ $# -ge 2 ]] || { echo "Missing value for --algo" >&2; exit 1; }
      algo_version="$2"
      shift 2
      ;;
    --pipelines)
      [[ $# -ge 2 ]] || { echo "Missing value for --pipelines" >&2; exit 1; }
      pipelines_version="$2"
      shift 2
      ;;
    --eval)
      [[ $# -ge 2 ]] || { echo "Missing value for --eval" >&2; exit 1; }
      eval_version="$2"
      shift 2
      ;;
    --suite)
      [[ $# -ge 2 ]] || { echo "Missing value for --suite" >&2; exit 1; }
      suite_version="$2"
      shift 2
      ;;
    --no-push)
      no_push=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ "$default_version" == "-" ]]; then
  echo "Default version cannot be '-'" >&2
  exit 1
fi

api_version="${api_version:-$default_version}"
algo_version="${algo_version:-$default_version}"
pipelines_version="${pipelines_version:-$default_version}"
eval_version="${eval_version:-$default_version}"
suite_version="${suite_version:-$default_version}"

git checkout master
git pull --ff-only origin master
git fetch --tags origin

tags_to_push=()

create_tag() {
  local prefix="$1"
  local version="$2"
  local tag
  if [[ "$version" == "-" ]]; then
    echo "Skipping ${prefix}"
    return
  fi
  if [[ -z "$version" ]]; then
    echo "Version for ${prefix} resolved to empty. Set --default or --${prefix#recsys-} explicitly." >&2
    exit 1
  fi
  tag="${prefix}/${version}"
  git tag -a "$tag" -m "Release $tag"
  tags_to_push+=("$tag")
}

create_tag "api" "$api_version"
create_tag "recsys-algo" "$algo_version"
create_tag "recsys-pipelines" "$pipelines_version"
create_tag "recsys-eval" "$eval_version"
create_tag "recsys-suite" "$suite_version"

if [[ "$no_push" == "true" ]]; then
  echo "Skipping push (--no-push)."
elif [[ ${#tags_to_push[@]} -gt 0 ]]; then
  git push origin "${tags_to_push[@]}"
else
  echo "No tags selected, nothing to push."
fi
