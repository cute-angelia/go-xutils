package wework

import (
	"fmt"
	"github.com/cute-angelia/go-xutils/third_party/wework/sendMessage"
	"github.com/fastwego/wxwork/corporation"
	apiMessage "github.com/fastwego/wxwork/corporation/apis/message"
	"log"
	"sync"
)

type Component struct {
	config *config
}

var Pools sync.Map

func newComponent(config *config) *Component {
	comp := &Component{}
	comp.config = config

	// init
	comp.initWeWork()

	return comp
}

func (c *Component) String() string {
	return ComponentName
}

// InitWeWork 初始化
func (c *Component) initWeWork() {
	log.Println(ComponentName, "InitWeWork", c.config.CorpId, c.config.AgentId)
	Corp := corporation.New(corporation.Config{
		Corpid: c.config.CorpId,
	})
	App := Corp.NewApp(corporation.AppConfig{
		AgentId: c.config.AgentId,
		Secret:  c.config.Secret,
	})
	App.SetAccessTokenCacheDriver(NewRedis(c.config.redisClient))

	_, loaded := Pools.LoadOrStore(c.config.AgentId, App)
	if loaded {
		log.Println("命中池")
	} else {
		log.Println("初始化")
	}
}

func GetApp(agentId string) (*corporation.App, error) {
	if v, ok := Pools.Load(agentId); ok {
		return v.(*corporation.App), nil
	} else {
		return nil, fmt.Errorf("%s: App不存在,请初始化配置", agentId)
	}
}

// SendMessage 发送消息
// https://developer.work.weixin.qq.com/document/path/90236
func SendMessage(agentId string, msg sendMessage.IMessage) error {
	if weworkApp, err := GetApp(agentId); err != nil {
		return err
	} else {
		if resp, err := apiMessage.Send(weworkApp, msg.GetMessage()); err != nil {
			log.Println("企业微信发送消息错误❌", err)
			return err
		} else {
			log.Println("企业微信发送消息成功", string(resp))
			return nil
		}

	}
}
