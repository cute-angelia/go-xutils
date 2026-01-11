package weworkrobot

import (
	"github.com/guonaihong/gout"
	"log"
	"time"
)

type Component struct {
	config *config
}

func newComponent(cfg *config) *Component {
	return &Component{
		config: cfg,
	}
}

func (c *Component) generateContent(content string) string {
	// 2026 实践：使用结构体指针接收器 (c *Component) 减少拷贝开销
	res := ""
	if c.config.WithTime {
		res += time.Now().Format("2006-01-02 15:04:05") + " "
	}

	if len(c.config.From) > 0 {
		res += "[" + c.config.From + "] "
	}

	if len(c.config.Topic) > 0 {
		res += "[" + c.config.Topic + "] "
	}

	return res + content
}

func (c *Component) SendText(content string) error {
	fullContent := c.generateContent(content)

	// 企业微信文本消息支持艾特
	return gout.POST(c.config.Uri).SetJSON(gout.H{
		"msgtype": "text",
		"text": gout.H{
			"content":               fullContent,
			"mentioned_list":        c.config.MentionedList,
			"mentioned_mobile_list": c.config.MentionedMobileList,
		},
	}).Debug(c.config.Debug).F().Retry().Attempt(c.config.Retry).WaitTime(time.Second).Do()
}

func (c *Component) SendMarkDown(content string) error {
	fullContent := c.generateContent(content)

	// 2026 提醒：企业微信 Markdown 消息体里不支持 mentioned_list 字段
	// 如果需要艾特，请在 content 中加入 <@userid> 或 <@all>
	return gout.POST(c.config.Uri).SetJSON(gout.H{
		"msgtype": "markdown",
		"markdown": gout.H{
			"content": fullContent,
		},
	}).Debug(c.config.Debug).F().Retry().Attempt(c.config.Retry).WaitTime(time.Second).Do()
}

func logError(key string, err error) {
	// 假设 PackageName 是包内定义的常量
	log.Printf("[%s] error at %s: %v", "weworkrobot", key, err)
}
