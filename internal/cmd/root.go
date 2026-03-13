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

// cliFlags holds all flag values for a single command invocation.
type cliFlags struct {
	concentrate    bool
	dpi            uint
	edgestyle      enum
	filename       string
	layers         bool
	legend         bool
	maxLabelLength uint
	nodesep        float64
	output         enum
	ranksep        float64
	scratch        enum
	separate       []string
	target         []string
	unflatten      uint
	version        bool
}

// dfgWriter is a writer that prints to stdout. When testing, we
// replace this with a writer that prints to a buffer.
type dfgWriter struct{}

func (d dfgWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

// NewRootCmd creates a new root command.
func NewRootCmd(
	w io.Writer, inputFS afero.Fs, dotCmd string,
) *cobra.Command {
	f := cliFlags{}

	rootCmd := &cobra.Command{
		Use:   "dockerfilegraph",
		Short: "Visualize your multi-stage Dockerfile",
		Long: `dockerfilegraph visualizes your multi-stage Dockerfile.
It creates a visual graph representation of the build process.`,
		Args: cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return checkFlags(f.maxLabelLength)
		},
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			if f.version {
				return printVersion(w)
			}

			// Make sure that graphviz is installed.
			_, err = exec.LookPath(dotCmd)
			if err != nil {
				return
			}

			// Load and parse the Dockerfile.
			dockerfile, err := dockerfile2dot.LoadAndParseDockerfile(
				inputFS,
				f.filename,
				dockerfile2dot.ParseOptions{
					MaxLabelLength: int(f.maxLabelLength),
					ScratchMode:    dockerfile2dot.ScratchModeFromString(f.scratch.String()),
					SeparateImages: f.separate,
					Targets:        f.target,
				},
			)
			if err != nil {
				return
			}

			dotFile, err := os.CreateTemp("", "dockerfile.*.dot")
			if err != nil {
				return
			}
			defer os.Remove(dotFile.Name())

			dotFileContent, err := dockerfile2dot.BuildDotFile(
				dockerfile,
				dockerfile2dot.BuildOptions{
					Concentrate:    f.concentrate,
					EdgeStyle:      f.edgestyle.String(),
					Layers:         f.layers,
					Legend:         f.legend,
					MaxLabelLength: int(f.maxLabelLength),
					NodeSep:        f.nodesep,
					RankSep:        f.ranksep,
				},
			)
			if err != nil {
				return
			}

			_, err = dotFile.Write([]byte(dotFileContent))
			if err != nil {
				return
			}

			err = dotFile.Close()
			if err != nil {
				return
			}

			if f.unflatten > 0 {
				err = runUnflatten(dotFile.Name(), w, f.unflatten)
				if err != nil {
					return
				}
				var b []byte
				b, err = os.ReadFile(dotFile.Name())
				if err != nil {
					return
				}
				dotFileContent = string(b)
			}

			filename := "Dockerfile." + f.output.String()

			if f.output.String() == "raw" {
				err = os.Rename(dotFile.Name(), filename)
				if err != nil {
					return
				}
				fmt.Fprintf(w, "Successfully created %s\n", filename)
				return
			}

			dotArgs := []string{
				"-T" + f.output.String(),
				"-o" + filename,
			}
			if f.output.String() == "png" {
				dotArgs = append(dotArgs, "-Gdpi="+fmt.Sprint(f.dpi))
			}
			dotArgs = append(dotArgs, dotFile.Name())

			out, err := exec.Command(dotCmd, dotArgs...).CombinedOutput()
			if err != nil {
				fmt.Fprintf(w,
					"Oh no, something went wrong while generating the graph!\n\n"+
						"This is the Graphviz file that was generated:\n\n"+
						"%s\n"+
						"The following error was reported by Graphviz:\n\n"+
						"%s",
					dotFileContent, string(out),
				)
				return
			}

			fmt.Fprintf(w, "Successfully created %s\n", filename)

			return
		},
	}

	// Flags
	rootCmd.Flags().BoolVarP(
		&f.concentrate,
		"concentrate",
		"c",
		false,
		"concentrate the edges (default false)",
	)

	rootCmd.Flags().UintVarP(
		&f.dpi,
		"dpi",
		"d",
		96, // the default dpi setting of Graphviz
		"dots per inch of the PNG export",
	)

	f.edgestyle = newEnum("default", "solid")
	rootCmd.Flags().VarP(
		&f.edgestyle,
		"edgestyle",
		"e",
		"style of the graph edges, one of: "+strings.Join(f.edgestyle.AllowedValues(), ", "),
	)

	rootCmd.Flags().StringVarP(
		&f.filename,
		"filename",
		"f",
		"Dockerfile",
		"name of the Dockerfile",
	)

	rootCmd.Flags().BoolVar(
		&f.layers,
		"layers",
		false,
		"display all layers (default false)",
	)

	rootCmd.Flags().BoolVar(
		&f.legend,
		"legend",
		false,
		"add a legend (default false)",
	)

	rootCmd.Flags().UintVarP(
		&f.maxLabelLength,
		"max-label-length",
		"m",
		20,
		"maximum length of the node labels, must be at least 4",
	)

	rootCmd.Flags().Float64VarP(
		&f.nodesep,
		"nodesep",
		"n",
		1,
		"minimum space between two adjacent nodes in the same rank",
	)

	f.output = newEnum("pdf", "canon", "dot", "png", "raw", "svg")
	rootCmd.Flags().VarP(
		&f.output,
		"output",
		"o",
		"output file format, one of: "+strings.Join(f.output.AllowedValues(), ", "),
	)

	rootCmd.Flags().Float64VarP(
		&f.ranksep,
		"ranksep",
		"r",
		0.5,
		"minimum separation between ranks",
	)

	f.scratch = newEnum("collapsed", "separated", "hidden")
	rootCmd.Flags().Var(
		&f.scratch,
		"scratch",
		"how to handle scratch images, one of: "+strings.Join(f.scratch.AllowedValues(), ", "),
	)

	rootCmd.Flags().StringSliceVar(
		&f.separate,
		"separate",
		nil,
		"external images to display as separate nodes per usage (e.g. --separate ubuntu,alpine)",
	)

	rootCmd.Flags().StringSliceVar(
		&f.target,
		"target",
		nil,
		"only show stages required to build the given target(s) (e.g. --target release,app)",
	)

	rootCmd.Flags().UintVarP(
		&f.unflatten,
		"unflatten",
		"u",
		0, // turned off
		"stagger length of leaf edges between [1,u] (default 0)",
	)

	rootCmd.Flags().BoolVar(
		&f.version,
		"version",
		false,
		"display the version of dockerfilegraph",
	)

	return rootCmd
}

func runUnflatten(dotPath string, w io.Writer, maxStagger uint) (err error) {
	unflattenFile, err := os.CreateTemp("", "dockerfile.*.dot")
	if err != nil {
		return
	}
	unflattenPath := unflattenFile.Name()
	err = unflattenFile.Close()
	if err != nil {
		os.Remove(unflattenPath)
		return
	}

	unflattenCmd := exec.Command(
		"unflatten",
		"-l", strconv.FormatUint(uint64(maxStagger), 10),
		"-o", unflattenPath, dotPath,
	)
	unflattenCmd.Stdout = w
	unflattenCmd.Stderr = w
	err = unflattenCmd.Run()
	if err != nil {
		os.Remove(unflattenPath)
		return
	}

	// Safely replace dotPath using a backup in case the final rename fails.
	// os.Rename cannot replace an existing file on Windows, so we move the
	// original aside first and only remove it after a successful swap.
	backupPath := dotPath + ".bak"
	os.Remove(backupPath) // best-effort removal of any stale backup
	err = os.Rename(dotPath, backupPath)
	if err != nil {
		os.Remove(unflattenPath)
		return
	}
	err = os.Rename(unflattenPath, dotPath)
	if err != nil {
		_ = os.Rename(backupPath, dotPath) // best-effort restore
		os.Remove(unflattenPath)
		return
	}
	os.Remove(backupPath)

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

func checkFlags(maxLabelLength uint) error {
	if maxLabelLength < 4 {
		return fmt.Errorf("--max-label-length must be at least 4")
	}
	return nil
}
