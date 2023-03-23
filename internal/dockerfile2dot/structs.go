package dockerfile2dot

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
	Name   string  // the part after the AS in the FROM line
	Layers []Layer // the layers of the stage
}

// Layer stores the changes compared to the image it’s based on within a
// multi-stage Dockerfile.
type Layer struct {
	Label   string  // the command and truncated args
	WaitFor WaitFor // stage or external image for which this layer needs to wait
}

// ExternalImage holds the name of an external image.
type ExternalImage struct {
	Name string
}

type waitForType int

const (
	waitForCopy waitForType = iota
	waitForFrom
	waitForCache
)

// WaitFor holds the name of the stage or external image for which the builder
// has to wait, and the type, i.e. the reason why it has to wait for it
type WaitFor struct {
	Name string      // the name of the stage or external image for which the builder has to wait
	Type waitForType // one of "from", "copy" or "runMountTypeCache"
}
