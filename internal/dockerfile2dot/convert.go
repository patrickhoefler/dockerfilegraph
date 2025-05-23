package dockerfile2dot

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/aquilax/truncate"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// ArgReplacement holds a key-value pair for ARG variable substitution in Dockerfiles.
type ArgReplacement struct {
	Key   string
	Value string
}

const (
	instructionFrom = "FROM"
	instructionCopy = "COPY"
	instructionRun  = "RUN"
	instructionArg  = "ARG"
)

var (
	dollarVarRegex = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	bracedVarRegex = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	fromFlagRegex  = regexp.MustCompile("--from=(.+)")
	mountFlagRegex = regexp.MustCompile("--mount=.*from=(.+?)(?:,| |$)")
)

// newLayer creates a new layer object with a modified label.
func newLayer(
	node *parser.Node, argReplacements []ArgReplacement, maxLabelLength int,
) (layer Layer) {
	// Replace argument variables in the original label.
	label := replaceArgVars(node.Original, argReplacements)

	// Replace double quotes with single quotes.
	label = strings.ReplaceAll(label, "\"", "'")

	// Collapse multiple spaces into a single space.
	label = strings.Join(strings.Fields(label), " ")

	// Truncate the label if it exceeds the maximum length.
	if len(label) > maxLabelLength {
		label = truncate.Truncate(label, maxLabelLength, "...", truncate.PositionEnd)
	}

	// Set the label of the layer object.
	layer.Label = label

	return
}

func dockerfileToSimplifiedDockerfile(
	content []byte,
	maxLabelLength int,
) (simplifiedDockerfile SimplifiedDockerfile, err error) {
	result, err := parser.Parse(bytes.NewReader(content))
	if err != nil {
		return
	}

	// Set that holds all stage IDs
	stages := make(map[string]struct{})

	stageIndex := -1
	layerIndex := -1

	argReplacements := make([]ArgReplacement, 0)

	for _, node := range result.AST.Children {
		switch strings.ToUpper(node.Value) {
		case instructionFrom:
			// Create a new stage
			stageIndex++
			stage := Stage{}

			// If there is an "AS" alias, set is at the name
			if node.Next.Next != nil {
				stage.Name = node.Next.Next.Next.Value
				stages[stage.Name] = struct{}{}
			}

			simplifiedDockerfile.Stages = append(simplifiedDockerfile.Stages, stage)

			// Add a new layer
			layerIndex = 0
			layer := newLayer(node, argReplacements, maxLabelLength)

			// Set the waitFor ID
			layer.WaitFors = []WaitFor{{
				Name: replaceArgVars(node.Next.Value, argReplacements),
				Type: waitForType(waitForFrom),
			}}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		case instructionCopy:
			// Add a new layer
			layerIndex++
			layer := newLayer(node, argReplacements, maxLabelLength)

			// If there is a "--from" option, set the waitFor ID
			for _, flag := range node.Flags {
				result := fromFlagRegex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					layer.WaitFors = []WaitFor{{
						Name: string(result[1]),
						Type: waitForType(waitForCopy),
					}}
				}
			}

			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		case instructionRun:
			// Add a new layer
			layerIndex++
			layer := newLayer(node, argReplacements, maxLabelLength)

			// If there is a "--mount=(.*)from=..." option, set the waitFor ID
			for _, flag := range node.Flags {
				matches := mountFlagRegex.FindAllSubmatch([]byte(flag), -1)
				for _, match := range matches {
					if len(match) > 1 {
						layer.WaitFors = append(layer.WaitFors, WaitFor{
							Name: string(match[1]),
							Type: waitForType(waitForMount),
						})
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
			layer := newLayer(node, argReplacements, maxLabelLength)

			if stageIndex == -1 {
				simplifiedDockerfile.BeforeFirstStage = append(
					simplifiedDockerfile.BeforeFirstStage,
					layer,
				)

				// NOTE: Currently, only global ARGs (defined before the first FROM instruction)
				// are processed for variable substitution. Stage-specific ARGs are not yet fully supported.
				if strings.ToUpper(node.Value) == instructionArg {
					key, value, valueProvided := strings.Cut(node.Next.Value, "=")
					if valueProvided {
						argReplacements = appendAndResolveArgReplacement(argReplacements, ArgReplacement{Key: key, Value: value})
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
			for _, waitFor := range layer.WaitFors {

				// Check if the layer waits for a stage
				if _, ok := stages[waitFor.Name]; ok {
					continue
				}

				// Check if we already added the external image
				externalImageAlreadyAdded := false
				for _, externalImage := range simplifiedDockerfile.ExternalImages {
					if externalImage.Name == waitFor.Name {
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
					ExternalImage{Name: waitFor.Name},
				)
			}
		}
	}
}

// appendAndResolveArgReplacement appends a new ARG and resolves its value using already-resolved previous ARGs.
func appendAndResolveArgReplacement(
	argReplacements []ArgReplacement,
	newArgReplacement ArgReplacement,
) []ArgReplacement {
	// Resolve the new ARG using previous, already-resolved ARGs
	resolvedValue := newArgReplacement.Value
	for _, prevArg := range argReplacements {
		resolvedValue = strings.ReplaceAll(resolvedValue, "$"+prevArg.Key, prevArg.Value)
		resolvedValue = strings.ReplaceAll(resolvedValue, "${"+prevArg.Key+"}", prevArg.Value)
	}
	// Remove any remaining ARG patterns
	resolvedValue = stripRemainingArgPatterns(resolvedValue)
	return append(argReplacements, ArgReplacement{Key: newArgReplacement.Key, Value: resolvedValue})
}

// stripRemainingArgPatterns replaces any remaining $VAR or ${VAR} patterns in s with an empty string.
// It's intended to be called after defined ARGs have already been substituted into s.
func stripRemainingArgPatterns(s string) string {
	s = dollarVarRegex.ReplaceAllString(s, "")
	s = bracedVarRegex.ReplaceAllString(s, "")
	return s
}

// replaceArgVars replaces ARG variables in a string using fully resolved replacements.
func replaceArgVars(baseImage string, resolvedReplacements []ArgReplacement) string {
	result := baseImage
	for _, r := range resolvedReplacements {
		result = strings.ReplaceAll(result, "$"+r.Key, r.Value)
		result = strings.ReplaceAll(result, "${"+r.Key+"}", r.Value)
	}
	return result
}
