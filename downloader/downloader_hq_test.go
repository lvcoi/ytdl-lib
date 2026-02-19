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

	video, err := testDownloader.Client.GetVideoContext(ctx, "dQw4w9WgXcQ")
	require.NoError(err)
	require.NoError(testDownloader.DownloadComposite(ctx, "", video, "hd1080", "mp4", ""))
}
