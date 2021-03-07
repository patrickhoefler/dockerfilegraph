package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/patrickhoefler/dockerfilegraph/internal/dockerfile2dot"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	output *enum

	rootCmd = &cobra.Command{
		Use:   "dockerfilegraph",
		Short: "Visualize your multi-stage Dockerfile",
		Long: `dockerfilegraph visualizes your multi-stage Dockerfile.
It outputs a graph representation of the build process.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.SetFlags(0)

			dockerfile, err := dockerfile2dot.LoadAndParseDockerfile()
			if err != nil {
				log.Fatal(err)
			}

			dotFile, err := ioutil.TempFile("", "dockerfile.*.dot")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(dotFile.Name())

			dotFileContent := dockerfile2dot.BuildDotFile(dockerfile)

			_, err = dotFile.Write([]byte(dotFileContent))
			if err != nil {
				log.Fatal(err)
			}

			err = dotFile.Close()
			if err != nil {
				log.Fatal(err)
			}

			filename := "Dockerfile." + output.String()

			resolutionFlag := ""
			if output.String() == "png" {
				resolutionFlag = "-Gdpi=300"
			}

			out, err := exec.Command(
				"dot",
				"-T"+output.String(),
				resolutionFlag,
				"-o"+filename,
				dotFile.Name(),
			).CombinedOutput()
			if err != nil {
				log.Println("Oh no, something went wrong!")
				log.Println()
				log.Println("This is the Graphviz file that was generated:")
				log.Println()
				log.Println(dotFileContent)
				log.Println("The following error was reported by Graphviz:")
				log.Println()
				log.Fatal(string(out))
			}

			fmt.Println("Successfully created " + filename)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	output = newEnum("pdf", "png")
	rootCmd.Flags().VarP(
		output,
		"output",
		"o",
		"Output file format. One of: "+strings.Join(output.AllowedValues(), ", "),
	)
}
