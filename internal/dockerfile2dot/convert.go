package dockerfile2dot

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func dockerfileToSimplifiedDockerfile(content []byte) SimplifiedDockerfile {
	result, err := parser.Parse(bytes.NewReader(content))
	if err != nil {
		log.Fatal(err)
	}

	simplifiedDockerfile := SimplifiedDockerfile{}
	stages := make(map[string]struct{})

	// Add all stages
	stageIndex := -1
	for _, child := range result.AST.Children {
		switch child.Value {
		case "from":
			stageIndex++

			stage := Stage{}
			stage.ID = fmt.Sprint(stageIndex)
			stage.WaitFor = []string{child.Next.Value}

			if child.Next.Next != nil {
				stages[child.Next.Next.Next.Value] = struct{}{}
				stage.Name = child.Next.Next.Next.Value
			} else {
				stage.Name = child.Next.Value
			}

			simplifiedDockerfile.Stages = append(simplifiedDockerfile.Stages, stage)

		case "copy":
			for _, flag := range child.Flags {
				regex := regexp.MustCompile("--from=(.+)")
				result := regex.FindSubmatch([]byte(flag))
				if len(result) > 1 {
					simplifiedDockerfile.Stages[stageIndex].WaitFor = append(
						simplifiedDockerfile.Stages[stageIndex].WaitFor,
						string(result[1]),
					)
				}
			}
		}
	}

	fmt.Println(simplifiedDockerfile)

	// Remove WaitFors that are actually external images
	for index, stage := range simplifiedDockerfile.Stages {
		for waitForIndex, waitFor := range stage.WaitFor {
			if _, ok := stages[waitFor]; !ok {
				simplifiedDockerfile.Stages[index].WaitFor[waitForIndex] = ""
			}
		}
	}

	fmt.Println(simplifiedDockerfile.Stages)
	return simplifiedDockerfile
}
