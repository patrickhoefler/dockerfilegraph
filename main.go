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

	_, err = dotFile.Write([]byte(dockerfile2dot.BuildDotFile(dockerfile)))
	if err != nil {
		log.Fatal(err)
	}

	err = dotFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("dot", "-Tpdf", "-odockerfile.pdf", dotFile.Name())
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully created dockerfile.pdf")
}
