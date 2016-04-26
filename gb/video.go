package gb

import (
	"bytes"
	"fmt"
)

const bufferSizeX uint = 256
const bufferSizeY uint = 256
const lcdSizeX uint = 160
const lcdSizeY uint = 144

const lcdcRegAddr uint16 = 0xff40
const statRegAddr uint16 = 0xff41
const scyRegAddr uint16 = 0xff42
const scxRegAddr uint16 = 0xff43

// This value should only be from 0 to 3.
type Pixel uint8

/*
 * The vsync rate is around 9198 Hz, which gives us a good round 456 cycles per
 * hsync. Vsync is 59.73 Hz, and doing the math on that gives us 154 hsync
 * cycles, which is enough to write the 144 lines and then some
 * (which makes sense).
 *
 * This means that a full video cycle should take 70224 cycles.
 */
const hCycles int = 456
const totalCycles int = 70224
const vblankCycles int = 65664

const mode2Length int = 80
const mode3Length int = 172
const mode0Length int = 204

type Video struct {
	swap     SwapFunc
	videoRAM RAM
	oam      RAM
	devs     []BusDev

	buf [lcdSizeX * lcdSizeY]Pixel

	// Registers
	lcdc MemRegister       // FF40h
	stat LCDStatusRegister // FF41h
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

	currentCycle int
}

const oamAddr uint16 = 0xfe00
const oamSize uint16 = 40

type SwapFunc func(pixels [lcdSizeX * lcdSizeY]Pixel)

func NewVideo(swap SwapFunc) *Video {
	v := &Video{}
	v.swap = swap
	v.videoRAM = *NewRAM(0x8000, 0x9fff)
	v.oam = *NewRAM(0xfe00, 0xfe9f)
	v.lcdc = *NewMemRegister(0xff40)
	v.stat = LCDStatusRegister{v, 0}
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

type LCDStatusRegister struct {
	video *Video
	v     uint8
}

func (l *LCDStatusRegister) val() uint8 {
	// The actual timings seem to suggest that at the beginning of a line
	// we'll be in mode 2 for 80 cycles, followed by mode 3 for 172 cycles,
	// followed by mode 0 for 204 cycles, which adds up nicely to 456.
	flags := uint8(0)
	if l.video.currentCycle >= vblankCycles {
		flags = 1
	}
	hcycleNum := l.video.currentCycle % hCycles
	if hcycleNum < mode2Length {
		flags = 2
	} else if hcycleNum < mode2Length+mode3Length {
		flags = 3
	}
	if l.video.lyc.val() == l.video.regLY() {
		flags |= 0x4
	}
	// If none of these cases are true we're in mode 0, which lasts 204 cycles

	return l.v | flags
}

func (l *LCDStatusRegister) R(_ uint16) uint8 {
	return l.val()
}

func (l *LCDStatusRegister) W(addr uint16, v uint8) {
	// Mask off the bits that are read-only
	l.v = v & 0xf8
}

func (l *LCDStatusRegister) Asserts(addr uint16) bool {
	return addr == 0xff41
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
	lcdc := v.lcdc.val()
	if lcdc&0x80 == 0 {
		// We're disabled, so don't step at all!
		return
	}
	stat := v.stat.val()
	if v.currentCycle == vblankCycles {
		sys.RaiseInterrupt(VBlankInterrupt)
		// Apparently we can have LCDStatus fire for vsync too
		if stat&(1<<4) != 0 {
			sys.RaiseInterrupt(LCDStatInterrupt)
		}
		// We're done drawing lines, so send send the output up to
		// gl so it can munge it into a gl texture
		v.swap(v.buf)
	}
	// Interrupt for mode 2 OAM (which occurs at the beginning of a new line)
	if v.currentCycle%hCycles == 0 && stat&(1<<5) != 0 && v.currentCycle < vblankCycles {
		sys.RaiseInterrupt(LCDStatInterrupt)
	}
	// Interrupt for LCY==LY
	if v.currentCycle%hCycles == 0 && stat&(1<<6) != 0 && v.regLY() == v.lyc.val() {
		sys.RaiseInterrupt(LCDStatInterrupt)
	}
	if v.currentCycle%hCycles == mode2Length && v.currentCycle < vblankCycles {
		// We've reached the end of mode 2, so transition to mode 3 and
		// zap a line in place!
		v.drawLine(sys)
	}
	v.currentCycle++
	v.currentCycle %= totalCycles
}

func (v *Video) State(sys *Sys) string {
	o := bytes.Buffer{}
	o.WriteString(fmt.Sprintf("Registers:\n"))
	o.WriteString(fmt.Sprintf("  LY: %02Xh\n", sys.Rb(0xff44)))
	o.WriteString(fmt.Sprintf("Values:\n"))
	o.WriteString(fmt.Sprintf("  currentCycle: %v\n", v.currentCycle))

	return o.String()
}

func (v *Video) regLY() uint8 {
	return uint8(v.currentCycle / hCycles)
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

const bgMapWidth uint = 32
const bgMapHeight uint = 32
const tileWidth uint = 8
const tileHeight uint = 8

func (v *Video) bgMap(sys *Sys) []byte {
	return sys.ReadBytes(0x9800, 0x400)
}

func (v *Video) chrTiles(sys *Sys) []byte {
	return sys.ReadBytes(0x8000, 256*16)
}

func tilePix(chrTile []byte, x uint, y uint) Pixel {
	b := chrTile[y*2+x/8]
	inBIdx := x % 8
	return Pixel((b >> (6 - inBIdx)) & 0x03)
}

// Draw the current line to the buffer. We do this at the beginning of mode 3
// since that's when the gameboy no longer expects to be able to write to the
// OAM or video RAM (although we let it do so anyway).
func (v *Video) drawLine(sys *Sys) {
	ly := uint(v.ly.val())
	y := uint(v.scy.val()) + ly
	xStart := uint(v.scx.val())
	// Handle the background first
	tileRow := (uint(y) / bgMapHeight) % bgMapHeight
	tileY := uint(y) % tileHeight
	bgMap := v.bgMap(sys)
	chrTiles := v.chrTiles(sys)
	// Oh god is this ugly...
	for x := uint(0); x < lcdSizeX; x++ {
		tileColumn := ((xStart + x) / bgMapWidth) % bgMapWidth
		tileX := (xStart + x) % tileWidth
		tileNum := bgMap[tileRow*bgMapWidth+tileColumn]

		tileIdx := uint(tileNum) * 16
		tile := chrTiles[tileIdx : tileIdx+16]

		v.buf[ly*lcdSizeX+x] = tilePix(tile, tileX, tileY)
	}
}
