### Release image
FROM ubuntu:focal-20220302@sha256:8ae9bafbb64f63a50caab98fd3a5e37b3eb837a3e0780b78e5218e63193961f9

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
