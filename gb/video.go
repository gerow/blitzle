package gb

const bufferSizeX int = 256
const bufferSizeY int = 256
const lcdSizeX int = 160
const lcdSizeY int = 144

// This value should only be from 0 to 3.
type Pixel uint8

type Video struct {
	videoRAM RAM
	oam      RAM
	devs     []BusDev

	buf [bufferSizeX * bufferSizeY]Pixel
	out [lcdSizeX * lcdSizeY]Pixel
}

func NewVideo() *Video {
	v := &Video{}
	v.videoRAM = *NewRAM(0x8000, 13)
	v.oam = *NewRAM(0xfe00, 9)
	v.devs = []BusDev{&v.videoRAM, &v.oam}

	return v
}

func (v *Video) getHandler(addr uint16) BusDev {
	for _, bd := range v.devs {
		if bd.Asserts(addr) {
			return bd
		}
	}
	return nil
}

func (v *Video) Rb(addr uint16) uint8 {
	return v.getHandler(addr).Rb(addr)
}

func (v *Video) Wb(addr uint16, val uint8) {
	v.getHandler(addr).Wb(addr, val)
}

func (v *Video) Rs(addr uint16) uint16 {
	return v.getHandler(addr).Rs(addr)
}

func (v *Video) Ws(addr uint16, val uint16) {
	v.getHandler(addr).Ws(addr, val)
}

func (v *Video) Asserts(addr uint16) bool {
	return v.getHandler(addr) != nil
}
