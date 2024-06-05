### Release image
FROM ubuntu:noble-20240530@sha256:e3f92abc0967a6c19d0dfa2d55838833e947b9d74edbcb0113e48535ad4be12a

LABEL org.opencontainers.image.source="https://github.com/patrickhoefler/dockerfilegraph"

# renovate: datasource=repology depName=ubuntu_24_04/fonts-dejavu versioning=loose
ENV FONTS_DEJAVU_VERSION="2.37-8"

# renovate: datasource=repology depName=ubuntu_24_04/graphviz versioning=loose
ENV GRAPHVIZ_VERSION="2.42.2-9build1"

RUN \
  apt-get update \
  && apt-get install -y --no-install-recommends \
  fonts-dejavu="${FONTS_DEJAVU_VERSION}" \
  graphviz="${GRAPHVIZ_VERSION}" \
  && rm -rf /var/lib/apt/lists/* \
  \
  # Add a non-root user
  && useradd app

# Run as non-root user
USER app

# This only works after running `make build-linux`
# or when using goreleaser
COPY dockerfilegraph /

ENTRYPOINT ["/dockerfilegraph"]
