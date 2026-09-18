### Release image
FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

LABEL org.opencontainers.image.source="https://github.com/patrickhoefler/dockerfilegraph"

# renovate: datasource=repology depName=alpine_3_22/font-dejavu versioning=loose
ENV FONT_DEJAVU_VERSION="2.37-r6"

# renovate: datasource=repology depName=alpine_3_22/graphviz versioning=loose
ENV GRAPHVIZ_VERSION="12.2.1-r0"

RUN apk add --update --no-cache \
  font-dejavu="${FONT_DEJAVU_VERSION}" \
  graphviz="${GRAPHVIZ_VERSION}" \
  \
  # Add a non-root user
  && adduser -D app

# Run as non-root user
USER app

# This only works after running `task build-linux`
# or when using goreleaser
COPY dockerfilegraph /

ENTRYPOINT ["/dockerfilegraph"]
