package gb

type Ram struct {
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

func NewRam(startAddr uint16, addrBits uint8) *Ram {
	mask := uint16((1 << addrBits) - 1)
	size := 2 << addrBits
	return &Ram{startAddr, make([]byte, size), mask}
}

func (r *Ram) Rb(addr uint16) uint8 {
	addr &= r.mask
	return r.data[addr]
}

func (r *Ram) Wb(addr uint16, val uint8) {
	addr &= r.mask
	r.data[addr] = val
}

func (r *Ram) Rs(addr uint16) uint16 {
	addr &= r.mask
	return uint16(r.data[addr] | (r.data[addr+1] << 8))
}

func (r *Ram) Ws(addr uint16, val uint16) {
	addr &= r.mask
	r.data[addr] = uint8(val & 0xff)
	r.data[addr+1] = uint8((val >> 8) & 0xff)
}

func (r *Ram) Asserts(addr uint16) bool {
	return addr&r.mask == r.startAddr
}

func NewHiRam() *Ram {
	return NewRam(0xff80, 7)
}

/*
 * System RAM needs to assert specially in order to properly do the mirrioring
 * nonsense.
 */
type SystemRam struct {
	ram Ram
}

func NewSystemRam() *SystemRam {
	return &SystemRam{*NewRam(0xc000, 13)}
}

func (sr *SystemRam) Rb(addr uint16) uint8 {
	return sr.ram.Rb(addr)
}

func (sr *SystemRam) Wb(addr uint16, val uint8) {
	sr.ram.Wb(addr, val)
}

func (sr *SystemRam) Rs(addr uint16) uint16 {
	return sr.ram.Rs(addr)
}

func (sr *SystemRam) Ws(addr uint16, val uint16) {
	sr.ram.Ws(addr, val)
}

func (sr *SystemRam) Asserts(addr uint16) bool {
	/* We need to skip some bits for the OAM */
	return addr >= 0xc000 && addr < 0xfe00
}
