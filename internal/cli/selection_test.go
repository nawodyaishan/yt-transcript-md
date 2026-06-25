package cli

import (
	"testing"

	"github.com/nawodyaishan/yt-transcript-md/internal/app"
)

func TestParseClipboardSelectionFlag(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    app.ClipboardSelection
		wantErr bool
	}{
		{
			name: "all",
			raw:  "all",
			want: app.ClipboardSelection{Mode: app.ClipboardSelectionAll},
		},
		{
			name: "one",
			raw:  "one:2",
			want: app.ClipboardSelection{Mode: app.ClipboardSelectionOne, Index: 2},
		},
		{
			name: "recent",
			raw:  "recent:3",
			want: app.ClipboardSelection{Mode: app.ClipboardSelectionRecent, Count: 3},
		},
		{name: "unknown", raw: "latest:3", wantErr: true},
		{name: "zero index", raw: "one:0", wantErr: true},
		{name: "negative count", raw: "recent:-1", wantErr: true},
		{name: "non numeric", raw: "one:two", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseClipboardSelectionFlag(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseClipboardSelectionFlag() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseClipboardSelectionFlag() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
