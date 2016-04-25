package gb

const bufferSizeX int = 256
const bufferSizeY int = 256
const lcdSizeX int = 160
const lcdSizeY int = 144

const lcdcRegAddr uint16 = 0xff40
const statRegAddr uint16 = 0xff41
const scyRegAddr uint16 = 0xff42
const scxRegAddr uint16 = 0xff43

// This value should only be from 0 to 3.
type Pixel uint8

type Video struct {
	videoRAM RAM
	oam      RAM
	devs     []BusDev

	buf [bufferSizeX * bufferSizeY]Pixel
	out [lcdSizeX * lcdSizeY]Pixel

	// Registers
	lcdc MemRegister // FF40h
	//stat uint8       // FF41h
	scy  MemRegister       // FF42h
	scx  MemRegister       // FF43h
	ly   ReadOnlyRegister  // FF44h
	lyc  MemRegister       // FF45h
	dma  WriteOnlyRegister // FF46h
	bgp  MemRegister       // FF47h
	obp0 MemRegister       // FF48h
	obp1 MemRegister       // FF49h
	wy   MemRegister       // FF4ah
	wx   MemRegister       // FF4bh

	doDma  bool
	dmaSrc uint16
}

const oamAddr uint16 = 0xfe00
const oamSize uint16 = 40

func NewVideo() *Video {
	v := &Video{}
	v.videoRAM = *NewRAM(0x8000, 0x9fff)
	v.oam = *NewRAM(0xfe00, 0xfe9f)
	v.lcdc = *NewMemRegister(0xff40)
	v.lcdc.set(0x91)
	v.scy = *NewMemRegister(0xff42)
	v.scx = *NewMemRegister(0xff43)
	v.ly = ReadOnlyRegister{0xff44, v.regLY}
	v.lyc = *NewMemRegister(0xff45)
	v.dma = WriteOnlyRegister{0xff46, v.dmaW}
	v.bgp = *NewMemRegister(0xff47)
	v.bgp.set(0xfc)
	v.obp0 = *NewMemRegister(0xff48)
	v.obp0.set(0xff)
	v.obp1 = *NewMemRegister(0xff49)
	v.obp1.set(0xff)
	v.wy = *NewMemRegister(0xff4a)
	v.wx = *NewMemRegister(0xff4b)
	v.devs = []BusDev{
		&v.videoRAM,
		&v.oam,
		&v.lcdc,
		&v.scy,
		&v.scx,
		&v.ly,
		&v.lyc,
		&v.dma,
		&v.bgp,
	}

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

func (v *Video) R(addr uint16) uint8 {
	return v.getHandler(addr).R(addr)
}

func (v *Video) W(addr uint16, val uint8) {
	v.getHandler(addr).W(addr, val)
}

func (v *Video) Asserts(addr uint16) bool {
	return v.getHandler(addr) != nil
}

func (v *Video) Step(sys *Sys) {
	if v.doDma {
		v.doDma = false
		b := sys.ReadBytes(v.dmaSrc, oamSize)
		sys.WriteBytes(b, oamAddr)
	}
}

func (v *Video) regLY() uint8 {
	return 0
}

func (v *Video) dmaW(val uint8) {
	v.doDma = true
	v.dmaSrc = uint16(val) * 0x100
}

const nOAMblocks uint = 40
const OAMblockSize uint = 4

type OAMblock struct {
	y       uint8
	x       uint8
	pattern uint8
	flags   uint8
}

func (o *OAMblock) priority() bool {
	return o.flags&0x80 != 0
}

func (o *OAMblock) yFlip() bool {
	return o.flags&0x40 != 0
}

func (o *OAMblock) xFlip() bool {
	return o.flags&0x20 != 0
}

func (o *OAMblock) palette() bool {
	return o.flags&0x10 != 0
}

func (v *Video) oamBlocks() *[nOAMblocks]OAMblock {
	i := uint(0)
	out := [nOAMblocks]OAMblock{}
	for n := uint(0); n < nOAMblocks; n++ {
		out[n] = OAMblock{
			v.oam.data[i],
			v.oam.data[i+1],
			v.oam.data[i+2],
			v.oam.data[i+3]}
		i += OAMblockSize
	}

	return &out
}
