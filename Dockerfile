### Release image
FROM ubuntu:focal-20220302@sha256:6f433a8355cb83b9d526759e50dfed2a7321b84401107e05a3dbfff9e46e6b54

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
