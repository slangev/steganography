package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"golang.org/x/image/bmp"
)

var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
var wordPtr *string = flag.String("s", "foo", "a string")
var bDecode *bool = flag.Bool("d", false, "decode flag")
var imagePath *string = flag.String("i", "./images/outimage.bmp", "path to image")
var asciiBitLength = 8

type Changeable interface {
	Set(x, y int, c color.Color)
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}

func convertTextToBitStream(covert string) []byte {
	bitStream := make([]byte, 0)
	for _, char := range *wordPtr {
		single := char & 0xFF
		count := 0
		for count < asciiBitLength {
			bit := byte(single&0x80) >> 7
			bitStream = append(bitStream, bit)
			single = (single << 1) & 0xFF
			count++
		}
	}
	//Add the zero byte as end of message
	for i := 0; i < 8; i++ {
		bitStream = append(bitStream, 0)
	}
	return bitStream
}

func createImageFile(img image.Image) {
	f, err := os.Create("./images/outimage.bmp")
	if err != nil {
		Error.Fatal(err)
	}
	defer f.Close()

	opt := bmp.Encode(f, img)
	if opt != nil {
		Error.Fatal(opt)
	}
}

func handleBitEncode(bit byte, currValue uint32) uint32 {
	if bit == 0 {
		if currValue%2 != 0 {
			if currValue == 0xFFFF {
				currValue--
			} else {
				currValue++
			}
		}
	} else {
		if currValue%2 == 0 {
			if currValue != 0 {
				currValue++
			} else {
				currValue--
			}
		}
	}
	return currValue
}

func encodeStream(bitStream []byte, img image.Image, cimg Changeable) {
	x := 0
	y := 0
	count := 0
	r, g, b, a := uint32(0), uint32(0), uint32(0), uint32(255)
	for len(bitStream) != 0 {
		bit := bitStream[0]
		bitStream = bitStream[1:]
		if count == 0 {
			rgb := img.At(x, y)
			r, g, b, a = rgb.RGBA()
		}

		//Handle first bit
		if count == 0 {
			r = handleBitEncode(bit, r)
		}

		//Handle second bit
		if count == 1 {
			g = handleBitEncode(bit, g)
		}

		//Handle third bit
		if count == 2 {
			b = handleBitEncode(bit, b)
		}

		count++

		c := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
		cimg.Set(x, y, c)
		if count == 3 {
			count = 0
			x++
		}

		if x == img.Bounds().Dx() {
			y++
			x = 0
		}
	}
}

func decodeStream(bitStream []byte) string {
	result := ""
	letterCount := len(bitStream) / 8
	for i := 0; i < letterCount; i++ {
		var letter byte
		for j := 0; j < asciiBitLength; j++ {
			bit := bitStream[0]
			bitStream = bitStream[1:]
			letter = (letter<<1 | bit)
		}
		result += string(letter)
	}

	return result
}

func getBitStreamImage(img image.Image) []byte {
	result := make([]byte, 0)
	countZero := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			rgb := img.At(x, y)
			r, g, b, _ := rgb.RGBA()
			if r%2 == 0 {
				countZero++
				result = append(result, 0)
			} else {
				countZero = 0
				result = append(result, 1)
			}
			if countZero == 8 {
				return result
			}
			if g%2 == 0 {
				countZero++
				result = append(result, 0)
			} else {
				countZero = 0
				result = append(result, 1)
			}
			if countZero == 8 {
				return result
			}
			if b%2 == 0 {
				countZero++
				result = append(result, 0)
			} else {
				countZero = 0
				result = append(result, 1)
			}
			if countZero == 8 {
				return result
			}
		}
	}

	return make([]byte, 0)
}

func main() {
	flag.Parse()

	fmt.Println("word:", *wordPtr)
	fmt.Println("decode:", *bDecode)
	img, err := getImageFromFilePath(*imagePath)
	if err != nil {
		Error.Fatal(err)
	}
	fmt.Println(img.Bounds())
	fmt.Println(img.Bounds().Dx())
	fmt.Println(img.Bounds().Dy())
	if *bDecode {
		bitStream := getBitStreamImage(img)
		message := decodeStream(bitStream)
		fmt.Println(message)
	} else {
		bitStream := convertTextToBitStream(*wordPtr)
		if cimg, ok := img.(Changeable); ok {
			encodeStream(bitStream, img, cimg)
			// when done, save img as usual
			createImageFile(img)
		}
	}
}
