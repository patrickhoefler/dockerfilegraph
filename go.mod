module github.com/patrickhoefler/dockerfilegraph

go 1.15

require (
	github.com/awalterschulze/gographviz v2.0.3+incompatible
	github.com/moby/buildkit v0.6.4
	github.com/spf13/afero v1.5.1
)

replace (
	github.com/containerd/containerd v1.3.0-0.20190507210959-7c1e88399ec0 => github.com/containerd/containerd v1.3.0
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c => github.com/docker/docker v1.4.2-0.20200227233006-38f52c9fec82
)
