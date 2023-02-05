#!/usr/bin/env bash

# Change to the root directory of the project.
cd $(git rev-parse --show-toplevel)

go run . -f examples/dockerfiles/Dockerfile --legend -o svg \
  && mv Dockerfile.svg examples/images/Dockerfile-legend.svg

go run . -f examples/dockerfiles/Dockerfile --layers -o svg \
  && mv Dockerfile.svg examples/images/Dockerfile-layers.svg

go run . -f examples/dockerfiles/Dockerfile.large -c -n 0.3 -o svg -u 4 \
  && mv Dockerfile.svg examples/images/Dockerfile-large.svg
