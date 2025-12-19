package idownload

import (
	"context"
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"log"
	"net/http"
	"testing"
	"time"
)

type datas struct {
	Uri    string
	Socket string
}

func TestLocation(t *testing.T) {
	uri := "https://videocelebs.net/get_file/1/011ca90b914025264f75527a7b1529085280883d0e/7000/7311/7311.mp4/"
	idown := New()
	header := http.Header{}
	err := idown.getGoHttpClient(uri, "HEAD").Debug(true).BindHeader(&header).Do()
	log.Println(err)

	// gout.HEAD(uri).Debug(true).BindHeader(&header).Do()

	log.Println(header)

}

func TestRosemm(t *testing.T) {
	uri := "http://rs.jinyemimi.com/jpg/3981-BJLdDOGQ/033.rosi"
	idown := New(
		WithReferer("rs.jinyemimi.com"),
		WithHost("rs.jinyemimi.com"),
		WithUserAgent("rosi app 1.0.3"),
	)
	// findo, err := idown.DownloadToByte(uri)
	findo, err := idown.Download(uri, "/tmp/1.jpg")
	log.Println(ijson.Pretty(findo))
	log.Println(err)
}

func TestToptoon(t *testing.T) {
	uri := "https://twattraction.akamaized.net/www_v1/imgComic/ep_content/1281_36208_1606901405.7415.jpg?__token__=exp=1651743096~acl=/*~hmac=1017ea8986e6e1b249f03bf06805f39be74c00ef091fbeaf2c172d0d55b618e9"
	idown := New(
		WithDebug(true),
		WithReferer("https://www.toptoon.net/"),
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"),
	)
	// findo, err := idown.DownloadToByte(uri)
	findo, err := idown.Download(uri, "/tmp/1.jpg")
	log.Println(ijson.Pretty(findo))
	log.Println(err)
}

func TestV2(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.Ltime)
	// uri := "https://cdn.v2ph.com/photos/czjDRlLlhvf-robP.jpg"
	uri := "https://cdn.v2ph.com/photos/4yUuLsfbGgf7dSdf.jpg"
	idown := New(
		WithDebug(true),
		WithReferer("https://www.v2ph.com/album/XIAOYU-555?page=2"),
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"),
		WithCookie("frontend=54300d6621fe8bf51aa0bd41bac06da6; frontend-rmu=SIKqjQhmcTsvgdVhaeg4rdj7EC8%3D; frontend-rmt=iWCNVa30N8MEds0vju3LY9QKfnIPrNIgvZ45wtWJnCtsPtvgQCodEKj3RAqyd5cy; _gid=GA1.2.1309613440.1651131130; fpestid=UflGBDjSIrZXUuUkg9YCWQigFBXg9H5TVbykLToxcXpsc5zMwRXfjUSlegirc6fyzbK6rw; __cf_bm=YcWcL8MOMBKJYiDAXWLk6U0f4MbJU0x0twkWApJeQaI-1651138216-0-AUUO2wc3zy5DVtt8H4y8QdyCUha8yGwpS3g87QpXspOiPNv+Xm/ZgNEK1yisl9yhKSS3UCV7WR8Vgv8xze0zU4AO3xsw8CZ6skvQQe9PtInuA9vX50Bs9zGLseAcNPnuhA==; _ga_170M3FX3HZ=GS1.1.1651137021.3.1.1651138218.47; _ga=GA1.2.1638894544.1651131130"),
	)
	findo, err := idown.Download(uri, "/tmp/1.jpg")
	log.Println(ijson.Pretty(findo))
	log.Println(err)
}

func TestV3(t *testing.T) {
	uri := "https://wx1.sinaimg.cn/large/008oKRrcgy1gv7y8kragcj60u0140goh02.jpg"
	idown := New(
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"),
	)
	findo, err := idown.Download(uri, "/tmp/2.jpg")
	log.Println(ijson.Pretty(findo))
	log.Println(err)
}

func TestHuaBan(t *testing.T) {
	uri := "https://hbimg.b0.upaiyun.com/55fd47126effca1653ca4a0f1536c020d4f8bb8469016-G9lg4Y"
	idown := New(
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"),
	)
	findo, err := idown.Download(uri, "/tmp/2.jpg")
	log.Println(ijson.Pretty(findo))
	log.Println(err)
}

func TestDownloadDouyin(t *testing.T) {
	uri := "https://aweme.snssdk.com/aweme/v1/play/?video_id=v0200fg10000c6b2tlrc77ub0qkvip1g&line=0&ratio=1080p&media_type=4&vr_type=0&improve_bitrate=0&is_play_url=1&is_support_h265=0&source=PackSourceEnum_PUBLISH"
	idown := New()
	findo, err := idown.Download(uri, "/tmp/1.mp4")
	log.Println(ijson.Pretty(findo))
	log.Println(err)
}

func TestDownloads(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)
	var data []datas
	data = append(data,
		datas{
			Uri:    "https://wx1.sinaimg.cn/large/008oKRrcgy1gv7y8kragcj60u0140goh02.jpg",
			Socket: "",
		},
		datas{
			Uri:    "https://telegra.ph/file/956de9b5ca3c41703eb52.jpg",
			Socket: "socks5://host-bwg-new.aaqq.in:8096",
		},
		datas{
			Uri:    "https://go.dev/dl/go1.17.7.src.tar.gz",
			Socket: "",
		},
	)
	for _, datum := range data {
		idown := New()
		if len(datum.Socket) > 0 {
			idown = New(WithDebug(true), WithProxySocks5(datum.Socket))
		}

		// 下载文件
		//newName := ifile.NewFileName(datum.Uri).GetNameOrigin()
		//if fileinfo, err := idown.Download(datum.Uri, "/tmp/"+newName); err == nil {
		//	log.Println(ijson.Pretty(fileinfo))
		//} else {
		//	log.Println("获取图片失败：❌", err)
		//}

		if filebyte, err := idown.DownloadToByte(datum.Uri); err != nil {
			log.Println("获取图片失败：❌", err)
		} else {
			log.Println(len(filebyte))
		}
	}
}

func TestDownloadHeader(t *testing.T) {
	fileuri := "http://ali2.a.kwimgs.com/ufile/atlas/NTIwMzM0NjQ3NjI4NzQyODkwM18xNjMyMjMyOTczMzMx_0.jpg"
	idown := New(
		WithDebug(true),
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"),
	)
	if filebyte, err := idown.DownloadToByte(fileuri); err != nil {
		log.Println("获取图片失败：❌", err)
	} else {
		log.Println(len(filebyte))
	}
}

// go test -v -run TestDownload3
func TestDownload3(t *testing.T) {
	fileuri := "https://cdn.v2ph.com/photos/P0DcKbgkeL39x5Ir.jpg"
	idown := New(
		WithDebug(true),
		WithReferer("https://www.v2ph.com/"),
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36"),
	)
	if filebyte, err := idown.DownloadToByte(fileuri); err != nil {
		log.Println("获取图片失败：❌", err)
	} else {
		log.Println("获取图片成功", len(filebyte))
	}
}

// go test -v -run TestDownLargeFile
func TestDownLargeFile(t *testing.T) {
	fileuri := "https://ttq.jinyemimi.com/2021/0782/202201010831.zip"
	idown := New(
		WithDebug(true),
	)
	newname := ifile.NewFileName(fileuri).GetNameSnowFlow()
	if filebyte, err := idown.Download(fileuri, "/tmp/"+newname); err == nil {
		log.Println(ijson.Pretty(filebyte))
	} else {
		log.Println(err)
	}
}

// 定时器测试， 必须 reset
func TestRetryTimer(t *testing.T) {
	timer := time.NewTimer(time.Second * 5)
	for i := 0; i < 12; i++ {
		timer.Reset(time.Second * 2)
		select {
		case <-timer.C:
			log.Println("z...", i)
		}
	}
	timer.Stop()
}

// 重试方法
func TestRetryFuc(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	NewRetry(3, time.Second*3).Func(func() error {
		log.Println("i, ok")
		return ErrRetry
	}).Do(ctx)
}

// TestTwitter go test -v --run TestTwitter
func TestTwitter(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	idownloader := New(
		WithTimeout(time.Minute),
		WithProxySocks5("socks5://23.56.107.58:38153"),
		WithHost("twitter.com"),
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"),
		WithFileMax(300*1024*1024),
	)

	// info, err := idownloader.Download("https://video.twimg.com/ext_tz_video/1671201090497576960/pu/vid/1280x720/LAtCzKi_8NkCMZ_0.mp4?tag=12", "/tmp/1.mp4")
	info, err := idownloader.Download("https://video.twimg.com/amplixy_video/1669825038521110528/vid/1920x1080/4Z3t98204cgh06Qo.mp4?tag=16", "/tmp/1.mp4")

	log.Println(ijson.Pretty(info))
	log.Println(err)
}
