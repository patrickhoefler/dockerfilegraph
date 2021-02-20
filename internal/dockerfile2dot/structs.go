package dockerfile2dot

// SimplifiedDockerfile contains the parts of the Dockerfile
// that are relevant for generating the multi-stage build graph
type SimplifiedDockerfile struct {
	Stages []Stage
}

// Stage contains the parts of a single build stage within a multi-stage Dockerfile
// that are relevant for generating the multi-stage build graph
type Stage struct {
	ID      string
	Name    string
	Args    []string
	WaitFor []string `yaml:"waitFor"`
}
