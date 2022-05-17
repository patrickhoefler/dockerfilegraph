package dockerfile2dot

import (
	"errors"
	"io/fs"

	"github.com/spf13/afero"
)

// LoadAndParseDockerfile looks for the Dockerfile and returns a SimplifiedDockerfile
func LoadAndParseDockerfile(inputFS afero.Fs) (SimplifiedDockerfile, error) {
	for _, filename := range []string{"Dockerfile"} {
		content, err := afero.ReadFile(inputFS, filename)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			} else {
				panic(err)
			}
		}
		return dockerfileToSimplifiedDockerfile(content)
	}

	return SimplifiedDockerfile{}, errors.New("could not find any Dockerfile in the current working directory")
}
