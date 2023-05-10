#!/bin/bash -e

cd "$(dirname "$0")"

. ./utils
. ./build_utils

trap teardown EXIT

export GO_VERSION="${1:-"1.18"}"

# Spin up Conjur environment
source ./start-conjur.sh

announce "Building test containers..."
docker-compose build "conjur_test"
echo "Done!"

# generate output folder locally, if needed
output_dir="../output/$GO_VERSION"
mkdir -p $output_dir

failed() {
  announce "TESTS FAILED"
  exit 1
}

# Golang container version to use: `1.18` or `1.19`
# announce "Running tests for Go version: $GO_VERSION...";
docker-compose run \
  -e GO_VERSION \
  "conjur_test" bash -c 'set -o pipefail;
           output_dir="./output/$GO_VERSION"
           echo "from TCS $(output_dir)"
           go test -coverprofile="$output_dir/c.out" -v ./... | tee "$output_dir/junit.output";
           exit_code=$?;
           echo "Tests finished - aggregating results...";
           cat "$output_dir/junit.output" | go-junit-report > "$output_dir/junit.xml";
           gocov convert "$output_dir/c.out" | gocov-xml > "$output_dir/coverage.xml";'

