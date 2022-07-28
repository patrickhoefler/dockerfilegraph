### TLS root certs and non-root user
FROM ubuntu:latest@sha256:b6b83d3c331794420340093eb706a6f152d9c1fa51b262d9bf34594887c2c7ac AS ubuntu

RUN \
  # Note that the lack of a "lock" mechanism for apt dependencies
  # currently prevents a fully reproducible build
  apt-get update \
  && apt-get install -y --no-install-recommends \
  # Install TLS root certificates
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# ---

FROM golang:1.18.4@sha256:f3d3d6903d92ff29a3b87f55a97a72cb6828055161eb10cfd20772dbbf346770 AS build
RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build

# ---

### Release image
FROM scratch AS release

# Copy the TLS certificates for encrypted network communication
COPY --from=ubuntu /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build . .

ENTRYPOINT ["/example"]
