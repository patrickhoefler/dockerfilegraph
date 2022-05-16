package dockerfile2dot

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func storeDataInLayer(layerIndex int, child *parser.Node) Layer {
	layer := Layer{}
	layer.ID = fmt.Sprintf("%02d", layerIndex)
	layer.Name = child.Value + "..."
	return layer
}

func dockerfileToSimplifiedDockerfile(content []byte) SimplifiedDockerfile {
	result, err := parser.Parse(bytes.NewReader(content))
	if err != nil {
		panic(err)
	}

	simplifiedDockerfile := SimplifiedDockerfile{}

	// Set that holds all stage IDs
	stages := make(map[string]struct{})

	// Add all stages
	stageIndex := -1

	// Add all layers
	layerIndex := -1

	for _, child := range result.AST.Children {
		switch strings.ToUpper(child.Value) {
		case "FROM":
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

			// set layer index as stage index
			layerIndex++
			layer := storeDataInLayer(layerIndex, child)
			simplifiedDockerfile.Stages[stageIndex].Layers = append(simplifiedDockerfile.Stages[stageIndex].Layers, layer)

		case "COPY":
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

			// creates a layer struct with the child data
			layerIndex++
			layer := storeDataInLayer(layerIndex, child)
			simplifiedDockerfile.Stages[stageIndex].Layers = append(simplifiedDockerfile.Stages[stageIndex].Layers, layer)

		case "RUN":
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

			// creates a layer struct with the child data
			layerIndex++
			layer := storeDataInLayer(layerIndex, child)
			simplifiedDockerfile.Stages[stageIndex].Layers = append(simplifiedDockerfile.Stages[stageIndex].Layers, layer)

		default:
			// creates a layer struct with the child data
			layerIndex++
			layer := storeDataInLayer(layerIndex, child)

			// check if the current stages array is empty
			// this usually happens when ARGs are provided before a FROM statement
			if len(simplifiedDockerfile.Stages) == 0 {
				simplifiedDockerfile.LayersNotStage = append(simplifiedDockerfile.LayersNotStage, layer)
				break
			}
			simplifiedDockerfile.Stages[stageIndex].Layers = append(simplifiedDockerfile.Stages[stageIndex].Layers, layer)
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
