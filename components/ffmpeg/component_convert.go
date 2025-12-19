package ffmpeg

import (
	"errors"
	"github.com/cute-angelia/go-xutils/syntax/icmd"
	"strings"
)

func (c Component) Convert(input string, savePath string) error {
	status := icmd.Exec(c.config.FfmpegPath, []string{
		"-loglevel", "error",
		"-i", input,
		"-c:v", "copy", "-c:a", "copy",
		savePath,
	}, c.config.Timeout)

	// log.Println(ijson.Pretty(status))

	if len(status.Stderr) > 0 {
		err := errors.New(strings.Join(status.Stderr, ""))
		return err
	}

	return nil
}
