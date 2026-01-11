package store

import (
	"context"
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
	"io"
	"log"
	"net/http"
	"os"
)

type Qiniu struct {
	Ak            string
	Sk            string
	Bucket        string
	Prefix        string
	Cmd           string
	FileType      int
	Zone          string
	BucketManager *storage.BucketManager
	QiniuConfig   storage.Config
	Domain        string
}

func NewQiNiu(ak, sk, bucket string, zone string) *Qiniu {
	mac := auth.New(ak, sk)
	cfg := storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS:      false,
		UseCdnDomains: false,
	}
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	if zone == "" {
		cfg.Zone = &storage.ZoneHuanan
	}

	bucketManager := storage.NewBucketManager(mac, &cfg)

	return &Qiniu{
		Ak:            ak,
		Sk:            sk,
		Bucket:        bucket,
		BucketManager: bucketManager,
		QiniuConfig:   cfg,
	}
}

func (self *Qiniu) SetConfig(prefix, cmd string, fileType int) *Qiniu {
	self.Prefix = prefix
	self.Cmd = cmd
	self.FileType = fileType
	return self
}

func (self *Qiniu) UploadByLocalFile(path string, key string) (string, error) {
	putPolicy := storage.PutPolicy{
		Scope: self.Bucket + ":" + key,
	}

	mac := auth.New(self.Ak, self.Sk)
	upToken := putPolicy.UploadToken(mac)

	//设置代理
	// proxyURL := "http://localhost:8888"
	// proxyURI, _ := url.Parse(proxyURL)

	//构建代理client对象
	client := http.Client{
		Transport: &http.Transport{
			// Proxy: http.ProxyURL(proxyURI),
		},
	}

	// 构建表单上传的对象
	formUploader := storage.NewFormUploaderEx(&self.QiniuConfig, &storage.Client{Client: &client})
	ret := storage.PutRet{}
	// 可选配置
	//putExtra := storage.PutExtra{
	//	Params: map[string]string{
	//		"x:name": "github logo",
	//	},
	//}
	//putExtra.NoCrc32Check = true

	if err := formUploader.PutFile(context.Background(), &ret, upToken, key, path, &storage.PutExtra{}); err != nil {
		log.Println(err)
		return "", err
	}
	// log.Println(ret.Key, ret.Hash)

	// 上传成功后删除文件
	os.Remove(path)

	log.Println("七牛上传成功：", path, ret.Key)
	return ret.Key, nil
}

func (self *Qiniu) UploadByForm(f io.Reader, filesize int64, key string) (string, error) {
	putPolicy := storage.PutPolicy{
		Scope: self.Bucket + ":" + key,
	}

	mac := auth.New(self.Ak, self.Sk)
	upToken := putPolicy.UploadToken(mac)

	//设置代理
	// proxyURL := "http://localhost:8888"
	// proxyURI, _ := url.Parse(proxyURL)

	//构建代理client对象
	client := http.Client{
		Transport: &http.Transport{
			// Proxy: http.ProxyURL(proxyURI),
		},
	}

	// 构建表单上传的对象
	formUploader := storage.NewFormUploaderEx(&self.QiniuConfig, &storage.Client{Client: &client})
	ret := storage.PutRet{}
	// 可选配置
	//putExtra := storage.PutExtra{
	//	Params: map[string]string{
	//		"x:name": "github logo",
	//	},
	//}
	//putExtra.NoCrc32Check = true
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(f)

	if err := formUploader.Put(context.Background(), &ret, upToken, key, f, filesize, &storage.PutExtra{}); err != nil {
		log.Println(err)
		return "", err
	}
	// log.Println(ret.Key, ret.Hash)

	// 上传成功后删除文件
	log.Println("七牛上传成功：", ret.Key)
	return ret.Key, nil
}

//func (self *Qiniu) UploadByUrl(url string, key string) (string, error) {
//	body, err := ifile.GetFileWithSrc(url)
//	if err != nil {
//		return "", err
//	}
//
//	putPolicy := storage.PutPolicy{
//		Scope: self.Bucket + ":" + key,
//	}
//
//	mac := auth.New(self.Ak, self.Sk)
//	upToken := putPolicy.UploadToken(mac)
//
//	//设置代理
//	// proxyURL := "http://localhost:8888"
//	// proxyURI, _ := url.Parse(proxyURL)
//
//	//构建代理client对象
//	client := http.Client{
//		Transport: &http.Transport{
//			// Proxy: http.ProxyURL(proxyURI),
//		},
//	}
//
//	// 构建表单上传的对象
//	formUploader := storage.NewFormUploaderEx(&self.QiniuConfig, &storage.Client{Client: &client})
//	ret := storage.PutRet{}
//	// 可选配置
//	//putExtra := storage.PutExtra{
//	//	Params: map[string]string{
//	//		"x:name": "github logo",
//	//	},
//	//}
//	//putExtra.NoCrc32Check = true
//
//	data := bytes.NewReader(body)
//	dataLen := int64(len(body))
//
//	if err := formUploader.Put(context.Background(), &ret, upToken, key, data, dataLen, &storage.PutExtra{}); err != nil {
//		log.Println(err)
//		return "", err
//	}
//	log.Println("七牛上传成功：", url, ret.Key)
//	return ret.Key, nil
//}

// 列出 keys
func (self *Qiniu) ListKeys() {
	limit := 1000
	prefix := self.Prefix
	delimiter := ""
	//初始列举marker为空
	marker := ""

	for {
		keys := []string{}

		entries, _, nextMarker, hashNext, err := self.BucketManager.ListFiles(self.Bucket, prefix, delimiter, marker, limit)
		if err != nil {
			log.Println("list error,", err)
			break
		}
		//print entries
		for _, entry := range entries {
			log.Println(entry.Key)
			keys = append(keys, entry.Key)
		}

		// dosomething
		// 处理
		switch self.Cmd {
		case "delete":
			self.DeleteKeys(keys)
		case "changeType":
			self.ChangeType(keys)
		}

		if hashNext {
			marker = nextMarker
		} else {
			//list end
			break
		}
	}
}

// 删除
func (self *Qiniu) DeleteKeys(keys []string) {
	deleteOps := make([]string, 0, len(keys))
	for _, key := range keys {
		deleteOps = append(deleteOps, storage.URIDelete(self.Bucket, key))
	}
	rets, err := self.BucketManager.Batch(deleteOps)
	if err != nil {
		// 遇到错误
		if _, ok := err.(*storage.ErrorInfo); ok {
			for _, ret := range rets {
				log.Println(ret.Code, ret.Data)
			}
		} else {
			fmt.Printf("batch error, %s", err)
		}
	} else {
		// 完全成功
		for _, ret := range rets {
			log.Println(ret.Code, ret.Data)
		}
	}
}

func (self *Qiniu) Delete(key string) error {
	return self.BucketManager.Delete(self.Bucket, key)
}

// 更改类型
func (self *Qiniu) ChangeType(keys []string) {
	fileType := self.FileType // 0 表示普通存储，1表示低频存储 2归档
	for _, key := range keys {
		err := self.BucketManager.ChangeType(self.Bucket, key, fileType)
		if err != nil {
			log.Println(key, err)
			continue
		}
	}
}
