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

const bankSize uint = 0x4000
const ramBankSize uint = 0x2000

var ramSizeMap map[byte]uint = map[byte]uint{
	0: 0,
	1: 0x800,
	2: 0x2000,
	3: 0x8000,
	4: 0x20000,
	5: 0x10000,
}

type ROM struct {
	data           []byte
	title          string
	sgbSupport     bool
	cartType       byte
	romSize        byte
	ramSize        byte
	banks          [][]byte
	currentBank    uint
	ram            []byte
	ramBanks       [][]byte
	currentRAMBank uint
	ramEnabled     bool
}

func LoadROMFromFile(fn string) (*ROM, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return LoadROM(data)
}

func LoadROM(data []byte) (*ROM, error) {
	var r ROM
	r.data = data
	r.title = string(r.data[0x0134:0x0144])
	r.sgbSupport = r.data[0x0146] == 0x03
	r.cartType = r.data[0x0147]
	r.romSize = r.data[0x0148]
	r.ramSize = r.data[0x0149]
	if uint(len(r.data))%bankSize != 0 {
		return nil, fmt.Errorf("wrongly sized banks")
	}

	// Record ROM banks
	nBanks := uint(len(r.data)) / bankSize
	r.banks = make([][]byte, nBanks)
	for i := uint(0); i < nBanks; i++ {
		addr := uint(i * bankSize)
		r.banks[i] = r.data[addr : addr+bankSize]
	}
	r.currentBank = 1

	// Record RAM banks
	//r.ram = make([]byte, ramSizeMap[r.ramSize])
	r.ram = make([]byte, ramSizeMap[5])
	nRamBanks := uint(len(r.ram)) / ramBankSize
	r.ramBanks = make([][]byte, nRamBanks)
	// Special case here since $01 is smaller than a full bank
	if nRamBanks == 1 {
		r.ramBanks[0] = r.ram
	} else {
		for i := uint(0); i < nRamBanks; i++ {
			addr := uint(i * ramBankSize)
			r.ramBanks[i] = r.ram[addr : addr+ramBankSize]
		}
	}

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
	for _, b := range r.data[:0x014e] {
		x += int(b)
	}

	for _, b := range r.data[0x0150:] {
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
	o.WriteString(fmt.Sprintf("4k banks in ROM file: %d\n", len(r.banks)))

	return o.String()
}

func (r *ROM) R(addr uint16) uint8 {
	if addr < 0x4000 {
		return r.data[addr]
	} else if addr < 0x8000 {
		return r.banks[r.currentBank][addr-0x4000]
	}
	//if !r.ramEnabled {
	//	log.Printf("!!! Attempt to read from cart RAM when disabled\n")
	//	return 0xff
	//}
	log.Printf("Read of %04Xh\n", addr)
	return r.ramBanks[r.currentRAMBank][addr-0xa000]
}

func (r *ROM) W(addr uint16, val uint8) {
	// Only works with MBC3 now
	if addr >= 0x0000 && addr < 0x2000 {
		if r.ramEnabled {
			log.Printf("Cartridge RAM was enabled\n")
		} else {
			log.Printf("Cartridge RAM was disabled\n")
		}
		r.ramEnabled = val&0x0a == 0x0a
		if r.ramEnabled {
			log.Printf("Cartridge RAM now enabled\n")
		} else {
			log.Printf("Cartridge RAM now disabled\n")
		}
		return
	}
	if addr >= 0x2000 && addr < 0x4000 {
		newBank := uint(val)
		if newBank == 0 {
			newBank = 1
		}
		if newBank/uint(len(r.banks)) != 0 {
			log.Printf("!!! Attempt to switch to bank beyond number in cart %04Xh\n", val)
		}
		r.currentBank = newBank % uint(len(r.banks))
		return
	}
	if addr >= 0x4000 && addr < 0x6000 {
		if len(r.ramBanks) == 0 {
			log.Printf("!!! Attempt to write to RAM bank on ROM with no RAM")
			return
		}
		newBank := uint(val)
		if newBank > uint(len(r.ramBanks))/ramBankSize {
			log.Printf("!!! Attempt to switch to ROM bank beyond number in cart %04Xh\n", val)
		}
		r.currentRAMBank = newBank % uint(len(r.ramBanks))
		return
	}
	if addr >= 0xa000 && addr < 0xc000 {
		//if !r.ramEnabled {
		//	log.Printf("!!! Attempt to write to cart RAM when disabled\n")
		//	return
		//}
		r.ramBanks[r.currentRAMBank][addr-0xa000] = val
		return
	}
	log.Printf("Attempt to write to ROM at %04Xh with val %02Xh ignored", addr, val)
}

func (r *ROM) Asserts(addr uint16) bool {
	return addr < 0x8000 || (addr >= 0xa000 && addr < 0xc000)
}

func (r *ROM) Dump() {
	for addr, b := range r.data {
		fmt.Printf("%04Xh: %02Xh\n", uint16(addr), b)
	}
}
