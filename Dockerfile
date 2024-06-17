### Release image
FROM ubuntu:noble-20240605@sha256:3a467b6a8258f51400fcc363a78de84fed3c0346243b563e2385971dd6a00008

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
