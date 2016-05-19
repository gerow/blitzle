package frontend

import (
	"fmt"
	"github.com/gerow/blitzle/gb"
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"log"
	"unsafe"
)

type Frontend struct {
	updateButtonser  gb.UpdateButtonser
	window           *sdl.Window
	renderer         *sdl.Renderer
	texture          *sdl.Texture
	eventWatchHandle sdl.EventWatchHandle
	buttonState      gb.ButtonState
}

var colorMap map[gb.Pixel]uint32 = map[gb.Pixel]uint32{
	0: 0xffffffff,
	1: 0x6e6e6eff,
	2: 0x404040ff,
	3: 0x000000ff,
}

func NewFrontend(updateButtonser gb.UpdateButtonser) (*Frontend, error) {
	window, err := sdl.CreateWindow(
		"Blitzle", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	renderer, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		return nil, err
	}
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_STREAMING, int(gb.LCDSizeX), int(gb.LCDSizeY))
	f := &Frontend{updateButtonser, window, renderer, texture, 0, gb.ButtonState{}}
	f.eventWatchHandle = sdl.AddEventWatch(f.FilterEvent)
	return f, nil
}

func getColor(p gb.Pixel) uint32 {
	return colorMap[p]
}

func (f *Frontend) VideoSwap(pixels [gb.LCDSizeX * gb.LCDSizeY]gb.Pixel) {
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
		}
	}
	f.texture.Unlock()
	f.renderer.Copy(f.texture, nil, nil)
	f.renderer.Present()
}

func (f *Frontend) Close() {
	sdl.DelEventWatch(f.eventWatchHandle)
}

var buttonName = map[sdl.Scancode]string{
	sdl.SCANCODE_DOWN:   "down",
	sdl.SCANCODE_UP:     "up",
	sdl.SCANCODE_LEFT:   "left",
	sdl.SCANCODE_RIGHT:  "right",
	sdl.SCANCODE_RETURN: "return",
	sdl.SCANCODE_RSHIFT: "rshift",
	sdl.SCANCODE_X:      "X",
	sdl.SCANCODE_Z:      "Z",
}

func (f *Frontend) getKey(code sdl.Scancode) *bool {
	switch code {
	case sdl.SCANCODE_DOWN:
		return &f.buttonState.Down
	case sdl.SCANCODE_UP:
		return &f.buttonState.Up
	case sdl.SCANCODE_LEFT:
		return &f.buttonState.Left
	case sdl.SCANCODE_RIGHT:
		return &f.buttonState.Right
	case sdl.SCANCODE_RETURN:
		return &f.buttonState.Start
	case sdl.SCANCODE_RSHIFT:
		return &f.buttonState.SelectButton
	case sdl.SCANCODE_X:
		return &f.buttonState.B
	case sdl.SCANCODE_Z:
		return &f.buttonState.A
	}
	return nil
}

func (f *Frontend) FilterEvent(e sdl.Event) bool {
	switch v := e.(type) {
	case *sdl.KeyDownEvent:
		if k := f.getKey(v.Keysym.Scancode); k != nil {
			*k = true
			fmt.Printf("%s pressed\n", buttonName[v.Keysym.Scancode])
			f.updateButtonser.UpdateButtons(f.buttonState)
		}
	case *sdl.KeyUpEvent:
		if k := f.getKey(v.Keysym.Scancode); k != nil {
			*k = false
			fmt.Printf("%s released\n", buttonName[v.Keysym.Scancode])
			f.updateButtonser.UpdateButtons(f.buttonState)
		}
	}
	return false
}

type WriterSerialSwapper struct {
	Writer io.Writer
}

func (w *WriterSerialSwapper) SerialSwap(out uint8) uint8 {
	realOut := []byte{byte(out)}
	_, err := w.Writer.Write(realOut)
	if err != nil {
		log.Printf("Failed to swap serial: %v\n", err)
	}
	return 0xff
}
