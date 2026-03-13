package dockerfile2dot

import (
	"fmt"
	"strconv"
	"strings"
)

// filterToTargets returns a new SimplifiedDockerfile containing only the
// stages that are transitively needed to build any of the named targets.
// External images referenced by the retained stages are also retained.
// Returns an error if any target name does not correspond to a stage.
func filterToTargets(sdf SimplifiedDockerfile, targets []string) (SimplifiedDockerfile, error) {
	targetIndices, err := resolveTargetIndices(sdf.Stages, targets)
	if err != nil {
		return SimplifiedDockerfile{}, err
	}

	reachable := collectReachableIndices(sdf.Stages, targetIndices)

	// Build sorted list of retained indices (preserving original order).
	retained := make([]int, 0, len(reachable))
	for i := range sdf.Stages {
		if _, ok := reachable[i]; ok {
			retained = append(retained, i)
		}
	}

	// Build old-index → new-index mapping for remapping numeric WaitFor IDs.
	oldToNew := make(map[int]int, len(retained))
	for newIdx, oldIdx := range retained {
		oldToNew[oldIdx] = newIdx
	}

	filteredStages := buildFilteredStages(sdf.Stages, retained, oldToNew)
	filteredExternal := filterExternalImages(sdf.ExternalImages, filteredStages)

	return SimplifiedDockerfile{
		BeforeFirstStage: sdf.BeforeFirstStage,
		Stages:           filteredStages,
		ExternalImages:   filteredExternal,
	}, nil
}

// resolveTargetIndices validates target names and returns their stage indices.
// Whitespace is trimmed from each target; empty entries are skipped.
func resolveTargetIndices(stages []Stage, targets []string) ([]int, error) {
	indices := make([]int, 0, len(targets))
	for _, target := range targets {
		trimmed := strings.TrimSpace(target)
		if trimmed == "" {
			continue
		}
		idx, found := findStageIndex(stages, trimmed)
		if !found {
			return nil, fmt.Errorf("target %q not found in Dockerfile", trimmed)
		}
		indices = append(indices, idx)
	}
	if len(indices) == 0 {
		return nil, fmt.Errorf("no valid targets specified")
	}
	return indices, nil
}

// collectReachableIndices performs a BFS from target indices, following WaitFor
// dependencies backwards through the stage graph.
func collectReachableIndices(stages []Stage, targetIndices []int) map[int]struct{} {
	reachable := make(map[int]struct{})
	queue := make([]int, len(targetIndices))
	copy(queue, targetIndices)

	for len(queue) > 0 {
		idx := queue[0]
		queue = queue[1:]

		if _, seen := reachable[idx]; seen {
			continue
		}
		reachable[idx] = struct{}{}

		for _, layer := range stages[idx].Layers {
			for _, waitFor := range layer.WaitFors {
				if depIdx, found := findStageIndex(stages, waitFor.ID); found {
					queue = append(queue, depIdx)
				}
			}
		}
	}
	return reachable
}

// buildFilteredStages copies the retained stages, remapping any numeric WaitFor IDs.
func buildFilteredStages(stages []Stage, retained []int, oldToNew map[int]int) []Stage {
	filtered := make([]Stage, 0, len(retained))
	for _, oldIdx := range retained {
		filtered = append(filtered, remapStage(stages[oldIdx], oldToNew))
	}
	return filtered
}

// remapStage returns a copy of stage with numeric WaitFor IDs updated to new indices.
func remapStage(stage Stage, oldToNew map[int]int) Stage {
	newLayers := make([]Layer, len(stage.Layers))
	for li, layer := range stage.Layers {
		newLayers[li] = remapLayer(layer, oldToNew)
	}
	return Stage{Name: stage.Name, Layers: newLayers}
}

// remapLayer returns a copy of layer with numeric WaitFor IDs updated to new indices.
func remapLayer(layer Layer, oldToNew map[int]int) Layer {
	newWaitFors := make([]WaitFor, len(layer.WaitFors))
	for wi, wf := range layer.WaitFors {
		if origIdx, err := strconv.Atoi(wf.ID); err == nil {
			if newIdx, ok := oldToNew[origIdx]; ok {
				wf.ID = strconv.Itoa(newIdx)
			}
		}
		newWaitFors[wi] = wf
	}
	return Layer{Label: layer.Label, WaitFors: newWaitFors}
}

// filterExternalImages retains only external images referenced by the filtered stages.
// WaitFor IDs that resolve to internal stages are excluded.
func filterExternalImages(allImages []ExternalImage, filteredStages []Stage) []ExternalImage {
	referencedIDs := make(map[string]struct{})
	for _, stage := range filteredStages {
		for _, layer := range stage.Layers {
			for _, wf := range layer.WaitFors {
				// Only count as an external image reference if it doesn't resolve to an internal stage.
				if _, found := findStageIndex(filteredStages, wf.ID); !found {
					referencedIDs[wf.ID] = struct{}{}
				}
			}
		}
	}
	filtered := make([]ExternalImage, 0, len(allImages))
	for _, img := range allImages {
		if _, ok := referencedIDs[img.ID]; ok {
			filtered = append(filtered, img)
		}
	}
	return filtered
}
