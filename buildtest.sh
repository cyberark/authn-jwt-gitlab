#!/bin/bash
set -ex


#   docker build --tag "secretless-broker-coverage:${FULL_VERSION_TAG}" \
#                --tag "secretless-broker-coverage:latest" \
#                $DOCKER_FLAGS \
#                --file "$TOPLEVEL_DIR/Dockerfile.coverage" \
#                "$TOPLEVEL_DIR"


docker build --build-arg "VERSION=test" -t "conjur-test:test" -f Dockerfile .