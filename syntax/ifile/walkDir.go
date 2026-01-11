package ifile

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

/*
性能：将正则表达式移出 Less 函数，避免排序时成千上万次的重复编译开销。
安全性：在 GetParentDirectory 中使用了 filepath.Clean，防止因为路径末尾带 / 导致获取父目录失败。
规范性：引入了 fs.DirEntry 相关的现代 API（Go 1.16+ 标准），这是 2026 年处理文件系统的首选方式。
健壮性：GetDepthOnePathsAndFilesIncludeExt 增强了对扩展名输入格式（带点或不带点）的容错处理。
*/

// 预编译正则，提升排序性能
var (
	reNonAlpha = regexp.MustCompile("[^a-zA-Z0-9]")
	reNum      = regexp.MustCompile(`\d+`)
)

// GetParentDirectory 修正路径分隔符逻辑
func GetParentDirectory(directory string) string {
	absPath, err := filepath.Abs(directory)
	if err != nil {
		return ""
	}
	parent := filepath.Dir(filepath.Clean(absPath))
	base := filepath.Base(parent)
	if base == "." || base == string(os.PathSeparator) {
		return ""
	}
	return base
}

// GetDepthOnePathsAndFilesIncludeExt 优化排序逻辑
func GetDepthOnePathsAndFilesIncludeExt(dirPath string, exts ...string) (dirPaths []string, filePaths []string, err error) {
	var targetExt string
	if len(exts) > 0 {
		targetExt = strings.ToLower(exts[0])
		if !strings.HasPrefix(targetExt, ".") {
			targetExt = "." + targetExt
		}
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, nil, err
	}

	// 排序
	sort.Sort(byNumber(files))

	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())
		if file.IsDir() {
			dirPaths = append(dirPaths, fullPath)
		} else {
			if targetExt != "" && strings.ToLower(filepath.Ext(file.Name())) != targetExt {
				continue
			}
			if file.Name() == ".DS_Store" {
				continue
			}
			filePaths = append(filePaths, fullPath)
		}
	}
	return
}

// 使用 WalkDir 优化全文遍历
func GetFilelist(searchDir string) []string {
	fileList := []string{}
	// WalkDir 比 Walk 性能更高，因为它不调用 Lstat
	_ = filepath.WalkDir(searchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if d.Name() != ".DS_Store" {
			fileList = append(fileList, path)
		}
		return nil
	})
	return fileList
}

// 优化后的排序逻辑
type byNumber []os.DirEntry

func (a byNumber) Len() int      { return len(a) }
func (a byNumber) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byNumber) Less(i, j int) bool {
	iname := a[i].Name()
	jname := a[j].Name()

	cleanI := reNonAlpha.ReplaceAllString(iname, "")
	cleanJ := reNonAlpha.ReplaceAllString(jname, "")

	iNumStr := reNum.FindString(cleanI)
	jNumStr := reNum.FindString(cleanJ)

	if iNumStr != "" && jNumStr != "" {
		// 先尝试按数值比较
		iNum, errI := strconv.ParseUint(iNumStr, 10, 64)
		jNum, errJ := strconv.ParseUint(jNumStr, 10, 64)

		if errI == nil && errJ == nil {
			if iNum != jNum {
				return iNum < jNum
			}
		} else {
			// 如果数字太大溢出了，按长度和字符串比较
			if len(iNumStr) != len(jNumStr) {
				return len(iNumStr) < len(jNumStr)
			}
			if iNumStr != jNumStr {
				return iNumStr < jNumStr
			}
		}
	}

	// 兜底：原始文件名排序
	return iname < jname
}
