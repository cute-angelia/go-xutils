package ifileutil

import (
	"github.com/cute-angelia/go-xutils/syntax/istrings"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

type Component struct {
	config *config
}

func newComponent(config *config) *Component {
	comp := &Component{}
	comp.config = config
	return comp
}

func (self *Component) isBlackDir(filePath string) bool {
	blacked := false
	// 排除文件夹
	if len(self.config.DirDeclude) > 0 && istrings.StringInSlice(filePath, self.config.DirDeclude) {
		blacked = true
	}
	// 选中文件夹
	if len(self.config.DirInclude) > 0 && !istrings.StringInSlice(filePath, self.config.DirInclude) {
		blacked = true
	}
	return blacked
}

func (self *Component) isBlackExt(ext string) bool {
	blacked := false
	fileExt := strings.ToLower(ext)
	// 排除后缀
	if len(self.config.ExtDeclude) > 0 && istrings.StringInSlice(fileExt, self.config.ExtDeclude) {
		blacked = true
	}
	// 包含后缀
	if len(self.config.ExtInclude) > 0 && !istrings.StringInSlice(fileExt, self.config.ExtInclude) {
		blacked = true
	}
	return blacked
}

// GetFileListCurrentDir 获取【当前文件夹】列表
func (self *Component) GetFileListCurrentDir(dir string) (dirPaths []string, filePaths []string, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			if !self.isBlackDir(file.Name()) {
				dirPaths = append(dirPaths, filepath.Join(dir, file.Name()))
			}
		} else {
			if !self.isBlackExt(path.Ext(file.Name())) {
				filePaths = append(filePaths, filepath.Join(dir, file.Name()))
			}
		}
	}
	return
}

// GetFileList 获取文件列表
func (self *Component) GetFileList() (dirPaths []string, filePaths []string, err error) {
	dirPaths, filePaths, err = self.GetFileListCurrentDir(self.config.Dir)
	if err != nil {
		return
	}

	// 读取子文件夹下文件和文件夹
	for _, dirPath2 := range dirPaths {
		dirPaths2, filePaths2, err := self.GetFileListCurrentDir(dirPath2)
		if err != nil {
			return nil, nil, err
		}
		dirPaths = append(dirPaths, dirPaths2...)
		filePaths = append(filePaths, filePaths2...)
	}
	return
}

// IsEmptyDir 检查是否为空文件夹
func (self *Component) IsEmptyDir() bool {
	s, _ := ioutil.ReadDir(self.config.Dir)
	if len(s) == 0 {
		return true
	}
	return false
}
