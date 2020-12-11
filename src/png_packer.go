package main

import (
	"bufio"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)
type Metadata struct {
	textureName string
	with int
	height int
	pixelFormat string
}

type Rect struct {
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

func run(info *Info) {
	newImg := image.NewNRGBA(image.Rect(0, 0, info.metadata.with, info.metadata.height))
	for _, frame := range info.frames {
		f, err := os.Open(info.pngPath + "/" + frame.name)
		if err != nil {
			log.Fatal(err)
			return
		}
		if img, _, err := image.Decode(bufio.NewReader(f)); err == nil {
			dx := img.Bounds().Dx()
			dy := img.Bounds().Dy()
			for x := 0; x < dx; x++ {
				for y := 0; y < dy; y++ {
					if frame.rotated {
						log.Println("rotated.", info.metadata.textureName, frame.name)
						newImg.Set(x+frame.rect.x, y+frame.rect.y, img.At(y, x))
					} else {
						newImg.Set(x+frame.rect.x, y+frame.rect.y, img.At(x, y))
					}
				}
			}
		} else {
			log.Fatal(err)
			return
		}
		f.Close()
	}
	outPath := info.OutPngPath + "/" + info.metadata.textureName
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
		resultRotated := regRotated.FindStringSubmatch(result[2])
		if result1Rect != nil && resultRotated != nil {
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
			info.frames = append(info.frames, Frame{
				name: result[1],
				rect: Rect{
					x: x,
					y: y,
					w: w,
					h: h,
				},
				rotated: resultRotated[1] == "true",
			})
		}
	}
	run(&info)
	//fmt.Println("info:", info)
	//fmt.Println(text)
}