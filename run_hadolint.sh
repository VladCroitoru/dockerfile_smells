#!/bin/bash
# Usage: ./run_hadolint.sh < <file_w\_dockerfile_paths>
# Example: ./run_hadolint.sh < dockerfiles_paths_sort.txt     
# To generate <file_w\_dockerfile_paths> file, use the `find` command:
# find <dir> -type f -iname "*dockerfile" > <file_w\_dockerfile_paths>

function runHadolint() {
  local dockerfile="$1"
  echo "hadolint -f $dockerfile"
  hadolint "$dockerfile" --format json > "$dockerfile.hadolint.json"
}

function main() {
  while read -r line; do
    runHadolint "$line"
  done < /dev/stdin
} 

main "$@"
