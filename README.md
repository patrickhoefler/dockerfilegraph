# dockerfilegraph

[![Go Report Card](https://goreportcard.com/badge/github.com/patrickhoefler/dockerfilegraph)](https://goreportcard.com/report/github.com/patrickhoefler/dockerfilegraph)
[![Maintainability](https://api.codeclimate.com/v1/badges/472d7a3637297d07773d/maintainability)](https://codeclimate.com/github/patrickhoefler/dockerfilegraph/maintainability)

`dockerfilegraph` visualizes your multi-stage Dockerfiles.

It outputs a PDF with a graph representation of the build process. The graph contains the following nodes:

- All _build stages_
- The _default build target_ highlighted in grey
- _External base images_ with dashed borders

The edges of the graph represent:

- _FROM dependencies_ with a full arrow head
- _COPY dependencies_ with an empty arrow head

## Example Output

![Example graph](example/Dockerfile.png)

## Getting Started

### Prerequisites

- A multi-stage [Dockerfile](https://docs.docker.com/engine/reference/builder/) file in your current working directory

### Installation and Usage

Running `dockerfilegraph` without any arguments will create a `Dockerfile.pdf` in your current working directory. This PDF contains a visual graph representation of your multi-stage Dockerfile.

#### Docker

```text
docker run \
  --rm \
  --workdir /workspace \
  --mount type=bind,source="$(pwd)",target=/workspace \
  ghcr.io/patrickhoefler/dockerfilegraph
```

#### Homebrew

```text
brew install patrickhoefler/tap/dockerfilegraph
dockerfilegraph
```

### More Options

```text
$ dockerfilegraph --help
dockerfilegraph visualizes your multi-stage Dockerfile.
It outputs a graph representation of the build process.

Usage:
  dockerfilegraph [flags]

Flags:
  -h, --help     help for dockerfilegraph
  -o, --output   Output file format. One of: pdf, png (default pdf)
```

## Build

```text
go build
./dockerfilegraph
```

## License

[MIT](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)
