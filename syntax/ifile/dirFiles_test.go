package ifile

import (
	"fmt"
	"github.com/cute-angelia/go-utils/syntax/ijson"
	pool "github.com/cute-angelia/go-xutils/utils/pool/antspool"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// go test -v -run TestDirFiles
func TestDirFiles(t *testing.T) {

	str := "FilterFilename\nx\t @ daw哦"
	log.Println(FilterFilename(str))

	// rootPath := "/Users/vanilla/Downloads"
	// t.Log(GetAllPaths(rootPath))
	t.Log(Name("/Users/vanilla/Downloads/财务管理学/课件/财务管理学课件精讲四王天娇.pdf"))
	t.Log(Name2("/Users/vanilla/Downloads/财务管理学/课件/财务管理学课件精讲四王天娇.pdf"))

	rootPath := "/Users/vanilla/Downloads"

	dirPaths, filePaths, err := GetDepthOnePathsAndFilesIncludeExt(rootPath)
	log.Println(ijson.Pretty(dirPaths))
	log.Println(ijson.Pretty(filePaths))
	log.Println(err)

}

func TestSyncFiles(t *testing.T) {
	dirPath := "/Users/vanilla/Downloads/MZZP"

	// 读取文件和文件夹
	// readCurrentDir(dirPath)
	readWalkDir(dirPath)
}

// 读取当前文件夹
var badDir []string

func readCurrentDir(dirPath string) {
	// 读取文件和文件夹
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}
	for _, file := range files {
		// log.Println(badDir)
		// log.Println("读取文件夹:--->", dirPath+"/"+file.Name())
		if file.IsDir() {
			next := true
			for _, s := range badDir {
				if strings.Contains(dirPath, s) {
					//  发现忽略文件
					next = false
				}
			}
			if next {
				readCurrentDir(dirPath + "/" + file.Name())
				log.Println("读取:--->", dirPath+"/"+file.Name())
				writeDir(dirPath + "/" + file.Name())
			}
		} else {
			// log.Println("dif:",dirPath + "/" + file.Name())
			if file.Name() == ".ignore" {
				log.Println("发现忽略文件：当前文件夹", dirPath+"/")
				badDir = append(badDir, dirPath)
			}
		}
	}
}

func readWalkDir(dirPath string) {
	ipool := pool.MustNewPoolAnts(30, false)
	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}
		files, _ := ioutil.ReadDir(path)
		if len(files) > 0 {
			// 是否全是文件夹
			allIsDir := true
			for _, file := range files {
				// 文件包含 特殊 ignore
				if strings.Contains(file.Name(), "ignore") {
					badDir = append(badDir, path)
					return nil
				}
				// log.Println(path + "/" + file.Name())

				if !file.IsDir() {
					allIsDir = false
				}
			}

			// 简单判断最后一级
			if !allIsDir {
				findIgnore := false
				for _, s := range badDir {
					if strings.Contains(path, s) {
						//  发现忽略文件
						findIgnore = true
					}
				}
				if !findIgnore {
					// info, _ := d.Info()
					ipool.SubmitTask(func() {
						// this.insert(path, info)
						log.Println("insert:----->", path)
					})
				}
			}
		}
		return nil
	})

	ipool.RunningTask()
	ipool.Stop()
}

func writeDir(dirPath string) {
	next := true
	for _, s := range badDir {
		if strings.Contains(dirPath, s) {
			//  发现忽略文件
			next = false
		}
	}
	log.Println("写", dirPath, next)
}

// go test -v -run TestMapList
func TestMapList(t *testing.T) {
	// inpath := "/Users/vanilla/Downloads/けんけんぱ(けんけ) fantia 2020.03-2021.08"
	basePath := "/Users/vanilla/Downloads/けんけんぱ(けんけ) fantia 2020.03-2021.08"

	data := GetFileMapList(basePath, map[string][]string{})
	// t.Log(ijson.Pretty(data))
	DoPutImage(basePath, data, 1)
}

// 处理图片 保留多少step级子集
func DoPutImage(basePath string, data map[string][]string, step int) {
	baseStepCount := strings.Count(basePath, "/")
	// 其他测试 处理 子集图片，保留一个子集
	for key, val := range data {
		log.Println("===>", key, len(val))
		if len(val) > 0 {
			for _, v := range val {
				// 保留多少级
				baseStepSonCount := strings.Count(key, "/")
				if baseStepSonCount-step <= baseStepCount {
					continue
				}

				filename := path.Base(v)
				//log.Println("---")
				//log.Println(baseStepCount, baseStepSonCount)
				//log.Println(key)
				//log.Println(v)
				//log.Println(filename)

				// 假如是顶级，不处理
				tempv := strings.Replace(v, basePath, "", -1)
				if len(strings.Split(tempv, "/")) == 2 {
					continue
				}
				// 将子集移动到父级
				nfilename := strings.Replace(filename, "fulibl.net", "", -1)
				nfilename = strings.TrimSpace(nfilename)
				newname := fmt.Sprintf("%s_%s", key, nfilename)
				log.Println("--")
				log.Println(v)
				log.Println(newname)
				Mv(v, newname)
			}
		} else {
			log.Println("empty:", key)
		}
	}

	// 清理空文件夹
	dirs, _, _ := GetAllPaths(basePath)

	for _, dir := range dirs {
		if CheckIsEmptyDir(dir) {
			log.Println("empty:", dir)
			os.Remove(dir)
		}
	}

}

// go test -v -run TestEmptyDir
func TestEmptyDir(t *testing.T) {
	inpath := "/Users/vanilla/Downloads/けんけんぱ(けんけ) fantia 2020.03-2021.08/2020.11"
	t.Log(CheckIsEmptyDir(inpath))
}
