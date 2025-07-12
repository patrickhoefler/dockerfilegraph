package dockerfile2dot

import (
	"strings"
	"testing"
)

func TestBuildDotFile(t *testing.T) {
	type args struct {
		simplifiedDockerfile SimplifiedDockerfile
		concentrate          bool
		edgestyle            string
		layers               bool
		legend               bool
		maxLabelLength       int
		nodesep              string
		ranksep              string
	}
	tests := []struct {
		name         string
		args         args
		wantContains string
	}{
		{
			name: "legend",
			args: args{
				simplifiedDockerfile: SimplifiedDockerfile{
					BeforeFirstStage: []Layer{
						{
							Label: "ARG...",
						},
					},
					ExternalImages: []ExternalImage{
						{ID: "build", Name: "build"},
						{ID: "release", Name: "release"},
					},
					Stages: []Stage{
						{
							Layers: []Layer{
								{
									Label: "FROM...",
									WaitFors: []WaitFor{{
										ID:   "build",
										Type: waitForType(waitForFrom),
									}},
								},
							},
						},
					},
				},
				edgestyle:      "default",
				legend:         true,
				maxLabelLength: 20,
				nodesep:        "0.5",
				ranksep:        "0.5",
			},
			wantContains: "release",
		},
		{
			name: "layers",
			args: args{
				simplifiedDockerfile: SimplifiedDockerfile{
					BeforeFirstStage: []Layer{
						{
							Label: "ARG...",
						},
					},
					ExternalImages: []ExternalImage{
						{ID: "build", Name: "build"},
						{ID: "release", Name: "release"},
					},
					Stages: []Stage{
						{
							Layers: []Layer{
								{
									Label: "FROM...",
									WaitFors: []WaitFor{{
										ID:   "build",
										Type: waitForType(waitForFrom),
									}},
								},
							},
						},
					},
				},
				edgestyle:      "default",
				layers:         true,
				maxLabelLength: 20,
				nodesep:        "0.5",
				ranksep:        "0.5",
			},
			wantContains: "release",
		},
		{
			name: "separate scratch images show correct labels",
			args: args{
				simplifiedDockerfile: SimplifiedDockerfile{
					ExternalImages: []ExternalImage{
						{ID: "scratch-0", Name: "scratch"},
						{ID: "scratch-1", Name: "scratch"},
					},
					Stages: []Stage{
						{
							Name: "app1",
							Layers: []Layer{
								{
									Label: "FROM scratch AS app1",
									WaitFors: []WaitFor{{
										ID:   "scratch-0",
										Type: waitForType(waitForFrom),
									}},
								},
							},
						},
						{
							Name: "app2",
							Layers: []Layer{
								{
									Label: "FROM scratch AS app2",
									WaitFors: []WaitFor{{
										ID:   "scratch-1",
										Type: waitForType(waitForFrom),
									}},
								},
							},
						},
					},
				},
				concentrate:    false,
				edgestyle:      "default",
				layers:         false,
				legend:         false,
				maxLabelLength: 20,
				nodesep:        "0.5",
				ranksep:        "0.5",
			},
			wantContains: `external_image_0 [ color=grey20, fontcolor=grey20, label="scratch", shape=box, ` +
				`style="dashed,rounded", width=2 ];
	external_image_1 [ color=grey20, fontcolor=grey20, label="scratch"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildDotFile(
				tt.args.simplifiedDockerfile,
				tt.args.concentrate,
				tt.args.edgestyle,
				tt.args.layers,
				tt.args.legend,
				tt.args.maxLabelLength,
				tt.args.nodesep,
				tt.args.ranksep,
			); !strings.Contains(got, tt.wantContains) {
				t.Errorf(
					"BuildDotFile() = %v, did not contain %v", got, tt.wantContains,
				)
			}
		})
	}
}
