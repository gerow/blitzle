package frontend

import (
	"github.com/gerow/blitzle/gb"
	"github.com/veandco/go-sdl2/sdl"
	"unsafe"
)

type Frontend struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
}

var colorMap map[gb.Pixel]uint32 = map[gb.Pixel]uint32{
	0: 0xffffffff,
	1: 0x6e6e6eff,
	2: 0x404040ff,
	3: 0x000000ff,
}

func NewFrontend() (*Frontend, error) {
	window, err := sdl.CreateWindow(
		"Blitzle", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int(gb.LCDSizeX), int(gb.LCDSizeY), sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	renderer, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		return nil, err
	}
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_STREAMING, int(gb.LCDSizeX), int(gb.LCDSizeY))
	return &Frontend{window, renderer, texture}, nil
}

func getColor(p gb.Pixel) uint32 {
	return colorMap[p]
}

func (f *Frontend) Swap(pixels [gb.LCDSizeX * gb.LCDSizeY]gb.Pixel) {
	//out := image.NewRGBA(
	//	image.Rectangle{
	//		image.Point{0, 0},
	//		image.Point{int(gb.LCDSizeX), int(gb.LCDSizeY)}})
	var texPixels unsafe.Pointer
	var pitch int
	err := f.texture.Lock(nil, &texPixels, &pitch)
	if err != nil {
		panic(err)
	}
	out := (*[gb.LCDSizeX * gb.LCDSizeY]uint32)(texPixels)
	for x := uint(0); x < gb.LCDSizeX; x++ {
		for y := uint(0); y < gb.LCDSizeY; y++ {
			out[y*gb.LCDSizeX+x] = getColor(pixels[y*gb.LCDSizeX+x])
			(*[gb.LCDSizeX * gb.LCDSizeY]uint32)(texPixels)[y*uint(pitch/4)+x] = getColor(pixels[y*gb.LCDSizeX+x])
		}
	}
	/*
		newSurface, err := sdl.CreateRGBSurfaceFrom(
			unsafe.Pointer(&out.Pix[0]), out.Rect.Max.X, out.Rect.Max.Y,
			32, out.Stride, 0x000000FF, 0x0000FF00, 0x00FF0000, 0xFF000000)
		if err != nil {
			panic(err)
		}
		defer newSurface.Free()
	*/
	f.texture.Unlock()
	f.renderer.Copy(f.texture, nil, nil)
	f.renderer.Present()
	/*
		rect := sdl.Rect{0, 0, int32(gb.LCDSizeX), int32(gb.LCDSizeY)}
		newSurface.Blit(&rect, f.surface, &rect)
		f.window.UpdateSurface()
	*/
}
