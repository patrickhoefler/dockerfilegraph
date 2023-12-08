### Build
FROM golang:alpine AS build

RUN apk add make graphviz
WORKDIR /go/src
COPY . .
RUN make build-linux


### Release image
FROM ubuntu:jammy-20231128@sha256:8eab65df33a6de2844c9aefd19efe8ddb87b7df5e9185a4ab73af936225685bb

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

COPY --from=build /go/src/dockerfilegraph .

ENTRYPOINT ["/dockerfilegraph"]
