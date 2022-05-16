package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/patrickhoefler/dockerfilegraph/internal/dockerfile2dot"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	dpi    int
	legend bool
	layers bool
	output enum
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
			dockerfile, err := dockerfile2dot.LoadAndParseDockerfile(inputFS)
			if err != nil {
				return
			}

			dotFile, err := ioutil.TempFile("", "dockerfile.*.dot")
			if err != nil {
				return
			}
			defer os.Remove(dotFile.Name())

			dotFileContent := dockerfile2dot.BuildDotFile(dockerfile, legend, layers)

			_, err = dotFile.Write([]byte(dotFileContent))
			if err != nil {
				return
			}

			err = dotFile.Close()
			if err != nil {
				return
			}

			filename := "Dockerfile." + output.String()

			dotArgs := []string{
				"-T" + output.String(),
				"-o" + filename,
			}
			if output.String() == "png" {
				dotArgs = append(dotArgs, "-Gdpi="+fmt.Sprint(dpi))
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

			fmt.Fprintf(dfgWriter, "Successfully created %s", filename)

			return
		},
	}

	rootCmd.Flags().IntVarP(
		&dpi,
		"dpi",
		"d",
		96, // the default dpi setting of Graphviz
		"dots per inch of the PNG export",
	)

	rootCmd.Flags().BoolVar(
		&legend,
		"legend",
		false,
		"add a legend (default false)",
	)

	output = newEnum("pdf", "canon", "dot", "png")
	rootCmd.Flags().VarP(
		&output,
		"output",
		"o",
		"output file format, one of: "+strings.Join(output.AllowedValues(), ", "),
	)

	rootCmd.Flags().BoolVar(
		&layers,
		"layers",
		false,
		"display all layers (default false)",
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
