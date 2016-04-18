package gb

import "log"

type Sys struct {
	rom       ROM
	systemRAM SystemRAM
	hiRAM     RAM
	video     Video
	cpu       CPU
	devs      []BusDev
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
	return &Sys{rom, *systemRAM, *hiRAM, *video, *cpu, devs}
}

func (s *Sys) IER() uint8 {
	return s.Rb(0xffff)
}

func (s *Sys) Run() {
	for {
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

func (s *Sys) Rb(addr uint16) uint8 {
	rv := s.getHandler(addr).Rb(addr)
	log.Printf("Read attempt of %04Xh returned %02Xh\n", addr, rv)
	return rv
}

func (s *Sys) Wb(addr uint16, val uint8) {
	s.getHandler(addr).Wb(addr, val)
}

func (s *Sys) Rs(addr uint16) uint16 {
	return s.getHandler(addr).Rs(addr)
}

func (s *Sys) Ws(addr uint16, val uint16) {
	s.getHandler(addr).Ws(addr, val)
}
