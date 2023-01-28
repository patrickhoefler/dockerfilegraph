package dockerfile2dot

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/aquilax/truncate"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func newLayer(
	child *parser.Node, replacements map[string]string,
) (layer Layer) {
	label := replaceArgVars(child.Original, replacements)
	label = strings.Replace(label, "\"", "'", -1)
	if len(label) > maxLabelLength {
		label = truncate.Truncate(
			label, maxLabelLength, "...", truncate.PositionEnd,
		)
	}
	layer.Label = label

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

	argReplacements := make(map[string]string)

	for _, child := range result.AST.Children {
		switch strings.ToUpper(child.Value) {
		case "FROM":
			// Create a new stage
			stageIndex++
			stage := Stage{}

			// If there is an "AS" alias, set is at the name
			if child.Next.Next != nil {
				stage.Name = child.Next.Next.Next.Value
				stages[stage.Name] = struct{}{}
			}

			simplifiedDockerfile.Stages = append(simplifiedDockerfile.Stages, stage)

			// Add a new layer
			layerIndex = 0
			layer := newLayer(child, argReplacements)

			// Set the waitFor ID
			layer.WaitFor = WaitFor{
				Name: replaceArgVars(child.Next.Value, argReplacements),
				Type: waitForType(from),
			}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		case "COPY":
			// Add a new layer
			layerIndex++
			layer := newLayer(child, argReplacements)

			// If there is a "--from" option, set the waitFor ID
			for _, flag := range child.Flags {
				regex := regexp.MustCompile("--from=(.+)")
				result := regex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					layer.WaitFor = WaitFor{
						Name: string(result[1]),
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
			layer := newLayer(child, argReplacements)

			// If there is a "--from" option, set the waitFor ID
			for _, flag := range child.Flags {
				regex := regexp.MustCompile("--mount=type=cache,.*from=(.+?)(?:,| |$)")
				result := regex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					layer.WaitFor = WaitFor{
						Name: string(result[1]),
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
			layer := newLayer(child, argReplacements)

			if stageIndex == -1 {
				simplifiedDockerfile.BeforeFirstStage = append(
					simplifiedDockerfile.BeforeFirstStage,
					layer,
				)

				if child.Value == "ARG" {
					key, value, valueProvided := strings.Cut(child.Next.Value, "=")
					if valueProvided {
						argReplacements[key] = value
					}
				}

				break
			}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)
		}
	}

	addExternalImages(&simplifiedDockerfile, stages)

	return
}

func addExternalImages(
	simplifiedDockerfile *SimplifiedDockerfile, stages map[string]struct{},
) {
	for _, stage := range simplifiedDockerfile.Stages {
		for _, layer := range stage.Layers {
			// Check if the layer waits for anything
			if layer.WaitFor.Name == "" {
				continue
			}

			// Check if the layer waits for a stage
			if _, ok := stages[layer.WaitFor.Name]; ok {
				continue
			}

			// Check if we already added the external image
			externalImageAlreadyAdded := false
			for _, externalImage := range simplifiedDockerfile.ExternalImages {
				if externalImage.Name == layer.WaitFor.Name {
					externalImageAlreadyAdded = true
					break
				}
			}
			if externalImageAlreadyAdded {
				continue
			}

			// Add the external image
			simplifiedDockerfile.ExternalImages = append(
				simplifiedDockerfile.ExternalImages,
				ExternalImage{Name: layer.WaitFor.Name},
			)
		}
	}
}

func replaceArgVars(baseImage string, replacements map[string]string) string {
	for k, v := range replacements {
		baseImage = strings.ReplaceAll(baseImage, "$"+k, v)
		baseImage = strings.ReplaceAll(baseImage, "${"+k+"}", v)
	}

	return baseImage
}
