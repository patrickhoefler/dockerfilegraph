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

### Including `--legend`

![Example output including a legend](https://user-images.githubusercontent.com/547220/169665156-09cb79a9-8441-48a7-b2af-4e010eec4b13.png)

---

### Including `--layers`

![Example output including layers](https://user-images.githubusercontent.com/547220/169665172-0a083ae4-6b9c-4900-ad91-aa9085e21b52.png)

## Getting Started

### Prerequisites

- A multi-stage [Dockerfile](https://docs.docker.com/engine/reference/builder/)

### Installation and Usage

Running `dockerfilegraph` without any arguments will create a `Dockerfile.pdf` in your current working directory. This PDF contains a visual graph representation of your multi-stage Dockerfile.

#### Docker / [nerdctl](https://github.com/containerd/nerdctl)

##### Image based on Ubuntu 22.10

```shell
docker run \
  --rm \
  --user "$(id -u):$(id -g)" \
  --workdir /workspace \
  --volume "$(pwd)":/workspace \
  ghcr.io/patrickhoefler/dockerfilegraph
```

##### Image based on Alpine Linux

```shell
docker run \
  --rm \
  --user "$(id -u):$(id -g)" \
  --workdir /workspace \
  --volume "$(pwd)":/workspace \
  ghcr.io/patrickhoefler/dockerfilegraph:alpine
```

#### [Homebrew](https://brew.sh/)

```text
brew install patrickhoefler/tap/dockerfilegraph
dockerfilegraph
```

#### [toolctl](https://toolctl.io/)

```text
toolctl install dockerfilegraph
dockerfilegraph
```

#### Build from Source

Make sure that [Graphviz](https://graphviz.org/) is installed locally.

Then:

```text
make build
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
  -d, --dpi int           dots per inch of the PNG export (default 96)
  -f, --filename string   name of the Dockerfile (default "Dockerfile")
  -h, --help              help for dockerfilegraph
      --layers            display all layers (default false)
      --legend            add a legend (default false)
  -o, --output            output file format, one of: canon, dot, pdf, png (default pdf)
      --version           display the version of dockerfilegraph
```

## License

[MIT](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)
