GITVERSION := $(shell git describe --tags --always)
GITCOMMIT := $(shell git log -1 --pretty=format:"%H")
BUILDDATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS += -s -w
LDFLAGS += -X github.com/patrickhoefler/dockerfilegraph/internal/cmd.gitVersion=$(GITVERSION)
LDFLAGS += -X github.com/patrickhoefler/dockerfilegraph/internal/cmd.gitCommit=$(GITCOMMIT)
LDFLAGS += -X github.com/patrickhoefler/dockerfilegraph/internal/cmd.buildDate=$(BUILDDATE)
FLAGS = -ldflags "$(LDFLAGS)"

build:
	go build $(FLAGS)

build-linux:
	GOOS=linux go build $(FLAGS)

example-images:
	# Change to the root directory of the project.
	cd $(git rev-parse --show-toplevel)

	go run . -f examples/dockerfiles/Dockerfile --legend -o svg \
		&& mv Dockerfile.svg examples/images/Dockerfile-legend.svg

	go run . -f examples/dockerfiles/Dockerfile --layers -o svg \
		&& mv Dockerfile.svg examples/images/Dockerfile-layers.svg

	go run . -f examples/dockerfiles/Dockerfile.large -c -n 0.3 -o svg -u 4 \
		&& mv Dockerfile.svg examples/images/Dockerfile-large.svg

lint:
	# https://github.com/golangci/golangci-lint needs to be installed.
	golangci-lint run

test:
	go test ./... --coverprofile=cover.out
	go install github.com/patrickhoefler/gocovergate@latest
	gocovergate
