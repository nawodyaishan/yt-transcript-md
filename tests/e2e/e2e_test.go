package e2e

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	flag.Parse()

	exitCode := func() int {
		// Build the binary
		tmpDir, err := os.MkdirTemp("", "yt-transcript-md-e2e-*")
		if err != nil {
			fmt.Printf("failed to create temp dir: %v\n", err)
			return 1
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		binName := "yt-transcript-md"
		if runtime.GOOS == "windows" {
			binName += ".exe"
		}
		binPath = filepath.Join(tmpDir, binName)

		cmd := exec.Command("go", "build", "-tags", "test", "-o", binPath, "../../cmd/yt-transcript-md")
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("failed to build binary: %v\n%s\n", err, output)
			return 1
		}

		return m.Run()
	}()

	os.Exit(exitCode)
}

func TestE2E_Export(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "transcripts.md")

	t.Run("basic export", func(t *testing.T) {
		cmd := exec.Command(binPath, "export", "--links", "dQw4w9WgXcQ", "--out", outPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}

		if _, err := os.Stat(outPath); os.IsNotExist(err) {
			t.Fatal("output file was not created")
		}

		content, _ := os.ReadFile(outPath)
		if !contains(string(content), "Video `dQw4w9WgXcQ`") {
			t.Errorf("output missing video ID")
		}
	})

	t.Run("export via root command", func(t *testing.T) {
		outPath2 := filepath.Join(tempDir, "root.md")
		cmd := exec.Command(binPath, "--links", "dQw4w9WgXcQ", "--out", outPath2)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}

		if _, err := os.Stat(outPath2); os.IsNotExist(err) {
			t.Fatal("output file was not created")
		}
	})

	t.Run("strict mode failure", func(t *testing.T) {
		cmd := exec.Command(binPath, "export", "--links", "fail1234567", "--strict")
		if err := cmd.Run(); err == nil {
			t.Fatal("command should have failed in strict mode")
		}
	})
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
