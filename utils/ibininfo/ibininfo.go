package ibininfo

import (
	"fmt"
	"runtime"
	"strings"
)

/*
# 获取信息
TAG=$(git describe --tags --always)
COMMIT=$(git log -1 --format='%h %s' | sed "s/'/\"/g")
STATUS=$(git status --porcelain)
TIME=$(date "+%Y-%m-%d %H:%M:%S")
GVER=$(go version)

# 注入并编译
go build -ldflags "-X 'ibininfo.GitTag=$TAG' \
                   -X 'ibininfo.GitCommitLog=$COMMIT' \
                   -X 'ibininfo.GitStatus=$STATUS' \
                   -X 'ibininfo.BuildTime=$TIME' \
                   -X 'ibininfo.BuildGoVersion=$GVER'" -o main .

*/

var (
	// 初始化为 unknown，如果编译时没有传入这些值，则为 unknown
	GitTag         = "unknown"
	GitCommitLog   = "unknown"
	GitStatus      = "unknown"
	BuildTime      = "unknown"
	BuildGoVersion = "unknown"
	Version        = "unknown"
)

// 返回单行格式
func StringifySingleLine() string {
	return fmt.Sprintf("Version=%s. GitTag=%s. GitCommitLog=%s. GitStatus=%s. BuildTime=%s. GoVersion=%s. runtime=%s/%s.",
		Version,
		GitTag, GitCommitLog, GitStatus, BuildTime, BuildGoVersion, runtime.GOOS, runtime.GOARCH)
}

// 返回多行格式
func StringifyMultiLine() string {
	return fmt.Sprintf("Version=%s\nGitTag=%s\nGitCommitLog=%s\nGitStatus=%s\nBuildTime=%s\nGoVersion=%s\nruntime=%s/%s\n",
		Version,
		GitTag, GitCommitLog, GitStatus, BuildTime, BuildGoVersion, runtime.GOOS, runtime.GOARCH)
}

// 对一些值做美化处理
func beauty() {
	if GitStatus == "" {
		// GitStatus 为空时，说明本地源码与最近的 commit 记录一致，无修改
		// 为它赋一个特殊值
		GitStatus = "cleanly"
	} else {
		// 将多行结果合并为一行
		replacer := strings.NewReplacer("\r\n", " |", "\n", " |")
		GitStatus = replacer.Replace(GitStatus)
	}
}

func init() {
	if BuildGoVersion == "unknown" {
		BuildGoVersion = runtime.Version()
	}
	beauty()
}

// 增加一个返回 map[string]string 的方法，方便集成到 HTTP 健康检查接口（JSON 格式）中：
func ToMap() map[string]string {
	return map[string]string{
		"version":    Version,
		"git_tag":    GitTag,
		"git_commit": GitCommitLog,
		"git_status": GitStatus,
		"build_time": BuildTime,
		"go_version": BuildGoVersion,
		"os_arch":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
