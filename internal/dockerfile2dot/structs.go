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
