package gb

type RAM struct {
	startAddr uint16
	data      []byte
	mask      uint16
}

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
:010 A
1011 B
1100 C
1101 D
1110 E
1111 F
*/

func NewRAM(startAddr uint16, addrBits uint8) *RAM {
	mask := uint16((1 << addrBits) - 1)
	size := 2 << addrBits
	return &RAM{startAddr, make([]byte, size), mask}
}

func (r *RAM) Rb(addr uint16) uint8 {
	addr &= r.mask
	return r.data[addr]
}

func (r *RAM) Wb(addr uint16, val uint8) {
	addr &= r.mask
	r.data[addr] = val
}

func (r *RAM) Rs(addr uint16) uint16 {
	addr &= r.mask
	return uint16(r.data[addr] | (r.data[addr+1] << 8))
}

func (r *RAM) Ws(addr uint16, val uint16) {
	addr &= r.mask
	r.data[addr] = uint8(val & 0xff)
	r.data[addr+1] = uint8((val >> 8) & 0xff)
}

func (r *RAM) Asserts(addr uint16) bool {
	return addr&r.mask == r.startAddr
}

func NewHiRAM() *RAM {
	return NewRAM(0xff80, 7)
}

/*
 * System RAM needs to assert specially in order to properly do the mirrioring
 * nonsense.
 */
type SystemRAM struct {
	ram RAM
}

func NewSystemRAM() *SystemRAM {
	return &SystemRAM{*NewRAM(0xc000, 13)}
}

func (sr *SystemRAM) Rb(addr uint16) uint8 {
	return sr.ram.Rb(addr)
}

func (sr *SystemRAM) Wb(addr uint16, val uint8) {
	sr.ram.Wb(addr, val)
}

func (sr *SystemRAM) Rs(addr uint16) uint16 {
	return sr.ram.Rs(addr)
}

func (sr *SystemRAM) Ws(addr uint16, val uint16) {
	sr.ram.Ws(addr, val)
}

func (sr *SystemRAM) Asserts(addr uint16) bool {
	/* We need to skip some bits for the OAM */
	return addr >= 0xc000 && addr < 0xfe00
}
