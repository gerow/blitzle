package main

import (
	"flag"
	"fmt"
	"github.com/gerow/blitzle/gb"
	"github.com/veandco/go-sdl2/sdl"
	"image"
	"image/color"
	"log"
	"os"
	"runtime"
	"unsafe"
)

var imgNum int = 0
var colorMap map[gb.Pixel]*color.RGBA = map[gb.Pixel]*color.RGBA{
	0: &color.RGBA{255, 255, 255, 255},
	1: &color.RGBA{110, 110, 110, 255},
	2: &color.RGBA{64, 64, 64, 255},
	3: &color.RGBA{0, 0, 0, 255},
}
var surface *sdl.Surface
var window *sdl.Window
var sys *gb.Sys

func getColor(p gb.Pixel) *color.RGBA {
	return colorMap[p]
}

func Swap(pixels [gb.LCDSizeX * gb.LCDSizeY]gb.Pixel) {
	fmt.Printf("wall: %d\n", sys.Wall)
	out := image.NewRGBA(
		image.Rectangle{
			image.Point{0, 0},
			image.Point{int(gb.LCDSizeX), int(gb.LCDSizeY)}})
	for x := uint(0); x < gb.LCDSizeX; x++ {
		for y := uint(0); y < gb.LCDSizeY; y++ {
			out.Set(int(x), int(y), getColor(pixels[y*gb.LCDSizeX+x]))
		}
	}
	newSurface, err := sdl.CreateRGBSurfaceFrom(
		unsafe.Pointer(&out.Pix[0]), out.Rect.Max.X, out.Rect.Max.Y, 32, out.Stride,
		0x000000FF, 0x0000FF00, 0x00FF0000, 0xFF000000)
	if err != nil {
		panic(err)
	}
	defer newSurface.Free()

	rect := sdl.Rect{0, 0, int32(gb.LCDSizeX), int32(gb.LCDSizeY)}
	newSurface.Blit(&rect, surface, &rect)
	window.UpdateSurface()
}

func main() {
	runtime.LockOSThread()
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

	sys = gb.NewSys(r, Swap)

	sdl.Init(sdl.INIT_EVERYTHING)
	window, err = sdl.CreateWindow("Blitzle", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int(gb.LCDSizeX), int(gb.LCDSizeY), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err = window.GetSurface()
	if err != nil {
		panic(err)
	}
	sys.Run()
}
