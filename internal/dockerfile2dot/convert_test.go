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
				ExternalImages: []ExternalImage{
					{Name: "scratch"},
				},
				Stages: []Stage{
					{
						Layers: []Layer{
							{
								Label:   "FROM scratch",
								WaitFor: WaitFor{Name: "scratch", Type: waitForType(from)}},
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
				ExternalImages: []ExternalImage{
					{Name: "ubuntu"},
					{Name: "scratch"},
					{Name: "buildcache"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:   "FROM ubuntu as base",
								WaitFor: WaitFor{Name: "ubuntu", Type: waitForType(from)},
							},
						},
					},
					{
						Layers: []Layer{
							{
								Label:   "FROM scratch",
								WaitFor: WaitFor{Name: "scratch", Type: waitForType(from)},
							},
							{
								Label:   "COPY --from=base . .",
								WaitFor: WaitFor{Name: "base", Type: waitForType(copy)},
							},
							{
								Label:   "RUN --mount=type=...",
								WaitFor: WaitFor{Name: "buildcache", Type: waitForType(runMountTypeCache)},
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
			ARG UBUNTU_VERSION=22.04
			FROM ubuntu:${UBUNTU_VERSION} as base
			USER app
			FROM scratch
			COPY --from=base . .
			RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build
			`)},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: "ubuntu:${UBUNTU_VERSION}"},
					{Name: "scratch"},
					{Name: "buildcache"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:   "FROM ubuntu:${UBU...",
								WaitFor: WaitFor{Name: "ubuntu:${UBUNTU_VERSION}", Type: waitForType(from)},
							},
							{
								Label: "USER app",
							},
						},
					},
					{
						Layers: []Layer{
							{
								Label:   "FROM scratch",
								WaitFor: WaitFor{Name: "scratch", Type: waitForType(from)},
							},
							{
								Label:   "COPY --from=base . .",
								WaitFor: WaitFor{Name: "base", Type: waitForType(copy)},
							},
							{
								Label:   "RUN --mount=type=...",
								WaitFor: WaitFor{Name: "buildcache", Type: waitForType(runMountTypeCache)},
							},
						},
					},
				},
				BeforeFirstStage: []Layer{
					{Label: "ARG UBUNTU_VERSIO..."},
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
