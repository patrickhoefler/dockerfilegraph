### TLS root certs
FROM ubuntu:latest@sha256:b6b83d3c331794420340093eb706a6f152d9c1fa51b262d9bf34594887c2c7ac AS ubuntu

RUN \
  apt-get update \
  && apt-get install -y --no-install-recommends \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# ---

FROM golang:1.18.4@sha256:6e10f44d212b24a3611280c8035edaf9d519f20e3f4d8f6f5e26796b738433da AS build-tool-dependencies
RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build

# ---

FROM scratch AS release

COPY --from=ubuntu /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-tool-dependencies . .

ENTRYPOINT ["/example"]
