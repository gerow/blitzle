package rom

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

type Ram struct {
	data []byte
}

var ramMask uint16 = 0x1fff
var hiRamMask uint16 = 0x007f

/*
0000 0
0001 1
0010 2
0011 3
0100 4
0101 5
0110 6
0111 7
1000 8
1001 9
1010 A
1011 B
1100 C
1101 D
1110 E
1111 F
*/

func (r *Ram) Rb(uint16 addr) uint8 {
	addr &= ramMask
	return r.data[addr]
}

func (r *Ram) Wb(uint16 addr, uint8 val) {
	addr &= ramMask
	r.data[addr] = val
}

func (r *Ram) Rs(uint16 addr) uint16 {
	addr &= ramMask
	return r.data[addr] | (r.data[addr+1] << 8)
}

func (r *Ram) Ws(uint16 addr, uint16 val) {
	addr &= ramMask
	r.data[addr] = val & 0xff
	r.data[addr+1] = (val >> 8) & 0xff
}

type HiRam struct {
	data []byte
}

func (h *HiRam) Rb(uint16 addr) uint8 {
	addr &= hiRamMask
	return h.data[addr]
}

func (h *HiRam) Wb(uint16 addr, uint8 val) {
	addr &= hiRamMask
	h.data[addr] = val
}

func (h *HiRam) Rs(uint16 addr) uint16 {
	addr &= hiRamMask
	return h.data[addr] | (r.data[addr+1] << 8)
}

func (h *HiRam) Ws(uint16 addr, uint16 val) {
	addr &= hiRamMask
	h.data[addr] = val & 0xff
	h.data[addr+1] = (val >> 8) & 0xff
}
