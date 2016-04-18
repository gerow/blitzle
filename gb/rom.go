package gb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
)

var expectedLogo []byte = []byte{
	0xCE, 0xED, 0x66, 0x66, 0xCC, 0x0D, 0x00, 0x0B, 0x03, 0x73, 0x00, 0x83,
	0x00, 0x0C, 0x00, 0x0D, 0x00, 0x08, 0x11, 0x1F, 0x88, 0x89, 0x00, 0x0E,
	0xDC, 0xCC, 0x6E, 0xE6, 0xDD, 0xDD, 0xD9, 0x99, 0xBB, 0xBB, 0x67, 0x63,
	0x6E, 0x0E, 0xEC, 0xCC, 0xDD, 0xDC, 0x99, 0x9F, 0xBB, 0xB9, 0x33, 0x3E,
}

var romMask uint16 = (1 << 15) - 1

type ROM struct {
	data       []byte
	title      string
	sgbSupport bool
	cartType   byte
	romSize    byte
	ramSize    byte
}

func LoadROM(fn string) (*ROM, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	var r ROM
	r.data = data
	r.title = string(r.data[0x0134:0x0144])
	r.sgbSupport = r.data[0x0146] == 0x03
	r.cartType = r.data[0x0147]
	r.romSize = r.data[0x0148]
	r.ramSize = r.data[0x0149]
	return &r, nil
}

func (r *ROM) HeaderChecksum() byte {
	x := 0
	for _, b := range r.data[0x0134:0x014d] {
		x = x - int(b) - 1
	}

	return byte(x & 0xff)
}

func (r *ROM) GlobalChecksum() uint16 {
	x := 0
	for _, b := range append(r.data[:0x014e], r.data[0x0150:]...) {
		x += int(b)
	}

	return uint16(x & 0xffff)
}

func (r *ROM) Info() string {
	o := bytes.Buffer{}
	l := len(r.data)
	o.WriteString(fmt.Sprintf("Title: %s\n", r.title))
	o.WriteString(fmt.Sprintf("Size: %d (0x%x)\n", l, l))
	logoCheck := "✗"
	if bytes.Equal(r.data[0x0104:0x0134], expectedLogo) {
		logoCheck = "✓"
	}
	o.WriteString(fmt.Sprintf("Logo: %s\n", logoCheck))
	sgbCheck := "✗"
	if r.sgbSupport {
		sgbCheck = "✓"
	}
	o.WriteString(fmt.Sprintf("Super Gameboy support: %s\n", sgbCheck))
	o.WriteString(fmt.Sprintf("Cartridge type: %02Xh\n", r.cartType))
	o.WriteString(fmt.Sprintf("ROM size: %02Xh\n", r.romSize))
	o.WriteString(fmt.Sprintf("RAM size: %02Xh\n", r.ramSize))
	destination := "Japanese"
	if r.data[0x14a] == 0x01 {
		destination = "Non-Japanese"
	}
	o.WriteString(fmt.Sprintf("Destination: %s\n", destination))
	o.WriteString(fmt.Sprintf("Mask ROM version: %02Xh\n", r.data[0x014c]))
	headerCheck := "✗"
	if r.HeaderChecksum() == r.data[0x014d] {
		headerCheck = "✓"
	}
	o.WriteString(fmt.Sprintf("Header checksum: %s\n", headerCheck))
	globalChecksum := (uint16(r.data[0x014e]) << 8) | (uint16(r.data[0x014f]))
	globalCheck := "✗"
	if r.GlobalChecksum() == globalChecksum {
		globalCheck = "✓"
	}
	o.WriteString(fmt.Sprintf("Global checksum: %s\n", globalCheck))

	return o.String()
}

func (r *ROM) Rb(addr uint16) uint8 {
	log.Printf("Read attempt of %04Xh in ROM\n", addr)
	return r.data[addr]
}

func (r *ROM) Wb(addr uint16, val uint8) {
	log.Println("Attempt to write to ROM at %04Xh with val %02Xh ignored", addr, val)
}

func (r *ROM) Rs(addr uint16) uint16 {
	return uint16(r.data[addr] | (r.data[addr+1] << 8))
}

func (r *ROM) Ws(addr uint16, val uint16) {
	log.Println("Attempt to write to ROM at %04Xh with val %04Xh ignored", addr, val)
}

func (r *ROM) Asserts(addr uint16) bool {
	return addr&^romMask == 0x0000
}
