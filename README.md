# dockerfilegraph

[![Go Report Card](https://goreportcard.com/badge/github.com/patrickhoefler/dockerfilegraph)](https://goreportcard.com/report/github.com/patrickhoefler/dockerfilegraph)
[![Maintainability](https://api.codeclimate.com/v1/badges/472d7a3637297d07773d/maintainability)](https://codeclimate.com/github/patrickhoefler/dockerfilegraph/maintainability)

`dockerfilegraph` visualizes your multi-stage Dockerfiles.

It creates a visual graph representation of the build process. The graph contains the following nodes:

- All _build stages_
- The _default build target_ highlighted in grey
- _External base images_ with dashed borders

The edges of the graph represent:

- _FROM_ dependencies with a full arrow head
- _COPY --from=..._ dependencies with an empty arrow head
- _RUN --mount=type=cache,from=..._ dependencies with an empty diamond arrow head

You can add an optional legend to the graph and change the output format and resolution. For all the details, see the [options](#more-options) below.

## Example Output

### Default

![Example output](https://user-images.githubusercontent.com/547220/120117354-2037a800-c18d-11eb-8750-ce954c5529df.png)

---

### Including a Legend

![Example output including a legend](https://user-images.githubusercontent.com/547220/143328161-dd80b5ef-5960-4023-b74e-e7b28cd31dcb.png)

## Getting Started

### Prerequisites

- A multi-stage [Dockerfile](https://docs.docker.com/engine/reference/builder/) file in your current working directory

### Installation and Usage

Running `dockerfilegraph` without any arguments will create a `Dockerfile.pdf` in your current working directory. This PDF contains a visual graph representation of your multi-stage Dockerfile.

#### Docker / [nerdctl](https://github.com/containerd/nerdctl)

##### Image based on Ubuntu 20.04

```shell
docker run \
  --rm \
  --workdir /workspace \
  --volume "$(pwd)":/workspace \
  ghcr.io/patrickhoefler/dockerfilegraph
```

##### Image based on Alpine Linux

```shell
docker run \
  --rm \
  --workdir /workspace \
  --volume "$(pwd)":/workspace \
  ghcr.io/patrickhoefler/dockerfilegraph:alpine
```

#### [Homebrew](https://brew.sh/)

```text
brew install patrickhoefler/tap/dockerfilegraph
dockerfilegraph
```

#### Build from Source

```text
go build
./dockerfilegraph
```

### More Options

```text
❯ dockerfilegraph --help
dockerfilegraph visualizes your multi-stage Dockerfile.
It outputs a graph representation of the build process.

Usage:
  dockerfilegraph [flags]

Flags:
  -d, --dpi int   Dots per inch of the PNG export (default 96)
  -h, --help      help for dockerfilegraph
  -l, --legend    Add a legend (default false)
  -o, --output    Output file format. One of: pdf, png (default pdf)
```

## License

[MIT](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)
