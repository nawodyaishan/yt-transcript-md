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

	t.Run("default export via clipboard", func(t *testing.T) {
		clipboardOut := filepath.Join(tempDir, "clipboard.md")
		cmd := exec.Command(binPath)
		cmd.Dir = tempDir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=https://youtu.be/dQw4w9WgXcQ",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}

		content, err := os.ReadFile(clipboardOut)
		if err != nil {
			t.Fatalf("clipboard output was not written: %v", err)
		}

		savedPath := filepath.Join(tempDir, "transcripts.md")
		saved, err := os.ReadFile(savedPath)
		if err != nil {
			t.Fatalf("default output file was not written: %v", err)
		}
		if string(saved) != string(content) {
			t.Errorf("saved output and clipboard output differ")
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") {
			t.Errorf("clipboard output missing video ID")
		}
		if !contains(string(content), "Test transcript snippet") {
			t.Errorf("clipboard output missing transcript")
		}
		if !contains(string(content), "Test Video") {
			t.Errorf("clipboard output missing metadata")
		}
	})

	t.Run("strict mode failure", func(t *testing.T) {
		cmd := exec.Command(binPath, "export", "--links", "fail1234567", "--strict")
		if err := cmd.Run(); err == nil {
			t.Fatal("command should have failed in strict mode")
		}
	})
}

func TestE2E_Help(t *testing.T) {
	t.Run("root help leads with clipboard workflow", func(t *testing.T) {
		cmd := exec.Command(binPath, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("help command failed: %v\n%s", err, output)
		}

		help := string(output)
		clipboardIndex := strings.Index(help, "Default workflow: read clipboard")
		flagsIndex := strings.Index(help, "--links")
		if clipboardIndex == -1 {
			t.Fatalf("root help missing clipboard workflow:\n%s", help)
		}
		if flagsIndex == -1 {
			t.Fatalf("root help missing flags:\n%s", help)
		}
		if clipboardIndex > flagsIndex {
			t.Fatalf("clipboard workflow should appear before advanced flags:\n%s", help)
		}
	})

	t.Run("export help explains advanced file export", func(t *testing.T) {
		cmd := exec.Command(binPath, "export", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("export help command failed: %v\n%s", err, output)
		}

		help := string(output)
		if !contains(help, "Advanced file and batch transcript export") {
			t.Fatalf("export help missing advanced export summary:\n%s", help)
		}
		if !contains(help, "For the simplest workflow, copy a YouTube link and run yt-transcript-md") {
			t.Fatalf("export help missing default workflow pointer:\n%s", help)
		}
	})
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
