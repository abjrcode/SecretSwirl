#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

docker build -t swervo_builder .
docker run --rm -v $(pwd)/build/bin:/artifacts swervo_builder