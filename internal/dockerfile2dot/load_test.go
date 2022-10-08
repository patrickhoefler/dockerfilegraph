package dockerfile2dot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func TestLoadAndParseDockerfile(t *testing.T) {
	type args struct {
		inputFS  afero.Fs
		filename string
	}

	dockerfileFS := afero.NewMemMapFs()
	_ = afero.WriteFile(dockerfileFS, "Dockerfile", []byte(`FROM scratch`), 0644)

	dockerfile2FS := afero.NewMemMapFs()
	_ = afero.WriteFile(dockerfile2FS, "Dockerfile2", []byte(`FROM docker`), 0644)

	tests := []struct {
		name    string
		args    args
		want    SimplifiedDockerfile
		wantErr bool
	}{
		{
			name: "no Dockerfile found",
			args: args{
				inputFS:  afero.NewMemMapFs(),
				filename: "Dockerfile",
			},
			wantErr: true,
		},
		{
			name: "wrong filename",
			args: args{
				inputFS:  dockerfileFS,
				filename: "Dockerfile2",
			},
			wantErr: true,
		},
		{
			name: "Dockerfile found",
			args: args{
				inputFS:  dockerfileFS,
				filename: "Dockerfile",
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
			name: "custom Dockerfile found",
			args: args{
				inputFS:  dockerfile2FS,
				filename: "Dockerfile2",
			},
			want: SimplifiedDockerfile{
				ExternalImages: []ExternalImage{{Name: "docker"}},
				Stages: []Stage{
					{
						Layers: []Layer{
							{
								Label:   "FROM docker",
								WaitFor: WaitFor{Name: "docker", Type: waitForType(from)}},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadAndParseDockerfile(tt.args.inputFS, tt.args.filename)
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
