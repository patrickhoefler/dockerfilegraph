package dockerfile2dot

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func newLayer(layerIndex int, child *parser.Node) (layer Layer) {
	layer.ID = fmt.Sprint(layerIndex)
	layer.Name = child.Value + "..."
	return
}

func dockerfileToSimplifiedDockerfile(content []byte) (
	simplifiedDockerfile SimplifiedDockerfile, err error,
) {
	result, err := parser.Parse(bytes.NewReader(content))
	if err != nil {
		return
	}

	// Set that holds all stage IDs
	stages := make(map[string]struct{})

	stageIndex := -1
	layerIndex := -1

	for _, child := range result.AST.Children {
		switch strings.ToUpper(child.Value) {
		case "FROM":
			// Create a new stage
			stageIndex++
			stage := Stage{
				ID: fmt.Sprint(stageIndex),
			}

			// If there is an "AS" alias, set is at the name
			if child.Next.Next != nil {
				stages[child.Next.Next.Next.Value] = struct{}{}
				stage.Name = child.Next.Next.Next.Value
			}

			simplifiedDockerfile.Stages = append(simplifiedDockerfile.Stages, stage)

			// Add a new layer
			layerIndex = 0
			layer := newLayer(layerIndex, child)

			// Set the waitFor ID
			layer.WaitFor = WaitFor{
				ID:   child.Next.Value,
				Type: waitForType(from),
			}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		case "COPY":
			// Add a new layer
			layerIndex++
			layer := newLayer(layerIndex, child)

			// If there is a "--from" option, set the waitFor ID
			for _, flag := range child.Flags {
				regex := regexp.MustCompile("--from=(.+)")
				result := regex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					layer.WaitFor = WaitFor{
						ID:   string(result[1]),
						Type: waitForType(copy),
					}
				}
			}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		case "RUN":
			// Add a new layer
			layerIndex++
			layer := newLayer(layerIndex, child)

			// If there is a "--from" option, set the waitFor ID
			for _, flag := range child.Flags {
				regex := regexp.MustCompile("--mount=type=cache,.*from=(.+?)[, ]")
				result := regex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					layer.WaitFor = WaitFor{
						ID:   string(result[1]),
						Type: waitForType(runMountTypeCache),
					}
				}
			}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		default:
			// Add a new layer
			layerIndex++
			layer := newLayer(layerIndex, child)

			if stageIndex == -1 {
				simplifiedDockerfile.BeforeFirstStage = append(
					simplifiedDockerfile.BeforeFirstStage,
					layer,
				)
				break
			}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)
		}
	}

	// Set that holds all external base images
	baseImages := make(map[string]struct{})

	// Add external base images
	for _, stage := range simplifiedDockerfile.Stages {
		for _, layer := range stage.Layers {
			if layer.WaitFor.ID == "" {
				continue
			}
			if _, ok := stages[layer.WaitFor.ID]; !ok {
				baseImages[layer.WaitFor.ID] = struct{}{}
				simplifiedDockerfile.BaseImages = append(
					simplifiedDockerfile.BaseImages,
					BaseImage{ID: layer.WaitFor.ID},
				)
			}
		}
	}

	return
}
