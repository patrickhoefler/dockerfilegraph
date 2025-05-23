package dockerfile2dot

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_dockerfileToSimplifiedDockerfile(t *testing.T) {
	type args struct {
		content        []byte
		maxLabelLength int
	}
	tests := []struct {
		name string
		args args
		want SimplifiedDockerfile
	}{
		{
			name: "Most minimal Dockerfile",
			args: args{
				content:        []byte("FROM scratch"),
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: "scratch"},
				},
				Stages: []Stage{
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{Name: "scratch", Type: waitForType(waitForFrom)}},
							},
						},
					},
				},
			},
		},
		{
			name: "All waitFor types",
			args: args{
				content: []byte(`
# syntax=docker/dockerfile:1
FROM ubuntu as base
FROM scratch
COPY --from=base . .
RUN --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ go build
`),
				maxLabelLength: 20,
			},
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
								Label:    "FROM ubuntu as base",
								WaitFors: []WaitFor{{Name: "ubuntu", Type: waitForType(waitForFrom)}},
							},
						},
					},
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{Name: "scratch", Type: waitForType(waitForFrom)}},
							},
							{
								Label:    "COPY --from=base . .",
								WaitFors: []WaitFor{{Name: "base", Type: waitForType(waitForCopy)}},
							},
							{
								Label:    "RUN --mount=type=...",
								WaitFors: []WaitFor{{Name: "buildcache", Type: waitForType(waitForMount)}},
							},
						},
					},
				},
			},
		},
		{
			name: "Wait for multiple mounts",
			args: args{
				content: []byte(`
# syntax=docker/dockerfile:1
FROM ubuntu as base
RUN \
  --mount=type=cache,from=buildcache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/ \
  --mount=from=artifacts,source=/artifacts/embeddata,target=/artifacts/embeddata go build
`),
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: "ubuntu"},
					{Name: "buildcache"},
					{Name: "artifacts"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:    "FROM ubuntu as base",
								WaitFors: []WaitFor{{Name: "ubuntu", Type: waitForType(waitForFrom)}},
							},
							{
								Label: "RUN --mount=type=...",
								WaitFors: []WaitFor{
									{Name: "buildcache", Type: waitForType(waitForMount)},
									{Name: "artifacts", Type: waitForType(waitForMount)},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "bind mount",
			args: args{
				content: []byte(`
# syntax=docker/dockerfile:1
FROM scratch
RUN --mount=from=build,source=/build/,target=/build/ go build
`),
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: "scratch"},
					{Name: "build"},
				},
				Stages: []Stage{
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{Name: "scratch", Type: waitForType(waitForFrom)}},
							},
							{
								Label:    "RUN --mount=from=...",
								WaitFors: []WaitFor{{Name: "build", Type: waitForType(waitForMount)}},
							},
						},
					},
				},
			},
		},
		{
			name: "ARGs before FROM",
			args: args{
				content: []byte(`
# syntax=docker/dockerfile:1
ARG UBUNTU_VERSION=22.04
ARG PHP_VERSION=8.0
ARG ALPINE_VERSION=3.15

FROM ubuntu:$UBUNTU_VERSION as base
USER app

FROM php:${PHP_VERSION}-fpm-alpine${ALPINE_VERSION} as php

FROM scratch
COPY --from=base . .
RUN --mount=type=cache,source=/go/pkg/mod/cache/,target=/go/pkg/mod/cache/,from=buildcache go build
`),
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: "ubuntu:22.04"},
					{Name: "php:8.0-fpm-alpine3.15"},
					{Name: "scratch"},
					{Name: "buildcache"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:    "FROM ubuntu:22.04...",
								WaitFors: []WaitFor{{Name: "ubuntu:22.04", Type: waitForType(waitForFrom)}},
							},
							{
								Label: "USER app",
							},
						},
					},
					{
						Name: "php",
						Layers: []Layer{
							{
								Label: "FROM php:8.0-fpm-...",
								WaitFors: []WaitFor{{
									Name: "php:8.0-fpm-alpine3.15",
									Type: waitForType(waitForFrom),
								}},
							},
						},
					},
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{Name: "scratch", Type: waitForType(waitForFrom)}},
							},
							{
								Label:    "COPY --from=base . .",
								WaitFors: []WaitFor{{Name: "base", Type: waitForType(waitForCopy)}},
							},
							{
								Label:    "RUN --mount=type=...",
								WaitFors: []WaitFor{{Name: "buildcache", Type: waitForType(waitForMount)}},
							},
						},
					},
				},
				BeforeFirstStage: []Layer{
					{Label: "ARG UBUNTU_VERSIO..."},
					{Label: "ARG PHP_VERSION=8.0"},
					{Label: "ARG ALPINE_VERSIO..."},
				},
			},
		},
		{
			name: "External image used in multiple stages",
			args: args{
				content: []byte(`
# syntax=docker/dockerfile:1.4

FROM scratch AS download-node-setup
ADD https://deb.nodesource.com/setup_16.x ./

FROM scratch AS download-get-pip
ADD https://bootstrap.pypa.io/get-pip.py ./

FROM alpine AS final
COPY --from=download-node-setup setup_16.x ./
COPY --from=download-get-pip get-pip.py ./
`),
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: "scratch"},
					{Name: "alpine"},
				},
				Stages: []Stage{
					{
						Name: "download-node-setup",
						Layers: []Layer{
							{
								Label: "FROM scratch AS d...",
								WaitFors: []WaitFor{{
									Name: "scratch",
									Type: waitForType(waitForFrom),
								}},
							},
							{Label: "ADD https://deb.n..."},
						},
					},
					{
						Name: "download-get-pip",
						Layers: []Layer{
							{
								Label: "FROM scratch AS d...",
								WaitFors: []WaitFor{{
									Name: "scratch",
									Type: waitForType(waitForFrom),
								}},
							},
							{Label: "ADD https://boots..."},
						},
					},
					{
						Name: "final",
						Layers: []Layer{
							{
								Label: "FROM alpine AS final",
								WaitFors: []WaitFor{{
									Name: "alpine",
									Type: waitForType(waitForFrom),
								}},
							},
							{
								Label: "COPY --from=downl...",
								WaitFors: []WaitFor{{
									Name: "download-node-setup",
									Type: waitForType(waitForCopy),
								}},
							},
							{
								Label: "COPY --from=downl...",
								WaitFors: []WaitFor{{
									Name: "download-get-pip",
									Type: waitForType(waitForCopy),
								}},
							},
						},
					},
				},
			},
		},
		{
			name: "Nested ARG variable substitution",
			args: args{
				content: []byte(`
ARG WORLD=world
ARG IMAGE1=hello-${WORLD}-1
ARG IMAGE2=hello-${WORLD}-2

FROM ${IMAGE1}:latest AS stage1
RUN echo "Stage 1"

FROM ${IMAGE2}:latest AS stage2
RUN echo "Stage 2"
`),
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: "hello-world-1:latest"},
					{Name: "hello-world-2:latest"},
				},
				Stages: []Stage{
					{
						Name: "stage1",
						Layers: []Layer{
							{
								Label: "FROM hello-world-...",
								WaitFors: []WaitFor{{
									Name: "hello-world-1:latest",
									Type: waitForType(waitForFrom),
								}},
							},
							{Label: "RUN echo 'Stage 1'"},
						},
					},
					{
						Name: "stage2",
						Layers: []Layer{
							{
								Label: "FROM hello-world-...",
								WaitFors: []WaitFor{{
									Name: "hello-world-2:latest",
									Type: waitForType(waitForFrom),
								}},
							},
							{Label: "RUN echo 'Stage 2'"},
						},
					},
				},
				BeforeFirstStage: []Layer{
					{Label: "ARG WORLD=world"},
					{Label: "ARG IMAGE1=hello-..."},
					{Label: "ARG IMAGE2=hello-..."},
				},
			},
		},
		{
			// This test verifies that an ARG referenced before its definition resolves to an empty string.
			// This aligns with the Docker specification's behavior for ARG variable substitution,
			// where only previously defined ARGs are considered for replacement.
			name: "ARG referencing later ARG (should not resolve)",
			args: args{
				content: []byte(`
ARG IMAGE1=$IMAGE2
ARG IMAGE2=scratch
FROM $IMAGE1
FROM $IMAGE2
`),
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{Name: ""},
					{Name: "scratch"},
				},
				Stages: []Stage{
					{
						Layers: []Layer{{
							Label:    "FROM",
							WaitFors: []WaitFor{{Name: "", Type: waitForType(waitForFrom)}},
						}},
					},
					{
						Layers: []Layer{{
							Label:    "FROM scratch",
							WaitFors: []WaitFor{{Name: "scratch", Type: waitForType(waitForFrom)}},
						}},
					},
				},
				BeforeFirstStage: []Layer{
					{Label: "ARG IMAGE1=$IMAGE2"},
					{Label: "ARG IMAGE2=scratch"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerfileToSimplifiedDockerfile(
				tt.args.content,
				tt.args.maxLabelLength,
			)
			if tt.name == "Wait for multiple mounts" {
				fmt.Printf("%q", got.Stages[0].Layers[1])
			}
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
