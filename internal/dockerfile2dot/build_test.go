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
						{Name: "build"},
						{Name: "release"},
					},
					Stages: []Stage{
						{
							Layers: []Layer{
								{
									Label: "FROM...",
									WaitFor: WaitFor{
										Name: "build",
										Type: waitForType(from),
									},
								},
							},
						},
					},
				},
				edgestyle:      "default",
				legend:         true,
				maxLabelLength: 20,
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
						{Name: "build"},
						{Name: "release"},
					},
					Stages: []Stage{
						{
							Layers: []Layer{
								{
									Label: "FROM...",
									WaitFor: WaitFor{
										Name: "build",
										Type: waitForType(from),
									},
								},
							},
						},
					},
				},
				edgestyle:      "default",
				layers:         true,
				maxLabelLength: 20,
			},
			wantContains: "release",
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
			); !strings.Contains(got, tt.wantContains) {
				t.Errorf(
					"BuildDotFile() = %v, did not contain %v", got, tt.wantContains,
				)
			}
		})
	}
}
