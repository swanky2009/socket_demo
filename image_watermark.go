package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"math"
	"strconv"
)

func main() {
	//原始图片
	imgb, err := os.Open("d:/web/images/test.jpg")
	img, _ := jpeg.Decode(imgb)
	defer imgb.Close()

	if err != nil {
		fmt.Println("read img error:" + err.Error())
		return
	}

	wmb, err1 := os.Open("d:/web/images/watermark.png")
	watermark, _ := png.Decode(wmb)
	fmt.Println("img height : " + strconv.Itoa(img.Bounds().Dy()))
	waterSize := int(math.Floor(float64(img.Bounds().Dy()/10)))
	waterOffset := waterSize/10
	fmt.Println("img offset : " + strconv.Itoa(waterOffset))
	watermark = imaging.Resize(watermark, waterSize, waterSize, imaging.Lanczos)
	defer wmb.Close()
  
	if err1 != nil {
		fmt.Println("read watermark error:" + err1.Error())
		return
	}
	//把水印写到右下角，并向0坐标各偏移10个像素
	offset := image.Pt(img.Bounds().Dx()-watermark.Bounds().Dx()-waterOffset, img.Bounds().Dy()-watermark.Bounds().Dy()-waterOffset)
	b := img.Bounds()
	m := image.NewNRGBA(b)

	draw.Draw(m, b, img, image.ZP, draw.Src)
	draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)

	//生成新图片new.jpg，并设置图片质量..
	imgw, _ := os.Create("d:/web/images/" + GetGuid() + ".jpg")
	jpeg.Encode(imgw, m, &jpeg.Options{100})

	defer imgw.Close()

	fmt.Println("水印添加成功")
}

//生成Guid字串
func GetGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
