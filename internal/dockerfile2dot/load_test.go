package dockerfile2dot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func TestLoadAndParseDockerfile(t *testing.T) {
	type args struct {
		inputFS afero.Fs
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
			name: "no Dockerfile found",
			args: args{
				inputFS: afero.NewMemMapFs(),
			},
			wantErr: true,
		},
		{
			name: "Dockerfile found",
			args: args{
				inputFS: dockerfileFS,
			},
			want: SimplifiedDockerfile{
				BaseImages: []BaseImage{{ID: "scratch"}},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadAndParseDockerfile(tt.args.inputFS)
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
