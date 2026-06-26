//go:build docker

package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	testImage = "yt-transcript-md-e2e"
	testURL1  = "https://youtu.be/pLXb6TDyIUI"
	testURL2  = "https://youtu.be/jNQXAC9IVRw"
)

var (
	linuxBin string
	goArch   string
)

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Fprintln(os.Stderr, "[SKIP] docker not found in PATH")
		os.Exit(0)
	}

	goArch = runtime.GOARCH

	tmp, err := os.MkdirTemp("", "yt-docker-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	linuxBin = filepath.Join(tmp, "yt-transcript-md")

	fmt.Fprintf(os.Stderr, "[BUILD] cross-compiling linux/%s binary...\n", goArch)
	cmd := exec.Command("go", "build", "-o", linuxBin, "../../cmd/yt-transcript-md")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+goArch)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "cross-compile failed: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// dockerRun mounts the Linux binary into the test image and runs yt-transcript-md with args.
func dockerRun(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmdArgs := []string{
		"run", "--rm",
		"--platform", "linux/" + goArch,
		"-v", linuxBin + ":/usr/local/bin/yt-transcript-md",
		testImage,
		"yt-transcript-md",
	}
	cmdArgs = append(cmdArgs, args...)
	out, err := exec.Command("docker", cmdArgs...).CombinedOutput()
	return string(out), err
}

func TestDockerLinux_SinglePositionalURL(t *testing.T) {
	out, err := dockerRun(t, testURL1)
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, out)
	}
	for _, want := range []string{"pLXb6TDyIUI", "[OK]", "1 succeeded", "[SAVED]"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestDockerLinux_MultiplePositionalURLs(t *testing.T) {
	out, err := dockerRun(t, testURL1, testURL2)
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, out)
	}
	for _, want := range []string{"pLXb6TDyIUI", "jNQXAC9IVRw", "2 succeeded"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestDockerLinux_OutputFile(t *testing.T) {
	outDir := t.TempDir()
	args := []string{
		"run", "--rm",
		"--platform", "linux/" + goArch,
		"-v", linuxBin + ":/usr/local/bin/yt-transcript-md",
		"-v", outDir + ":/out",
		testImage,
		"yt-transcript-md", testURL1, "--out", "/out/transcript.md",
	}
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, out)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "transcript.md"))
	if err != nil {
		t.Fatalf("output file not written: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(data), "pLXb6TDyIUI") {
		t.Errorf("output file missing video ID:\n%s", string(data))
	}
}

func TestDockerLinux_InvalidURLFails(t *testing.T) {
	out, err := dockerRun(t, "not-a-youtube-link-at-all")
	if err == nil {
		t.Fatalf("command should have failed:\n%s", out)
	}
	if !strings.Contains(out, "input error") {
		t.Errorf("expected input error in output:\n%s", out)
	}
}

func TestDockerLinux_ConflictFlagRejected(t *testing.T) {
	out, err := dockerRun(t, testURL1, "--links", "dQw4w9WgXcQ")
	if err == nil {
		t.Fatalf("command should have failed:\n%s", out)
	}
	if !strings.Contains(out, "--links") {
		t.Errorf("expected --links conflict error:\n%s", out)
	}
}
