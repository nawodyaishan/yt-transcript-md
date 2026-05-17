package input

import (
	"reflect"
	"testing"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{"raw ID", "dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"youtu.be link", "https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"youtube.com watch link", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"m.youtube.com watch link with timestamp", "https://m.youtube.com/watch?v=dQw4w9WgXcQ&t=12", "dQw4w9WgXcQ", false},
		{"shorts link", "https://www.youtube.com/shorts/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"embed link", "https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"live link", "https://www.youtube.com/live/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"angle bracket wrapped", "<https://youtu.be/dQw4w9WgXcQ>", "dQw4w9WgXcQ", false},
		{"invalid host", "https://example.com/not-youtube", "", true},
		{"too short ID", "abc123abc1", "", true},
		{"too long ID", "abc123abc123", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractVideoID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractVideoID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractVideoID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitInputLinks(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{"comma separated", "a,b,c", []string{"a", "b", "c"}},
		{"newline separated", "a\nb\nc", []string{"a", "b", "c"}},
		{"mixed", "a,b\nc", []string{"a", "b", "c"}},
		{"with spaces", " a , b \n c ", []string{"a", "b", "c"}},
		{"empty entries", "a,,b\n\nc", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitInputLinks(tt.raw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitInputLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseVideoInputs(t *testing.T) {
	raw := "dQw4w9WgXcQ,https://youtu.be/dQw4w9WgXcQ"
	want := []models.VideoInput{
		{Original: "dQw4w9WgXcQ", VideoID: "dQw4w9WgXcQ"},
	}

	got, err := ParseVideoInputs(raw)
	if err != nil {
		t.Fatalf("ParseVideoInputs() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseVideoInputs() = %v, want %v", got, want)
	}
}
