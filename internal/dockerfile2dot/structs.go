package dockerfile2dot

// SimplifiedDockerfile contains the parts of the Dockerfile
// that are relevant for generating the multi-stage build graph
type SimplifiedDockerfile struct {
	BaseImages     []BaseImage
	Stages         []Stage
	LayersNotStage []Layer
}

// Stage contains the parts of a single build stage within a multi-stage Dockerfile
// that are relevant for generating the multi-stage build graph
type Stage struct {
	ID      string    // numeric index based on the order of appearance in the Dockerfile
	Name    string    // the part after the AS in the FROM line
	WaitFor []WaitFor // dependencies of the stage
	Layers  []Layer   // layers per stage
}

// Layer stores the changes compared to the image it’s based on within a multi-stage Dockerfile
type Layer struct {
	ID   string // numeric index based on the order of appearance in the stage
	Name string // display the command store in the layer
}

// BaseImage contains the ID of an external base images that a build stage depends on
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
	Type waitForType // either "from" or "copy"
}
