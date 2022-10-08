// Package dockerfile2dot provides the functionality for loading a Dockerfile
// and converting it into a GraphViz DOT file.
package dockerfile2dot

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"
)

// LoadAndParseDockerfile looks for the Dockerfile and returns a
// SimplifiedDockerfile.
func LoadAndParseDockerfile(inputFS afero.Fs, filename string) (SimplifiedDockerfile, error) {
	content, err := afero.ReadFile(inputFS, filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			absFilePath, err := filepath.Abs(filename)
			if err != nil {
				panic(err)
			}
			return SimplifiedDockerfile{}, errors.New("could not find a Dockerfile at " + absFilePath)
		}
		panic(err)
	}
	return dockerfileToSimplifiedDockerfile(content)
}
