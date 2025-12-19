package iimage

import (
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/ifile"
	"image"
	"image/color"
	"os"
)

// Keep it DRY so don't have to repeat opening file and decode
func OpenAndDecode(filepath string) (image.Image, string, error) {
	imgFile, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()
	img, format, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}
	return img, format, nil
}

// Create a struct to deal with pixel
type Pixel struct {
	Point image.Point
	Color color.Color
}

// Decode image.Image's pixel data into []*Pixel
func DecodePixelsFromImage(img image.Image, offsetX, offsetY int) []*Pixel {
	pixels := []*Pixel{}
	for y := 0; y <= img.Bounds().Max.Y; y++ {
		for x := 0; x <= img.Bounds().Max.X; x++ {
			p := &Pixel{
				Point: image.Point{x + offsetX, y + offsetY},
				Color: img.At(x, y),
			}
			pixels = append(pixels, p)
		}
	}
	return pixels
}

// LimitWidthHeightUseIsNot 限制图片大小
func LimitWidthHeightUseIsNot(localFile string, width, height int) error {
	if f, err := ifile.OpenLocalFile(localFile); err != nil {
		return err
	} else {
		defer f.Close()
		if i, _, err := image.DecodeConfig(f); err != nil {
			return err
		} else {
			if width > 0 {
				if i.Width < width {
					os.Remove(localFile)
					return fmt.Errorf(fmt.Sprintf("限制图片大小:小于规定宽度:%d", width))
				}
			}
			if height > 0 {
				if i.Height < height {
					os.Remove(localFile)
					return fmt.Errorf(fmt.Sprintf("限制图片大小:小于规定高度:%d", height))
				}
			}
		}

		return nil
	}
}
