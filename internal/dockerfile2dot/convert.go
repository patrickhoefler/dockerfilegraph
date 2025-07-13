package dockerfile2dot

import (
	"bytes"
	"fmt"
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
	scratchMode string,
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
			stage, layer := processFromInstruction(node, argReplacements, maxLabelLength, scratchMode, stages)
			simplifiedDockerfile.Stages = append(simplifiedDockerfile.Stages, stage)

			// Add a new layer
			layerIndex = 0
			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		case instructionCopy:
			// Add a new layer
			layerIndex++
			layer := processCopyInstruction(node, argReplacements, maxLabelLength, scratchMode)
			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		case instructionRun:
			// Add a new layer
			layerIndex++
			layer := processRunInstruction(node, argReplacements, maxLabelLength, scratchMode)
			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)

		default:
			// Add a new layer
			layerIndex++

			if stageIndex == -1 {
				layer := processBeforeFirstStage(node, &argReplacements, maxLabelLength)
				simplifiedDockerfile.BeforeFirstStage = append(
					simplifiedDockerfile.BeforeFirstStage,
					layer,
				)
				break
			}

			layer := newLayer(node, argReplacements, maxLabelLength)
			simplifiedDockerfile.Stages[stageIndex].Layers = append(
				simplifiedDockerfile.Stages[stageIndex].Layers,
				layer,
			)
		}
	}

	addExternalImages(&simplifiedDockerfile, stages, scratchMode)

	return
}

// shouldSkipScratchWaitFor returns true if scratch WaitFors should be skipped in hidden mode
func shouldSkipScratchWaitFor(scratchMode, waitForID string) bool {
	return scratchMode == "hidden" && waitForID == "scratch"
}

// processFromInstruction handles FROM instruction parsing
func processFromInstruction(
	node *parser.Node,
	argReplacements []ArgReplacement,
	maxLabelLength int,
	scratchMode string,
	stages map[string]struct{},
) (Stage, Layer) {
	stage := Stage{}

	// If there is an "AS" alias, set it as the name
	if node.Next.Next != nil {
		stage.Name = node.Next.Next.Next.Value
		stages[stage.Name] = struct{}{}
	}

	layer := newLayer(node, argReplacements, maxLabelLength)

	// Set the waitFor ID (skip scratch in hidden mode)
	waitForID := replaceArgVars(node.Next.Value, argReplacements)
	if !shouldSkipScratchWaitFor(scratchMode, waitForID) {
		layer.WaitFors = []WaitFor{{
			ID:   waitForID,
			Type: waitForType(waitForFrom),
		}}
	}

	return stage, layer
}

// processCopyInstruction handles COPY instruction parsing
func processCopyInstruction(
	node *parser.Node,
	argReplacements []ArgReplacement,
	maxLabelLength int,
	scratchMode string,
) Layer {
	layer := newLayer(node, argReplacements, maxLabelLength)

	// If there is a "--from" option, set the waitFor ID (skip scratch in hidden mode)
	for _, flag := range node.Flags {
		result := fromFlagRegex.FindSubmatch([]byte(flag))
		if len(result) > 1 {
			fromID := string(result[1])
			if !shouldSkipScratchWaitFor(scratchMode, fromID) {
				layer.WaitFors = []WaitFor{{
					ID:   fromID,
					Type: waitForType(waitForCopy),
				}}
			}
		}
	}

	return layer
}

// processRunInstruction handles RUN instruction parsing
func processRunInstruction(
	node *parser.Node,
	argReplacements []ArgReplacement,
	maxLabelLength int,
	scratchMode string,
) Layer {
	layer := newLayer(node, argReplacements, maxLabelLength)

	// If there is a "--mount=(.*)from=..." option, set the waitFor ID (skip scratch in hidden mode)
	for _, flag := range node.Flags {
		matches := mountFlagRegex.FindAllSubmatch([]byte(flag), -1)
		for _, match := range matches {
			if len(match) > 1 {
				mountID := string(match[1])
				if !shouldSkipScratchWaitFor(scratchMode, mountID) {
					layer.WaitFors = append(layer.WaitFors, WaitFor{
						ID:   mountID,
						Type: waitForType(waitForMount),
					})
				}
			}
		}
	}

	return layer
}

// processBeforeFirstStage handles instructions before the first FROM
func processBeforeFirstStage(
	node *parser.Node,
	argReplacements *[]ArgReplacement,
	maxLabelLength int,
) Layer {
	layer := newLayer(node, *argReplacements, maxLabelLength)

	// NOTE: Currently, only global ARGs (defined before the first FROM instruction)
	// are processed for variable substitution. Stage-specific ARGs are not yet fully supported.
	if strings.ToUpper(node.Value) == instructionArg {
		key, value, valueProvided := strings.Cut(node.Next.Value, "=")
		if valueProvided {
			*argReplacements = appendAndResolveArgReplacement(*argReplacements, ArgReplacement{Key: key, Value: value})
		}
	}

	return layer
}

// addExternalImages processes all layers and identifies external images.
func addExternalImages(
	simplifiedDockerfile *SimplifiedDockerfile,
	stages map[string]struct{},
	scratchMode string,
) {
	// Counter to generate unique IDs for separate scratch instances
	scratchCounter := 0

	// Iterate through all stages and layers to find external image dependencies
	for stageIndex, stage := range simplifiedDockerfile.Stages {
		for layerIndex, layer := range stage.Layers {
			// Process WaitFors for external image collection
			for waitForIndex, waitFor := range layer.WaitFors {
				// Skip if this references an internal stage (not an external image)
				if _, ok := stages[waitFor.ID]; ok {
					continue
				}

				imageID := waitFor.ID
				originalName := waitFor.ID

				// Handle scratch image modes
				if originalName == "scratch" {
					switch scratchMode {
					case "separated":
						// Generate unique IDs while preserving display name
						imageID = fmt.Sprintf("scratch-%d", scratchCounter)
						scratchCounter++
						// Update the layer's waitFor reference to use the unique ID for graph connections
						simplifiedDockerfile.Stages[stageIndex].Layers[layerIndex].WaitFors[waitForIndex].ID = imageID
					case "hidden":
						// Skip adding to external images entirely
						continue
					case "collapsed":
						// Default behavior - use original ID (no changes needed)
					default:
						// Default to collapsed for unknown modes
					}
				}

				// Add to external images if not already present
				addExternalImageIfNotExists(&simplifiedDockerfile.ExternalImages, imageID, originalName)
			}
		}
	}
}

// addExternalImageIfNotExists adds an external image if it doesn't already exist
func addExternalImageIfNotExists(externalImages *[]ExternalImage, imageID, originalName string) {
	// Avoid duplicate external image entries (based on unique ID)
	for _, externalImage := range *externalImages {
		if externalImage.ID == imageID {
			return // Already exists
		}
	}

	// Add the external image with proper ID/Name separation
	*externalImages = append(*externalImages, ExternalImage{
		ID:   imageID,      // Unique identifier for graph connections
		Name: originalName, // Display name for graph labels
	})
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
