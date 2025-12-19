package iupload

import (
	"errors"
	"github.com/cute-angelia/go-xutils/syntax/islice"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

const componentName = "component.store.iupload"

type Component struct {
	name   string
	config *config
	locker sync.Mutex
}

type FileType int

const (
	FileTypeImage FileType = iota + 1 // 图片
	FileTypeVideo                     // 视频
)

// UploadFile 文件对象
type UploadFile struct {
	Name string   // 文件名称
	Type FileType // 文件类型
	Size int64    // 文件大小
	Ext  string   // 文件扩展
	Uri  string   // 文件路径
}

// newComponent ...
func newComponent(config *config) *Component {
	return &Component{
		name:   componentName,
		config: config,
	}
}

func (that *Component) Upload(file *multipart.FileHeader, folder string, fileName string, fileType FileType) (uf *UploadFile, e error) {
	if e = that.checkFile(file, fileType); e != nil {
		return
	}
	// 映射目录
	directory := that.config.UploadDirectory
	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return uf, errors.New("打开文件失败!" + err.Error())
	}
	defer src.Close()
	// 文件信息
	savePath := path.Join(directory, folder, path.Dir(fileName))
	saveFilePath := path.Join(directory, folder, fileName)
	// 创建目录
	err = os.MkdirAll(savePath, 0755)
	if err != nil && !os.IsExist(err) {
		return uf, errors.New("创建上传目录失败!" + err.Error())
	}
	// 创建目标文件
	out, err := os.Create(saveFilePath)
	if err != nil {
		return uf, errors.New("创建文件失败!" + err.Error())
	}
	defer out.Close()
	// 写入目标文件
	_, err = io.Copy(out, src)
	if err != nil {
		return uf, errors.New("上传文件失败: " + err.Error())
	}

	fileRelPath := path.Join(folder, fileName)
	return &UploadFile{
		Name: file.Filename,
		Type: fileType,
		Size: file.Size,
		Ext:  strings.ToLower(strings.Replace(path.Ext(file.Filename), ".", "", 1)),
		Uri:  fileRelPath,
	}, nil
}

// checkFile 文件验证
func (that *Component) checkFile(file *multipart.FileHeader, fileType FileType) (e error) {
	fileName := file.Filename
	fileExt := strings.ToLower(strings.Replace(path.Ext(fileName), ".", "", 1))
	fileSize := file.Size

	if fileType == FileTypeImage {
		// 图片文件
		if !islice.InSlice(that.config.UploadImageExt, fileExt) {
			return errors.New("不被支持的图片扩展: " + fileExt)
		}

		if fileSize > that.config.UploadImageSize {
			return errors.New("上传图片不能超出限制: " + strconv.FormatInt(that.config.UploadImageSize/1024/1024, 10) + "M")
		}
	} else if fileType == FileTypeVideo {
		// 视频文件
		if !islice.InSlice(that.config.UploadVideoExt, fileExt) {
			return errors.New("不被支持的视频扩展: " + fileExt)
		}

		if fileSize > that.config.UploadVideoSize {
			return errors.New("上传视频不能超出限制: " + strconv.FormatInt(that.config.UploadVideoSize/1024/1024, 10) + "M")
		}
	} else {
		return errors.New("上传文件类型错误" + fileExt)
	}

	return nil
}
