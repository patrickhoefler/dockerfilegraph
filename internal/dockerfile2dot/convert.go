package dockerfile2dot

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func dockerfileToSimplifiedDockerfile(content []byte) SimplifiedDockerfile {
	result, err := parser.Parse(bytes.NewReader(content))
	if err != nil {
		log.Fatal(err)
	}

	simplifiedDockerfile := SimplifiedDockerfile{}

	// Set that holds all stage IDs
	stages := make(map[string]struct{})

	// Add all stages
	stageIndex := -1
	for _, child := range result.AST.Children {
		switch child.Value {
		case "from":
			stageIndex++

			stage := Stage{}
			stage.ID = fmt.Sprint(stageIndex)
			stage.WaitFor = []WaitFor{
				{
					ID:   child.Next.Value,
					Type: waitForType(from),
				},
			}

			// If there is an "AS" alias, set is at the name
			if child.Next.Next != nil {
				stages[child.Next.Next.Next.Value] = struct{}{}
				stage.Name = child.Next.Next.Next.Value
			}

			simplifiedDockerfile.Stages = append(simplifiedDockerfile.Stages, stage)

		case "copy":
			for _, flag := range child.Flags {
				regex := regexp.MustCompile("--from=(.+)")
				result := regex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					simplifiedDockerfile.Stages[stageIndex].WaitFor = append(
						simplifiedDockerfile.Stages[stageIndex].WaitFor,
						WaitFor{
							ID:   string(result[1]),
							Type: waitForType(copy),
						},
					)
				}
			}

		case "run":
			for _, flag := range child.Flags {
				regex := regexp.MustCompile("--mount=type=cache,.*from=(.+?)[, ]")
				result := regex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					simplifiedDockerfile.Stages[stageIndex].WaitFor = append(
						simplifiedDockerfile.Stages[stageIndex].WaitFor,
						WaitFor{
							ID:   string(result[1]),
							Type: waitForType(runMountTypeCache),
						},
					)
				}
			}
		}
	}

	// Set that holds all external base images
	baseImages := make(map[string]struct{})

	// Add external base images
	for _, stage := range simplifiedDockerfile.Stages {
		for _, waitFor := range stage.WaitFor {
			if _, ok := stages[waitFor.ID]; !ok {
				// simplifiedDockerfile.Stages[index].WaitFor[waitForIndex] = ""
				baseImages[waitFor.ID] = struct{}{}
				simplifiedDockerfile.BaseImages = append(
					simplifiedDockerfile.BaseImages,
					BaseImage{ID: waitFor.ID},
				)
			}
		}
	}

	return simplifiedDockerfile
}
