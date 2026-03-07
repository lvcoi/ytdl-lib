package youtube

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYoutube_GetItagInfo(t *testing.T) {
	if os.Getenv("YTDL_RUN_LIVE_IT") != "1" {
		t.Skip("skipping live YouTube integration test (set YTDL_RUN_LIVE_IT=1 to run)")
	}
	require := require.New(t)
	client := Client{}

	// url from issue #25
	url := "https://www.youtube.com/watch?v=rFejpH_tAHM"
	video, err := client.GetVideo(url)
	require.NoError(err)
	require.GreaterOrEqual(len(video.Formats), 24)
}
