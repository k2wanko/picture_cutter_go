package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"strconv"

	"github.com/nfnt/resize"

	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	const imageMax float64 = 128.0
	const splitMax float64 = 4.0

	// image file path
	filePath := os.Args[1]
	base := filepath.Base(filePath)
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)
	name := base[0 : len(base)-len(ext)]

	imageFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer imageFile.Close()

	var decodedImage image.Image

	// decide file type from file extention
	if ext == ".jpg" || ext == ".jpeg" {
		decodedImage, err = jpeg.Decode(imageFile)
		if err != nil {
			log.Fatal(err)
		}
	} else if ext == ".png" {
		decodedImage, err = png.Decode(imageFile)
		if err != nil {
			log.Fatal(err)
		}
	} else if ext == ".gif" {
		decodedImage, err = gif.Decode(imageFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		return
	}

	width := float64(decodedImage.Bounds().Max.X)
	height := float64(decodedImage.Bounds().Max.Y)

	var ratio float64
	if width > height && width > imageMax*splitMax {
		ratio = imageMax * splitMax / width
	} else if height > imageMax*splitMax {
		ratio = imageMax * splitMax / height
	} else {
		ratio = 1
	}

	if ratio != 1 {
		decodedImage = resize.Resize(uint(math.Floor(width*ratio)), uint(math.Floor(height*ratio)),
			decodedImage, resize.Lanczos3)
	}

	widthCount := int(math.Ceil(width / imageMax))
	heightCount := int(math.Ceil(height / imageMax))
	dispTxt := &bytes.Buffer{}
	for h := 1; h <= heightCount; h++ {
		for w := 1; w <= widthCount; w++ {
			fileName := name + "_" + strconv.Itoa(int(h)) + "_" + strconv.Itoa(int(w))
			fmt.Fprintf(dispTxt, ":%s:", fileName)
			tmpFile, err := os.Create(fileName + ext)
			if err != nil {
				log.Fatal(err)
			}
			defer tmpFile.Close()

			tmpHeight := int(imageMax) * h
			if tmpHeight > decodedImage.Bounds().Max.Y {
				tmpHeight = decodedImage.Bounds().Max.Y
			}
			tmpWidth := int(imageMax) * w
			if tmpWidth > decodedImage.Bounds().Max.X {
				tmpWidth = decodedImage.Bounds().Max.X
			}
			var tmpBounds image.Rectangle
			tmpBounds = image.Rect(int(imageMax)*(w-1), int(imageMax)*(h-1), tmpWidth, tmpHeight)
			subImage := decodedImage.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(tmpBounds)

			if tmpHeight < int(imageMax)*h || tmpWidth < int(imageMax)*w {
				tmpBounds = image.Rect(int(imageMax)*(w-1), int(imageMax)*(h-1), int(imageMax)*w, int(imageMax)*h)
				tmpRect := image.Rect(0, 0, int(imageMax), int(imageMax))
				tmpImage := image.NewRGBA(tmpRect)
				b := subImage.Bounds()
				m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
				draw.Draw(m, m.Bounds(), subImage, b.Min, draw.Src)
				if ext == ".jpg" || ext == ".jpeg" {
					draw.Draw(tmpImage, tmpRect, &image.Uniform{color.White}, image.ZP, draw.Over)
				} else {
					draw.Draw(tmpImage, tmpRect, &image.Uniform{color.Transparent}, image.ZP, draw.Over)
				}
				draw.Draw(tmpImage, tmpRect, m, image.ZP, draw.Src)
				subImage = tmpImage
			}
			if ext == ".jpg" || ext == ".jpeg" {
				jpeg.Encode(tmpFile, subImage, nil)
			} else if ext == ".png" {
				png.Encode(tmpFile, subImage)
			} else if ext == ".gif" {
				gif.Encode(tmpFile, subImage, nil)
			}

		}
		dispTxt.Write([]byte("\n"))
	}

	fmt.Print(dispTxt.String())
}
