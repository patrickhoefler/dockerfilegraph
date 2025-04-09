### Release image
FROM ubuntu:oracular-20250225@sha256:aadf9a3f5eda81295050d13dabe851b26a67597e424a908f25a63f589dfed48f

LABEL org.opencontainers.image.source="https://github.com/patrickhoefler/dockerfilegraph"

# renovate: datasource=repology depName=ubuntu_24_04/fonts-dejavu versioning=loose
ENV FONTS_DEJAVU_VERSION="2.37-8"

# renovate: datasource=repology depName=ubuntu_24_04/graphviz versioning=loose
ENV GRAPHVIZ_VERSION="2.42.4-2build2"

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
