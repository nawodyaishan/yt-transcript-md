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

	t.Run("clipboard multi-link selection all", func(t *testing.T) {
		dir := filepath.Join(tempDir, "clipboard-all")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--clipboard-selection", "all")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=https://youtu.be/dQw4w9WgXcQ\nhttps://youtu.be/jNQXAC9IVRw",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}

		content, err := os.ReadFile(filepath.Join(dir, "transcripts.md"))
		if err != nil {
			t.Fatalf("default output file was not written: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") || !contains(string(content), "Video `jNQXAC9IVRw`") {
			t.Fatalf("output missing selected videos:\n%s", content)
		}
	})

	t.Run("clipboard multi-link selection one", func(t *testing.T) {
		dir := filepath.Join(tempDir, "clipboard-one")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--clipboard-selection", "one:2")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=https://youtu.be/dQw4w9WgXcQ\nhttps://youtu.be/jNQXAC9IVRw",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}

		content, err := os.ReadFile(filepath.Join(dir, "transcripts.md"))
		if err != nil {
			t.Fatalf("default output file was not written: %v", err)
		}
		if contains(string(content), "Video `dQw4w9WgXcQ`") {
			t.Fatalf("output contains unselected first video:\n%s", content)
		}
		if !contains(string(content), "Video `jNQXAC9IVRw`") {
			t.Fatalf("output missing selected second video:\n%s", content)
		}
	})

	t.Run("clipboard multi-link selection recent", func(t *testing.T) {
		dir := filepath.Join(tempDir, "clipboard-recent")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--clipboard-selection", "recent:2")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=Videos: https://youtu.be/dQw4w9WgXcQ, https://youtu.be/jNQXAC9IVRw, https://youtu.be/BaW_jenozKc",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}

		content, err := os.ReadFile(filepath.Join(dir, "transcripts.md"))
		if err != nil {
			t.Fatalf("default output file was not written: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") || !contains(string(content), "Video `jNQXAC9IVRw`") {
			t.Fatalf("output missing recent videos:\n%s", content)
		}
		if contains(string(content), "Video `BaW_jenozKc`") {
			t.Fatalf("output contains unselected third video:\n%s", content)
		}
	})

	t.Run("clipboard history selection all", func(t *testing.T) {
		dir := filepath.Join(tempDir, "history-all")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--history-source", "copyq", "--history-limit", "10", "--clipboard-selection", "all")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
			"YT_TRANSCRIPT_MD_TEST_HISTORY_SOURCE=copyq",
			"YT_TRANSCRIPT_MD_TEST_HISTORY=https://youtu.be/dQw4w9WgXcQ\n---\nhttps://youtu.be/jNQXAC9IVRw",
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}
		content, err := os.ReadFile(filepath.Join(dir, "transcripts.md"))
		if err != nil {
			t.Fatalf("default output file was not written: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") || !contains(string(content), "Video `jNQXAC9IVRw`") {
			t.Fatalf("output missing history videos:\n%s", content)
		}
	})

	t.Run("clipboard history limit", func(t *testing.T) {
		dir := filepath.Join(tempDir, "history-limit")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--history-source", "copyq", "--history-limit", "1", "--clipboard-selection", "all")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
			"YT_TRANSCRIPT_MD_TEST_HISTORY_SOURCE=copyq",
			"YT_TRANSCRIPT_MD_TEST_HISTORY=https://youtu.be/dQw4w9WgXcQ\n---\nhttps://youtu.be/jNQXAC9IVRw",
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}
		content, err := os.ReadFile(filepath.Join(dir, "transcripts.md"))
		if err != nil {
			t.Fatalf("default output file was not written: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") {
			t.Fatalf("output missing first history video:\n%s", content)
		}
		if contains(string(content), "Video `jNQXAC9IVRw`") {
			t.Fatalf("output contains history entry beyond limit:\n%s", content)
		}
	})

	t.Run("no history uses current clipboard only", func(t *testing.T) {
		dir := filepath.Join(tempDir, "no-history")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--no-history")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=https://youtu.be/dQw4w9WgXcQ",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
			"YT_TRANSCRIPT_MD_TEST_HISTORY_SOURCE=copyq",
			"YT_TRANSCRIPT_MD_TEST_HISTORY=https://youtu.be/jNQXAC9IVRw",
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}
		content, err := os.ReadFile(filepath.Join(dir, "transcripts.md"))
		if err != nil {
			t.Fatalf("default output file was not written: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") || contains(string(content), "Video `jNQXAC9IVRw`") {
			t.Fatalf("output should contain only current clipboard video:\n%s", content)
		}
	})

	t.Run("clipboard multi-link non-interactive requires selection", func(t *testing.T) {
		dir := filepath.Join(tempDir, "clipboard-no-selection")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=https://youtu.be/dQw4w9WgXcQ\nhttps://youtu.be/jNQXAC9IVRw",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
		)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed without selection:\n%s", output)
		}
		if !contains(string(output), "clipboard selection is required") {
			t.Fatalf("error should explain selection requirement:\n%s", output)
		}
		if _, statErr := os.Stat(filepath.Join(dir, "transcripts.md")); !os.IsNotExist(statErr) {
			t.Fatalf("output file should not be created without selection")
		}
	})

	t.Run("clipboard selection flag rejected with explicit root links", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "selection-rejected.md")
		cmd := exec.Command(binPath, "--links", "dQw4w9WgXcQ", "--clipboard-selection", "all", "--out", outPath)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed with explicit links and selection flag:\n%s", output)
		}
		if !contains(string(output), "--clipboard-selection can only be used with the default clipboard workflow") {
			t.Fatalf("error should explain selection flag scope:\n%s", output)
		}
		if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
			t.Fatalf("output file should not be created after rejected selection flag")
		}
	})

	t.Run("history flag rejected with explicit root links", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "history-rejected.md")
		cmd := exec.Command(binPath, "--links", "dQw4w9WgXcQ", "--history-source", "copyq", "--out", outPath)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed with explicit links and history flag:\n%s", output)
		}
		if !contains(string(output), "--history-source, --history-limit, and --no-history can only be used with the default clipboard workflow") {
			t.Fatalf("error should explain history flag scope:\n%s", output)
		}
		if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
			t.Fatalf("output file should not be created after rejected history flag")
		}
	})

	t.Run("clipboard selection flag rejected with export command", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "selection-export-rejected.md")
		cmd := exec.Command(binPath, "export", "--links", "dQw4w9WgXcQ", "--clipboard-selection", "all", "--out", outPath)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed with export command and selection flag:\n%s", output)
		}
		if !contains(string(output), "unknown flag: --clipboard-selection") {
			t.Fatalf("error should explain selection flag is unavailable for export:\n%s", output)
		}
		if _, statErr := os.Stat(outPath); !os.IsNotExist(statErr) {
			t.Fatalf("output file should not be created after rejected export selection flag")
		}
	})

	t.Run("invalid clipboard selection flag fails before output", func(t *testing.T) {
		dir := filepath.Join(tempDir, "clipboard-invalid-selection")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--clipboard-selection", "one:0")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=https://youtu.be/dQw4w9WgXcQ\nhttps://youtu.be/jNQXAC9IVRw",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
		)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed with invalid selection flag:\n%s", output)
		}
		if !contains(string(output), "invalid video index") {
			t.Fatalf("error should explain invalid selection flag:\n%s", output)
		}
		if _, statErr := os.Stat(filepath.Join(dir, "transcripts.md")); !os.IsNotExist(statErr) {
			t.Fatalf("output file should not be created after invalid selection flag")
		}
	})

	t.Run("out of range clipboard selection fails for single video", func(t *testing.T) {
		dir := filepath.Join(tempDir, "clipboard-out-of-range-selection")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		clipboardOut := filepath.Join(dir, "clipboard.md")
		cmd := exec.Command(binPath, "--clipboard-selection", "one:2")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD=https://youtu.be/dQw4w9WgXcQ",
			"YT_TRANSCRIPT_MD_TEST_CLIPBOARD_OUT="+clipboardOut,
		)
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed with out-of-range selection flag:\n%s", output)
		}
		if !contains(string(output), "video index 2 is out of range") {
			t.Fatalf("error should explain out-of-range selection:\n%s", output)
		}
		if _, statErr := os.Stat(filepath.Join(dir, "transcripts.md")); !os.IsNotExist(statErr) {
			t.Fatalf("output file should not be created after out-of-range selection flag")
		}
	})

	t.Run("strict mode failure", func(t *testing.T) {
		cmd := exec.Command(binPath, "export", "--links", "fail1234567", "--strict")
		if err := cmd.Run(); err == nil {
			t.Fatal("command should have failed in strict mode")
		}
	})
}

func TestE2E_PositionalArgs(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("single positional URL via root command", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "single.md")
		cmd := exec.Command(binPath, "https://youtu.be/dQw4w9WgXcQ", "--out", outPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}
		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("output file was not created: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") {
			t.Errorf("output missing video ID:\n%s", content)
		}
	})

	t.Run("multiple positional URLs via root command", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "multi.md")
		cmd := exec.Command(binPath, "https://youtu.be/dQw4w9WgXcQ", "https://youtu.be/jNQXAC9IVRw", "--out", outPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}
		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("output file was not created: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") || !contains(string(content), "Video `jNQXAC9IVRw`") {
			t.Errorf("output missing one or more video IDs:\n%s", content)
		}
	})

	t.Run("positional URL via export subcommand", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "export-positional.md")
		cmd := exec.Command(binPath, "export", "https://youtu.be/dQw4w9WgXcQ", "--out", outPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}
		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("output file was not created: %v", err)
		}
		if !contains(string(content), "Video `dQw4w9WgXcQ`") {
			t.Errorf("output missing video ID:\n%s", content)
		}
	})

	t.Run("positional URL conflicts with --links", func(t *testing.T) {
		cmd := exec.Command(binPath, "https://youtu.be/dQw4w9WgXcQ", "--links", "jNQXAC9IVRw")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed: %s", output)
		}
		if !contains(string(output), "--links") {
			t.Errorf("error should mention --links:\n%s", output)
		}
	})

	t.Run("positional URL conflicts with --input-file", func(t *testing.T) {
		cmd := exec.Command(binPath, "https://youtu.be/dQw4w9WgXcQ", "--input-file", "links.txt")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed: %s", output)
		}
		if !contains(string(output), "--input-file") {
			t.Errorf("error should mention --input-file:\n%s", output)
		}
	})

	t.Run("positional URL conflicts with clipboard workflow flag", func(t *testing.T) {
		cmd := exec.Command(binPath, "https://youtu.be/dQw4w9WgXcQ", "--history-source", "copyq")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed: %s", output)
		}
		if !contains(string(output), "--history-source") {
			t.Errorf("error should mention --history-source:\n%s", output)
		}
	})

	t.Run("duplicate positional URLs are deduplicated", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "dedup.md")
		cmd := exec.Command(binPath, "https://youtu.be/dQw4w9WgXcQ", "https://youtu.be/dQw4w9WgXcQ", "--out", outPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command failed: %v\n%s", err, output)
		}
		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("output file was not created: %v", err)
		}
		if count := strings.Count(string(content), "Video `dQw4w9WgXcQ`"); count != 1 {
			t.Errorf("expected video to appear once, got %d times:\n%s", count, content)
		}
	})

	t.Run("invalid positional argument fails clearly", func(t *testing.T) {
		cmd := exec.Command(binPath, "not-a-youtube-link-at-all")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("command should have failed: %s", output)
		}
		if !contains(string(output), "input error") {
			t.Errorf("error should describe input problem:\n%s", output)
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
		if !contains(help, "clipboard-history managers") {
			t.Fatalf("root help missing clipboard history language:\n%s", help)
		}
		if !contains(help, "--clipboard-selection") {
			t.Fatalf("root help missing clipboard selection flag:\n%s", help)
		}
		if !contains(help, "--history-source") || !contains(help, "--history-limit") || !contains(help, "--no-history") {
			t.Fatalf("root help missing clipboard history flags:\n%s", help)
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
