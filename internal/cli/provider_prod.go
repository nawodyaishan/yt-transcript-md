//go:build !test

package cli

import (
	"github.com/nawodyaishan/yt-transcript-md/internal/app"
	"github.com/nawodyaishan/yt-transcript-md/internal/clipboard"
	"github.com/nawodyaishan/yt-transcript-md/internal/metadata"
	"github.com/nawodyaishan/yt-transcript-md/internal/transcript"
)

func getProvider() transcript.Provider {
	return transcript.NewYouTubeProvider()
}

func getClipboard() app.Clipboard {
	return clipboard.System{}
}

func getMetadataProvider() metadata.Provider {
	return metadata.NewOEmbedProvider()
}
