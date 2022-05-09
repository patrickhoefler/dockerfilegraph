package dockerfile2dot

import (
	"strings"
	"testing"
)

func TestBuildDotFile(t *testing.T) {
	type args struct {
		simplifiedDockerfile SimplifiedDockerfile
		legend               bool
		layers               bool
	}
	tests := []struct {
		name         string
		args         args
		wantContains string
	}{
		{
			name: "simple test",
			args: args{
				simplifiedDockerfile: SimplifiedDockerfile{
					BaseImages: []BaseImage{
						{ID: "build"},
						{ID: "release"},
					},
					Stages: []Stage{
						{
							ID: "release",
							WaitFor: []WaitFor{
								{
									ID:   "build",
									Type: waitForType(from),
								},
							},
						},
					},
				},
				legend: true,
				layers: true,
			},
			wantContains: "release",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildDotFile(tt.args.simplifiedDockerfile, tt.args.legend, tt.args.layers); !strings.Contains(got, tt.wantContains) {
				t.Errorf("BuildDotFile() = %v, did not contain %v", got, tt.wantContains)
			}
		})
	}
}
