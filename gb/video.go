package gb

type Video struct {
	videoRam Ram
	oam      Ram
	devs     []BusDev
}

func NewVideo() *Video {
	videoRam := NewRam(0x8000, 13)
	oam := NewRam(0xfe00, 9)
	devs := []BusDev{videoRam, oam}
	return &Video{*videoRam, *oam, devs}
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
