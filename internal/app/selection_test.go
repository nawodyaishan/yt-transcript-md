package app

import (
	"errors"
	"reflect"
	"testing"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

func TestApplyClipboardSelection(t *testing.T) {
	videos := []models.VideoInput{
		{Original: "one", VideoID: "one11111111"},
		{Original: "two", VideoID: "two22222222"},
		{Original: "three", VideoID: "three333333"},
	}

	tests := []struct {
		name      string
		selection ClipboardSelection
		want      []models.VideoInput
		wantErr   bool
	}{
		{
			name:      "all",
			selection: ClipboardSelection{Mode: ClipboardSelectionAll},
			want:      videos,
		},
		{
			name:      "one",
			selection: ClipboardSelection{Mode: ClipboardSelectionOne, Index: 2},
			want:      []models.VideoInput{videos[1]},
		},
		{
			name:      "recent",
			selection: ClipboardSelection{Mode: ClipboardSelectionRecent, Count: 2},
			want:      videos[:2],
		},
		{
			name:      "one zero index",
			selection: ClipboardSelection{Mode: ClipboardSelectionOne, Index: 0},
			wantErr:   true,
		},
		{
			name:      "one too large index",
			selection: ClipboardSelection{Mode: ClipboardSelectionOne, Index: 4},
			wantErr:   true,
		},
		{
			name:      "recent zero count",
			selection: ClipboardSelection{Mode: ClipboardSelectionRecent, Count: 0},
			wantErr:   true,
		},
		{
			name:      "recent too large count",
			selection: ClipboardSelection{Mode: ClipboardSelectionRecent, Count: 4},
			wantErr:   true,
		},
		{
			name:      "unknown mode",
			selection: ClipboardSelection{Mode: "unknown"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyClipboardSelection(videos, tt.selection)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ApplyClipboardSelection() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyClipboardSelection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyClipboardSelectionCancel(t *testing.T) {
	_, err := ApplyClipboardSelection(nil, ClipboardSelection{Mode: ClipboardSelectionCancel})
	if !errors.Is(err, ErrClipboardSelectionCanceled) {
		t.Fatalf("ApplyClipboardSelection() error = %v, want canceled", err)
	}
}
