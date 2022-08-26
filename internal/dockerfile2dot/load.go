// Package dockerfile2dot provides the functionality for loading a Dockerfile
// and converting it into a GraphViz DOT file.
package dockerfile2dot

import (
	"errors"
	"io/fs"

	"github.com/spf13/afero"
)

// LoadAndParseDockerfile looks for the Dockerfile and returns a
// SimplifiedDockerfile.
func LoadAndParseDockerfile(inputFS afero.Fs) (SimplifiedDockerfile, error) {
	content, err := afero.ReadFile(inputFS, "Dockerfile")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return SimplifiedDockerfile{}, errors.New("could not find any Dockerfile in the current working directory")
		}
		panic(err)
	}
	return dockerfileToSimplifiedDockerfile(content)
}
