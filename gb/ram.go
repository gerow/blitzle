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
1010 A
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

type MemRegister struct {
	RAM
}

func NewMemRegister(addr uint16) *MemRegister {
	m := &MemRegister{}
	m.RAM = *NewRAM(addr, addr)

	return m
}

func (m *MemRegister) val() uint8 {
	return m.RAM.data[0]
}

func (m *MemRegister) set(val uint8) {
	m.RAM.data[0] = val
}

func NewHiRAM() *RAM {
	return NewRAM(0xff80, 0xfffe)
}

type ReadOnlyRegister struct {
	addr     uint16
	readFunc func() uint8
}

func (r *ReadOnlyRegister) R(_ uint16) uint8 {
	return r.val()
}

func (r *ReadOnlyRegister) W(_ uint16, _ uint8) {
}

func (r *ReadOnlyRegister) Asserts(addr uint16) bool {
	return addr == r.addr
}

func (r *ReadOnlyRegister) val() uint8 {
	return r.readFunc()
}

type WriteOnlyRegister struct {
	addr      uint16
	writeFunc func(uint8)
}

func (w *WriteOnlyRegister) R(_ uint16) uint8 {
	return 0xff
}

func (w *WriteOnlyRegister) W(_ uint16, val uint8) {
	w.writeFunc(val)
}

func (w *WriteOnlyRegister) Asserts(addr uint16) bool {
	return addr == w.addr
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
