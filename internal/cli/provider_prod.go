//go:build !test

package cli

import (
	"github.com/nawodyaishan/yt-transcript-md/internal/transcript"
)

func getProvider() transcript.Provider {
	return transcript.NewYouTubeProvider()
}
