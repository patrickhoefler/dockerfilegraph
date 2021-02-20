FROM ubuntu:latest@sha256:703218c0465075f4425e58fac086e09e1de5c340b12976ab9eb8ad26615c3715

LABEL org.opencontainers.image.source="https://github.com/patrickhoefler/dockerfilegraph"

# Because there is no "lock" mechanism for apt dependencies,
# this step prevents a fully reproducible build
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
  graphviz \
  && rm -rf /var/lib/apt/lists/*

# Run as non-root user
RUN groupadd --gid 1000 dockerfilegraph \
  && useradd --uid 1000 --gid dockerfilegraph --shell /bin/bash --create-home dockerfilegraph
USER dockerfilegraph

# This currently only works with goreleaser
# or if you manually copy the binary into the main project directory
COPY dockerfilegraph /

WORKDIR /dockerfile

ENTRYPOINT ["/dockerfilegraph"]
