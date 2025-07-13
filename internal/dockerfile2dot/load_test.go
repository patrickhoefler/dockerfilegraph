package dockerfile2dot

import (
	"testing"

	"github.com/spf13/afero"
)

// TestLoadAndParseDockerfile tests the file loading functionality of LoadAndParseDockerfile.
// This test focuses on file I/O concerns (file not found, path resolution) rather than
// Dockerfile parsing logic, which is tested in convert_test.go.
func TestLoadAndParseDockerfile(t *testing.T) {
	type args struct {
		inputFS  afero.Fs
		filename string
	}

	dockerfileFS := afero.NewMemMapFs()
	_ = afero.WriteFile(dockerfileFS, "Dockerfile", []byte(`FROM scratch`), 0o644)

	tests := []struct {
		name    string
		args    args
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
				inputFS:  dockerfileFS,
				filename: "Dockerfile",
			},
			wantErr: false,
		},
		{
			name: "should work in any directory",
			args: args{
				inputFS:  dockerfileFS,
				filename: "subdir/../Dockerfile",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadAndParseDockerfile(
				tt.args.inputFS,
				tt.args.filename,
				20,          // Default maxLabelLength
				"collapsed", // Default scratchMode
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAndParseDockerfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
