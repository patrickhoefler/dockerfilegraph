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

![Example graph](example/dockerfile.png)

## Getting Started

### Prerequisites

- A multi-stage [Dockerfile](https://docs.docker.com/engine/reference/builder/) file in your current working directory

### Installation and Usage

Running `dockerfilegraph` will create a `Dockerfile.pdf` in your current working directory that contains a visual graph representation of your multi-stage Dockerfile.

#### Docker

```shell
docker run --rm --workdir /workspace --mount type=bind,source="$(pwd)",target=/workspace ghcr.io/patrickhoefler/dockerfilegraph
```

#### Homebrew

```shell
brew install patrickhoefler/tap/dockerfilegraph
dockerfilegraph
```

#### Build

```shell
go build
./dockerfilegraph
```

## License

[MIT](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)
