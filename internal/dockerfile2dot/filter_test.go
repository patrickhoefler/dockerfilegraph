package dockerfile2dot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// stageFrom is a helper that builds a Stage with a single FROM layer.
func stageFrom(name, image string, wfType waitForType) Stage {
	return Stage{
		Name: name,
		Layers: []Layer{{
			Label:    "FROM " + image,
			WaitFors: []WaitFor{{ID: image, Type: wfType}},
		}},
	}
}

func Test_filterToTargets(t *testing.T) {
	tests := []struct {
		name    string
		sdf     SimplifiedDockerfile
		targets []string
		want    SimplifiedDockerfile
		wantErr bool
	}{
		{
			name: "single target retains only that stage and its ancestors",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					stageFrom("mid", "base", waitForFrom),
					stageFrom("final", "mid", waitForFrom),
					stageFrom("unrelated", "alpine", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
				},
			},
			targets: []string{"final"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					stageFrom("mid", "base", waitForFrom),
					stageFrom("final", "mid", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
		},
		{
			name: "single target with no deps retains only that stage",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("standalone", "alpine", waitForFrom),
					stageFrom("other", "ubuntu", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "alpine", Name: "alpine"},
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
			targets: []string{"standalone"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("standalone", "alpine", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "alpine", Name: "alpine"},
				},
			},
		},
		{
			name: "two targets union their ancestors",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					stageFrom("app1", "base", waitForFrom),
					stageFrom("app2", "alpine", waitForFrom),
					stageFrom("unused", "scratch", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
					{ID: "scratch", Name: "scratch"},
				},
			},
			targets: []string{"app1", "app2"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					stageFrom("app1", "base", waitForFrom),
					stageFrom("app2", "alpine", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
				},
			},
		},
		{
			name: "two targets with shared ancestor",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("shared", "ubuntu", waitForFrom),
					stageFrom("app1", "shared", waitForFrom),
					stageFrom("app2", "shared", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
			targets: []string{"app1", "app2"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("shared", "ubuntu", waitForFrom),
					stageFrom("app1", "shared", waitForFrom),
					stageFrom("app2", "shared", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
		},
		{
			name: "COPY and RUN mount dependencies are followed",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("builder", "ubuntu", waitForFrom),
					stageFrom("cache", "alpine", waitForFrom),
					{
						Name: "final",
						Layers: []Layer{
							{Label: "FROM scratch", WaitFors: []WaitFor{{ID: "scratch", Type: waitForFrom}}},
							{Label: "COPY --from=builder", WaitFors: []WaitFor{{ID: "builder", Type: waitForCopy}}},
							{Label: "RUN --mount=from=cache", WaitFors: []WaitFor{{ID: "cache", Type: waitForMount}}},
						},
					},
					stageFrom("unused", "debian", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
					{ID: "scratch", Name: "scratch"},
					{ID: "debian", Name: "debian"},
				},
			},
			targets: []string{"final"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("builder", "ubuntu", waitForFrom),
					stageFrom("cache", "alpine", waitForFrom),
					{
						Name: "final",
						Layers: []Layer{
							{Label: "FROM scratch", WaitFors: []WaitFor{{ID: "scratch", Type: waitForFrom}}},
							{Label: "COPY --from=builder", WaitFors: []WaitFor{{ID: "builder", Type: waitForCopy}}},
							{Label: "RUN --mount=from=cache", WaitFors: []WaitFor{{ID: "cache", Type: waitForMount}}},
						},
					},
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
					{ID: "scratch", Name: "scratch"},
				},
			},
		},
		{
			// Stages: 0=base, 1=unused, 2=final (COPY --from=0)
			// After filtering to "final": stages become 0=base, 1=final
			// WaitFor "0" stays "0" (base is still at index 0 in filtered result)
			name: "numeric WaitFor IDs are remapped after stage elision (no shift)",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					stageFrom("unused", "alpine", waitForFrom),
					{
						Name: "final",
						Layers: []Layer{
							{Label: "FROM scratch", WaitFors: []WaitFor{{ID: "scratch", Type: waitForFrom}}},
							{Label: "COPY --from=0", WaitFors: []WaitFor{{ID: "0", Type: waitForCopy}}},
						},
					},
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
					{ID: "scratch", Name: "scratch"},
				},
			},
			targets: []string{"final"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					{
						Name: "final",
						Layers: []Layer{
							{Label: "FROM scratch", WaitFors: []WaitFor{{ID: "scratch", Type: waitForFrom}}},
							{Label: "COPY --from=0", WaitFors: []WaitFor{{ID: "0", Type: waitForCopy}}},
						},
					},
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "scratch", Name: "scratch"},
				},
			},
		},
		{
			// Stages: 0=elided, 1=base, 2=final (COPY --from=1)
			// After filtering to "final": stages become 0=base, 1=final
			// WaitFor "1" in final becomes "0"
			name: "numeric WaitFor IDs are remapped when earlier stage is elided",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("elided", "debian", waitForFrom),
					stageFrom("base", "ubuntu", waitForFrom),
					{
						Name: "final",
						Layers: []Layer{
							{Label: "FROM scratch", WaitFors: []WaitFor{{ID: "scratch", Type: waitForFrom}}},
							{Label: "COPY --from=1", WaitFors: []WaitFor{{ID: "1", Type: waitForCopy}}},
						},
					},
				},
				ExternalImages: []ExternalImage{
					{ID: "debian", Name: "debian"},
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "scratch", Name: "scratch"},
				},
			},
			targets: []string{"final"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					{
						Name: "final",
						Layers: []Layer{
							{Label: "FROM scratch", WaitFors: []WaitFor{{ID: "scratch", Type: waitForFrom}}},
							{Label: "COPY --from=1", WaitFors: []WaitFor{{ID: "0", Type: waitForCopy}}},
						},
					},
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "scratch", Name: "scratch"},
				},
			},
		},
		{
			name: "BeforeFirstStage is always preserved",
			sdf: SimplifiedDockerfile{
				BeforeFirstStage: []Layer{
					{Label: "ARG VERSION=1.0"},
				},
				Stages: []Stage{
					stageFrom("app", "alpine", waitForFrom),
					stageFrom("other", "ubuntu", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "alpine", Name: "alpine"},
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
			targets: []string{"app"},
			want: SimplifiedDockerfile{
				BeforeFirstStage: []Layer{
					{Label: "ARG VERSION=1.0"},
				},
				Stages: []Stage{
					stageFrom("app", "alpine", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "alpine", Name: "alpine"},
				},
			},
		},
		{
			name: "invalid target returns error",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					{Name: "app", Layers: []Layer{{Label: "FROM alpine"}}},
				},
			},
			targets: []string{"nonexistent"},
			wantErr: true,
		},
		{
			name: "target that is the first and only stage",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("only", "scratch", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "scratch", Name: "scratch"},
				},
			},
			targets: []string{"only"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("only", "scratch", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "scratch", Name: "scratch"},
				},
			},
		},
		{
			name: "external images not referenced by kept stages are elided",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("a", "img-a", waitForFrom),
					stageFrom("b", "img-b", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "img-a", Name: "img-a"},
					{ID: "img-b", Name: "img-b"},
				},
			},
			targets: []string{"a"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("a", "img-a", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "img-a", Name: "img-a"},
				},
			},
		},
		{
			// Stage "alpine" depends on internal stage "base" (same name as a common image).
			// The external image "alpine" from an elided stage must not be incorrectly retained
			// because "base" happens to match a stage name.
			name: "internal stage WaitFor IDs are not treated as external image references",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					{
						Name: "app",
						Layers: []Layer{{
							Label:    "FROM base",
							WaitFors: []WaitFor{{ID: "base", Type: waitForFrom}},
						}},
					},
					stageFrom("other", "alpine", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
					{ID: "alpine", Name: "alpine"},
				},
			},
			targets: []string{"app"},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("base", "ubuntu", waitForFrom),
					{
						Name: "app",
						Layers: []Layer{{
							Label:    "FROM base",
							WaitFors: []WaitFor{{ID: "base", Type: waitForFrom}},
						}},
					},
				},
				ExternalImages: []ExternalImage{
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
		},
		{
			name: "target with leading/trailing whitespace is normalized",
			sdf: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("app", "alpine", waitForFrom),
					stageFrom("other", "ubuntu", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "alpine", Name: "alpine"},
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
			targets: []string{" app ", "  other  "},
			want: SimplifiedDockerfile{
				Stages: []Stage{
					stageFrom("app", "alpine", waitForFrom),
					stageFrom("other", "ubuntu", waitForFrom),
				},
				ExternalImages: []ExternalImage{
					{ID: "alpine", Name: "alpine"},
					{ID: "ubuntu", Name: "ubuntu"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterToTargets(tt.sdf, tt.targets)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterToTargets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("filterToTargets() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
