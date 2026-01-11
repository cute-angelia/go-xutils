package izip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"github.com/cute-angelia/go-xutils/utils/iprogressbar"
)

// ZipFiles 压缩文件列表
func ZipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	// 2026 实践：显式 Close 以便捕获写入索引时的错误
	defer func() {
		if err := zipWriter.Close(); err != nil {
			fmt.Printf("close zip writer error: %v\n", err)
		}
	}()

	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

// ZipBytes 压缩字节数据
func ZipBytes(archiveName string, entryName string, data []byte) error {
	archive, err := os.Create(archiveName)
	if err != nil {
		return err
	}
	defer archive.Close()

	zw := zip.NewWriter(archive)
	w, err := zw.Create(entryName)
	if err != nil {
		return err
	}

	if _, err = io.Copy(w, bytes.NewReader(data)); err != nil {
		return err
	}
	return zw.Close() // 显式关闭以确保数据刷入
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// 2026 建议：统一使用正斜杠，并只保留基础文件名防止路径污染
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// 进度条处理：MultiWriter
	bar := iprogressbar.GetProgressbar(int(info.Size()), "zip:"+filename)
	_, err = io.Copy(io.MultiWriter(writer, bar), fileToZip)
	return err
}

// Unzip 解压文件，修复 Zip Slip 漏洞
func Unzip(archive, targetDir string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, ifile.DefaultDirPerm); err != nil {
		return err
	}

	for _, f := range reader.File {
		// 1. 安全性检查：修复 Zip Slip 漏洞
		fullPath, err := sanitizeExtractPath(f.Name, targetDir)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fullPath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(fullPath), ifile.DefaultDirPerm); err != nil {
			return err
		}

		// 解压文件
		if err := extractFile(f, fullPath); err != nil {
			return err
		}
	}
	return nil
}

// sanitizeExtractPath 防止路径穿越攻击
func sanitizeExtractPath(filePath, targetDir string) (string, error) {
	dest := filepath.Join(targetDir, filePath)
	// 检查解压后的绝对路径是否以 targetDir 开头
	if !strings.HasPrefix(dest, filepath.Clean(targetDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("illegal file path (Zip Slip): %s", filePath)
	}
	return dest, nil
}

func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}
