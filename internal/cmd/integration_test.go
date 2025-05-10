package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/patrickhoefler/dockerfilegraph/internal/cmd"
	"github.com/spf13/afero"
)

func TestIntegrationCLIGeneratesOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	dockerfilePath := copyExampleDockerfile(t, tempDir)

	originalWorkingDirectory, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWorkingDirectory); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	inputFS := afero.NewOsFs()
	outputFile := filepath.Join(tempDir, "Dockerfile.pdf")
	buf := runCLI(t, inputFS, dockerfilePath)

	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("expected output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("output file is empty: %s", outputFile)
	}

	if !bytes.Contains(buf.Bytes(), []byte("Successfully created Dockerfile.pdf")) {
		t.Errorf("CLI output does not contain success message. Output: %s", buf.String())
	}

	checkGoldenFile(t, outputFile)
}

func copyExampleDockerfile(t *testing.T, tempDir string) string {
	dockerfileSrc := filepath.Join("..", "..", "examples", "dockerfiles", "Dockerfile")
	content, err := os.ReadFile(dockerfileSrc)
	if err != nil {
		t.Fatalf("failed to read example Dockerfile: %v", err)
	}
	dockerfileDst := filepath.Join(tempDir, "Dockerfile")
	if err := os.WriteFile(dockerfileDst, content, 0644); err != nil {
		t.Fatalf("failed to write Dockerfile to temp dir: %v", err)
	}
	return dockerfileDst
}

func runCLI(t *testing.T, inputFS afero.Fs, dockerfilePath string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	command := cmd.NewRootCmd(buf, inputFS, "dot")
	command.SetArgs([]string{"--filename", filepath.Base(dockerfilePath), "--output", "pdf"})
	command.SetOut(buf)
	command.SetErr(buf)

	// Set SOURCE_DATE_EPOCH=0 for deterministic PDF output
	oldEnv := os.Getenv("SOURCE_DATE_EPOCH")
	if err := os.Setenv("SOURCE_DATE_EPOCH", "0"); err != nil {
		t.Fatalf("failed to set SOURCE_DATE_EPOCH: %v", err)
	}
	defer func() {
		_ = os.Setenv("SOURCE_DATE_EPOCH", oldEnv)
	}()

	if err := command.Execute(); err != nil {
		t.Fatalf("CLI execution failed: %v\nOutput: %s", err, buf.String())
	}
	return buf
}

func checkGoldenFile(t *testing.T, outputFile string) {
	_, thisFile, _, _ := runtime.Caller(0)
	goldenDir := filepath.Join(filepath.Dir(thisFile), "testdata")
	goldenFile := filepath.Join(goldenDir, "Dockerfile.golden.pdf")
	outputBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read generated output file: %v", err)
	}
	if _, err := os.Stat(goldenFile); os.IsNotExist(err) {
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenFile, outputBytes, 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("golden file did not exist, created: %s", goldenFile)
	} else {
		goldenBytes, err := os.ReadFile(goldenFile)
		if err != nil {
			t.Fatalf("failed to read golden file: %v", err)
		}
		if !bytes.Equal(outputBytes, goldenBytes) {
			t.Errorf("output PDF does not match golden file. To update, delete %s and re-run the test.", goldenFile)
		}
	}
}
