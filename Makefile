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

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(FLAGS)
