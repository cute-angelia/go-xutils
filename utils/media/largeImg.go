package media

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// GetLargeImg 获取大图
func GetLargeImg(imgURL string) string {
	if imgURL != "" {
		// Weibo
		if strings.Contains(imgURL, "sinaimg.cn") {
			u, _ := url.Parse(imgURL)
			parts := strings.Split(u.Path, "/")
			parts[1] = "large"
			imgURL = u.Scheme + "://" + u.Host + strings.Join(parts, "/")
		}

		// Twitter
		if strings.Contains(imgURL, "twimg.com") {
			// 移除查询参数
			if queryPos := strings.Index(imgURL, "?"); queryPos != -1 {
				imgURL = imgURL[:queryPos]
			}

			// 获取扩展名（包含点号）
			ext := path.Ext(imgURL)
			if ext == "" {
				// 没有扩展名时添加默认处理
				ext = ".jpg"
			}

			// 确保格式参数正确添加
			imgURL = fmt.Sprintf("%s?format=%s&name=orig", imgURL, strings.ToLower(strings.TrimPrefix(ext, ".")))
		}

	}

	return imgURL
}
