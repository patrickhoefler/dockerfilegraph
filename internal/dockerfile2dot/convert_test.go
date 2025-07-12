package dockerfile2dot

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_dockerfileToSimplifiedDockerfile(t *testing.T) {
	type args struct {
		content         []byte
		maxLabelLength  int
		separateScratch bool
	}
	tests := []struct {
		name string
		args args
		want SimplifiedDockerfile
	}{
		{
			name: "Most minimal Dockerfile",
			args: args{
				content:         []byte("FROM scratch"),
				maxLabelLength:  20,
				separateScratch: false,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{ID: "scratch", Name: "scratch"},
				},
				Stages: []Stage{
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{ID: "scratch", Type: waitForType(waitForFrom)}},
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
				maxLabelLength:  20,
				separateScratch: false,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "scratch", Name: "scratch"},
					{ID: "buildcache", Name: "buildcache"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:    "FROM ubuntu as base",
								WaitFors: []WaitFor{{ID: "ubuntu", Type: waitForType(waitForFrom)}},
							},
						},
					},
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{ID: "scratch", Type: waitForType(waitForFrom)}},
							},
							{
								Label:    "COPY --from=base . .",
								WaitFors: []WaitFor{{ID: "base", Type: waitForType(waitForCopy)}},
							},
							{
								Label:    "RUN --mount=type=...",
								WaitFors: []WaitFor{{ID: "buildcache", Type: waitForType(waitForMount)}},
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
				maxLabelLength:  20,
				separateScratch: false,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "buildcache", Name: "buildcache"},
					{ID: "artifacts", Name: "artifacts"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:    "FROM ubuntu as base",
								WaitFors: []WaitFor{{ID: "ubuntu", Type: waitForType(waitForFrom)}},
							},
							{
								Label: "RUN --mount=type=...",
								WaitFors: []WaitFor{
									{ID: "buildcache", Type: waitForType(waitForMount)},
									{ID: "artifacts", Type: waitForType(waitForMount)},
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
					{ID: "scratch", Name: "scratch"},
					{ID: "build", Name: "build"},
				},
				Stages: []Stage{
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{ID: "scratch", Type: waitForType(waitForFrom)}},
							},
							{
								Label:    "RUN --mount=from=...",
								WaitFors: []WaitFor{{ID: "build", Type: waitForType(waitForMount)}},
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
					{ID: "ubuntu:22.04", Name: "ubuntu:22.04"},
					{ID: "php:8.0-fpm-alpine3.15", Name: "php:8.0-fpm-alpine3.15"},
					{ID: "scratch", Name: "scratch"},
					{ID: "buildcache", Name: "buildcache"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:    "FROM ubuntu:22.04...",
								WaitFors: []WaitFor{{ID: "ubuntu:22.04", Type: waitForType(waitForFrom)}},
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
									ID:   "php:8.0-fpm-alpine3.15",
									Type: waitForType(waitForFrom),
								}},
							},
						},
					},
					{
						Layers: []Layer{
							{
								Label:    "FROM scratch",
								WaitFors: []WaitFor{{ID: "scratch", Type: waitForType(waitForFrom)}},
							},
							{
								Label:    "COPY --from=base . .",
								WaitFors: []WaitFor{{ID: "base", Type: waitForType(waitForCopy)}},
							},
							{
								Label:    "RUN --mount=type=...",
								WaitFors: []WaitFor{{ID: "buildcache", Type: waitForType(waitForMount)}},
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
					{ID: "scratch", Name: "scratch"},
					{ID: "alpine", Name: "alpine"},
				},
				Stages: []Stage{
					{
						Name: "download-node-setup",
						Layers: []Layer{
							{
								Label: "FROM scratch AS d...",
								WaitFors: []WaitFor{{
									ID:   "scratch",
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
									ID:   "scratch",
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
									ID:   "alpine",
									Type: waitForType(waitForFrom),
								}},
							},
							{
								Label: "COPY --from=downl...",
								WaitFors: []WaitFor{{
									ID:   "download-node-setup",
									Type: waitForType(waitForCopy),
								}},
							},
							{
								Label: "COPY --from=downl...",
								WaitFors: []WaitFor{{
									ID:   "download-get-pip",
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
					{ID: "hello-world-1:latest", Name: "hello-world-1:latest"},
					{ID: "hello-world-2:latest", Name: "hello-world-2:latest"},
				},
				Stages: []Stage{
					{
						Name: "stage1",
						Layers: []Layer{
							{
								Label: "FROM hello-world-...",
								WaitFors: []WaitFor{{
									ID:   "hello-world-1:latest",
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
									ID:   "hello-world-2:latest",
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
					{ID: "", Name: ""},
					{ID: "scratch", Name: "scratch"},
				},
				Stages: []Stage{
					{
						Layers: []Layer{{
							Label:    "FROM",
							WaitFors: []WaitFor{{ID: "", Type: waitForType(waitForFrom)}},
						}},
					},
					{
						Layers: []Layer{{
							Label:    "FROM scratch",
							WaitFors: []WaitFor{{ID: "scratch", Type: waitForType(waitForFrom)}},
						}},
					},
				},
				BeforeFirstStage: []Layer{
					{Label: "ARG IMAGE1=$IMAGE2"},
					{Label: "ARG IMAGE2=scratch"},
				},
			},
		},
		{
			name: "Separate scratch images",
			args: args{
				content: []byte(`
FROM scratch AS app1
COPY app1.txt /app1.txt

FROM scratch AS app2
COPY app2.txt /app2.txt
`),
				maxLabelLength:  20,
				separateScratch: true,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{ID: "scratch-0", Name: "scratch"},
					{ID: "scratch-1", Name: "scratch"},
				},
				Stages: []Stage{
					{
						Name: "app1",
						Layers: []Layer{
							{
								Label:    "FROM scratch AS app1",
								WaitFors: []WaitFor{{ID: "scratch-0", Type: waitForType(waitForFrom)}},
							},
							{Label: "COPY app1.txt /ap..."},
						},
					},
					{
						Name: "app2",
						Layers: []Layer{
							{
								Label:    "FROM scratch AS app2",
								WaitFors: []WaitFor{{ID: "scratch-1", Type: waitForType(waitForFrom)}},
							},
							{Label: "COPY app2.txt /ap..."},
						},
					},
				},
			},
		},
		{
			name: "Single scratch image with separate flag",
			args: args{
				content: []byte(`FROM scratch AS app
COPY app.txt /app.txt`),
				maxLabelLength:  20,
				separateScratch: true,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{ID: "scratch-0", Name: "scratch"},
				},
				Stages: []Stage{
					{
						Name: "app",
						Layers: []Layer{
							{
								Label:    "FROM scratch AS app",
								WaitFors: []WaitFor{{ID: "scratch-0", Type: waitForType(waitForFrom)}},
							},
							{Label: "COPY app.txt /app..."},
						},
					},
				},
			},
		},
		{
			name: "No scratch images with separate flag enabled",
			args: args{
				content: []byte(`FROM ubuntu AS base
COPY app.txt /app.txt

FROM alpine AS final
COPY --from=base /app.txt /final.txt`),
				maxLabelLength:  20,
				separateScratch: true,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
				},
				Stages: []Stage{
					{
						Name: "base",
						Layers: []Layer{
							{
								Label:    "FROM ubuntu AS base",
								WaitFors: []WaitFor{{ID: "ubuntu", Type: waitForType(waitForFrom)}},
							},
							{Label: "COPY app.txt /app..."},
						},
					},
					{
						Name: "final",
						Layers: []Layer{
							{
								Label:    "FROM alpine AS final",
								WaitFors: []WaitFor{{ID: "alpine", Type: waitForType(waitForFrom)}},
							},
							{
								Label:    "COPY --from=base ...",
								WaitFors: []WaitFor{{ID: "base", Type: waitForType(waitForCopy)}},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dockerfileToSimplifiedDockerfile(
				tt.args.content,
				tt.args.maxLabelLength,
				tt.args.separateScratch,
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
