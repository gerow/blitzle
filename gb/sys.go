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
	devs      []BusDev
	Stop      bool
}

type BusDev interface {
	Rb(addr uint16) uint8
	Wb(addr uint16, val uint8)
	Rs(addr uint16) uint16
	Ws(addr uint16, val uint16)
	Asserts(addr uint16) bool
}

type BusHole struct {
	startAddr uint16
	mask      uint16
}

func NewBusHole(startAddr uint16, addrBits uint8) *BusHole {
	mask := uint16((1 << addrBits) - 1)
	return &BusHole{startAddr, mask}
}

func (b *BusHole) Rb(addr uint16) uint8 {
	return 0xff
}

func (b *BusHole) Wb(addr uint16, val uint8) {
}

func (b *BusHole) Rs(addr uint16) uint16 {
	return 0xffff
}

func (b *BusHole) Ws(addr uint16, val uint16) {
}

func (b *BusHole) Asserts(addr uint16) bool {
	return addr&^b.mask == b.startAddr
}

func NewSys(rom ROM) *Sys {
	systemRAM := NewSystemRAM()
	hiRAM := NewHiRAM()
	video := NewVideo()
	cpu := NewCPU()
	bh1 := NewBusHole(0xa000, 13)
	devs := []BusDev{&rom, systemRAM, hiRAM, video, bh1}

	s := &Sys{rom, *systemRAM, *hiRAM, *video, *cpu, devs, false}
	s.SetPostBootloaderState()

	return s
}

func (s *Sys) IER() uint8 {
	return s.Rb(0xffff)
}

func (s *Sys) Run() {
	for {
		fmt.Print(s.cpu.State(s))
		s.cpu.Step(s)
	}
}

func (s *Sys) getHandler(addr uint16) BusDev {
	for _, bd := range s.devs {
		if bd.Asserts(addr) {
			return bd
		}
	}
	return nil
}

func (s *Sys) RbLog(addr uint16, l bool) uint8 {
	rv := s.getHandler(addr).Rb(addr)
	if l {
		log.Printf("R1 (%04Xh) => %02Xh\n", addr, rv)
	}
	return rv
}

func (s *Sys) WbLog(addr uint16, val uint8, l bool) {
	if l {
		log.Printf("W1 %02Xh => (%04Xh)\n", val, addr)
	}
	s.getHandler(addr).Wb(addr, val)
}

func (s *Sys) RsLog(addr uint16, l bool) uint16 {
	rv := s.getHandler(addr).Rs(addr)
	if l {
		log.Printf("R2 (%04Xh) => %04Xh\n", addr, rv)
	}
	return rv
}

func (s *Sys) WsLog(addr uint16, val uint16, l bool) {
	if l {
		log.Printf("W2 %04Xh => (%04Xh)\n", val, addr)
	}
	s.getHandler(addr).Ws(addr, val)
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
		s.Wb(addr, b)
		addr++
	}
}
