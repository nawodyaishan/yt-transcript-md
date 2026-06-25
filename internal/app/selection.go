package app

import (
	"errors"
	"fmt"

	"github.com/nawodyaishan/yt-transcript-md/internal/models"
)

type ClipboardSelectionMode string

const (
	ClipboardSelectionAll    ClipboardSelectionMode = "all"
	ClipboardSelectionOne    ClipboardSelectionMode = "one"
	ClipboardSelectionRecent ClipboardSelectionMode = "recent"
	ClipboardSelectionCancel ClipboardSelectionMode = "cancel"
)

var ErrClipboardSelectionCanceled = errors.New("clipboard selection canceled")

type ClipboardSelection struct {
	Mode  ClipboardSelectionMode
	Index int
	Count int
}

type ClipboardSelector interface {
	Select(videos []models.VideoInput) (ClipboardSelection, error)
}

type FixedClipboardSelector interface {
	ClipboardSelector
	FixedSelection() ClipboardSelection
}

func ApplyClipboardSelection(videos []models.VideoInput, selection ClipboardSelection) ([]models.VideoInput, error) {
	switch selection.Mode {
	case ClipboardSelectionAll:
		return videos, nil
	case ClipboardSelectionOne:
		if selection.Index < 1 || selection.Index > len(videos) {
			return nil, fmt.Errorf("video index %d is out of range", selection.Index)
		}
		return []models.VideoInput{videos[selection.Index-1]}, nil
	case ClipboardSelectionRecent:
		if selection.Count < 1 {
			return nil, fmt.Errorf("recent count must be greater than zero")
		}
		if selection.Count > len(videos) {
			return nil, fmt.Errorf("recent count %d is greater than detected video count %d", selection.Count, len(videos))
		}
		return videos[:selection.Count], nil
	case ClipboardSelectionCancel:
		return nil, ErrClipboardSelectionCanceled
	default:
		return nil, fmt.Errorf("unknown clipboard selection mode: %s", selection.Mode)
	}
}
