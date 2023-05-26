#!/bin/bash
set -ex

docker build --build-arg "VERSION=test" -t "conjur-test:test" -f Dockerfile .