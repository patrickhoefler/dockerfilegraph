package dockerfile2dot

// SimplifiedDockerfile contains the parts of the Dockerfile
// that are relevant for generating the multi-stage build graph.
type SimplifiedDockerfile struct {
	// External base images
	BaseImages []BaseImage
	// Args set before the first stage, see
	// https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact
	BeforeFirstStage []Layer
	// Build stages
	Stages []Stage
}

// Stage contains the parts of a single build stage within a multi-stage Dockerfile
// that are relevant for generating the multi-stage build graph.
type Stage struct {
	ID     string  // numeric index based on the order of appearance in the Dockerfile
	Name   string  // the part after the AS in the FROM line
	Layers []Layer // layers per stage
}

// Layer stores the changes compared to the image it’s based on within a
// multi-stage Dockerfile.
type Layer struct {
	ID      string  // numeric index based on the order of appearance in the stage
	Name    string  // the command and truncated args
	WaitFor WaitFor // stage or external base image for which this layer needs to wait
}

// BaseImage contains the ID of an external base image used for dependencies.
type BaseImage struct {
	ID string // full repo:tag@sha
}

type waitForType int

const (
	copy waitForType = iota
	from
	runMountTypeCache
)

// WaitFor provides the ID of a stage or base image that the builder
// has to wait for and the type, i.e. the reason why it has to wait for it
type WaitFor struct {
	ID   string      // the ID of the base image or stage
	Type waitForType // one of "from", "copy" or "runMountTypeCache"
}
