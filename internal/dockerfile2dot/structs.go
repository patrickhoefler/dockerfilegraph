package dockerfile2dot

import "strconv"

// SimplifiedDockerfile contains the parts of the Dockerfile
// that are relevant for generating the multi-stage build graph.
type SimplifiedDockerfile struct {
	// Args set before the first stage, see
	// https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact
	BeforeFirstStage []Layer
	// Build stages
	Stages []Stage
	// External images
	ExternalImages []ExternalImage
}

// Stage represents a single build stage within the multi-stage Dockerfile or
// an external image.
type Stage struct {
	Name   string  // The part after the AS in the FROM line
	Layers []Layer // The layers of the stage
}

// Layer stores the changes compared to the image it's based on within a
// multi-stage Dockerfile.
type Layer struct {
	Label    string    // The command and truncated args
	WaitFors []WaitFor // Stages or external images for which this layer needs to wait
}

// ExternalImage holds the name of an external image.
type ExternalImage struct {
	ID   string // Unique identifier for this external image instance
	Name string // The original name of the external image
}

// ScratchMode controls how scratch base images are rendered in the graph.
type ScratchMode int

// ScratchMode values control how scratch base images appear in the graph.
const (
	ScratchCollapsed ScratchMode = iota // All scratch references share a single node
	ScratchSeparated                    // Each scratch reference gets its own node
	ScratchHidden                       // Scratch references are omitted from the graph
)

// waitForType represents the type of dependency between stages or images.
type waitForType int

const (
	waitForCopy  waitForType = iota // COPY dependency from another stage
	waitForFrom                     // FROM dependency on another stage or image
	waitForMount                    // MOUNT dependency for build cache
)

// WaitFor holds the name of the stage or external image for which the builder
// has to wait, and the type, i.e. the reason why it has to wait for it.
type WaitFor struct {
	ID   string      // The unique identifier of the stage or external image for which the builder has to wait
	Type waitForType // The reason why it has to wait
}

// findStageIndex returns the index of the stage identified by nameOrID (a stage
// name or a decimal numeric index string) and true if found. For a numeric
// index that parses but is out of range, it returns that index and false. For
// a non-numeric name that is not found, it returns -1 and false.
func findStageIndex(stages []Stage, nameOrID string) (int, bool) {
	if idx, err := strconv.Atoi(nameOrID); err == nil {
		if idx >= 0 && idx < len(stages) {
			return idx, true
		}
		return idx, false
	}
	for i, s := range stages {
		if s.Name == nameOrID {
			return i, true
		}
	}
	return -1, false
}

// ScratchModeFromString converts a validated string to a ScratchMode constant.
// The empty string and any unrecognized value returns ScratchCollapsed.
func ScratchModeFromString(s string) ScratchMode {
	switch s {
	case "separated":
		return ScratchSeparated
	case "hidden":
		return ScratchHidden
	default:
		return ScratchCollapsed
	}
}

// ParseOptions controls how a Dockerfile is parsed into a SimplifiedDockerfile.
type ParseOptions struct {
	MaxLabelLength int
	ScratchMode    ScratchMode
	SeparateImages []string
	Targets        []string
}

// BuildOptions controls how a SimplifiedDockerfile is rendered into a DOT file.
type BuildOptions struct {
	Concentrate    bool
	EdgeStyle      string
	Layers         bool
	Legend         bool
	MaxLabelLength int
	NodeSep        float64
	RankSep        float64
}
