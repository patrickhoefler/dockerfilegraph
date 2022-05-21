package cmd

import (
	"encoding/json"
	"fmt"
	"io"
)

var (
	shortFlag  bool
	gitVersion = "v0.0.0-dev"
	gitCommit  = "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	buildDate  = "0000-00-00T00:00:00Z"
)

// VersionInfo holds the version information for a build of dockerfilegraph.
type VersionInfo struct {
	GitVersion string
	GitCommit  string
	BuildDate  string
}

func printVersion(dfgWriter io.Writer) (err error) {
	if shortFlag {
		fmt.Fprintf(dfgWriter, "%s\n", gitVersion)
	} else {
		var versionInfo []byte
		versionInfo, err = json.Marshal(VersionInfo{
			GitVersion: gitVersion,
			GitCommit:  gitCommit,
			BuildDate:  buildDate,
		})
		if err != nil {
			return
		}

		fmt.Fprintf(dfgWriter, "%s\n", string(versionInfo))
	}

	return
}
