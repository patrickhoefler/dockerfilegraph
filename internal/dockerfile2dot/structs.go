package dockerfile2dot

// SimplifiedDockerfile contains the parts of the Dockerfile
// that are relevant for generating the multi-stage build graph
type SimplifiedDockerfile struct {
	BaseImages []BaseImage
	Stages     []Stage
}

// Stage contains the parts of a single build stage within a multi-stage Dockerfile
// that are relevant for generating the multi-stage build graph
type Stage struct {
	ID      string   // numeric index based on the order of appearance in the Dockerfile
	Name    string   // alias (the part after the AS) for internal stages
	WaitFor []string `yaml:"waitFor"` // IDs of stages or external images that the stage depends on
}

// BaseImage contains the ID of an external base images that a build stage depends on
type BaseImage struct {
	ID string // full repo:tag@sha
}
