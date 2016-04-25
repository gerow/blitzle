package gb

import (
	"fmt"
	"log"
)

type Sys struct {
	rom       ROM
	systemRAM SystemRAM
	hiRAM     RAM
	video     Video
	cpu       CPU
	/* Interrupt controller registers */
	ieReg MemRegister
	ifReg MemRegister
	devs  []BusDev
	Stop  bool

	wall    int
	cpuWait int
}

type BusDev interface {
	R(addr uint16) uint8
	W(addr uint16, val uint8)
	Asserts(addr uint16) bool
}

type BusHole struct {
	startAddr uint16
	endAddr   uint16
}

func NewBusHole(startAddr uint16, endAddr uint16) *BusHole {
	return &BusHole{startAddr, endAddr}
}

func (b *BusHole) R(addr uint16) uint8 {
	return 0xff
}

func (b *BusHole) W(addr uint16, val uint8) {
}

func (b *BusHole) Asserts(addr uint16) bool {
	return addr >= b.startAddr && addr <= b.endAddr
}

func NewSys(rom ROM) *Sys {
	systemRAM := NewSystemRAM()
	hiRAM := NewHiRAM()
	video := NewVideo()
	cpu := NewCPU()
	bh1 := NewBusHole(0xa000, 0xbfff)
	bh2 := NewBusHole(0xfea0, 0xff7f)
	ieReg := NewMemRegister(0xffff)
	ifReg := NewMemRegister(0xff0f)
	devs := []BusDev{&rom, systemRAM, hiRAM, video, bh1, bh2, ieReg, ifReg}

	s := &Sys{
		rom,
		*systemRAM,
		*hiRAM,
		*video,
		*cpu,
		*ieReg,
		*ifReg,
		devs,
		false,
		0,
		0}
	s.SetPostBootloaderState()

	return s
}

func (s *Sys) IER() uint8 {
	return s.Rb(0xffff)
}

func (s *Sys) Run() {
	for {
		s.Step()
	}
}

// Step one clock cycle.
func (s *Sys) Step() {
	if s.cpuWait == 0 {
		s.cpuWait = s.cpu.Step(s)
		fmt.Print(s.cpu.State(s))
	} else {
		s.cpuWait--
	}
	s.wall++
}

func (s *Sys) getHandler(addr uint16) BusDev {
	for _, bd := range s.devs {
		if bd.Asserts(addr) {
			return bd
		}
	}
	log.Fatalf("Couldn't find handler for addr %04Xh\n", addr)
	return nil
}

func (s *Sys) RbLog(addr uint16, l bool) uint8 {
	rv := s.getHandler(addr).R(addr)
	if l {
		log.Printf("R1 (%04Xh) => %02Xh\n", addr, rv)
	}
	return rv
}

func (s *Sys) WbLog(addr uint16, val uint8, l bool) {
	if l {
		log.Printf("W1 %02Xh => (%04Xh)\n", val, addr)
	}
	s.getHandler(addr).W(addr, val)
}

func (s *Sys) RsLog(addr uint16, l bool) uint16 {
	rv := uint16(s.RbLog(addr, false)) | uint16(s.RbLog(addr+1, false))<<8
	if l {
		log.Printf("R2 (%04Xh) => %04Xh\n", addr, rv)
	}
	return rv
}

func (s *Sys) WsLog(addr uint16, val uint16, l bool) {
	if l {
		log.Printf("W2 %04Xh => (%04Xh)\n", val, addr)
	}
	s.WbLog(addr, uint8(val), false)
	s.WbLog(addr+1, uint8(val<<8), false)
}

func (s *Sys) Rb(addr uint16) uint8 {
	return s.RbLog(addr, true)
}

func (s *Sys) Wb(addr uint16, val uint8) {
	s.WbLog(addr, val, true)
}

func (s *Sys) Rs(addr uint16) uint16 {
	return s.RsLog(addr, true)
}

func (s *Sys) Ws(addr uint16, val uint16) {
	s.WsLog(addr, val, true)
}

func (s *Sys) SetPostBootloaderState() {
	s.cpu.SetPostBootloaderState(s)
}

func (s *Sys) WriteBytes(bytes []byte, addr uint16) {
	for _, b := range bytes {
		s.WbLog(addr, b, false)
		addr++
	}
}
