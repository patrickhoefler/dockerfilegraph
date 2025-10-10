### Release image
FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

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

# This only works after running `make build-linux`
# or when using goreleaser
COPY dockerfilegraph /

ENTRYPOINT ["/dockerfilegraph"]
