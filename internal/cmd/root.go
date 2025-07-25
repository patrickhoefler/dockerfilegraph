// Package cmd contains the Cobra CLI.
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/patrickhoefler/dockerfilegraph/internal/dockerfile2dot"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	concentrateFlag    bool
	dpiFlag            uint
	edgestyleFlag      enum
	filenameFlag       string
	layersFlag         bool
	legendFlag         bool
	maxLabelLengthFlag uint
	nodesepFlag        float64
	outputFlag         enum
	ranksepFlag        float64
	scratchFlag        enum
	unflattenFlag      uint
	versionFlag        bool
)

// dfgWriter is a writer that prints to stdout. When testing, we
// replace this with a writer that prints to a buffer.
type dfgWriter struct{}

func (d dfgWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

// NewRootCmd creates a new root command.
func NewRootCmd(
	dfgWriter io.Writer, inputFS afero.Fs, dotCmd string,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dockerfilegraph",
		Short: "Visualize your multi-stage Dockerfile",
		Long: `dockerfilegraph visualizes your multi-stage Dockerfile.
It creates a visual graph representation of the build process.`,
		Args: cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) (err error) {
			return checkFlags()
		},
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			if versionFlag {
				return printVersion(dfgWriter)
			}

			// Make sure that graphviz is installed.
			_, err = exec.LookPath(dotCmd)
			if err != nil {
				return
			}

			// Determine scratch mode from flag
			scratchMode := scratchFlag.String()

			// Load and parse the Dockerfile.
			dockerfile, err := dockerfile2dot.LoadAndParseDockerfile(
				inputFS,
				filenameFlag,
				int(maxLabelLengthFlag),
				scratchMode,
			)
			if err != nil {
				return
			}

			dotFile, err := os.CreateTemp("", "dockerfile.*.dot")
			if err != nil {
				return
			}
			defer os.Remove(dotFile.Name())

			dotFileContent := dockerfile2dot.BuildDotFile(
				dockerfile,
				concentrateFlag,
				edgestyleFlag.String(),
				layersFlag,
				legendFlag,
				int(maxLabelLengthFlag),
				fmt.Sprintf("%.2f", nodesepFlag),
				fmt.Sprintf("%.2f", ranksepFlag),
			)

			_, err = dotFile.Write([]byte(dotFileContent))
			if err != nil {
				return
			}

			err = dotFile.Close()
			if err != nil {
				return
			}

			if unflattenFlag > 0 {
				err = unflatten(dotFile, dfgWriter)
				if err != nil {
					return
				}
			}

			filename := "Dockerfile." + outputFlag.String()

			if outputFlag.String() == "raw" {
				err = os.Rename(dotFile.Name(), filename)
				if err != nil {
					return
				}
				fmt.Fprintf(dfgWriter, "Successfully created %s\n", filename)
				return
			}

			dotArgs := []string{
				"-T" + outputFlag.String(),
				"-o" + filename,
			}
			if outputFlag.String() == "png" {
				dotArgs = append(dotArgs, "-Gdpi="+fmt.Sprint(dpiFlag))
			}
			dotArgs = append(dotArgs, dotFile.Name())

			out, err := exec.Command(dotCmd, dotArgs...).CombinedOutput()
			if err != nil {
				fmt.Fprintf(dfgWriter,
					`Oh no, something went wrong while generating the graph!

					This is the Graphviz file that was generated:

					%s
					The following error was reported by Graphviz:

					%s`,
					dotFileContent, string(out),
				)
				return
			}

			fmt.Fprintf(dfgWriter, "Successfully created %s\n", filename)

			return
		},
	}

	// Flags
	rootCmd.Flags().BoolVarP(
		&concentrateFlag,
		"concentrate",
		"c",
		false,
		"concentrate the edges (default false)",
	)

	rootCmd.Flags().UintVarP(
		&dpiFlag,
		"dpi",
		"d",
		96, // the default dpi setting of Graphviz
		"dots per inch of the PNG export",
	)

	edgestyleFlag = newEnum("default", "solid")
	rootCmd.Flags().VarP(
		&edgestyleFlag,
		"edgestyle",
		"e",
		"style of the graph edges, one of: "+strings.Join(edgestyleFlag.AllowedValues(), ", "),
	)

	rootCmd.Flags().StringVarP(
		&filenameFlag,
		"filename",
		"f",
		"Dockerfile",
		"name of the Dockerfile",
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

	rootCmd.Flags().UintVarP(
		&maxLabelLengthFlag,
		"max-label-length",
		"m",
		20,
		"maximum length of the node labels, must be at least 4",
	)

	rootCmd.Flags().Float64VarP(
		&nodesepFlag,
		"nodesep",
		"n",
		1,
		"minimum space between two adjacent nodes in the same rank",
	)

	outputFlag = newEnum("pdf", "canon", "dot", "png", "raw", "svg")
	rootCmd.Flags().VarP(
		&outputFlag,
		"output",
		"o",
		"output file format, one of: "+strings.Join(outputFlag.AllowedValues(), ", "),
	)

	rootCmd.Flags().Float64VarP(
		&ranksepFlag,
		"ranksep",
		"r",
		0.5,
		"minimum separation between ranks",
	)

	scratchFlag = newEnum("collapsed", "separated", "hidden")
	rootCmd.Flags().Var(
		&scratchFlag,
		"scratch",
		"how to handle scratch images, one of: "+strings.Join(scratchFlag.AllowedValues(), ", "),
	)

	rootCmd.Flags().UintVarP(
		&unflattenFlag,
		"unflatten",
		"u",
		0, // turned off
		"stagger length of leaf edges between [1,u] (default 0)",
	)

	rootCmd.Flags().BoolVar(
		&versionFlag,
		"version",
		false,
		"display the version of dockerfilegraph",
	)

	return rootCmd
}

func unflatten(dotFile *os.File, dfgWriter io.Writer) (err error) {
	var unflattenFile *os.File
	unflattenFile, err = os.CreateTemp("", "dockerfile.*.dot")
	if err != nil {
		return
	}
	defer os.Remove(unflattenFile.Name())

	unflattenCmd := exec.Command(
		"unflatten",
		"-l", strconv.FormatUint(uint64(unflattenFlag), 10),
		"-o", unflattenFile.Name(), dotFile.Name(),
	)
	unflattenCmd.Stdout = dfgWriter
	unflattenCmd.Stderr = dfgWriter
	err = unflattenCmd.Run()
	if err != nil {
		return
	}

	err = unflattenFile.Close()
	if err != nil {
		return
	}

	err = os.Rename(unflattenFile.Name(), dotFile.Name())
	if err != nil {
		return
	}

	return
}

// Execute executes the root command.
func Execute() {
	err := NewRootCmd(
		dfgWriter{}, afero.NewOsFs(), "dot",
	).Execute()
	if err != nil {
		// Cobra prints the error message
		os.Exit(1)
	}
}

func checkFlags() (err error) {
	if maxLabelLengthFlag < 4 {
		err = fmt.Errorf("--max-label-length must be at least 4")
		return
	}
	return
}
