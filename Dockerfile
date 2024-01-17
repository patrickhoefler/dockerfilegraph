### Release image
FROM ubuntu:jammy-20240111@sha256:054913d4e5ecc87e7b8ca504d5d440d82af21041f77695b154ea3f4eaf8548ef

LABEL org.opencontainers.image.source="https://github.com/patrickhoefler/dockerfilegraph"

# renovate: datasource=repology depName=ubuntu_22_04/fonts-dejavu versioning=loose
ENV FONTS_DEJAVU_VERSION="2.37-2build1"

# renovate: datasource=repology depName=ubuntu_22_04/graphviz versioning=loose
ENV GRAPHVIZ_VERSION="2.42.2-6"

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
