### Release image
FROM alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715

LABEL org.opencontainers.image.source="https://github.com/patrickhoefler/dockerfilegraph"

# renovate: datasource=repology depName=alpine_3_21/font-dejavu versioning=loose
ENV FONT_DEJAVU_VERSION="2.37-r5"

# renovate: datasource=repology depName=alpine_3_21/graphviz versioning=loose
ENV GRAPHVIZ_VERSION="12.2.0-r0"

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
