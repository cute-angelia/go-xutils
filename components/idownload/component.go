package idownload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	humanize "github.com/dustin/go-humanize"
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 全局 iclient
var iHttpClient *gout.Client

func init() {
	iHttpClient = gout.NewWithOpt(gout.WithInsecureSkipVerify())
}

var (
	ErrorNotFound = errors.New("404 file not found")
	ErrorHead     = errors.New("error head")
	ErrorUrl      = errors.New("url 不合法")
)

type FileInfo struct {
	SourceUrl string
	Path      string
}

type Component struct {
	config *config
	// 进度条 — 加锁保护并发赋值
	mu  sync.Mutex
	bar *progressbar.ProgressBar
}

// newComponent ...
func newComponent(config *config) *Component {
	comp := &Component{}
	comp.config = config
	return comp
}

func (d *Component) getHttpHeader() gout.H {
	gh := gout.H{}
	if len(d.config.UserAgent) > 0 {
		gh["User-Agent"] = d.config.UserAgent
	}
	if len(d.config.Referer) > 0 {
		gh["Referer"] = d.config.Referer
	}
	if len(d.config.Cookie) > 0 {
		gh["Cookie"] = d.config.Cookie
	}
	if len(d.config.Host) > 0 {
		gh["Host"] = d.config.Host
	}
	if len(d.config.Authorization) > 0 {
		gh["Authorization"] = "Bearer " + d.config.Authorization
	}
	return gh
}

// getGoHttpClient 业务定制头部
func (d *Component) getGoHttpClient(uri string, method string) *dataflow.DataFlow {
	var igout *dataflow.DataFlow
	switch method {
	case "GET":
		igout = iHttpClient.GET(uri)
	case "POST":
		igout = iHttpClient.POST(uri)
	case "PUT":
		igout = iHttpClient.PUT(uri)
	case "DELETE":
		igout = iHttpClient.DELETE(uri)
	case "HEAD":
		igout = iHttpClient.HEAD(uri)
	case "OPTIONS":
		igout = iHttpClient.OPTIONS(uri)
	default:
		igout = iHttpClient.GET(uri)
	}
	if d.config.Timeout > 0 {
		igout = igout.SetTimeout(d.config.Timeout)
	}
	if len(d.config.ProxySocks5) > 0 {
		igout = igout.SetSOCKS5(d.config.ProxySocks5)
	}
	if len(d.config.ProxyHttp) > 0 {
		igout = igout.SetProxy(d.config.ProxyHttp)
	}
	return igout.SetHeader(d.getHttpHeader()).Debug(d.config.Debug)
}

func (d *Component) validFileContentLength(strURL string) error {
	if d.config.FileMax != -1 {
		length := d.GetContentLength(strURL)
		if length > d.config.FileMax {
			return fmt.Errorf("链接：%s 未下载，大小：%s, 超过设置大小: %s",
				strURL,
				humanize.Bytes(uint64(length)),
				humanize.Bytes(uint64(d.config.FileMax)),
			)
		}
	}
	return nil
}

// GetContentLength 获取文件长度
func (d *Component) GetContentLength(strURL string) int {
	contentLength := 0
	header := http.Header{}
	var statusCode int
	if err := d.getGoHttpClient(strURL, "HEAD").BindHeader(&header).Code(&statusCode).Do(); err == nil {
		contentLength, _ = strconv.Atoi(header.Get("Content-Length"))
	} else {
		log.Println("err:", err)
	}
	return contentLength
}

// Download 下载文件
func (d *Component) Download(strURL, filename string) (fileInfo FileInfo, errResp error) {
	strURL = strings.TrimSpace(strURL)

	if !strings.Contains(strURL, "http") {
		return fileInfo, errors.New("Url 不合法：" + strURL)
	}

	if d.config.Debug {
		log.Println("下载地址：", strURL, "保存地址：", filename)
	}

	if err := d.validFileContentLength(strURL); err != nil {
		return fileInfo, err
	}

	if filename == "" {
		filename = path.Base(strURL)
	}

	header := http.Header{}
	var statusCode int

	ctx, cancel := context.WithTimeout(context.Background(), d.config.Timeout)
	defer cancel()

	err := d.getGoHttpClient(strURL, "HEAD").BindHeader(&header).Code(&statusCode).Do()
	if err != nil {
		log.Println("Head", err.Error())
	}

	if statusCode == http.StatusNotFound {
		return FileInfo{}, ErrorNotFound
	}

	// 下载地址切换
	if len(header["Location"]) > 0 {
		strURL = header["Location"][0]
	}

	// 是否分片下载
	if statusCode == http.StatusOK && header.Get("Accept-Ranges") == "bytes" && d.config.Concurrency > 0 {
		contentLength, _ := strconv.Atoi(header.Get("Content-Length"))
		fileInfo, errResp = d.multiDownload(strURL, filename, contentLength)
		if errResp != nil && d.config.RetryAttempt > 0 {
			NewRetry(d.config.RetryAttempt, d.config.RetryWaitTime).Func(func() error {
				log.Println("NewRetry multiDownload", strURL, filename)
				fileInfo, errResp = d.multiDownload(strURL, filename, contentLength)
				if errResp != nil {
					return ErrRetry
				}
				return nil
			}).Do(ctx)
		}
		return fileInfo, errResp
	}

	// 单例下载
	fileInfo, errResp = d.singleDownload(strURL, filename)
	if errResp != nil {
		log.Println("下载失败：错误：", strURL, errResp)
		if d.config.RetryAttempt > 0 {
			NewRetry(d.config.RetryAttempt, d.config.RetryWaitTime).Func(func() error {
				log.Println("NewRetry singleDownload", strURL, filename)
				fileInfo, errResp = d.singleDownload(strURL, filename)
				if errResp != nil {
					log.Println("singleDownload 下载失败：", errResp)
					return ErrRetry
				}
				return nil
			}).Do(ctx)
		}
	}
	return fileInfo, errResp
}

// DownloadToByteRetry 请求文件，返回字节（带重试）
func (d *Component) DownloadToByteRetry(src string, retry int) ([]byte, error) {
	src = strings.TrimSpace(src)

	if err := d.validFileContentLength(src); err != nil {
		return nil, err
	}

	var body []byte
	err := d.getGoHttpClient(src, "GET").Callback(func(c *dataflow.Context) error {
		switch c.Code {
		case 200:
			c.BindBody(&body)
			return nil
		case 404:
			return ErrorNotFound
		default:
			return fmt.Errorf(src+" error: %d", c.Code)
		}
	}).F().Retry().Attempt(retry).WaitTime(time.Second * 3).Do()

	if err != nil {
		log.Println(PackageName, "DownloadToByteRetry error -> ", src, err)
	}
	return body, err
}

// DownloadToByte 请求文件，流式读取返回字节，预分配容量避免大文件 OOM
func (d *Component) DownloadToByte(strURL string) ([]byte, error) {
	strURL = strings.TrimSpace(strURL)

	if err := d.validFileContentLength(strURL); err != nil {
		return nil, err
	}

	iClient := d.getGoHttpClient(strURL, "GET").Client()
	req, err := http.NewRequest("GET", strURL, nil)
	if err != nil {
		return nil, err
	}

	headers := d.getHttpHeader()
	for key, value := range headers {
		req.Header.Add(key, fmt.Sprintf("%v", value))
	}

	resp, err := iClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 预分配，避免 bytes.Buffer 多次扩容
	var buf bytes.Buffer
	if resp.ContentLength > 0 {
		buf.Grow(int(resp.ContentLength))
	}

	bufCache := make([]byte, 32*1024)

	if d.config.Progressbar {
		bar := d.newBar(int(resp.ContentLength), strURL)
		_, err = io.CopyBuffer(io.MultiWriter(&buf, bar), resp.Body, bufCache)
	} else {
		_, err = io.CopyBuffer(&buf, resp.Body, bufCache)
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RemoveFile 删除文件
func (d *Component) RemoveFile(filepath string) error {
	return os.Remove(filepath)
}

// CleanPartFiles 清理指定文件的所有分片临时文件，方便手动清理下载中断后的残留文件。
// 如果某个分片文件不存在则忽略，其余错误会聚合后返回。
func (d *Component) CleanPartFiles(filename string) error {
	var errs []string
	for i := 0; i < d.config.Concurrency; i++ {
		partFile := d.getPartFilename(filename, i)
		if err := os.Remove(partFile); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("删除分片 %s 失败: %v", partFile, err))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// multiDownload 并发分片下载
// 修复：使用 errgroup 收集 goroutine 错误；下载或合并失败时自动清理分片文件
func (d *Component) multiDownload(strURL, filename string, contentLen int) (FileInfo, error) {
	var info FileInfo

	bar := d.newBar(contentLen, strURL)

	partSize := contentLen / d.config.Concurrency
	partDir := d.getPartDir(filename)
	if err := os.MkdirAll(partDir, 0777); err != nil {
		return info, fmt.Errorf("创建分片目录失败: %w", err)
	}

	eg := new(errgroup.Group)
	rangeStart := 0

	for i := 0; i < d.config.Concurrency; i++ {
		i, rangeStart := i, rangeStart // 捕获循环变量

		rangeEnd := rangeStart + partSize
		if i == d.config.Concurrency-1 {
			rangeEnd = contentLen
		}

		eg.Go(func() error {
			downloaded := 0
			if d.config.Resume {
				partFileName := d.getPartFilename(filename, i)
				if content, err := os.ReadFile(partFileName); err == nil {
					downloaded = len(content)
				}
				if d.config.Progressbar && bar != nil {
					bar.Add(downloaded)
				}
			}
			return d.downloadPartial(strURL, filename, rangeStart+downloaded, rangeEnd, i, bar)
		})

		rangeStart += partSize + 1
	}

	if err := eg.Wait(); err != nil {
		// 分片下载失败，清理所有临时文件
		if cleanErr := d.CleanPartFiles(filename); cleanErr != nil {
			log.Println("清理分片文件失败:", cleanErr)
		}
		return info, fmt.Errorf("分片下载失败: %w", err)
	}

	if err := d.merge(filename); err != nil {
		// 合并失败，清理所有临时文件
		if cleanErr := d.CleanPartFiles(filename); cleanErr != nil {
			log.Println("清理分片文件失败:", cleanErr)
		}
		return info, fmt.Errorf("合并文件失败: %w", err)
	}

	info.SourceUrl = strURL
	info.Path = filename
	return info, nil
}

// singleDownload 单线程下载
func (d *Component) singleDownload(strURL, filename string) (FileInfo, error) {
	var info FileInfo

	iClient := d.getGoHttpClient(strURL, "GET").Client()
	req, err := http.NewRequest("GET", strURL, nil)
	if err != nil {
		return info, err
	}

	headers := d.getHttpHeader()
	for key, value := range headers {
		req.Header.Add(key, fmt.Sprintf("%v", value))
	}

	resp, err := iClient.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return info, err
	}
	defer f.Close()

	buf := make([]byte, 32*1024)

	if d.config.Progressbar {
		bar := d.newBar(int(resp.ContentLength), strURL)
		_, err = io.CopyBuffer(io.MultiWriter(f, bar), resp.Body, buf)
	} else {
		_, err = io.CopyBuffer(f, resp.Body, buf)
	}

	if err != nil {
		return info, err
	}

	info.SourceUrl = strURL
	info.Path = filename
	return info, nil
}

// newBar 创建进度条。
// 使用局部变量而非 Component.bar 共享字段，避免多 goroutine 并发赋值竞态。
// 同时更新 Component.bar（加锁），供外部需要访问当前 bar 的场景使用。
func (d *Component) newBar(length int, name string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(
		length,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Print("\n")
		}),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription("downloading "+name),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	d.mu.Lock()
	d.bar = bar
	d.mu.Unlock()

	return bar
}

// downloadPartial 下载单个分片，错误向上返回（不再静默丢弃）
func (d *Component) downloadPartial(strURL, filename string, rangeStart, rangeEnd, i int, bar *progressbar.ProgressBar) error {
	if rangeStart >= rangeEnd {
		return nil
	}

	iClient := d.getGoHttpClient(strURL, "GET").Client()
	req, err := http.NewRequest("GET", strURL, nil)
	if err != nil {
		return fmt.Errorf("分片 %d 创建请求失败: %w", i, err)
	}

	headers := d.getHttpHeader()
	for key, value := range headers {
		req.Header.Add(key, fmt.Sprintf("%v", value))
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd))
	resp, err := iClient.Do(req)
	if err != nil {
		return fmt.Errorf("分片 %d 请求失败: %w", i, err)
	}
	defer resp.Body.Close()

	flags := os.O_CREATE | os.O_WRONLY
	if d.config.Resume {
		flags |= os.O_APPEND
	}

	partFile, err := os.OpenFile(d.getPartFilename(filename, i), flags, 0666)
	if err != nil {
		return fmt.Errorf("分片 %d 打开文件失败: %w", i, err)
	}
	defer partFile.Close()

	buf := make([]byte, 32*1024)

	var writer io.Writer = partFile
	if d.config.Progressbar && bar != nil {
		writer = io.MultiWriter(partFile, bar)
	}

	if _, err = io.CopyBuffer(writer, resp.Body, buf); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("分片 %d 写入失败: %w", i, err)
	}
	return nil
}

func (d *Component) merge(filename string) error {
	destFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer destFile.Close()

	for i := 0; i < d.config.Concurrency; i++ {
		partFileName := d.getPartFilename(filename, i)
		partFile, err := os.Open(partFileName)
		if err != nil {
			return fmt.Errorf("打开分片 %d 失败: %w", i, err)
		}
		if _, err = io.Copy(destFile, partFile); err != nil {
			partFile.Close()
			return fmt.Errorf("合并分片 %d 失败: %w", i, err)
		}
		partFile.Close()
		os.Remove(partFileName)
	}
	return nil
}

// getPartDir 分片文件存放目录
func (d *Component) getPartDir(filename string) string {
	return path.Dir(filename)
}

// getPartFilename 构造分片文件名
func (d *Component) getPartFilename(filename string, partNum int) string {
	partDir := d.getPartDir(filename)
	return fmt.Sprintf("%s/%s-%d", partDir, path.Base(filename), partNum)
}
