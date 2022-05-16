package dockerfile2dot

import (
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"reflect"
	"testing"
)

func Test_dockerfileToSimplifiedDockerfile(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name string
		args args
		want SimplifiedDockerfile
	}{
		{
			name: "Most minimal Dockerfile",
			args: args{content: []byte(`
			FROM scratch
			`)},
			want: SimplifiedDockerfile{
				BaseImages: []BaseImage{
					{ID: "scratch"},
				},
				Stages: []Stage{
					{
						ID: "0", WaitFor: []WaitFor{
							{ID: "scratch", Type: waitForType(from)},
						},
						Layers: []Layer{
							{ID: "00", Name: "FROM..."},
						},
					},
				},
			},
		},
		{
			name: "All waitFor types",
			args: args{content: []byte(`
			# syntax=docker/dockerfile:1
			FROM ubuntu as base
			FROM scratch
			COPY --from=base . .
			RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build
			`)},
			want: SimplifiedDockerfile{
				BaseImages: []BaseImage{
					{ID: "ubuntu"},
					{ID: "scratch"},
					{ID: "buildcache"},
				},
				Stages: []Stage{
					{
						ID: "0", Name: "base", WaitFor: []WaitFor{
							{ID: "ubuntu", Type: waitForType(from)},
						},
						Layers: []Layer{
							{ID: "00", Name: "FROM..."},
						},
					},
					{
						ID: "1", WaitFor: []WaitFor{
							{ID: "scratch", Type: waitForType(from)},
							{ID: "base", Type: waitForType(copy)},
							{ID: "buildcache", Type: waitForType(runMountTypeCache)},
						},
						Layers: []Layer{
							{ID: "01", Name: "FROM..."}, {ID: "02", Name: "COPY..."}, {ID: "03", Name: "RUN..."},
						},
					},
				},
			},
		},
		{
			name: "ARGs before FROM",
			args: args{content: []byte(`
			# syntax=docker/dockerfile:1
			ARG UBUNTU_VERSION=20.04
			FROM ubuntu:${UBUNTU_VERSION} as base
			FROM scratch
			COPY --from=base . .
			RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build
			`)},
			want: SimplifiedDockerfile{
				BaseImages: []BaseImage{
					{ID: "ubuntu:${UBUNTU_VERSION}"},
					{ID: "scratch"},
					{ID: "buildcache"},
				},
				Stages: []Stage{
					{
						ID: "0", Name: "base", WaitFor: []WaitFor{
							{ID: "ubuntu:${UBUNTU_VERSION}", Type: waitForType(from)},
						},
						Layers: []Layer{
							{ID: "01", Name: "FROM..."},
						},
					},
					{
						ID: "1", WaitFor: []WaitFor{
							{ID: "scratch", Type: waitForType(from)},
							{ID: "base", Type: waitForType(copy)},
							{ID: "buildcache", Type: waitForType(runMountTypeCache)},
						},
						Layers: []Layer{
							{ID: "02", Name: "FROM..."}, {ID: "03", Name: "COPY..."}, {ID: "04", Name: "RUN..."},
						},
					},
				},
				LayersNotStage: []Layer{
					{ID: "00", Name: "ARG..."},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dockerfileToSimplifiedDockerfile(tt.args.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dockerfileToSimplifiedDockerfile()\ngot  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

// parser.Node definition at
// github.com/moby/buildkit@v0.10.3/frontend/dockerfile/parser/parser.go
func Test_storeDataInLayer(t *testing.T) {
	type args struct {
		layerIndex int
		child      *parser.Node
	}
	tests := []struct {
		name string
		args args
		want Layer
	}{
		{
			name: "Creation layer",
			args: args{
				layerIndex: 0,
				child: &parser.Node{
					Value:       "LABEL",
					Next:        nil,
					Children:    nil,
					Heredocs:    nil,
					Attributes:  nil,
					Original:    "LABEL org.opencontainers.image.source='https://github.com/patrickhoefler/dockerfilegraph'",
					Flags:       nil,
					StartLine:   4,
					EndLine:     4,
					PrevComment: nil,
				},
			},
			want: Layer{ID: "00", Name: "LABEL..."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := storeDataInLayer(tt.args.layerIndex, tt.args.child); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("storeDataInLayer()\ngot  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}
