package iminio

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/cute-angelia/go-xutils/components/idownload"
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"github.com/cute-angelia/go-xutils/utils/generator/hash"
	progress "github.com/markity/minio-progress"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

// CheckMode ✅ 移除废弃的 rand.Seed
func (e *Component) CheckMode(objectName string) (newObjectName string, canupload bool) {
	switch e.config.ReplaceMode {
	case ReplaceModeIgnore:
		return objectName, false
	case ReplaceModeReplace:
		return objectName, true
	case ReplaceModeTwo:
		// Go 1.20+ 全局 rand 已自动随机初始化，无需 Seed
		return fmt.Sprintf("bak_%d_%s", rand.Intn(100), objectName), true
	default:
		return objectName, false
	}
}

// CopyObject 复制对象
func (e *Component) CopyObject(dst minio.CopyDestOptions, src minio.CopySrcOptions) (uploadInfo minio.UploadInfo, err error) {
	uploadInfo, err = e.Client.CopyObject(context.Background(), dst, src)
	return
}

// PutObject 上传-按读取文件数据
// PutObject：流式上传时用 pipe，避免整体读入内存
func (e *Component) PutObject(bucket string, objectNameIn string, reader io.Reader, objectSize int64, objopt minio.PutObjectOptions) (minio.UploadInfo, error) {
	objectName, ok := e.CheckMode(objectNameIn)
	if !ok {
		return minio.UploadInfo{}, fmt.Errorf("模式未设置 %s", objectNameIn)
	}
	objectName = strings.ReplaceAll(objectName, "//", "/")

	// 流式上传：size=-1 时不设 PartSize（无效），改用 SendContentMd5 减少重传风险
	if objectSize <= 0 {
		// ✅ 必须设 PartSize，否则 SDK 全量缓冲
		if objopt.PartSize == 0 {
			objopt.PartSize = 32 * 1024 * 1024 // 32MB 每片，内存占用恒定
		}
		objopt.Progress = nil
	} else {
		// ✅ size 已知时，动态计算分片大小，保证不超过 10000 片
		if objopt.PartSize == 0 {
			partSize := uint64(objectSize) / 10000
			if partSize < 32*1024*1024 {
				partSize = 32 * 1024 * 1024 // 最小 32MB
			}
			objopt.PartSize = partSize
		}
		objopt.Progress = progress.NewUploadProgress(objectSize)
	}

	// ✅ 统一用带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	uploadInfo, err := e.Client.PutObject(ctx, bucket, objectName, reader, objectSize, objopt)
	if err != nil {
		log.Println("Upload Failed:", bucket, objectNameIn, err)
		return uploadInfo, err
	}
	if e.config.Debug {
		log.Printf("Successfully uploaded: %s/%s, Size: %d\n", bucket, objectName, uploadInfo.Size)
	}
	return uploadInfo, nil
}

// FPutObject：加上超时 context
func (e *Component) FPutObject(bucket string, objectNameIn string, filePath string, objopt minio.PutObjectOptions) (minio.UploadInfo, error) {
	objectName, ok := e.CheckMode(objectNameIn)
	if !ok {
		return minio.UploadInfo{}, fmt.Errorf("模式未设置 %s", objectNameIn)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0444)
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("获取文件信息失败: %w", err)
	}

	objopt.Progress = progress.NewUploadProgress(fileInfo.Size())

	// ✅ 加超时，避免大文件永久阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	uploadInfo, err := e.Client.PutObject(ctx, bucket, objectName, file, fileInfo.Size(), objopt)
	if err != nil {
		return uploadInfo, fmt.Errorf("上传失败: %w", err)
	}
	if e.config.Debug {
		log.Println("Successfully uploaded bytes: ", uploadInfo)
	}
	return uploadInfo, nil
}

// PutObjectBase64：✅ 用 streaming decode 避免双份内存
func (e *Component) PutObjectBase64(bucket string, objectNameIn string, base64File string, objopt minio.PutObjectOptions) (minio.UploadInfo, error) {
	objectName, ok := e.CheckMode(objectNameIn)
	if !ok {
		return minio.UploadInfo{}, fmt.Errorf("模式未设置 %s", objectNameIn)
	}

	// 跳过 data:image/xxx;base64, 前缀
	commaIdx := strings.IndexByte(base64File, ',')
	if commaIdx >= 0 {
		base64File = base64File[commaIdx+1:]
	}

	// ✅ 用 base64.NewDecoder 流式解码，不在内存中产生完整的解码副本
	srcReader := strings.NewReader(base64File)
	decodedReader := base64.NewDecoder(base64.StdEncoding, srcReader)

	// 流式上传，size=-1 表示未知大小
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	uploadInfo, err := e.Client.PutObject(ctx, bucket, objectName, decodedReader, -1, objopt)
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("base64 上传失败: %w", err)
	}
	if e.config.Debug {
		log.Println("Successfully uploaded bytes: ", uploadInfo)
	}
	return uploadInfo, nil
}

// PutObjectWithSrc 提供链接，上传到 minio
// return key & hash sha1 & error
func (e *Component) PutObjectWithSrc(dnComponent *idownload.Component, uri string, bucket string, objectName string, objopt minio.PutObjectOptions) (string, error) {
	if !strings.Contains(uri, "http") {
		return uri, errors.New("非链接地址:" + uri)
	}

	objectName = strings.ReplaceAll(objectName, "//", "/")

	tempname := ifile.NewFileName(uri).GetNameSnowFlow()
	if _, err := dnComponent.Download(uri, tempname); err != nil {
		if e.config.Debug {
			log.Println(PackageName, "获取文件失败：❌", uri, err)
		}
		return "", fmt.Errorf("获取文件失败：❌ %s  %w", uri, err)
	}
	defer os.Remove(tempname) // 下载成功后，函数退出时清理

	file, err := os.OpenFile(tempname, os.O_RDONLY, 0444)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	objopt.Progress = progress.NewUploadProgress(fileInfo.Size())

	info, err := e.Client.PutObject(context.TODO(), bucket, objectName, file, fileInfo.Size(), objopt)
	if err != nil {
		log.Println(PackageName, "上传失败：❌", err, bucket, objectName, uri)
		return "", fmt.Errorf("上传失败：❌ %w, %s %s %s", err, bucket, objectName, uri)
	}

	log.Println(PackageName, "上传成功：✅", uri, bucket+"/"+info.Key)
	return bucket + "/" + info.Key, nil
}

// DeleteObject ✅ 统一使用 log，移除 fmt.Printf
func (e *Component) DeleteObject(objectNameWithBucket string) error {
	decodedPath, err := url.PathUnescape(objectNameWithBucket)
	if err != nil {
		log.Println(PackageName, "路径解码失败:", err)
		return err
	}

	bucket, objectName := e.GetBucketAndObjectName(decodedPath)
	if len(bucket) == 0 || len(objectName) == 0 {
		return fmt.Errorf("invalid path: %s", objectNameWithBucket)
	}

	err = e.Client.RemoveObject(context.Background(), bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("%s 删除对象失败：❌ Bucket:%s; Object:%s; 原因：%v\n", PackageName, bucket, objectName, err)
		return err
	}

	log.Printf("%s 删除对象成功：✅ Bucket:%s; Object:%s\n", PackageName, bucket, objectName)
	return nil
}

// DeleteFolder 递归删除目录下所有对象
func (e *Component) DeleteFolder(objectNameWithBucket string) error {
	// 1. 解码并获取 Bucket 和 Prefix
	decodedPath, err := url.PathUnescape(objectNameWithBucket)
	if err != nil {
		log.Println(PackageName, "文件夹路径解码失败:", err)
		return err
	}

	bucket, folderPrefix := e.GetBucketAndObjectName(decodedPath)
	if len(bucket) == 0 || len(folderPrefix) == 0 {
		return fmt.Errorf("invalid folder path: %s", objectNameWithBucket)
	}

	// 2. 核心：必须确保前缀以 "/" 结尾，否则会误删前缀相似的对象
	// 例如：想删 "photos"，如果不加 /，会把 "photos_backup" 也删掉
	if !strings.HasSuffix(folderPrefix, "/") {
		folderPrefix += "/"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. 获取该前缀下的所有对象通道
	// ListObjects 会递归 (Recursive: true) 查找目录下所有子文件
	objectsCh := e.Client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    folderPrefix,
		Recursive: true,
	})

	// 4. 调用 RemoveObjects 批量删除通道中的所有对象
	// 这是比循环调用 RemoveObject 效率高得多的方式
	errorCh := e.Client.RemoveObjects(ctx, bucket, objectsCh, minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	})

	// 5. 监听错误通道
	hasError := false
	for err := range errorCh {
		log.Printf("%s 批量删除中发生错误: Object:%s, Error:%v\n", PackageName, err.ObjectName, err.Err)
		hasError = true
	}

	if hasError {
		return fmt.Errorf("partially failed to delete folder: %s", folderPrefix)
	}

	log.Printf("%s 目录及其内容已彻底删除：✅ Bucket:%s; Prefix:%s\n", PackageName, bucket, folderPrefix)
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
// ✅ 修复 switch case 穿透导致 png/jpg/svg 拿不到 contentType
func (e *Component) GetPutObjectOptionByExt(uri string) minio.PutObjectOptions {
	fileExt := path.Ext(uri)
	if len(uri) == 0 || fileExt == "" {
		fileExt = ".jpg"
	}

	var contentType string
	switch strings.ToLower(fileExt) {
	case ".png", ".jpg", ".jpeg", ".svg":
		contentType = "image/jpeg,image/png"
	case ".gif":
		contentType = "image/gif"
	case ".mp4":
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
