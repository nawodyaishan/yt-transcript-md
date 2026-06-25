package history

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeRunner struct {
	paths map[string]bool
	outs  map[string][]byte
	errs  map[string]error
	calls []string
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	if f.paths[file] {
		return "/bin/" + file, nil
	}
	return "", errors.New("not found")
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, key)
	if err, ok := f.errs[key]; ok {
		return nil, err
	}
	if out, ok := f.outs[key]; ok {
		return out, nil
	}
	return nil, errors.New("missing fake output: " + key)
}

func TestCollectCandidates(t *testing.T) {
	providers := []Provider{
		fakeProvider{name: SourceCopyQ, entries: []Entry{
			{Provider: SourceCopyQ, ID: "0", Text: "https://youtu.be/jNQXAC9IVRw", Rank: 0},
			{Provider: SourceCopyQ, ID: "1", Text: "duplicate dQw4w9WgXcQ", Rank: 1},
		}},
	}

	got, warnings, err := CollectCandidates(context.Background(), "current https://youtu.be/dQw4w9WgXcQ", providers, Options{Source: SourceAuto, Limit: 10})
	if err != nil {
		t.Fatalf("CollectCandidates() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("CollectCandidates() warnings = %v", warnings)
	}
	ids := make([]string, 0, len(got))
	for _, candidate := range got {
		ids = append(ids, candidate.Video.VideoID)
	}
	want := []string{"dQw4w9WgXcQ", "jNQXAC9IVRw"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("candidate IDs = %v, want %v", ids, want)
	}
}

func TestCollectCandidatesNoHistory(t *testing.T) {
	provider := fakeProvider{name: SourceCopyQ, entries: []Entry{{Text: "https://youtu.be/jNQXAC9IVRw"}}}
	got, _, err := CollectCandidates(context.Background(), "https://youtu.be/dQw4w9WgXcQ", []Provider{provider}, Options{NoHistory: true})
	if err != nil {
		t.Fatalf("CollectCandidates() error = %v", err)
	}
	if len(got) != 1 || got[0].Video.VideoID != "dQw4w9WgXcQ" {
		t.Fatalf("candidates = %v, want only current clipboard", got)
	}
}

func TestCopyQProvider(t *testing.T) {
	runner := &fakeRunner{
		paths: map[string]bool{"copyq": true},
		outs: map[string][]byte{
			"copyq read 0": []byte("https://youtu.be/dQw4w9WgXcQ\n"),
			"copyq read 1": []byte("https://youtu.be/jNQXAC9IVRw\n"),
		},
		errs: map[string]error{
			"copyq read 2": errors.New("missing"),
		},
	}
	provider := NewCopyQProvider(runner)
	if err := provider.Available(context.Background()); err != nil {
		t.Fatalf("Available() error = %v", err)
	}
	entries, err := provider.Entries(context.Background(), 5)
	if err != nil {
		t.Fatalf("Entries() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %v, want 2", entries)
	}
}

func TestCliphistProvider(t *testing.T) {
	runner := &fakeRunner{
		paths: map[string]bool{"cliphist": true},
		outs: map[string][]byte{
			"cliphist list":             []byte("1\tfirst\n2\tsecond\n"),
			"cliphist decode 1\tfirst":  []byte("https://youtu.be/dQw4w9WgXcQ\n"),
			"cliphist decode 2\tsecond": []byte("https://youtu.be/jNQXAC9IVRw\n"),
		},
	}
	entries, err := NewCliphistProvider(runner).Entries(context.Background(), 10)
	if err != nil {
		t.Fatalf("Entries() error = %v", err)
	}
	if len(entries) != 2 || entries[0].ID != "1" {
		t.Fatalf("entries = %v, want parsed cliphist entries", entries)
	}
}

func TestGPasteProvider(t *testing.T) {
	runner := &fakeRunner{
		paths: map[string]bool{"gpaste-client": true},
		outs: map[string][]byte{
			"gpaste-client history --oneline": []byte("0: first\n1: second\n"),
			"gpaste-client get 0 --raw":       []byte("https://youtu.be/dQw4w9WgXcQ\n"),
			"gpaste-client get 1 --raw":       []byte("https://youtu.be/jNQXAC9IVRw\n"),
		},
	}
	entries, err := NewGPasteProvider(runner).Entries(context.Background(), 10)
	if err != nil {
		t.Fatalf("Entries() error = %v", err)
	}
	if len(entries) != 2 || entries[1].ID != "1" {
		t.Fatalf("entries = %v, want parsed gpaste entries", entries)
	}
}

type fakeProvider struct {
	name    string
	entries []Entry
	err     error
}

func (p fakeProvider) Name() string { return p.name }
func (p fakeProvider) Available(ctx context.Context) error {
	return p.err
}
func (p fakeProvider) Entries(ctx context.Context, limit int) ([]Entry, error) {
	if p.err != nil {
		return nil, p.err
	}
	if limit < len(p.entries) {
		return p.entries[:limit], nil
	}
	return p.entries, nil
}
