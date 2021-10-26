### Release image
FROM ubuntu:focal-20211006@sha256:626ffe58f6e7566e00254b638eb7e0f3b11d4da9675088f4781a50ae288f3322

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
