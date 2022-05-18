package dockerfile2dot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
						ID: "0",
						Layers: []Layer{
							{
								ID:      "0",
								Name:    "FROM...",
								WaitFor: WaitFor{ID: "scratch", Type: waitForType(from)}},
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
						ID: "0", Name: "base",
						Layers: []Layer{
							{
								ID:      "0",
								Name:    "FROM...",
								WaitFor: WaitFor{ID: "ubuntu", Type: waitForType(from)},
							},
						},
					},
					{
						ID: "1",
						Layers: []Layer{
							{
								ID:      "0",
								Name:    "FROM...",
								WaitFor: WaitFor{ID: "scratch", Type: waitForType(from)},
							},
							{
								ID:      "1",
								Name:    "COPY...",
								WaitFor: WaitFor{ID: "base", Type: waitForType(copy)},
							},
							{
								ID:      "2",
								Name:    "RUN...",
								WaitFor: WaitFor{ID: "buildcache", Type: waitForType(runMountTypeCache)},
							},
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
			USER app
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
						ID:   "0",
						Name: "base",
						Layers: []Layer{
							{
								ID:      "0",
								Name:    "FROM...",
								WaitFor: WaitFor{ID: "ubuntu:${UBUNTU_VERSION}", Type: waitForType(from)},
							},
							{
								ID:   "1",
								Name: "USER...",
							},
						},
					},
					{
						ID: "1",
						Layers: []Layer{
							{
								ID:      "0",
								Name:    "FROM...",
								WaitFor: WaitFor{ID: "scratch", Type: waitForType(from)},
							},
							{
								ID:      "1",
								Name:    "COPY...",
								WaitFor: WaitFor{ID: "base", Type: waitForType(copy)},
							},
							{
								ID:      "2",
								Name:    "RUN...",
								WaitFor: WaitFor{ID: "buildcache", Type: waitForType(runMountTypeCache)},
							},
						},
					},
				},
				BeforeFirstStage: []Layer{
					{ID: "0", Name: "ARG..."},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerfileToSimplifiedDockerfile(tt.args.content)
			if err != nil {
				t.Errorf("dockerfileToSimplifiedDockerfile() error = %v", err)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
