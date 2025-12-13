package ifile

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// 检查是否为空文件夹
// CheckIsEmptyDir 检查指定路径是否为空文件夹（排除.DS_Store）
func CheckIsEmptyDir(dirpath string) bool {
	// 读取目录内容
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		// 读取失败默认非空
		return false
	}
	// 过滤.DS_Store文件
	for _, file := range files {
		if file.Name() != ".DS_Store" {
			return false
		}
	}
	return true
}

// GetParentDirectory 获取指定路径的上级目录名
func GetParentDirectory(directory string) string {
	// 获取绝对路径
	absPath, err := filepath.Abs(directory)
	if err != nil {
		log.Fatal(err)
	}
	// 分割路径并获取父目录
	parts := strings.Split(absPath, string(os.PathSeparator))
	if len(parts) < 2 {
		return "" // 根目录或无效路径
	}
	return parts[len(parts)-2]
}

// GetDepthOnePathsAndFilesIncludeExt 只获取 当前文件夹 匹配后缀名 的 文件和文件夹
// @param ext 示例：.go .jpg
func GetDepthOnePathsAndFilesIncludeExt(dirPath string, exts ...string) (dirPaths []string, filePaths []string, err error) {
	// 处理要过滤的后缀名
	var ext string
	if len(exts) > 0 {
		ext = path.Ext(exts[0])
		if ext == "" {
			err = errors.New("ext format incorrect, ext:" + exts[0])
			return
		}
	}

	// 读取文件和文件夹
	files, err := os.ReadDir(dirPath)

	// 按数字排序
	sort.Sort(byNumber(files))

	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			dirPaths = append(dirPaths, filepath.Join(dirPath, file.Name()))
		} else {
			if ext != "" && strings.ToLower(path.Ext(file.Name())) != strings.ToLower(ext) {
				continue
			}
			filePaths = append(filePaths, filepath.Join(dirPath, file.Name()))
		}
	}
	return
}

// GetAllPaths 获取所有文件和文件夹的路径，包含子文件夹下的文件和文件夹
// @param ext 过滤文件，只获取匹配后缀名的文件，示例：.go
func GetAllPaths(dirPath string, exts ...string) (dirPaths []string, filePaths []string, err error) {
	dirPaths, filePaths, err = GetDepthOnePathsAndFilesIncludeExt(dirPath, exts...)
	if err != nil {
		return
	}

	// 读取子文件夹下文件和文件夹
	for _, dirPath2 := range dirPaths {
		dirPaths2, filePaths2, err := GetAllPaths(dirPath2, exts...)
		if err != nil {
			return nil, nil, err
		}
		dirPaths = append(dirPaths, dirPaths2...)
		filePaths = append(filePaths, filePaths2...)
	}
	return
}

// 遍历文件夹
func GetFilelist(searchDir string) []string {
	fileList := []string{}
	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && f.Name() != ".DS_Store" {
			fileList = append(fileList, path)
		}
		return nil
	})

	return fileList
}

// 遍历文件夹获取文件夹列表Map （不包括空文件夹）
func GetFileMapList(searchDir string, data map[string][]string) map[string][]string {
	// log.SetFlags(log.Lshortfile)
	// log.Println("dir:",searchDir,path.Base(searchDir))
	files, err := os.ReadDir(searchDir)

	// 按数字排序
	sort.Sort(byNumber(files))

	if err != nil {
		log.Println("GetFileMapList error:", err.Error())
		return nil
	}
	for _, putFile := range files {
		if putFile.IsDir() {
			data = GetFileMapList(searchDir+"/"+putFile.Name(), data)
		} else {
			if putFile.Name() == ".DS_Store" {
				continue
			}
			data[searchDir] = append(data[searchDir], searchDir+"/"+putFile.Name())
		}
	}
	return data
}

// 按数字排序
type byNumber []os.DirEntry

func (a byNumber) Len() int      { return len(a) }
func (a byNumber) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byNumber) Less(i, j int) bool {

	iname := a[i].Name()
	jname := a[j].Name()

	re := regexp.MustCompile("[^a-zA-Z0-9]") // 匹配非字母数字字符
	iname = re.ReplaceAllString(iname, "")
	jname = re.ReplaceAllString(jname, "")

	// 自定义排序逻辑，提取文件名中的数字部分进行比较
	reNum := regexp.MustCompile(`\d+`)
	iNum, _ := strconv.Atoi(reNum.FindString(iname))
	jNum, _ := strconv.Atoi(reNum.FindString(jname))

	return iNum < jNum
}
