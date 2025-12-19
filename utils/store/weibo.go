package store

import (
	"encoding/base64"
	"fmt"
	"github.com/cute-angelia/go-xutils/components/caches/mem"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crypto/tls"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/pkg/errors"
	"hash/crc32"
	"log"
	"regexp"
)

type Weibo struct {
	FileLimit []string
	MaxSize   int

	Config SinaConfig
}

// Weibo
type SinaConfig struct {
	Status bool `json:"status"`
	//用户名
	UserName string `json:"user_name"`
	//密码
	PassWord string `json:"pass_word"`
	//新浪 Cookie 更新的频率,默认为3600s ,单位 s
	ResetSinaCookieTime int `json:"reset_sina_cookie_time"`
	//新浪图床默认使用的尺寸大小 square,thumb150,orj360,orj480,mw690,mw1024,mw2048,small,bmiddle,large 、默认为large
	DefultPicSize string `json:"defult_pic_size"`

	Proxy ProxyConf `json:"proxy"`
}

type SinaError struct {
	Retcode string `json:"retcode"`
	Reason  string `json:"reason"`
}

// Sina 图床 json
type SinaMsg struct {
	Code string   `json:"code"`
	Data SinaData `json:"data"`
}

type SinaData struct {
	Count int      `json:"count"`
	Data  string   `json:"data"`
	Pics  SinaPics `json:"pics"`
}

type SinaPics struct {
	Pic_1 picInfo `json:"pic_1"`
}

type picInfo struct {
	Width  int    `json:"width"`
	Size   int    `json:"size"`
	Ret    int    `json:"ret"`
	Height int    `json:"height"`
	Name   string `json:"name"`
	Pid    string `json:"pid"`
}

var picType = []string{"png", "jpg", "jpeg", "gif", "bmp"}
var memcache = mem.NewLRU(10, time.Hour)

// 实例对象
var WeiboStore *Weibo

func NewWeibo(config SinaConfig) *Weibo {

	size := "large"
	if len(config.DefultPicSize) > 0 {
		size = config.DefultPicSize
	}

	config.DefultPicSize = size

	return &Weibo{
		Config: config,
	}
}

func (s *Weibo) Upload(image *ImageParam) (ImageReturn, error) {
	var sinaAccount = s.Config
	if sinaAccount.PassWord == "" || sinaAccount.UserName == "" {
		err := errors.New("Sina Account is null")
		return ImageReturn{}, err
	}

	durl := "https://picupload.service.weibo.com/interface/pic_upload.php?mime=image%2Fjpeg&data=base64&url=0&markpos=1&logo=&nick=0&marks=1&app=miniblog&cb=http://weibo.com/aj/static/upimgback.html?_wv=5&callback=STK_ijax_1111"

	imgStr := base64.StdEncoding.EncodeToString(*image.Content)
	//构造 http 请求
	postData := make(url.Values)
	postData["b64_data"] = []string{imgStr}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}
	request, err := http.NewRequest("POST", durl, strings.NewReader(postData.Encode()))
	if err != nil {
		log.Println(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//设置 cookie
	unCookies := s.Login(sinaAccount.UserName, sinaAccount.PassWord)

	// log.Println("unCookies", unCookies)

	//需要进行断言转换
	cookies, ok := unCookies.([]*http.Cookie)
	if !ok {
		panic(ok)
	}
	for _, value := range cookies {
		request.AddCookie(value)
	}
	resp, err := client.Do(request)

	if err != nil {
		log.Println("resp, err := client.Do(request)", err)
		return ImageReturn{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll(resp.Body)", err)
		return ImageReturn{}, err
	}
	var proxy = ""
	if sinaAccount.Proxy.Status {
		proxy = sinaAccount.Proxy.Node
	}
	sinaUrl := s.getSinaUrl(body, image.Type)
	if sinaUrl != "" {
		sinaUrl = proxy + sinaUrl
	}
	return ImageReturn{
		Url: sinaUrl,
		ID:  2,
	}, nil

}

// 新浪图床登录
func (s *Weibo) Login(name string, pass string) interface{} {
	uri := "https://login.sina.com.cn/sso/login.php?client=ssologin.js(v1.4.15)&_=1403138799543"
	userInfo := make(map[string]string)
	userInfo["UserName"] = base64.StdEncoding.EncodeToString([]byte(name))
	userInfo["PassWord"] = pass
	cookie := s.getCookies(uri, userInfo)
	return cookie
}

// 获取新浪图床 Cookie
func (s *Weibo) getCookies(durl string, data map[string]string) interface{} {
	//尝试从缓存里面获取 Cookie
	if memcache.GetInterface("SinaCookies") != nil {
		return memcache.GetInterface("SinaCookies")
	}

	postData := make(url.Values)
	postData["entry"] = []string{"sso"}
	postData["gateway"] = []string{"1"}
	postData["from"] = []string{"null"}
	postData["savestate"] = []string{"30"}
	postData["uAddicket"] = []string{"0"}
	postData["pagerefer"] = []string{""}
	postData["vsnf"] = []string{"1"}
	postData["su"] = []string{data["UserName"]} //UserName
	postData["service"] = []string{"sso"}
	postData["sp"] = []string{data["PassWord"]} //PassWord
	postData["sr"] = []string{"1920*1080"}
	postData["encoding"] = []string{"UTF-8"}
	postData["cdult"] = []string{"3"}
	postData["domain"] = []string{"sina.com.cn"}
	postData["prelt"] = []string{"0"}
	postData["returntype"] = []string{"TEXT"}
	client := &http.Client{}
	request, err := http.NewRequest("POST", durl, strings.NewReader(postData.Encode()))
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		log.Print("err:", err.Error())
		return ""
	}
	body, _ := ioutil.ReadAll(resp.Body)
	sinaError := SinaError{}
	err = json.Unmarshal(body, &sinaError)
	if err != nil {
		log.Print("err:", err.Error())
		return ""
	}
	if sinaError.Retcode == "101" {
		logs.Alert("新浪图床上传错误:" + sinaError.Reason)
	}
	defer resp.Body.Close()
	cookie := resp.Cookies()
	//缓存 Cookie 缓存一个小时
	memcache.PutInterface("SinaCookies", cookie)
	return cookie
}

// 获取 Sina 图床 URL
func (s *Weibo) getSinaUrl(body []byte, imgType string) string {
	var sinaAccount = s.Config

	str := string(body)

	//正则获取
	pat := "({.*)"
	res := regexp.MustCompile(pat)
	jsons := res.FindAllString(str, -1)

	if jsons == nil {
		return ""
	}

	msg := SinaMsg{}
	json.Unmarshal([]byte(jsons[0]), &msg)

	//验证 pid 的合法性
	pid := msg.Data.Pics.Pic_1.Pid

	sinaUrl := s.CheckPid(pid, imgType, sinaAccount.DefultPicSize)
	if sinaUrl == "" {
		return ""
	}
	return sinaUrl
}

func (s *Weibo) CheckPid(pid string, imgType string, size string) string {
	if pid == "" {
		return ""
	}
	sinaNumber := fmt.Sprint((crc32.ChecksumIEEE([]byte(pid)) & 3) + 1)
	sinaUrl := "https://ww" + sinaNumber + ".sinaimg.cn/" + size + "/" + pid + imgType
	return sinaUrl
}
