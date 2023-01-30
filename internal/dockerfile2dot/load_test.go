package dockerfile2dot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func TestLoadAndParseDockerfile(t *testing.T) {
	type args struct {
		inputFS        afero.Fs
		filename       string
		maxLabelLength int
	}

	dockerfileFS := afero.NewMemMapFs()
	_ = afero.WriteFile(dockerfileFS, "Dockerfile", []byte(`FROM scratch`), 0644)

	tests := []struct {
		name    string
		args    args
		want    SimplifiedDockerfile
		wantErr bool
	}{
		{
			name: "Dockerfile not found",
			args: args{
				inputFS:  dockerfileFS,
				filename: "missing/Dockerfile",
			},
			wantErr: true,
		},
		{
			name: "should work in the current working directory",
			args: args{
				inputFS:        dockerfileFS,
				filename:       "Dockerfile",
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{{Name: "scratch"}},
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
			name: "should work in any directory",
			args: args{
				inputFS:        dockerfileFS,
				filename:       "subdir/../Dockerfile",
				maxLabelLength: 20,
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{{Name: "scratch"}},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadAndParseDockerfile(
				tt.args.inputFS,
				tt.args.filename,
				tt.args.maxLabelLength,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAndParseDockerfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LoadAndParseDockerfile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
