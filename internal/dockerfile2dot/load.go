// Package dockerfile2dot provides the functionality for loading a Dockerfile
// and converting it into a GraphViz DOT file.
package dockerfile2dot

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"
)

// LoadAndParseDockerfile looks for the Dockerfile and returns a
// SimplifiedDockerfile.
func LoadAndParseDockerfile(
	inputFS afero.Fs,
	filename string,
	maxLabelLength int,
	scratchMode string,
) (SimplifiedDockerfile, error) {
	content, err := afero.ReadFile(inputFS, filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			absFilePath, absErr := filepath.Abs(filename)
			if absErr != nil {
				return SimplifiedDockerfile{}, fmt.Errorf("could not get absolute path: %w", absErr)
			}
			return SimplifiedDockerfile{}, errors.New("could not find a Dockerfile at " + absFilePath)
		}
		return SimplifiedDockerfile{}, err
	}
	return dockerfileToSimplifiedDockerfile(content, maxLabelLength, scratchMode)
}
