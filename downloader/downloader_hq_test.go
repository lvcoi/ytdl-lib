//go:build integration
// +build integration

package downloader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownload_HighQuality(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	video, err := testDownloader.Client.GetVideoContext(ctx, "BaW_jenozKc")
	if err != nil {
		t.Skipf("Skipping test: video is not available: %v", err)
	}
	require.NoError(testDownloader.DownloadComposite(ctx, "", video, "hd1080", "mp4", ""))
}
