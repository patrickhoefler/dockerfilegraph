package dockerfile2dot

import (
	"github.com/awalterschulze/gographviz"
)

// BuildDotFile builds a GraphViz .dot file from a Google Cloud Build configuration
func BuildDotFile(simplifiedDockerfile SimplifiedDockerfile, legend bool, layers bool) string {
	graph := gographviz.NewEscape()
	_ = graph.SetName("G")
	_ = graph.SetDir(true)
	_ = graph.AddAttr("G", "rankdir", "LR")
	_ = graph.AddAttr("G", "nodesep", "1")

	if legend {
		_ = graph.AddSubGraph("G", "cluster_legend", nil)

		_ = graph.AddNode("cluster_legend", "key",
			map[string]string{
				"shape":    "plaintext",
				"fontname": "monospace",
				"fontsize": "10",
				"label": `<<table border="0" cellpadding="2" cellspacing="0" cellborder="0">
	<tr><td align="right" port="i0">FROM&nbsp;...&nbsp;</td></tr>
	<tr><td align="right" port="i1">COPY --from=...&nbsp;</td></tr>
	<tr><td align="right" port="i2">RUN --mount=type=cache,from=...&nbsp;</td></tr>
</table>>`,
			},
		)
		_ = graph.AddNode("cluster_legend", "key2",
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
		)

		_ = graph.AddPortEdge("key", "i0:e", "key2", "i0:w", true, nil)
		_ = graph.AddPortEdge("key", "i1:e", "key2", "i1:w", true, map[string]string{"arrowhead": "empty"})
		_ = graph.AddPortEdge("key", "i2:e", "key2", "i2:w", true, map[string]string{"arrowhead": "ediamond"})
	}

	for _, baseImage := range simplifiedDockerfile.BaseImages {
		_ = graph.AddNode("G", "\""+baseImage.ID+"\"", map[string]string{
			"shape":     "Mrecord",
			"width":     "2",
			"style":     "dashed",
			"color":     "grey20",
			"fontcolor": "grey20",
		})
	}

	for index, stage := range simplifiedDockerfile.Stages {
		attrs := map[string]string{
			"label": "\"" + getStageLabel(stage) + "\"",
			"shape": "Mrecord",
			"width": "2",
		}

		// Color the last stage, because it is the default build target
		if index == len(simplifiedDockerfile.Stages)-1 {
			attrs["style"] = "filled"
			attrs["fillcolor"] = "grey90"
		}

		_ = graph.AddNode("G", "\""+stage.ID+"\"", attrs)

		// draw layers per stage
		if layers {
			_ = graph.AddSubGraph("G", "cluster_"+stage.ID, map[string]string{
				"label": getStageLabel(stage),
				"style": "rounded",
			})
			attrs["style"] = "invis"
			_ = graph.AddNode("cluster_"+stage.ID, "\""+stage.ID+"\"", attrs)
			_ = graph.AddAttr("G", "nodesep", "0.03")

			for _, layer := range stage.Layers {
				attrs["label"] = "\"" + layer.Name + "\""
				attrs["style"] = "dashed"
				_ = graph.AddNode("cluster_"+stage.ID, "stage_"+stage.ID+"_layer_"+layer.ID, attrs)
			}
		}

		for _, waitFor := range stage.WaitFor {
			if waitFor.ID == "" {
				continue
			}

			edgeAttrs := map[string]string{}
			if waitFor.Type == waitForType(copy) {
				edgeAttrs["arrowhead"] = "empty"
			} else if waitFor.Type == waitForType(runMountTypeCache) {
				edgeAttrs["arrowhead"] = "ediamond"
			}

			_ = graph.AddEdge(
				"\""+getRealStageID(simplifiedDockerfile, waitFor.ID)+"\"",
				"\""+stage.ID+"\"",
				true,
				edgeAttrs,
			)
		}
	}
	if layers {
		if len(simplifiedDockerfile.LayersNotStage) > 0 {
			_ = graph.AddSubGraph("G", "cluster_layers_not_stage", map[string]string{"label": "Before First Stage"})
			for _, layerNotStage := range simplifiedDockerfile.LayersNotStage {
				_ = graph.AddNode("cluster_layers_not_stage", "layer_not_stage_"+layerNotStage.ID, map[string]string{
					"label": layerNotStage.Name,
					"shape": "Mrecord",
					"width": "2",
				})
			}
		}
	}

	return graph.String()
}

func getStageLabel(stage Stage) string {
	if stage.Name != "" {
		return stage.Name
	}
	return stage.ID
}

func getRealStageID(simplifiedDockerfile SimplifiedDockerfile, stageID string) string {
	// Look up the real stage id, could be either numeric or the "AS" alias
	for _, stage := range simplifiedDockerfile.Stages {
		if stageID == stage.ID || stageID == stage.Name {
			return stage.ID
		}
	}

	// It is actually an external base image, keep the ID as is
	return stageID
}
