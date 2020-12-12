package main

import (
	"bufio"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"github.com/robfig/graphics-go/graphics"
)

type Metadata struct {
	textureName string
	with int
	height int
	pixelFormat string
}

type Rect struct {
	offX int
	offY int
	x int
	y int
	w int
	h int
}

type Frame struct {
	name string
	rect Rect
	rotated bool
}

type Info struct {
	frames []Frame
	metadata Metadata
	pngPath string
	OutPngPath string
}

func trim(content []byte) string {
	text := string(content)
	if idx := strings.Index(text, "<key>frames</key>"); idx != -1 {
		text = text[idx + len("<key>frames</key>"):]
	}
	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.Replace(text, "\n", " ", -1)
	re, err := regexp.Compile("\\s+")
	if err != nil {
		log.Fatal(err)
	}
	text = re.ReplaceAllString(text, " ")
	return text
}

func clipImage(pngPath string, frame *Frame) *image.NRGBA {
	f, err := os.Open(pngPath + "/" + frame.name)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	if img, _, err := image.Decode(bufio.NewReader(f)); err == nil {
		offX := (img.Bounds().Dx()-frame.rect.w)/2 + frame.rect.offX
		offY := (img.Bounds().Dy()-frame.rect.h)/2 - frame.rect.offY
		newImg := image.NewNRGBA(image.Rect(0, 0, frame.rect.w, frame.rect.h))
		dstImg := image.NewNRGBA(image.Rect(0, 0, frame.rect.h, frame.rect.w))
		for x := 0; x < frame.rect.w; x++ {
			for y := 0; y < frame.rect.h; y++ {
				newImg.Set(x, y, img.At(offX+x, offY+y))
			}
		}
		graphics.Rotate(dstImg, newImg, &graphics.RotateOptions{
			Angle: math.Pi / 2,
		})
		return dstImg
	}
	return nil
}

func run(info *Info) {
	newImg := image.NewNRGBA(image.Rect(0, 0, info.metadata.with, info.metadata.height))
	for _, frame := range info.frames {
		f, err := os.Open(info.pngPath + "/" + frame.name)
		if err != nil {
			log.Fatal(err)
			return
		}
		if img, _, err := image.Decode(bufio.NewReader(f)); err == nil {
			if frame.rotated {
				log.Println("rotated.", info.metadata.textureName, frame.name)
				img := clipImage(info.pngPath, &frame)
				for x := 0; x < frame.rect.h; x++ {
					for y := 0; y < frame.rect.w; y++ {
						newImg.Set(x+frame.rect.x, y+frame.rect.y, img.At(x, y))
						//newImg.Set(x+frame.rect.x, y+frame.rect.y, img.At(offX+y, offY+frame.rect.offX-x-1))
					}
				}
			} else {
				offX := (img.Bounds().Dx() - frame.rect.w) / 2 + frame.rect.offX
				offY := (img.Bounds().Dy() - frame.rect.h) / 2 - frame.rect.offY
				for x := 0; x < frame.rect.w; x++ {
					for y := 0; y < frame.rect.h; y++ {
						newImg.Set(x+frame.rect.x, y+frame.rect.y, img.At(offX+x, offY+y))
					}
				}
			}
		} else {
			log.Fatal(err)
			return
		}
		f.Close()
	}
	outPath := info.OutPngPath + "/n_" + info.metadata.textureName
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	if png.Encode(outFile, newImg); err != nil {
		log.Fatal(err)
	}
}

func main()  {
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	text := trim(content)

	reg1, err := regexp.Compile("\\<key\\>([^\\>]+)\\</key\\> <dict>(.+?)(</dict>)")
	if err != nil {
		log.Fatal(err)
	}

	regRect, err := regexp.Compile("\\<key\\>textureRect</key\\> <string>{{(\\d+),(\\d+)},{(\\d+),(\\d+)}}(</string>)")
	if err != nil {
		log.Fatal(err)
	}

	regOffset, err := regexp.Compile("\\<key\\>spriteOffset</key\\> <string>{([-0-9]+),([-0-9]+)}(</string>)")
	if err != nil {
		log.Fatal(err)
	}

	regRotated, err := regexp.Compile("\\<key\\>textureRotated</key\\> \\<(\\w+)/\\>")
	if err != nil {
		log.Fatal(err)
	}

	regSize, err := regexp.Compile("\\<key\\>size</key\\> <string>{(\\d+),(\\d+)}(</string>)")
	if err != nil {
		log.Fatal(err)
	}

	regFileName, err := regexp.Compile("\\<key\\>realTextureFileName</key\\> <string>(.+?)(</string>)")
	if err != nil {
		log.Fatal(err)
	}

	info := Info{
		frames: make([]Frame, 0, 1024),
		pngPath : os.Args[2],
		OutPngPath : os.Args[3],
	}

	result1 := reg1.FindAllStringSubmatch(text, -1)
	for _, result := range result1 {
		if result[1] == "frames" {
			continue
		}
		if result[1] == "metadata" {
			resultSize := regSize.FindStringSubmatch(result[2])
			resultFileName := regFileName.FindStringSubmatch(result[2])
			w, err := strconv.Atoi(resultSize[1])
			if err != nil {
				log.Fatal(err)
			}
			h, err := strconv.Atoi(resultSize[2])
			if err != nil {
				log.Fatal(err)
			}
			info.metadata.with = w
			info.metadata.height = h
			info.metadata.textureName = resultFileName[1]
			continue
		}
		//fmt.Print(result[1])
		result1Rect := regRect.FindStringSubmatch(result[2])
		resultOffset := regOffset.FindStringSubmatch(result[2])
		resultRotated := regRotated.FindStringSubmatch(result[2])
		if result1Rect != nil && resultRotated != nil && resultOffset != nil {
			//fmt.Println("", result1Rect[1], result1Rect[2], result1Rect[3], result1Rect[4], resultRotated[1])
			x, err := strconv.Atoi(result1Rect[1])
			if err != nil {
				log.Fatal(err)
			}
			y, err := strconv.Atoi(result1Rect[2])
			if err != nil {
				log.Fatal(err)
			}
			w, err := strconv.Atoi(result1Rect[3])
			if err != nil {
				log.Fatal(err)
			}
			h, err := strconv.Atoi(result1Rect[4])
			if err != nil {
				log.Fatal(err)
			}
			offX, err := strconv.Atoi(resultOffset[1])
			if err != nil {
				log.Fatal(err)
			}
			offY, err := strconv.Atoi(resultOffset[2])
			if err != nil {
				log.Fatal(err)
			}
			info.frames = append(info.frames, Frame{
				name: result[1],
				rect: Rect{
					x: x,
					y: y,
					w: w,
					h: h,
					offX: offX,
					offY: offY,
				},
				rotated: resultRotated[1] == "true",
			})
		}
	}
	run(&info)
	//fmt.Println("info:", info)
	//fmt.Println(text)
}