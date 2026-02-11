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
		//log.Println("errors.New( key is empty )")
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

	// 1. 处理缓存逻辑
	hashkey := e.GenerateHashKey(1, bucket, key, urlOpt.Version)
	if urlOpt.Cache != nil && !urlOpt.Rebuild {
		if cachedData, err := urlOpt.Cache.Get(hashkey); err == nil && len(cachedData) > 0 {
			return cachedData
		}
	}

	var finalUrl string
	if urlOpt.Expiry > 0 {
		// 私有簽名邏輯
		reqParams := make(url.Values)
		if urlOpt.Version != "" {
			reqParams.Set("v", urlOpt.Version)
		}

		presignedURL, err := e.Client.PresignedGetObject(urlOpt.Context, bucket, key, urlOpt.Expiry, reqParams)
		if err != nil {
			log.Println("minio 簽名失敗 ", err)
			return ""
		}
		finalUrl = presignedURL.String()
	} else {
		// 公共地址邏輯
		baseUrl := strings.TrimSuffix(e.Client.EndpointURL().String(), "/")
		finalUrl = baseUrl + "/" + path.Join(bucket, key)
		if urlOpt.Version != "" {
			finalUrl = fmt.Sprintf("%s?v=%s", finalUrl, urlOpt.Version)
		}
	}

	// 3. 寫入快取 (關鍵修復)
	if urlOpt.Cache != nil && finalUrl != "" {
		cacheTTL := urlOpt.Expiry
		if urlOpt.Expiry > 0 {
			// 重要：快取時間要比簽名過期時間短，預留 5 分鐘緩衝
			buffer := 5 * time.Minute
			if cacheTTL > buffer {
				cacheTTL -= buffer
			} else {
				// 如果設定的有效期本來就很短，就不快取，避免失效
				return finalUrl
			}
		} else {
			// 公共地址快取 24 小時
			cacheTTL = 24 * time.Hour
		}
		urlOpt.Cache.Set(hashkey, finalUrl, cacheTTL)
	}

	return finalUrl
}

// GenerateHashKey 支持傳入任意數量的參數來生成唯一哈希
func (e *Component) GenerateHashKey(bucketType int32, bucket string, prefix string, version string) string {
	// 基礎組分
	base := fmt.Sprintf("%d:%s:%s:%s", bucketType, bucket, prefix, version)
	return hash.NewEncodeMD5(base)
}

// GetObjectsByPage minio 获取分页对象数据
// 1.分页
// 2.可以指定文件后缀获取
// 建議在文件上傳到 MinIO 時，文件名數字部分補零（例如：第009話），這樣 MinIO 默認的字典序就會是正確的自然排序，你原本的流式分頁代碼就能直接運行。
func (e *Component) GetObjectsByPage(bucket string, prefix string, page int32, perpage int32, fileExt []string) (objs []string, notall bool) {
	count := int32(0)
	offset := (page - 1) * perpage

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 严格按需索取：告诉 Minio 最多只返回到当前页为止的数据量
	opt := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
		MaxKeys:   int(offset + perpage),
	}
	objectCh := e.Client.ListObjects(ctx, bucket, opt)

	// 局部 map 过滤
	extMap := make(map[string]bool)
	for _, str := range fileExt {
		extMap[strings.ToLower(str)] = true
	}

	for object := range objectCh {
		if object.Err != nil {
			log.Println("object.Err", object.Err)
			continue
		}

		// 过滤逻辑
		objkeyname := path.Base(object.Key)
		if len(fileExt) > 0 {
			ext := strings.ToLower(path.Ext(objkeyname))
			if !extMap[ext] {
				continue
			}
		}

		// 分页逻辑
		if count >= offset {
			// 一旦满足当前页请求数量，立即 cancel 并退出
			if int32(len(objs)) >= perpage {
				notall = true
				cancel()
				break
			}
			objs = append(objs, bucket+"/"+object.Key)
		}
		count++
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

		// --- 核心修复 1：优化内存分配 ---
		if objectSize <= 0 {
			// 如果是流式上传（size为-1），手动限制分片大小为 10MB
			// 这样 SDK 内部的缓冲区就会被限制在 10MB 左右，而不是默认的数百MB
			if objopt.PartSize == 0 {
				objopt.PartSize = 10 * 1024 * 1024
			}
		}

		// --- 核心修复 2：安全初始化进度条 ---
		// 只有明确知道大小时才启用进度条，防止进度条逻辑因 -1 产生除零或计算异常
		if objectSize > 0 {
			objopt.Progress = progress.NewUploadProgress(objectSize)
		} else {
			// 确保流式上传时不带进度条
			objopt.Progress = nil
		}

		// 执行上传
		uploadInfo, err := e.Client.PutObject(context.Background(), bucket, objectName, reader, objectSize, objopt)

		if err != nil {
			log.Println("Upload Failed:", bucket, objectNameIn, err)
			return uploadInfo, err
		}

		if e.config.Debug {
			log.Printf("Successfully uploaded: %s/%s, Size: %d\n", bucket, objectName, uploadInfo.Size)
		}
		return uploadInfo, err
	}

	return minio.UploadInfo{}, fmt.Errorf("模式未设置 %s", objectNameIn)
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
	// 1. 先进行 URL Query 解码
	decodedPath, err := url.QueryUnescape(objectNameWithBucket)
	if err != nil {
		log.Println("删除解码失败", err)
		return err // 或者处理解码失败
	}
	opts := minio.RemoveObjectOptions{}
	bucket, objectName := e.GetBucketAndObjectName(decodedPath)
	if len(bucket) == 0 || len(objectName) == 0 {
		return nil
	}
	err = e.Client.RemoveObject(context.Background(), bucket, objectName, opts)
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
