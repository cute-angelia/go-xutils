package ffmpeg

import (
	"errors"
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/icmd"
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"log"
	"os"
	"strings"
	"time"
)

const (
	TempConcatTxt = "ConcatMovFiles.txt"
)

// getTempText 需要一个文件列表 txt
func (c *Component) getTempText() string {
	if len(c.config.FilesPath) > 0 {
		return c.config.FilesPath + "/" + TempConcatTxt
	} else {
		return fmt.Sprintf("%s/%s", ifile.GetHomeDir(), TempConcatTxt)
	}
}

func (c *Component) generateText(ext []string) (tempText string, err error) {
	text := c.getTempText()
	if ifile.IsExist(text) {
		// 删除旧文件
		ifile.DeleteFile(text)
	}

	// 获取文件夹
	if _, files, err := ifile.GetDepthOnePathsAndFilesIncludeExt(c.config.FilesPath, ext...); err != nil {
		log.Println("get path", c.config.FilesPath, ext, err)
		return "", err
	} else {
		// 生成一个文件用于合并： 格式
		// file ./name.mov
		// log.Println(c.getTempText())
		log.Println(ijson.Pretty(files))

		if itempText, err := ifile.OpenLocalFile(text); err != nil {
			return "", err
		} else {
			defer itempText.Close()
			for _, file := range files {
				itempText.WriteString(fmt.Sprintf("file '%s'\n", file))
			}
			return text, nil
		}
	}
}

// ConcatMovFiles 合并 MOV
// ffmpeg -safe 0 -f concat -i files_to_combine -vcodec copy -acodec copy merged.MOV
func (c *Component) ConcatMovFiles(ext []string, saveName string) error {
	if len(saveName) == 0 {
		if len(ext) > 0 {
			saveName = time.Now().Format("20060102-150405") + ext[0]
		} else {
			saveName = time.Now().Format("20060102-150405")
		}
	}

	log.Println("处理文件：", c.config.FilesPath, ext)

	if temptext, err := c.generateText(ext); err != nil {
		log.Println(err)
		return err
	} else {
		if !ifile.IsDir(c.config.FilesPath + "/success/") {
			ifile.Mkdir(c.config.FilesPath+"/success/", 0755)
		}

		status := icmd.Exec(c.config.FfmpegPath, []string{
			"-loglevel", "error",
			"-safe", "0",
			"-f", "concat",
			"-i", temptext,
			"-vcodec", "copy",
			"-acodec", "copy",
			c.config.FilesPath + "/success/" + saveName,
		}, c.config.Timeout)

		//log.Println(ijson.Pretty(status))

		if len(status.Stderr) > 0 {
			err := errors.New(strings.Join(status.Stderr, ""))
			return err
		} else {
			os.Remove(temptext)

			log.Println("success", c.config.FilesPath)
			return nil
		}
	}
}
