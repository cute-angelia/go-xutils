package oss

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cute-angelia/go-xutils/components/idownload"
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"io"
	"log"
	"net/url"
	"path"
	"strings"
	"time"
)

const PackageName = "components.aliyun-oss"

type Component struct {
	config *config
	Client *oss.Bucket
}

// newComponent ...
func newComponent(config *config) *Component {
	client, err := oss.New(config.Endpoint, config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		log.Println(err.Error())
	}
	bucket, err := client.Bucket(config.BucketName)
	if err != nil {
		log.Println("发生错误" + err.Error())
	}
	return &Component{
		config: config,
		Client: bucket,
	}
}

// GetClient
func (e Component) GetClient() *oss.Bucket {
	return e.Client
}

// 从完整链接中，获取 objKey
func (e Component) GetObjectKeyByUrl(uri string) string {
	if up, err := url.Parse(uri); err != nil {
		return uri
	} else {
		if path.IsAbs(up.Path) {
			return up.Path[1:]
		} else {
			return up.Path
		}
	}
}

// 拼接完整路径
func (e Component) JoinUrl(objectKey string) string {
	e.config.BucketHost = strings.TrimRight(e.config.BucketHost, "/")
	return fmt.Sprintf("%s/%s", e.config.BucketHost, e.CleanObjKey(objectKey))
}

func (e Component) CleanObjKey(objectKey string) string {
	if path.IsAbs(objectKey) {
		return objectKey[1:]
	} else {
		return objectKey
	}
}

// 上传文件
func (e Component) PutObject(objectNameIn string, reader io.Reader) (string, error) {
	return e.CleanObjKey(objectNameIn), e.Client.PutObject(objectNameIn, reader)
}

// 按文件上传
func (e Component) FPutObject(objectNameIn string, filePath string) (string, error) {
	return e.CleanObjKey(objectNameIn), e.Client.PutObjectFromFile(objectNameIn, filePath)
}

// PutObjectWithSrc 提供链接，上传, 返回key， filehash, error
func (e Component) PutObjectWithSrc(uri string, objectName string) (string, string, error) {
	// iweb 不处理
	if !strings.Contains(uri, "iweb") {
		return "", "", errors.New("链接提供不正确：" + uri)
	}
	// 更换图片到本地
	idown := idownload.New(
		idownload.WithProxySocks5(e.config.ProxySocks5),
		idownload.WithDebug(e.config.Debug),
		idownload.WithTimeout(time.Second*20),
	)
	if filebyte, err := idown.DownloadToByte(uri); err != nil {
		log.Println(PackageName, "获取图片失败：❌", err)
		return "", "", errors.New("获取图片失败：❌：" + uri)
	} else {
		// 打印日志
		if e.config.Debug {
			log.Printf(PackageName+"获取图片: %s, 代理：%s", uri, e.config.ProxySocks5)
		}

		if p, err := e.PutObject(objectName, bytes.NewReader(filebyte)); err != nil {
			log.Println(PackageName, "上传失败：❌", err, uri)
			return "", "", errors.New("获取图片失败：❌：" + uri)
		} else {
			log.Printf(PackageName+"上传成功：✅ %s => %s", uri, objectName)
			return p, ifile.FileHashSHA1(bytes.NewReader(filebyte)), nil
		}
	}
}

// PutObjectWithBase64 上传 - base64
func (e Component) PutObjectWithBase64(objectNameIn string, base64File string) (string, error) {
	b64data := base64File[strings.IndexByte(base64File, ',')+1:]
	if decode, err := base64.StdEncoding.DecodeString(b64data); err == nil {
		body := bytes.NewReader(decode)
		return e.PutObject(objectNameIn, body)
	} else {
		return "", err
	}
}

// 删除文件
func (e Component) DeleteObject(filePath string) error {
	return e.Client.DeleteObject(filePath)
}
