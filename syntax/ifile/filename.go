package ifile

import (
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/irandom"
	"github.com/cute-angelia/go-xutils/syntax/isnowflake"
	"path"
	"strconv"
	"strings"
	"time"
)

type fileName struct {
	uri    string
	suffix string
	prefix string
	ext    string
}

func NewFileName(uri string) *fileName {
	return &fileName{
		uri: uri,
	}
}

// SetSuffix 后缀
func (f *fileName) SetSuffix(suffix string) *fileName {
	f.suffix = suffix
	return f
}

// SetPrefix 前缀
func (f *fileName) SetPrefix(prefix string) *fileName {
	f.prefix = prefix
	return f
}

// SetExt 自定义 ext
func (f *fileName) SetExt(ext string) *fileName {
	if strings.Contains(ext, "?") {
		ext, _, _ = strings.Cut(ext, "?")
	}
	if len(ext) == 0 {
		ext = f.GetExt()
	}
	f.ext = ext
	return f
}

// GetExt 获取后缀，带点 如：.jpg
func (f *fileName) GetExt() string {
	if len(f.ext) > 0 {
		return f.ext
	}
	cleanz := f.CleanUrl()
	ext := path.Ext(cleanz)
	return strings.ToLower(ext)
}

func (f fileName) GetDir() string {
	return path.Dir(f.uri)
}

func (f fileName) CleanUrl() string {
	newName := f.uri
	if i := strings.Index(newName, "?"); i != -1 {
		newName = newName[:i]
	}
	// path.Clean 可以去除路径中的 ../ 或多余的斜杠
	return path.Clean(newName)
}

// name 保持原有名字
func (f fileName) GetNameOrigin() string {
	uri := f.CleanUrl()
	return fmt.Sprintf("%s%s%s%s", f.prefix, NameNoExt(uri), f.suffix, f.GetExt())
}

// name 按时间戳
func (f fileName) GetNameTimeline() string {
	return fmt.Sprintf("%s%d%s%s", f.prefix, time.Now().UnixNano(), f.suffix, f.GetExt())
}

// name 按时间戳 反序：minio 文件获取按文件名排序，需要反序时间戳
// 算法：未来时间减去当前时间， 为了防止串号，增加 nano 长度
func (f fileName) GetNameTimelineReverse(withDate bool) string {
	ext := f.GetExt()
	// 雪花算法
	id, _ := isnowflake.GetSnowflake().NextID()
	// 使用足够大的常数，或使用 2026 年推荐的基准：1<<63 - 1
	const maxUint64 = 18446744073709551615
	diffTime := maxUint64 - id

	respName := ""
	if withDate {
		respName = fmt.Sprintf("%d_%s", diffTime, time.Now().Format("20060102150405"))
	}

	return fmt.Sprintf("%s%s%s%s", f.prefix, respName, f.suffix, ext)
}

// name 按时间格式
func (f fileName) GetNameTimeDate() string {
	dname := time.Now().Format("20060102-150405") + "-" + irandom.RandString(6, irandom.LetterNum)
	return fmt.Sprintf("%s%s%s%s", f.prefix, dname, f.suffix, f.GetExt())
}

// name 按雪花算法
func (f fileName) GetNameSnowFlow() string {
	id, _ := isnowflake.GetSnowflake().NextID()
	// 使用 strconv 转换数字，使用 + 拼接字符串
	return f.prefix + strconv.FormatUint(id, 10) + f.suffix + f.GetExt()
}
