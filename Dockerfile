### Release image
FROM ubuntu:latest@sha256:64255397e256fd33d6c6ddbc371027093315f9822089a32b8eeb045d83bb91c4

LABEL org.opencontainers.image.source="https://github.com/patrickhoefler/dockerfilegraph"

RUN \
  # Note that the lack of a "lock" mechanism for apt dependencies
  # currently prevents a fully reproducible build
  apt-get update \
  && apt-get install -y --no-install-recommends \
  # Install Graphviz
  graphviz \
  && rm -rf /var/lib/apt/lists/* \
  \
  # Add a non-root user
  && useradd app

# Run as non-root user
USER app

# This currently only works with goreleaser
# or if you manually copy the binary into the main project directory
COPY dockerfilegraph /

ENTRYPOINT ["/dockerfilegraph"]
