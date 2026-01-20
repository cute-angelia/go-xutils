package ifile

import (
	"io/fs"
	"log"
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

// GetAllPaths 递归获取 basePath 下的所有文件夹和文件
// 专门为 CleanDir 设计：文件夹列表会按深度反转，确保先处理子目录
func GetAllPaths(basePath string) (dirs []string, files []string, err error) {
	err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == basePath {
			return nil
		}

		if d.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
		return nil
	})

	// 重要：反转目录顺序
	// 这样 dirs 的顺序会从 [a, a/b, a/b/c] 变成 [a/b/c, a/b, a]
	// 确保 CleanDir 循环时先删最深层的空目录
	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	return dirs, files, err
}

// GetDepthOnePathsAndFilesIncludeExt 优化排序逻辑
// 函数更适合用于 UI 文件列表展示（因为它有排序和后缀过滤）
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

// GetFileMapList 递归获取文件夹及其文件列表的映射（不包括空文件夹）
func GetFileMapList(searchDir string, data map[string][]string) map[string][]string {
	// 确保 map 已初始化
	if data == nil {
		data = make(map[string][]string)
	}

	files, err := os.ReadDir(searchDir)
	if err != nil {
		log.Printf("GetFileMapList 错误 [%s]: %v\n", searchDir, err)
		return data
	}

	// 使用你定义的 byNumber 进行排序
	sort.Sort(byNumber(files))

	for _, file := range files {
		// 使用 filepath.Join 适配不同操作系统的路径符
		fullPath := filepath.Join(searchDir, file.Name())

		if file.IsDir() {
			// 递归处理子目录
			data = GetFileMapList(fullPath, data)
		} else {
			// 排除系统隐藏文件
			if file.Name() == ".DS_Store" || file.Name() == "Thumbs.db" {
				continue
			}
			// 将文件路径加入当前目录的 key 中
			data[searchDir] = append(data[searchDir], fullPath)
		}
	}
	return data
}
