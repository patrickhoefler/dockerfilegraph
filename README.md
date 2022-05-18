# dockerfilegraph

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/patrickhoefler/dockerfilegraph/CI)](https://github.com/patrickhoefler/dockerfilegraph/actions?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/patrickhoefler/dockerfilegraph)](https://goreportcard.com/report/github.com/patrickhoefler/dockerfilegraph)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/patrickhoefler/dockerfilegraph)](https://github.com/patrickhoefler/dockerfilegraph/releases/latest)
[![GitHub](https://img.shields.io/github/license/patrickhoefler/dockerfilegraph)](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)

`dockerfilegraph` visualizes your multi-stage Dockerfiles.

It creates a visual graph representation of the build process. The graph contains the following nodes:

- All _build stages_
- The _default build target_ highlighted in grey
- _External images_ with dashed borders

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

---

### Including layers per stage

![Example output including layer per stage](https://user-images.githubusercontent.com/35694200/168588978-d0386eaf-e79b-45d8-a6e1-b99a4771ca05.png)

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

Make sure that [Graphviz](https://graphviz.org/) is installed locally.

Then:

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
  -d, --dpi int   dots per inch of the PNG export (default 96)
  -h, --help      help for dockerfilegraph
      --layers    display all layers (default false)
      --legend    add a legend (default false)
  -o, --output    output file format, one of: canon, dot, pdf, png (default pdf)
```

## License

[MIT](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)
