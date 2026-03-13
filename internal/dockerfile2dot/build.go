package dockerfile2dot

import (
	"fmt"
	"maps"
	"strconv"

	"github.com/aquilax/truncate"
	"github.com/awalterschulze/gographviz"
)

// BuildDotFile builds a GraphViz .dot file from a simplified Dockerfile
func BuildDotFile(
	simplifiedDockerfile SimplifiedDockerfile,
	opts BuildOptions,
) (string, error) {
	// Create a new graph
	graph := gographviz.NewEscape()

	var graphErr error
	set := func(err error) {
		if graphErr == nil {
			graphErr = err
		}
	}

	set(graph.SetName("G"))
	set(graph.SetDir(true))
	set(graph.AddAttr("G", "compound", "true")) // allow edges between clusters
	set(graph.AddAttr("G", "nodesep", fmt.Sprintf("%.2f", opts.NodeSep)))
	set(graph.AddAttr("G", "rankdir", "LR"))
	set(graph.AddAttr("G", "ranksep", fmt.Sprintf("%.2f", opts.RankSep)))
	if opts.Concentrate {
		set(graph.AddAttr("G", "concentrate", "true"))
	}

	if graphErr != nil {
		return "", graphErr
	}

	// Add the legend if requested
	if opts.Legend {
		if err := addLegend(graph, opts.EdgeStyle); err != nil {
			return "", err
		}
	}

	if err := addExternalImagesToGraph(graph, simplifiedDockerfile, opts.MaxLabelLength); err != nil {
		return "", err
	}

	if err := addStages(graph, simplifiedDockerfile, opts.MaxLabelLength, opts.Layers, opts.EdgeStyle); err != nil {
		return "", err
	}

	// Add the ARGS that appear before the first stage, if layers are requested
	if opts.Layers {
		if err := addBeforeFirstStage(graph, simplifiedDockerfile); err != nil {
			return "", err
		}
	}

	return graph.String(), nil
}

func addExternalImagesToGraph(
	graph *gographviz.Escape,
	simplifiedDockerfile SimplifiedDockerfile,
	maxLabelLength int,
) error {
	var graphErr error
	set := func(err error) {
		if graphErr == nil {
			graphErr = err
		}
	}

	for externalImageIndex, externalImage := range simplifiedDockerfile.ExternalImages {
		label := externalImage.Name
		if len(label) > maxLabelLength {
			truncatePosition := truncate.PositionMiddle
			if maxLabelLength < 5 {
				truncatePosition = truncate.PositionEnd
			}
			label = truncate.Truncate(label, maxLabelLength, "...", truncatePosition)
		}

		set(graph.AddNode(
			"G",
			fmt.Sprintf("external_image_%d", externalImageIndex),
			map[string]string{
				"label":     "\"" + label + "\"",
				"shape":     "box",
				"width":     "2",
				"style":     "\"dashed,rounded\"",
				"color":     "grey20",
				"fontcolor": "grey20",
			},
		))
	}

	return graphErr
}

func addStages(
	graph *gographviz.Escape,
	simplifiedDockerfile SimplifiedDockerfile,
	maxLabelLength int,
	layers bool,
	edgestyle string,
) error {
	var graphErr error
	set := func(err error) {
		if graphErr == nil {
			graphErr = err
		}
	}

	for stageIndex, stage := range simplifiedDockerfile.Stages {
		attrs := map[string]string{
			"label": "\"" + getStageLabel(stageIndex, stage, maxLabelLength) + "\"",
			"shape": "box",
			"style": "rounded",
			"width": "2",
		}

		// Add layers if requested
		if layers {
			if err := addStageWithLayers(graph, simplifiedDockerfile, stageIndex, stage, attrs); err != nil {
				return err
			}
		} else {
			// Add the build stages.
			// Color the last one, because it is the default build target.
			if stageIndex == len(simplifiedDockerfile.Stages)-1 {
				attrs["style"] = "\"filled,rounded\""
				attrs["fillcolor"] = "grey90"
			}

			set(graph.AddNode("G", fmt.Sprintf("stage_%d", stageIndex), attrs))
		}

		if graphErr != nil {
			return graphErr
		}

		// Add the edges for this build stage
		if err := addEdgesForStage(
			stageIndex, stage, graph, simplifiedDockerfile, layers, edgestyle,
		); err != nil {
			return err
		}
	}

	return graphErr
}

func addStageWithLayers(
	graph *gographviz.Escape,
	simplifiedDockerfile SimplifiedDockerfile,
	stageIndex int,
	stage Stage,
	attrs map[string]string,
) error {
	var graphErr error
	set := func(err error) {
		if graphErr == nil {
			graphErr = err
		}
	}

	cluster := fmt.Sprintf("cluster_stage_%d", stageIndex)

	clusterAttrs := map[string]string{
		"label":  "\"" + getStageLabel(stageIndex, stage, 0) + "\"",
		"margin": "16",
	}

	if stageIndex == len(simplifiedDockerfile.Stages)-1 {
		clusterAttrs["style"] = "filled"
		clusterAttrs["fillcolor"] = "grey90"
	}

	set(graph.AddSubGraph("G", cluster, clusterAttrs))

	for layerIndex, layer := range stage.Layers {
		attrs["label"] = "\"" + layer.Label + "\""
		attrs["penwidth"] = "0.5"
		attrs["style"] = "\"filled,rounded\""
		attrs["fillcolor"] = "white"
		set(graph.AddNode(
			cluster,
			fmt.Sprintf("stage_%d_layer_%d", stageIndex, layerIndex),
			attrs,
		))

		// Add edges between layers to guarantee the correct order
		if layerIndex > 0 {
			set(graph.AddEdge(
				fmt.Sprintf("stage_%d_layer_%d", stageIndex, layerIndex-1),
				fmt.Sprintf("stage_%d_layer_%d", stageIndex, layerIndex),
				true,
				nil,
			))
		}
	}

	return graphErr
}

func addBeforeFirstStage(
	graph *gographviz.Escape,
	simplifiedDockerfile SimplifiedDockerfile,
) error {
	if len(simplifiedDockerfile.BeforeFirstStage) == 0 {
		return nil
	}

	var graphErr error
	set := func(err error) {
		if graphErr == nil {
			graphErr = err
		}
	}

	set(graph.AddSubGraph(
		"G",
		"cluster_before_first_stage",
		map[string]string{"label": "\"Before First Stage\""},
	))
	for argIndex, arg := range simplifiedDockerfile.BeforeFirstStage {
		set(graph.AddNode(
			"cluster_before_first_stage",
			fmt.Sprintf("before_first_stage_%d", argIndex),
			map[string]string{
				"label": "\"" + arg.Label + "\"",
				"shape": "box",
				"style": "rounded",
				"width": "2",
			},
		))
	}

	return graphErr
}

func addEdgesForStage(
	stageIndex int, stage Stage, graph *gographviz.Escape,
	simplifiedDockerfile SimplifiedDockerfile, layers bool, edgestyle string,
) error {
	for layerIndex, layer := range stage.Layers {
		for _, waitFor := range layer.WaitFors {
			edgeAttrs := map[string]string{}
			if waitFor.Type == waitForType(waitForCopy) {
				edgeAttrs["arrowhead"] = "empty"
				if edgestyle == "default" {
					edgeAttrs["style"] = "dashed"
				}
			} else if waitFor.Type == waitForType(waitForMount) {
				edgeAttrs["arrowhead"] = "ediamond"
				if edgestyle == "default" {
					edgeAttrs["style"] = "dotted"
				}
			}

			sourceNodeID, additionalEdgeAttrs, err := getWaitForNodeID(
				simplifiedDockerfile, waitFor.ID, layers,
			)
			if err != nil {
				return err
			}
			maps.Copy(edgeAttrs, additionalEdgeAttrs)

			targetNodeID := fmt.Sprintf("stage_%d", stageIndex)
			if layers {
				targetNodeID = targetNodeID + fmt.Sprintf("_layer_%d", layerIndex)
			}

			if err := graph.AddEdge(sourceNodeID, targetNodeID, true, edgeAttrs); err != nil {
				return err
			}
		}
	}
	return nil
}

func addLegend(graph *gographviz.Escape, edgestyle string) error {
	var graphErr error
	set := func(err error) {
		if graphErr == nil {
			graphErr = err
		}
	}

	set(graph.AddSubGraph("G", "cluster_legend", nil))

	set(graph.AddNode("cluster_legend", "key",
		map[string]string{
			"shape":    "plaintext",
			"fontname": "monospace",
			"fontsize": "10",
			"label": `<<table border="0" cellpadding="2" cellspacing="0" cellborder="0">
	<tr><td align="right" port="i0">FROM&nbsp;...&nbsp;</td></tr>
	<tr><td align="right" port="i1">COPY --from=...&nbsp;</td></tr>
	<tr><td align="right" port="i2">RUN --mount=(.*)from=...&nbsp;</td></tr>
</table>>`,
		},
	))
	set(graph.AddNode("cluster_legend", "key2",
		map[string]string{
			"shape":    "plaintext",
			"fontname": "monospace",
			"fontsize": "10",
			"label": `<<table border="0" cellpadding="2" cellspacing="0" cellborder="0">
	<tr><td port="i0">&nbsp;</td></tr>
	<tr><td port="i1">&nbsp;</td></tr>
	<tr><td port="i2">&nbsp;</td></tr>
</table>>`,
		},
	))

	set(graph.AddPortEdge("key", "i0:e", "key2", "i0:w", true, nil))

	copyEdgeAttrs := map[string]string{"arrowhead": "empty"}
	if edgestyle == "default" {
		copyEdgeAttrs["style"] = "dashed"
	}
	set(graph.AddPortEdge(
		"key", "i1:e", "key2", "i1:w", true,
		copyEdgeAttrs,
	))

	mountEdgeAttrs := map[string]string{"arrowhead": "ediamond"}
	if edgestyle == "default" {
		mountEdgeAttrs["style"] = "dotted"
	}
	set(graph.AddPortEdge(
		"key", "i2:e", "key2", "i2:w", true,
		mountEdgeAttrs,
	))

	return graphErr
}

func getStageLabel(stageIndex int, stage Stage, maxLabelLength int) string {
	if maxLabelLength > 0 && len(stage.Name) > maxLabelLength {
		return truncate.Truncate(
			stage.Name, maxLabelLength, "...", truncate.PositionEnd,
		)
	}

	if stage.Name == "" {
		return fmt.Sprintf("%d", stageIndex)
	}

	return stage.Name
}

// getWaitForNodeID returns the ID of the node identified by the stage ID or
// name or the external image name.
func getWaitForNodeID(
	simplifiedDockerfile SimplifiedDockerfile, nameOrID string, layers bool,
) (string, map[string]string, error) {
	attrs := map[string]string{}

	// If it can be converted to an integer, it's a numeric stage reference
	if stageIndex, convertErr := strconv.Atoi(nameOrID); convertErr == nil {
		if stageIndex < 0 || stageIndex >= len(simplifiedDockerfile.Stages) {
			return "", nil, fmt.Errorf(
				"stage index %d out of range (have %d stages)",
				stageIndex, len(simplifiedDockerfile.Stages),
			)
		}
		return stageNodeID(simplifiedDockerfile, stageIndex, nameOrID, layers)
	}

	// Check if it's a stage name
	if stageIndex, found := findStageIndex(simplifiedDockerfile.Stages, nameOrID); found {
		return stageNodeID(simplifiedDockerfile, stageIndex, nameOrID, layers)
	}

	// Check if it's an external image ID
	for externalImageIndex, externalImage := range simplifiedDockerfile.ExternalImages {
		if nameOrID == externalImage.ID {
			return fmt.Sprintf("external_image_%d", externalImageIndex), attrs, nil
		}
	}

	return "", nil, fmt.Errorf(
		"could not resolve node ID for %q (expected stage index, stage name, or external image ID)",
		nameOrID,
	)
}

// stageNodeID returns the graph node ID for a stage, handling the layers case.
func stageNodeID(
	sdf SimplifiedDockerfile, stageIndex int, nameOrID string, layers bool,
) (string, map[string]string, error) {
	if layers {
		if len(sdf.Stages[stageIndex].Layers) == 0 {
			return "", nil, fmt.Errorf("stage %q has no layers", nameOrID)
		}
		lastLayer := len(sdf.Stages[stageIndex].Layers) - 1
		return fmt.Sprintf("stage_%d_layer_%d", stageIndex, lastLayer),
			map[string]string{"ltail": fmt.Sprintf("cluster_stage_%d", stageIndex)},
			nil
	}
	return fmt.Sprintf("stage_%d", stageIndex), map[string]string{}, nil
}
