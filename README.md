# dockerfilegraph

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/patrickhoefler/dockerfilegraph/ci.yml?branch=main)](https://github.com/patrickhoefler/dockerfilegraph/actions/workflows/ci.yml?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/patrickhoefler/dockerfilegraph)](https://goreportcard.com/report/github.com/patrickhoefler/dockerfilegraph)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/patrickhoefler/dockerfilegraph)](https://github.com/patrickhoefler/dockerfilegraph/releases/latest)
[![GitHub](https://img.shields.io/github/license/patrickhoefler/dockerfilegraph)](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)

`dockerfilegraph` visualizes your multi-stage Dockerfiles.

It creates a visual graph representation of the build process.
The graph contains the following nodes:

- All build stages
- The default build target (highlighted in grey)
- External images (with dashed borders)

The edges of the graph represent:

- `FROM ...` dependencies
  (with a solid line and a full arrow head)
- `COPY --from=...` dependencies
  (with a dashed line and an empty arrow head)
- `RUN --mount=type=cache,from=...` dependencies
  (with a dotted line and an empty diamond arrow head)

You can add an optional legend to the graph and change the output format and resolution.
For all the details, see the [options](#more-options) below.

## Example Output

### Including `--legend`

![Example output including a legend](https://user-images.githubusercontent.com/547220/215192032-e5553646-2095-4884-826d-64034e613395.png)

### Including `--layers`

![Example output including layers](https://user-images.githubusercontent.com/547220/215192059-ae3bf14f-432d-4fa0-b7d9-382e5777e862.png)

## Large Dockerfile including `--concentrate` and `--unflatten 4`

tbd

## Getting Started

### Prerequisites

- A multi-stage [Dockerfile](https://docs.docker.com/engine/reference/builder/)

### Installation and Usage

Running `dockerfilegraph` without any arguments will create a `Dockerfile.pdf` in your current working directory.
This PDF contains a visual graph representation of your multi-stage Dockerfile.

#### docker / [nerdctl](https://github.com/containerd/nerdctl)

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
  -c, --concentrate       concentrate the edges (default false)
  -d, --dpi int           dots per inch of the PNG export (default 96)
  -e, --edgestyle         style of the graph edges, one of: default, solid (default default)
  -f, --filename string   name of the Dockerfile (default "Dockerfile")
  -h, --help              help for dockerfilegraph
      --layers            display all layers (default false)
      --legend            add a legend (default false)
  -o, --output            output file format, one of: canon, dot, pdf, png, raw, svg (default pdf)
      --version           display the version of dockerfilegraph
```

## License

[MIT](https://github.com/patrickhoefler/dockerfilegraph/blob/main/LICENSE)
