package iminio

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/cute-angelia/go-xutils/components/idownload"
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"github.com/cute-angelia/go-xutils/utils/generator/hash"
	progress "github.com/markity/minio-progress"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type Component struct {
	name   string
	config *config
	locker sync.Mutex
	Client *minio.Client
}

// newComponent ...
func newComponent(config *config) *Component {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccesskeyId, config.SecretaccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		log.Println("发生错误" + err.Error())
	}
	return &Component{
		name:   PackageName,
		config: config,
		Client: minioClient,
	}
}

func (e *Component) GetUrl(bucket, key string, opts ...UrlOption) string {
	if key == "" {
		log.Println("errors.New( key is empty )")
		return ""
	}

	// 默认配置
	urlOpt := &UrlOptions{
		Context: context.Background(),
	}
	for _, opt := range opts {
		opt(urlOpt)
	}

	// 2. 自動路徑處理：處理 bucket 和 key 連在一起的情況
	if bucket == "" {
		fullPath := strings.Trim(key, "/")
		parts := strings.SplitN(fullPath, "/", 2)
		if len(parts) == 2 {
			bucket = parts[0]
			key = parts[1]
		}
	}

	var finalUrl string

	// 1. 处理缓存逻辑
	var hashkey string
	if urlOpt.Cache != nil && !urlOpt.Rebuild {
		hashkey = e.GenerateHashKey(1, bucket, key, urlOpt.Expiry, urlOpt.Version)
		if cachedData, err := urlOpt.Cache.Get(hashkey); err == nil && len(cachedData) > 0 {
			finalUrl = cachedData
		}
	}

	// 2. 生成基础 URL (如果缓存没中)
	if finalUrl == "" {
		if urlOpt.Expiry > 0 {
			// --- 修改点：将版本号纳入签名计算 ---
			var reqParams url.Values
			if urlOpt.Version != "" {
				reqParams = make(url.Values)
				// 将 v=xxx 放入参数，MinIO 签名时会将其包含进去
				reqParams.Set("v", fmt.Sprintf("%d", urlOpt.Version))
			}

			// 需要签名的私有地址
			presignedURL, err := e.Client.PresignedGetObject(urlOpt.Context, bucket, key, urlOpt.Expiry, nil)
			if err != nil {
				log.Println("minio 簽名失敗 ", err)
				return ""
			}
			finalUrl = presignedURL.String()
			// 既然已经包含在签名里了，清空 Version 避免最后重复拼接
			urlOpt.Version = ""
		} else {
			// 公共拼接地址
			baseUrl := strings.TrimSuffix(e.Client.EndpointURL().String(), "/")
			finalUrl = baseUrl + "/" + path.Join(bucket, key)
		}

		// 3. 写入缓存
		if urlOpt.Cache != nil && finalUrl != "" {
			// 缓存时间略短于签名时间
			ttl := urlOpt.Expiry
			if ttl > 5*time.Minute {
				ttl -= 5 * time.Minute
			}
			urlOpt.Cache.Set(hashkey, finalUrl, ttl)
		}
	}

	// 4. 处理版本号（仅针对公共地址，私有地址在上面已处理并清空）
	if urlOpt.Version != "" {
		connector := "?"
		if strings.Contains(finalUrl, "?") {
			connector = "&"
		}
		finalUrl = fmt.Sprintf("%s%sv=%s", finalUrl, connector, urlOpt.Version)
	}

	return finalUrl
}

// GenerateHashKey 支持傳入任意數量的參數來生成唯一哈希
func (e *Component) GenerateHashKey(bucketType int32, bucket string, prefix string, extras ...interface{}) string {
	// 基礎組分
	base := fmt.Sprintf("%d:%s:%s", bucketType, bucket, prefix)

	// 處理額外參數（如 Expiry, Version 等）
	if len(extras) > 0 {
		for _, extra := range extras {
			base += fmt.Sprintf(":%v", extra)
		}
	}

	return hash.NewEncodeMD5(base)
}

// GetObjectsByPage minio 获取分页对象数据
// 1.分页
// 2.可以指定文件后缀获取
func (e *Component) GetObjectsByPage(bucket string, prefix string, page int32, perpage int32, fileExt []string) (objs []string, notall bool) {
	// 控制流程
	count := int32(0)
	offset := (page - 1) * perpage

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opt := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}
	objectCh := e.Client.ListObjects(ctx, bucket, opt)

	// 后缀扩展
	extMap := sync.Map{}
	for _, str := range fileExt {
		extMap.Store(str, true)
	}

	for object := range objectCh {
		if object.Err == nil {
			// 名称
			objkeyname := path.Base(object.Key)

			// 内置删除文件
			if objkeyname == ".DS_Store" {
				e.Client.RemoveObject(context.Background(), bucket, object.Key, minio.RemoveObjectOptions{})
				continue
			}

			// 处理指定后缀文件
			if len(fileExt) > 0 {
				if _, ok := extMap.Load(strings.ToLower(path.Ext(objkeyname))); !ok {
					continue
				}
			}

			// log.Printf("---->1 count:%d, offset:%d, perpage:%d, %v", count, offset, perpage, count >= offset)
			// 小于当前游标
			if count >= offset {
				// 当前计数 - 游标
				// log.Printf("<---- count:%d, offset:%d, perpage:%d, %v false:继续", count, offset, perpage, count-offset >= perpage)
				if count-offset >= perpage {
					notall = true
					cancel()
					break
				}
				objs = append(objs, bucket+"/"+object.Key)
				count++
			} else {
				count++
			}
		} else {
			log.Println("object.Err", object.Err)
		}
	}
	return objs, notall
}

// GetObjectStat Objects 状态
func (e *Component) GetObjectStat(bucket string, objectName string) (minio.ObjectInfo, error) {
	objInfo, err := e.Client.StatObject(context.Background(), bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		log.Println(err)
	}
	return objInfo, err
}

func (e *Component) CheckMode(objectName string) (newObjectName string, canupload bool) {
	// 跳过
	if e.config.ReplaceMode == ReplaceModeIgnore {
		canupload = false
		newObjectName = objectName
	}
	if e.config.ReplaceMode == ReplaceModeReplace {
		canupload = true
		newObjectName = objectName
	}
	if e.config.ReplaceMode == ReplaceModeTwo {
		rand.Seed(time.Now().Unix())
		canupload = true
		newObjectName = fmt.Sprintf("bak_%d_%s", rand.Intn(100), objectName)
	}
	return
}

// CopyObject 复制对象
func (e *Component) CopyObject(dst minio.CopyDestOptions, src minio.CopySrcOptions) (uploadInfo minio.UploadInfo, err error) {
	uploadInfo, err = e.Client.CopyObject(context.Background(), dst, src)
	return
}

// PutObject 上传-按读取文件数据
func (e *Component) PutObject(bucket string, objectNameIn string, reader io.Reader, objectSize int64, objopt minio.PutObjectOptions) (minio.UploadInfo, error) {
	if objectName, ok := e.CheckMode(objectNameIn); ok {
		objectName = strings.Replace(objectName, "//", "/", -1)

		// 创建上传进度条对象
		objopt.Progress = progress.NewUploadProgress(objectSize)

		uploadInfo, err := e.Client.PutObject(context.Background(), bucket, objectName, reader, objectSize, objopt)
		if err != nil {
			log.Println(bucket, objectNameIn, err)
			return uploadInfo, err
		}
		if e.config.Debug {
			log.Println("Successfully uploaded bytes: ", uploadInfo)
		}
		return uploadInfo, err
	} else {
		return minio.UploadInfo{}, fmt.Errorf("模式未设置 %s", objectNameIn)
	}
}

// FPutObject 上传-按存在文件
func (e *Component) FPutObject(bucket string, objectNameIn string, filePath string, objopt minio.PutObjectOptions) (minio.UploadInfo, error) {
	if objectName, ok := e.CheckMode(objectNameIn); ok {

		// 打开文件
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0444)
		defer file.Close()
		if err != nil {
			log.Fatalf("打开文件失败:%v\n", err)
		}
		// 获取文件大小
		fileInfo, err := file.Stat()
		if err != nil {
			log.Fatalf("获取文件信息失败:%v\n", err)
		}
		tempFileSize := fileInfo.Size()

		// 创建上传进度条对象
		objopt.Progress = progress.NewUploadProgress(tempFileSize)

		ctx := context.TODO()
		uploadInfo, err := e.Client.PutObject(ctx, bucket, objectName, file, tempFileSize, objopt)
		if err != nil {
			log.Println(err)
			return uploadInfo, err
		}
		if e.config.Debug {
			log.Println("Successfully uploaded bytes: ", uploadInfo)
		}
		return uploadInfo, err
	} else {
		return minio.UploadInfo{}, fmt.Errorf("模式未设置 %s", objectNameIn)
	}
}

// PutObjectBase64 上传 - base64
func (e *Component) PutObjectBase64(bucket string, objectNameIn string, base64File string, objopt minio.PutObjectOptions) (minio.UploadInfo, error) {
	if objectName, ok := e.CheckMode(objectNameIn); ok {
		b64data := base64File[strings.IndexByte(base64File, ',')+1:]
		if decode, err := base64.StdEncoding.DecodeString(b64data); err == nil {
			body := bytes.NewReader(decode)
			if uploadInfo, err := e.Client.PutObject(context.Background(), bucket, objectName, body, body.Size(), objopt); err == nil {
				if e.config.Debug {
					log.Println("Successfully uploaded bytes: ", uploadInfo)
				}
				return uploadInfo, err
			} else {
				log.Println(bucket, objectNameIn, err)
				return uploadInfo, err
			}
		} else {
			return minio.UploadInfo{}, err
		}
	} else {
		return minio.UploadInfo{}, fmt.Errorf("模式未设置 %s", objectNameIn)
	}
}

// PutObjectWithSrc 提供链接，上传到 minio
// return key & hash sha1 & error
func (e *Component) PutObjectWithSrc(dnComponent *idownload.Component, uri string, bucket string, objectName string, objopt minio.PutObjectOptions) (string, error) {
	// http 不处理
	if !strings.Contains(uri, "http") {
		return uri, errors.New("非链接地址:" + uri)
	}

	objectName = strings.Replace(objectName, "//", "/", -1)

	// 得判断文件大小，过大文件下载后上传
	// limitMax, _ := humanize.ParseBytes("43 MB")
	// fileSize := uint64(dnComponent.GetContentLength(uri))
	// log.Println(fileSize > limitMax || fileSize == 0, " xx")

	if true {
		tempname := ifile.NewFileName(uri).GetNameSnowFlow()
		if _, err := dnComponent.Download(uri, tempname); err == nil {
			defer os.Remove(tempname)
			// 打开文件
			file, err := os.OpenFile(tempname, os.O_RDONLY, 0444)
			defer file.Close()
			if err != nil {
				log.Fatalf("打开文件失败:%v\n", err)
			}
			// 获取文件大小
			fileInfo, err := file.Stat()
			if err != nil {
				log.Fatalf("获取文件信息失败:%v\n", err)
			}
			tempFileSize := fileInfo.Size()

			// 创建上传进度条对象
			objopt.Progress = progress.NewUploadProgress(tempFileSize)

			ctx := context.TODO()
			if info, err := e.Client.PutObject(ctx, bucket, objectName, file, tempFileSize, objopt); err == nil {
				log.Println(PackageName, "上传成功：✅", uri, bucket+"/"+info.Key)
				return bucket + "/" + info.Key, nil
			} else {
				log.Println(PackageName, "上传失败：❌", err, bucket, objectName, uri)
				return "", fmt.Errorf("上传失败：❌ %v, %s %s %s", err, bucket, objectName, uri)
			}
		} else {
			if e.config.Debug {
				log.Println(PackageName, "获取图片失败：❌", uri, err)
			}
			return "", errors.New("获取图片失败：❌" + uri + "  " + err.Error())
		}
	} else {
		if filebyte, err := dnComponent.DownloadToByte(uri); err != nil {
			if e.config.Debug {
				log.Println(PackageName, "DownloadToByte 失败：❌", uri, err)
			}
			return "", errors.New("DownloadToByte 失败：❌" + uri + "  " + err.Error())
		} else {
			// 打印日志
			if e.config.Debug {
				log.Printf("获取地址: %s, 代理：%s", uri, e.config.ProxySocks5)
			}
			if info, err := e.Client.PutObject(context.TODO(), bucket, objectName, bytes.NewReader(filebyte), int64(len(filebyte)), objopt); err != nil {
				if e.config.Debug {
					log.Println(PackageName, "上传失败：❌", err, bucket, objectName, uri)
				}
				return "", fmt.Errorf("上传失败：❌ %v, %s %s %s", err, bucket, objectName, uri)
			} else {
				log.Println(PackageName, "上传成功：✅", uri, bucket+"/"+info.Key)
				return bucket + "/" + info.Key, nil
			}
		}
	}
}

// DeleteObject 删除文件
func (e *Component) DeleteObject(objectNameWithBucket string) error {
	opts := minio.RemoveObjectOptions{}
	bucket, objectName := e.GetBucketAndObjectName(objectNameWithBucket)
	if len(bucket) == 0 || len(objectName) == 0 {
		return nil
	}
	err := e.Client.RemoveObject(context.Background(), bucket, objectName, opts)
	if err != nil {
		log.Println(PackageName, "删除对象失败：❌", fmt.Sprintf("Bucket:%s; Object:%s; 失败原因：", bucket, objectName), err)
		return err
	}
	log.Println(PackageName, "删除对象成功：✅", fmt.Sprintf("Bucket:%s; Object:%s", bucket, objectName))
	return nil
}

// DeleteObjectWithBucketAndKey 删除文件
func (e *Component) DeleteObjectWithBucketAndKey(bucket, key string) error {
	opts := minio.RemoveObjectOptions{}

	if len(bucket) == 0 || len(key) == 0 {
		return nil
	}
	err := e.Client.RemoveObject(context.Background(), bucket, key, opts)
	if err != nil {
		log.Println(PackageName, "删除对象失败：❌", fmt.Sprintf("Bucket:%s; Object:%s; 失败原因：", bucket, key), err)
		return err
	}

	log.Println(PackageName, "删除对象成功：✅", fmt.Sprintf("Bucket:%s; Object:%s", bucket, key))
	return nil
}

// GetBucketAndObjectName 根据路径获取bucket 和 object name
func (e *Component) GetBucketAndObjectName(input string) (string, string) {
	if input == "" {
		return "", ""
	}

	path := input
	// 1. 如果是完整 URL，解析出路徑部分
	if strings.Contains(input, "://") {
		u, err := url.Parse(input)
		if err != nil {
			return "", ""
		}
		path = u.Path // 獲取 /bucket/object/name...
	}

	// 2. 清洗路徑：去除首尾斜槓並處理連續斜槓
	path = strings.Trim(path, "/")
	if path == "" {
		return "", ""
	}

	// 3. 分割 bucket 和 objectName
	parts := strings.SplitN(path, "/", 2) // 只分割成兩份
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	// 只有 bucket，沒有 object
	return parts[0], ""
}

// 快速
func (e *Component) GetPutObjectOptionByFlag(contentType ContentType) minio.PutObjectOptions {
	return minio.PutObjectOptions{ContentType: string(contentType)}
}

// GetPutObjectOptions 获取默认PutObjectOptions
// video/mp4,video/webm,video/ogg
func (e *Component) GetPutObjectOptions(contentType string) minio.PutObjectOptions {
	if len(contentType) > 0 {
		return minio.PutObjectOptions{ContentType: contentType}
	}
	return minio.PutObjectOptions{ContentType: "image/jpeg,image/png,image/jpeg"}
}

// GetPutObjectOptionByExt 根据类型获取对象
func (e *Component) GetPutObjectOptionByExt(uri string) minio.PutObjectOptions {
	fileExt := path.Ext(uri)
	if len(uri) == 0 {
		fileExt = ".jpg"
	}

	contentType := ""
	switch fileExt {
	case ".png":
	case ".jpg":
	case ".svg":
	case ".jpeg":
		contentType = "image/jpeg,image/png"
	case ".gif":
		contentType = "image/gif"
	case ".mp4":
		// contentType = "audio/mp4"
		contentType = "video/mp4,video/webm,video/ogg"
	case ".avi":
		contentType = "video/avi"
	case ".mp3":
		contentType = "audio/mp3"
	case ".pdf":
		contentType = "application/pdf"
	case ".txt":
		contentType = "text/plain"
	default:
		contentType = "application/octet-stream"
	}
	return minio.PutObjectOptions{ContentType: contentType}
}

func (e *Component) GetConfig() {
	log.Println(PackageName, "配置信息：", fmt.Sprintf("%+v", e.config))
}
