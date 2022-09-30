// Package cmd contains the Cobra CLI.
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/patrickhoefler/dockerfilegraph/internal/dockerfile2dot"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	dpiFlag     int
	legendFlag  bool
	layersFlag  bool
	outputFlag  enum
	versionFlag bool
	fileFlag    string
)

// dfgWriter is a writer that prints to stdout. When testing, we
// replace this with a writer that prints to a buffer.
type dfgWriter struct{}

func (d dfgWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

// NewRootCmd creates a new root command.
func NewRootCmd(dfgWriter io.Writer, inputFS afero.Fs) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dockerfilegraph",
		Short: "Visualize your multi-stage Dockerfile",
		Long: `dockerfilegraph visualizes your multi-stage Dockerfile.
It outputs a graph representation of the build process.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if versionFlag {
				return printVersion(dfgWriter)
			}

			dockerfile, err := dockerfile2dot.LoadAndParseDockerfile(inputFS, fileFlag)
			if err != nil {
				return
			}

			dotFile, err := os.CreateTemp("", "dockerfile.*.dot")
			if err != nil {
				return
			}
			defer os.Remove(dotFile.Name())

			dotFileContent := dockerfile2dot.BuildDotFile(dockerfile, legendFlag, layersFlag)

			_, err = dotFile.Write([]byte(dotFileContent))
			if err != nil {
				return
			}

			err = dotFile.Close()
			if err != nil {
				return
			}

			filename := "Dockerfile." + outputFlag.String()

			dotArgs := []string{
				"-T" + outputFlag.String(),
				"-o" + filename,
			}
			if outputFlag.String() == "png" {
				dotArgs = append(dotArgs, "-Gdpi="+fmt.Sprint(dpiFlag))
			}
			dotArgs = append(dotArgs, dotFile.Name())

			out, err := exec.Command("dot", dotArgs...).CombinedOutput()
			if err != nil {
				fmt.Fprintf(dfgWriter,
					`Oh no, something went wrong while generating the graph!

					This is the Graphviz file that was generated:

					%s
					The following error was reported by Graphviz:

					%s`,
					dotFileContent, string(out),
				)
				os.Exit(1)
			}

			fmt.Fprintf(dfgWriter, "Successfully created %s\n", filename)

			return
		},
	}

	// Flags
	rootCmd.Flags().IntVarP(
		&dpiFlag,
		"dpi",
		"d",
		96, // the default dpi setting of Graphviz
		"dots per inch of the PNG export",
	)

	rootCmd.Flags().BoolVar(
		&layersFlag,
		"layers",
		false,
		"display all layers (default false)",
	)

	rootCmd.Flags().BoolVar(
		&legendFlag,
		"legend",
		false,
		"add a legend (default false)",
	)

	outputFlag = newEnum("pdf", "canon", "dot", "png")
	rootCmd.Flags().VarP(
		&outputFlag,
		"output",
		"o",
		"output file format, one of: "+strings.Join(outputFlag.AllowedValues(), ", "),
	)

	rootCmd.Flags().BoolVar(
		&versionFlag,
		"version",
		false,
		"display the version of dockerfilegraph",
	)

	rootCmd.Flags().StringVarP(
		&fileFlag,
		"file",
		"f",
		"Dockerfile",
		"name of the Dockerfile",
	)

	return rootCmd
}

// Execute executes the root command.
func Execute() {
	err := NewRootCmd(dfgWriter{}, afero.NewOsFs()).Execute()
	if err != nil {
		// Cobra prints the error message
		os.Exit(1)
	}
}
