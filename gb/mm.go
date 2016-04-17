package mm

import (
	"github.com/gerow/blitzle/gb"
)

type MM struct {
	rom   rom.Rom
	ram   ram.Ram
	hiram ram.HiRam
}

type BusDev interface {
	Rb(int16 addr) uint8
	Wb(uint16 addr, uint8 val)
	Rs(int16 addr) uint16
	Ws(uint16 addr, uint16 val)
}

func (m *MM) getHandler(uint16 addr) *BusDev {
	if addr < 0x8000 {
		return &m.rom
	}
	if addr >= 0xc000 && addr < 0xfe00 {
		return &m.ram
	}
	if addr >= 0xff80 && addr < 0xffff {
		return &m.hiram
	}
}

func (m *MM) Rb(uint16 addr) uint8 {
	return getHandler(addr).Rb(addr)
}

func (m *MM) Wb(uint16 addr, uint8 val) {
	getHandler(addr).Wb(addr, val)
}

func (m *MM) Rs(uint16 addr) uint16 {
	return getHandler(addr).Rs(addr)
}

func (m *MM) Ws(uint16 addr, uint16 val) {
	getHandler(addr).Ws(addr, val)
}
