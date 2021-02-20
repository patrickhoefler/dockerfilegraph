package dockerfile2dot

import (
	"errors"
	"log"
	"os"

	"github.com/spf13/afero"
)

// LoadAndParseDockerfile looks for the Dockerfile and returns a SimplifiedDockerfile
func LoadAndParseDockerfile() (simplifiedDockerfile SimplifiedDockerfile, err error) {
	return loadAndParseDockerfile(afero.NewOsFs())
}

func loadAndParseDockerfile(AppFs afero.Fs) (SimplifiedDockerfile, error) {
	for _, filename := range []string{"Dockerfile"} {
		content, err := afero.ReadFile(AppFs, filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				log.Fatal(err)
			}
		}

		return dockerfileToSimplifiedDockerfile(content), nil
	}

	return SimplifiedDockerfile{}, errors.New("Could not find any Dockerfile in the current working directory")
}
