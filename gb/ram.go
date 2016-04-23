package gb

type RAM struct {
	startAddr uint16
	endAddr   uint16
	data      []byte
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

func NewRAM(startAddr uint16, endAddr uint16) *RAM {
	size := endAddr - startAddr + 1
	data := make([]byte, size)
	return &RAM{startAddr, endAddr, data}
}

func (r *RAM) R(addr uint16) uint8 {
	addr %= uint16(len(r.data))
	return r.data[addr]
}

func (r *RAM) W(addr uint16, val uint8) {
	addr %= uint16(len(r.data))
	r.data[addr] = val
}

func (r *RAM) Asserts(addr uint16) bool {
	return addr >= r.startAddr && addr <= r.endAddr
}

func NewHiRAM() *RAM {
	return NewRAM(0xff80, 0xffff)
}

/*
 * System RAM needs to assert specially in order to properly do the mirrioring
 * nonsense.
 */
type SystemRAM struct {
	ram RAM
}

func NewSystemRAM() *SystemRAM {
	return &SystemRAM{*NewRAM(0xc000, 0xdfff)}
}

func (sr *SystemRAM) R(addr uint16) uint8 {
	return sr.ram.R(addr)
}

func (sr *SystemRAM) W(addr uint16, val uint8) {
	sr.ram.W(addr, val)
}

func (sr *SystemRAM) Asserts(addr uint16) bool {
	/* We need to skip some bits for the OAM */
	return addr >= 0xc000 && addr < 0xfe00
}
