package dockerfile2dot

// SimplifiedDockerfile contains the parts of the Dockerfile
// that are relevant for generating the multi-stage build graph
type SimplifiedDockerfile struct {
	Stages []Stage
}

// Stage contains the parts of a single build stage within a multi-stage Dockerfile
// that are relevant for generating the multi-stage build graph
type Stage struct {
	ID       string   // numeric identifier for internal stages, full repo:tag@sha for external base images
	Name     string   // alias (the part after the AS) for internal stages
	Args     []string // not currently used, might be removed
	WaitFor  []string `yaml:"waitFor"` // IDs of stages or external images that the stage depends on
	External bool     // true if the stage is actually an external base image
}
