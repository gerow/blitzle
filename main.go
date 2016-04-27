package main

import (
	"flag"
	"fmt"
	"github.com/gerow/blitzle/gb"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

var imgNum int = 0
var colorMap map[gb.Pixel]uint8 = map[gb.Pixel]uint8{
	0: 255,
	1: 110,
	2: 64,
	3: 0,
}

func getColor(p gb.Pixel) *color.Gray {
	return &color.Gray{colorMap[p]}
}

func Swap(pixels [gb.LCDSizeX * gb.LCDSizeY]gb.Pixel) {
	if imgNum%30 != 0 {
		imgNum++
		return
	}
	out := image.NewGray(
		image.Rectangle{
			image.Point{0, 0},
			image.Point{int(gb.LCDSizeX), int(gb.LCDSizeY)}})
	for x := uint(0); x < gb.LCDSizeX; x++ {
		for y := uint(0); y < gb.LCDSizeY; y++ {
			out.Set(int(x), int(y), getColor(pixels[y*gb.LCDSizeX+x]))
		}
	}
	filename := fmt.Sprintf("%000000d.png", imgNum)
	imgNum++
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(file, out)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s ROM_FILE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	fn := flag.Args()[0]

	r, err := gb.LoadROMFromFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(r.Info())

	sys := gb.NewSys(*r, Swap)
	sys.Run()
}
