package pictran

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
)

// GetGrayPicRGBA get gray pgba
func GetGrayPicRGBA(picSource string) *image.RGBA {
	ff, err := ioutil.ReadFile(picSource)
	if err != nil {
		panic(err)
	}

	picBytes := bytes.NewBuffer(ff)
	m, _, err := image.Decode(picBytes)
	if err != nil {
		panic(err)
	}

	grayRet := hdImage(m)

	return grayRet
}

// GetPicRGBA trans png into RGBA
func GetPicRGBA(picSource string) *image.RGBA {
	ff, err := ioutil.ReadFile(picSource)
	if err != nil {
		panic(err)
	}

	picBytes := bytes.NewBuffer(ff)
	m, _, err := image.Decode(picBytes)
	if err != nil {
		panic(err)
	}

	bounds := m.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	newRgba := image.NewRGBA(bounds)
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			colorRgb := m.At(i, j)
			r, g, b, a := colorRgb.RGBA()
			newRgba.SetRGBA(i, j, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}
	return newRgba
}

// CreatePic create pic
func CreatePic(rgbaRet *image.RGBA, picTarget, formatName string) {
	f, _ := os.Create(picTarget)
	defer f.Close()
	encode(formatName, f, rgbaRet)
}

/* static funcs */

func hdImage(m image.Image) *image.RGBA {
	bounds := m.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	newRgba := image.NewRGBA(bounds)
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			colorRgb := m.At(i, j)
			_, g, _, a := colorRgb.RGBA()
			gUint8 := uint8(g >> 8)
			aUint8 := uint8(a >> 8)
			newRgba.SetRGBA(i, j, color.RGBA{gUint8, gUint8, gUint8, aUint8})
		}
	}
	return newRgba
}

//图片编码
func encode(inputName string, file *os.File, rgba *image.RGBA) {
	if strings.HasSuffix(inputName, "jpg") || strings.HasSuffix(inputName, "jpeg") {
		jpeg.Encode(file, rgba, nil)
	} else if strings.HasSuffix(inputName, "png") {
		png.Encode(file, rgba)
	} else if strings.HasSuffix(inputName, "gif") {
		gif.Encode(file, rgba, nil)
	} else {
		fmt.Errorf("不支持的图片格式")
	}
}
