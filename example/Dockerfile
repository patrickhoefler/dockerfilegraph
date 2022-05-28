### TLS root certs and non-root user
FROM ubuntu:latest@sha256:26c68657ccce2cb0a31b330cb0be2b5e108d467f641c62e13ab40cbec258c68d AS ubuntu

RUN \
  # Note that the lack of a "lock" mechanism for apt dependencies
  # currently prevents a fully reproducible build
  apt-get update \
  && apt-get install -y --no-install-recommends \
  # Install TLS root certificates
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# ---

FROM golang:1.18@sha256:deba77da21f1c852e87512c8fcb9d7c02b075712c895894af181ebb8cb7b497f AS build
RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build

# ---

### Release image
FROM scratch AS release

# Copy the TLS certificates for encrypted network communication
COPY --from=ubuntu /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build . .

ENTRYPOINT ["/example"]
