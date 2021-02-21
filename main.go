package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/patrickhoefler/dockerfilegraph/internal/dockerfile2dot"
)

func main() {
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

	out, err := exec.Command("dot", "-Tpdf", "-oDockerfile.pdf", dotFile.Name()).CombinedOutput()
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

	fmt.Println("Successfully created Dockerfile.pdf")
}
