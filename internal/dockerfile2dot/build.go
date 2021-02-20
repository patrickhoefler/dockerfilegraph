package dockerfile2dot

import (
	"github.com/awalterschulze/gographviz"
)

// BuildDotFile builds a GraphViz .dot file from a Google Cloud Build configuration
func BuildDotFile(simplifiedDockerfile SimplifiedDockerfile) string {
	graph := gographviz.NewEscape()
	graph.SetName("G")
	graph.SetDir(true)
	graph.AddAttr("G", "splines", "ortho")
	graph.AddAttr("G", "rankdir", "LR")
	graph.AddAttr("G", "nodesep", "1")

	for index, step := range simplifiedDockerfile.Stages {
		attrs := map[string]string{
			"label": "\"" + getStepLabel(step) + "\"",
			"shape": "Mrecord",
			"width": "2",
		}

		// Color the last stage, because it is the default build target
		if index == len(simplifiedDockerfile.Stages)-1 {
			attrs["style"] = "filled"
			attrs["fillcolor"] = "grey90"
		}

		graph.AddNode("G", step.ID, attrs)

		for _, waitForStepID := range step.WaitFor {
			if waitForStepID == "" {
				continue
			}

			graph.AddEdge(
				getRealStepID(simplifiedDockerfile, waitForStepID),
				step.ID,
				true,
				nil,
			)
		}
	}

	return graph.String()
}

func getStepLabel(stage Stage) string {
	if stage.Name != "" {
		return stage.Name
	}

	return stage.ID
}

func getRealStepID(simplifiedDockerfile SimplifiedDockerfile, stepID string) string {
	for _, step := range simplifiedDockerfile.Stages {
		if stepID == step.ID || stepID == step.Name {
			return step.ID
		}
	}
	// This should never happen
	return ""
}
