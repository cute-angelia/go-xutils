package ffmpeg

import (
	"errors"
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/icmd"
	"github.com/cute-angelia/go-xutils/syntax/itime"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (c Component) cutOne(sec string, mp4path string, savePic string) error {
	status := icmd.Exec(c.config.FfmpegPath, []string{
		"-loglevel", "error",
		"-y",
		"-ss", sec,
		"-t", "1",
		"-i", mp4path,
		"-vframes", "1",
		savePic,
	}, c.config.Timeout)

	// log.Println(ijson.Pretty(status))

	if len(status.Stderr) > 0 {
		err := errors.New(strings.Join(status.Stderr, ""))
		return err
	}

	return nil
}

// 截取视频第几秒图片
func (c *Component) GetCutPictureOneRandom(mp4path string, savePic string) error {
	rand.Seed(time.Now().UnixNano())
	videoLen, _ := c.GetVideoLength(mp4path)
	cutsec := rand.Intn(videoLen)
	sec := strconv.Itoa(cutsec)
	return c.cutOne(sec, mp4path, savePic)
}

func (c *Component) GetCutPictureOne(mp4path string, cutsec int, savePic string) error {
	videoLen, _ := c.GetVideoLength(mp4path)
	if cutsec > videoLen {
		return errors.New(fmt.Sprintf("视频长度(%d)小于剪辑秒数(%d)", videoLen, cutsec))
	}
	sec := strconv.Itoa(cutsec)
	return c.cutOne(sec, mp4path, savePic)
}

// 截取视频图片，按x秒
func (c *Component) GetCutPictures(mp4path string, cutsec int, saveDir string) error {
	var outputerror string
	videoLen, _ := c.GetVideoLength(mp4path)
	for i := 0; i < videoLen; i = i + cutsec {
		sec := strconv.Itoa(i)
		status := icmd.Exec(c.config.FfmpegPath, []string{
			"-loglevel", "error",
			"-y",
			"-ss", sec,
			"-t", "1",
			"-i", mp4path,
			"-vframes", "1",
			saveDir + "/" + strings.Join(itime.ConvertVideoSecToStr(int64(i)), "-") + ".jpg",
		}, c.config.Timeout)

		// log.Println(ijson.Pretty(status))
		if len(status.Stderr) > 0 {
			outputerror += strings.Join(status.Stderr, "")
		}
	}
	return errors.New(outputerror)
}

// 获取视频时长 s
func (c *Component) GetVideoLength(mp4path string) (int, error) {
	var length int

	// 查询长度
	status := icmd.Exec(c.config.FfmpegPath, []string{
		"-i",
		mp4path,
	}, c.config.Timeout)

	// Duration: 00:22:08.60
	video_length_regexp, _ := regexp.Compile(`Duration: (\d\d:\d\d:\d\d).\d\d`)
	str := video_length_regexp.FindStringSubmatch(strings.Join(status.Stderr, ""))
	if len(str) > 0 {
		// str = "2006-01-02" + strings.TrimPrefix(str, "Duration:")
		strtime := "2006-01-02 " + str[1]
		if videotime, err := time.Parse("2006-01-02 15:04:05", strtime); err != nil {
			length = 0
		} else {
			length = videotime.Hour()*3600 + videotime.Minute()*60 + videotime.Second()
		}
	}
	return length, nil
}
