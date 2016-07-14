package gb

import (
	"bytes"
	"fmt"
	"log"
)

type CPU struct {
	/* GP registers */
	b  uint8
	c  uint8
	d  uint8
	e  uint8
	h  uint8
	l  uint8
	a  uint8
	ip uint16
	sp uint16

	/* Flags */
	fz bool
	fn bool
	fh bool
	fc bool

	halt       bool
	interrupts bool
}

func (c *CPU) flags() uint8 {
	o := uint8(0)
	if c.fz {
		o |= 0x80
	}
	if c.fn {
		o |= 0x40
	}
	if c.fh {
		o |= 0x20
	}
	if c.fc {
		o |= 0x10
	}

	return o
}

func (c *CPU) setFlags(val uint8) {
	c.fz = val&0x80 != 0
	c.fn = val&0x40 != 0
	c.fh = val&0x20 != 0
	c.fc = val&0x10 != 0
}

func halfCarry(a uint8, b uint8) bool {
	return halfCarryWithC(a, b, false)
}

func halfCarryWithC(a uint8, b uint8, c bool) bool {
	var carryMod uint16
	if c {
		carryMod = 1
	} else {
		carryMod = 0
	}
	return uint16(a&0xf)+uint16(b&0xf)+carryMod > 0xf
}

func carry(a uint8, b uint8) bool {
	return carryWithC(a, b, false)
}

func carryWithC(a uint8, b uint8, c bool) bool {
	var carryMod uint16
	if c {
		carryMod = 1
	} else {
		carryMod = 0
	}
	a16 := uint16(a) + carryMod
	return uint16(a16)+uint16(b) > 0xff
}

func negate(a uint8) uint8 {
	return ^a + 1
}

func halfBorrow(a uint8, b uint8) bool {
	return halfBorrowWithC(a, b, false)
}

func halfBorrowWithC(a uint8, b uint8, c bool) bool {
	var carryMod uint16
	if c {
		carryMod = 1
	} else {
		carryMod = 0
	}
	return uint16(b&0xf)+carryMod > uint16(a&0xf)
}

func borrow(a uint8, b uint8) bool {
	return borrowWithC(a, b, false)
}

func borrowWithC(a uint8, b uint8, c bool) bool {
	var carryMod uint16
	if c {
		carryMod = 1
	} else {
		carryMod = 0
	}
	b16 := uint16(b) + carryMod
	if b16 > uint16(a) {
		return true
	}
	return false
}

func signExtend(a uint8) uint16 {
	if a&0x80 != 0 {
		return 0xff00 | uint16(a)
	}

	return uint16(a)
}

func NewCPU() *CPU {
	cpu := &CPU{}

	return cpu
}

func (c *CPU) SetPostBootloaderState(sys *Sys) {
	c.ip = 0x100

	c.a = 0x01
	c.fz = true
	c.fh = true
	c.fn = false
	c.fc = true

	c.b = 0x00
	c.c = 0x13

	c.d = 0x00
	c.e = 0xd8

	c.h = 0x01
	c.l = 0x4d

	c.sp = 0xfffe
}

type CPUCond int

const (
	condC CPUCond = iota
	condNC
	condZ
	condNZ
	condNone
)

func (c *CPU) cond(con CPUCond) bool {
	switch con {
	case condC:
		return c.fc
	case condNC:
		return !c.fc
	case condZ:
		return c.fz
	case condNZ:
		return !c.fz
	case condNone:
		return true
	default:
		panic("received invalid con")
	}
}

func (c *CPU) Step(sys *Sys) int {
	if c.interrupts {
		interrupt := sys.HandleInterrupt()
		if interrupt != nil {
			if c.halt {
				fmt.Printf("HALT ended\n")
			}
			c.halt = false
			c.interrupts = false
			addr := uint16(0x40 + 8*uint(*interrupt))
			return RST(addr, true)(c, sys)
		}
	}
	if c.halt {
		return 4
	}
	opcode := sys.Rb(c.ip)
	if opcode == 0xcb {
		opcode = sys.Rb(c.ip + 1)
		return cbops[opcode](c, sys)
	}
	return ops[opcode](c, sys)
}

func (c *CPU) State(sys *Sys) string {
	o := bytes.Buffer{}
	o.WriteString(fmt.Sprintf("Registers\n"))
	o.WriteString(fmt.Sprintf("  B:  %02Xh\n", c.b))
	o.WriteString(fmt.Sprintf("  C:  %02Xh\n", c.c))
	o.WriteString(fmt.Sprintf("  D:  %02Xh\n", c.d))
	o.WriteString(fmt.Sprintf("  E:  %02Xh\n", c.e))
	o.WriteString(fmt.Sprintf("  H:  %02Xh\n", c.h))
	o.WriteString(fmt.Sprintf("  L:  %02Xh\n", c.l))
	o.WriteString(fmt.Sprintf("  A:  %02Xh\n", c.a))
	o.WriteString("  F:  ZNHC0000\n")
	o.WriteString(fmt.Sprintf("      %08b\n", c.flags()))
	o.WriteString(fmt.Sprintf("  IP: %04Xh\n", c.ip))
	if c.ip > 0x8000 && c.ip < 0xff80 {
		o.WriteString("!! IP outside of ROM/DMA range !!\n")
	}
	if c.ip == 0x0038 {
		o.WriteString("!!??!!\n")
	}
	o.WriteString(fmt.Sprintf("  SP: %04Xh\n", c.sp))
	//o.WriteString(fmt.Sprintf("Area around IP\n"))

	//	for addr := c.ip - 10; addr < c.ip+10; addr++ {
	//		ipChar := "*"
	//		if addr != c.ip {
	//			ipChar = " "
	//		}
	//		o.WriteString(fmt.Sprintf("%s%04Xh: %02Xh\n", ipChar, addr, sys.RbLog(addr, false)))
	//	}
	o.WriteString(fmt.Sprintf("*%04Xh: %02Xh %02Xh %02Xh\n", c.ip,
		sys.RbLog(c.ip, false), sys.RbLog(c.ip+1, false), sys.RbLog(c.ip+2, false)))

	//o.WriteString(fmt.Sprintf("Last 10 items on stack (most recent first):\n"))
	//for addr := c.sp; addr < c.sp+20; addr += 2 {
	//		o.WriteString(fmt.Sprintf("%04Xh: %04Xh\n", addr, sys.RsLog(addr, false)))
	//	}

	return o.String()
}

type ByteRegister int

const (
	B ByteRegister = iota
	C
	D
	E
	H
	L
	A
	HLind
	Imm
)

var ByteRegisterNameMap map[ByteRegister]string = map[ByteRegister]string{
	B:     "B",
	C:     "C",
	D:     "D",
	E:     "E",
	H:     "H",
	L:     "L",
	A:     "A",
	HLind: "(HL)",
	Imm:   "d8",
}

/* Read register byte */
func (c *CPU) rrb(br ByteRegister) uint8 {
	switch br {
	case B:
		return c.b
	case C:
		return c.c
	case D:
		return c.d
	case E:
		return c.e
	case H:
		return c.h
	case L:
		return c.l
	case A:
		return c.a
	default:
		panic("received invalid br")
	}
}

func (c *CPU) wrb(br ByteRegister, v uint8) {
	switch br {
	case B:
		c.b = v
	case C:
		c.c = v
	case D:
		c.d = v
	case E:
		c.e = v
	case H:
		c.h = v
	case L:
		c.l = v
	case A:
		c.a = v
	default:
		panic("received invalid br")
	}
}

type ShortRegister int

const (
	BC ShortRegister = iota
	DE
	HL
	SP
	AF
)

func (c *CPU) rrs(sr ShortRegister) uint16 {
	switch sr {
	case BC:
		return uint16(c.c) | uint16(c.b)<<8
	case DE:
		return uint16(c.e) | uint16(c.d)<<8
	case HL:
		return uint16(c.l) | uint16(c.h)<<8
	case SP:
		return c.sp
	case AF:
		return uint16(c.flags()) | uint16(c.a)<<8
	default:
		panic("received invalid sr")
	}
}

type ALUOp int

const (
	ADD ALUOp = iota
	ADC
	SUB
	SBC
	AND
	XOR
	OR
	CP
)

func (c *CPU) wrs(sr ShortRegister, v uint16) {
	l := uint8(v & 0xff)
	u := uint8(v >> 8)
	switch sr {
	case BC:
		c.c = l
		c.b = u
	case DE:
		c.e = l
		c.d = u
	case HL:
		c.l = l
		c.h = u
	case SP:
		c.sp = v
	default:
		panic("received invalid sr")
	}
}

func (c *CPU) Push(sys *Sys, v uint16) {
	c.sp -= 2
	sys.Ws(c.sp, v)
}

func (c *CPU) Pop(sys *Sys) uint16 {
	rv := sys.Rs(c.sp)
	c.sp += 2
	return rv
}

type OpFunc func(cpu *CPU, sys *Sys) int

func NOP(cpu *CPU, sys *Sys) int {
	cpu.ip += 1
	return 4
}

func STOP(cpu *CPU, sys *Sys) int {
	sys.Stop = true

	/* This technically takes an immediate argument, but it doesn't do anything */
	cpu.ip += 2
	return 4
}

/* Load short immediate */
func LDSImm(sr ShortRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		cpu.wrs(sr, sys.Rs(cpu.ip+1))
		cpu.ip += 3
		return 12
	}
}

/* Load A indirectly */
func LDARegInd(br ShortRegister, mod int) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := cpu.rrs(br)
		sys.Wb(addr, cpu.rrb(A))
		if mod == 1 {
			cpu.wrs(br, addr+1)
		} else if mod == -1 {
			cpu.wrs(br, addr-1)
		}

		cpu.ip++
		return 8
	}
}

/* Increment or decrement short register */
func INCDECS(sr ShortRegister, mod int) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		v := cpu.rrs(sr)
		cpu.wrs(sr, uint16(int(v)+mod))

		cpu.ip++
		return 8
	}
}

/* Increment or decrement byte register */
func INCDECB(br ByteRegister, mod int) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		var v uint8
		var newVal uint8
		if br == HLind {
			v = sys.Rb(cpu.rrs(HL))
			newVal = uint8(int(v) + mod)
			sys.Wb(cpu.rrs(HL), newVal)
		} else {
			v = cpu.rrb(br)
			newVal = uint8(int(v) + mod)
			cpu.wrb(br, newVal)
		}
		cpu.fz = newVal == 0
		cpu.fn = mod == -1
		if mod == 1 {
			cpu.fh = halfCarry(v, 1)
		} else {
			cpu.fh = halfBorrow(v, 1)
		}

		cpu.ip++
		if br == HLind {
			return 12
		}
		return 4
	}
}

func JR(con CPUCond) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		j := signExtend(sys.Rb(cpu.ip + 1))
		duration := 0
		/* The relative amount is relative to where we would have been after this op */
		cpu.ip += 2
		if cpu.cond(con) {
			cpu.ip += j
			duration = 12
		} else {
			duration = 8
		}
		return duration
	}
}

/* Load byte immediate */
func LDBImm(br ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		duration := 0
		if br == HLind {
			sys.Wb(cpu.rrs(HL), sys.Rb(cpu.ip+1))
			duration = 12
		} else {
			cpu.wrb(br, sys.Rb(cpu.ip+1))
			duration = 8
		}
		cpu.ip += 2
		return duration
	}
}

func RLCA(cpu *CPU, sys *Sys) int {
	a := cpu.rrb(A)
	carry := a>>7 != 0
	oldA := a
	a <<= 1
	a |= oldA >> 7
	cpu.wrb(A, a)

	cpu.fz = a == 0
	cpu.fn = false
	cpu.fh = false
	cpu.fc = carry

	cpu.ip++
	return 4
}

func RLA(cpu *CPU, sys *Sys) int {
	a := cpu.rrb(A)
	carry := a>>7 != 0
	a <<= 1
	if cpu.fc {
		a |= 1
	}
	cpu.wrb(A, a)

	cpu.fz = a == 0
	cpu.fn = false
	cpu.fh = false
	cpu.fc = carry

	cpu.ip++
	return 4
}

/* Load SP via an imediate value that points to another value */
func LDSPImmInd(cpu *CPU, sys *Sys) int {
	sp := cpu.rrs(SP)
	addr := sys.Rs(cpu.ip + 1)
	sys.Ws(addr, sp)

	cpu.ip += 3
	return 20
}

/* Add short */
func ADDS(sr ShortRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		hl := cpu.rrs(HL)
		val := cpu.rrs(sr)

		cpu.wrs(HL, hl+val)

		cpu.fh = uint32(hl&0xfff)+uint32(val&0xfff) > 0xfff
		cpu.fc = uint32(hl)+uint32(val) > 0xffff
		cpu.fn = false

		cpu.ip++
		return 8
	}
}

func LDBInd(destReg ByteRegister, srcAddrReg ShortRegister, mod int) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := cpu.rrs(srcAddrReg)
		cpu.wrb(destReg, sys.Rb(addr))
		if mod != 0 {
			cpu.wrs(srcAddrReg, uint16(int(addr)+mod))
		}

		cpu.ip++
		return 8
	}
}

func RRCA(cpu *CPU, sys *Sys) int {
	a := cpu.rrb(A)
	carry := a&1 != 0
	oldA := a
	a >>= 1
	a |= oldA << 7
	cpu.wrb(A, a)

	cpu.fz = a == 0
	cpu.fn = false
	cpu.fh = false
	cpu.fc = carry

	cpu.ip++
	return 4
}

func RRA(cpu *CPU, sys *Sys) int {
	a := cpu.rrb(A)
	carry := a&1 != 0
	a >>= 1
	if cpu.fc {
		a |= 0x80
	}
	cpu.wrb(A, a)

	cpu.fz = a == 0
	cpu.fn = false
	cpu.fh = false
	cpu.fc = carry

	cpu.ip++
	return 4
}

/* Decimal adjust A */
func DAA(cpu *CPU, sys *Sys) int {
	// Aped from DParrott on nesdev because I can't figure out all
	// the undefined cases :(
	// <http://forums.nesdev.com/viewtopic.php?t=9088>
	a := uint16(cpu.rrb(A))

	if !cpu.fn {

		if cpu.fh || (a&0xf) > 9 {
			a += 0x06
		}
		if cpu.fc || a > 0x9f {
			a += 0x60
		}
	} else {
		if cpu.fh {
			a = (a - 6) & 0xff
		}
		if cpu.fc {
			a -= 0x60
		}
	}
	cpu.fh = false
	cpu.fz = false
	if (a & 0x100) == 0x100 {
		cpu.fc = true
	}
	a &= 0xff
	if a == 0 {
		cpu.fz = true
	}

	cpu.wrb(A, uint8(a))

	cpu.ip++
	return 4
}

func SCF(cpu *CPU, sys *Sys) int {
	cpu.fn = false
	cpu.fh = false
	cpu.fc = true

	cpu.ip++
	return 4
}

func CPL(cpu *CPU, sys *Sys) int {
	cpu.a = ^cpu.a

	cpu.fn = true
	cpu.fh = true

	cpu.ip++
	return 4
}

func CCF(cpu *CPU, sys *Sys) int {
	cpu.fn = false
	cpu.fh = false
	cpu.fc = !cpu.fc

	cpu.ip++
	return 4
}

/* Load between byte registers */
func LDB(destReg ByteRegister, srcReg ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		src := cpu.rrb(srcReg)
		cpu.wrb(destReg, src)

		cpu.ip++
		return 4
	}
}

func LDHLBindir(br ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		val := cpu.rrb(br)
		addr := cpu.rrs(HL)
		sys.Wb(addr, val)

		cpu.ip++
		return 8
	}
}

func HALT(cpu *CPU, sys *Sys) int {
	fmt.Printf("HALT started\n")
	cpu.halt = true

	cpu.ip++
	return 4
}

func ALU(op ALUOp, br ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		var val uint8
		var duration int
		var iSize uint16
		if br == HLind {
			val = sys.Rb(cpu.rrs(HL))
			duration = 8
			iSize = 1
		} else if br == Imm {
			val = sys.Rb(cpu.ip + 1)
			duration = 8
			iSize = 2
		} else {
			val = cpu.rrb(br)
			duration = 4
			iSize = 1
		}

		var carryMod uint8
		if cpu.fc {
			carryMod = 1
		} else {
			carryMod = 0
		}
		var res uint8
		switch op {
		case ADD:
			cpu.fn = false
			cpu.fh = halfCarry(cpu.a, val)
			cpu.fc = carry(cpu.a, val)

			cpu.a += val
		case ADC:
			cpu.fn = false
			cpu.fh = halfCarryWithC(cpu.a, val, cpu.fc)
			cpu.fc = carryWithC(cpu.a, val, cpu.fc)

			cpu.a += val + carryMod
		case SUB:
			cpu.fn = true
			cpu.fh = halfBorrow(cpu.a, val)
			cpu.fc = borrow(cpu.a, val)

			cpu.a -= val
		case SBC:
			cpu.fn = true
			cpu.fh = halfBorrowWithC(cpu.a, val, cpu.fc)
			cpu.fc = borrowWithC(cpu.a, val, cpu.fc)

			cpu.a -= val + carryMod
		case AND:
			cpu.fn = false
			cpu.fh = true
			cpu.fc = false

			cpu.a &= val
		case XOR:
			cpu.fn = false
			cpu.fh = false
			cpu.fc = false

			cpu.a ^= val
		case OR:
			cpu.fn = false
			cpu.fh = false
			cpu.fc = false

			cpu.a |= val
		case CP:
			cpu.fn = true
			cpu.fh = halfBorrow(cpu.a, val)
			cpu.fc = borrow(cpu.a, val)

			res = cpu.a - val
		default:
			panic("received invalid op")
		}

		if op == CP {
			cpu.fz = res == 0
		} else {
			cpu.fz = cpu.a == 0
		}

		cpu.ip += iSize
		return duration
	}
}

func RET(con CPUCond, enableInterrupts bool) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		if enableInterrupts {
			cpu.interrupts = true
		}
		if cpu.cond(con) {
			ra := cpu.Pop(sys)
			cpu.ip = ra
			if con == condNone {
				return 16
			}
			return 20
		} else {
			cpu.ip++
			return 8
		}
	}
}

func JP(con CPUCond) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := sys.Rs(cpu.ip + 1)
		if cpu.cond(con) {
			cpu.ip = addr
			return 16
		} else {
			cpu.ip += 3
			return 12
		}
	}
}

func CALL(con CPUCond) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := sys.Rs(cpu.ip + 1)
		if cpu.cond(con) {
			cpu.Push(sys, cpu.ip+3)
			cpu.ip = addr
			return 24
		} else {
			cpu.ip += 3
			return 12
		}
	}
}

func JPHLind(cpu *CPU, sys *Sys) int {
	addr := cpu.rrs(HL)
	cpu.ip = addr

	return 4
}

func POP(sr ShortRegister) OpFunc {
	if sr == AF {
		panic("no, that won't work!")
	}

	return func(cpu *CPU, sys *Sys) int {
		cpu.wrs(sr, cpu.Pop(sys))

		cpu.ip++
		return 12
	}
}

func POPAF(cpu *CPU, sys *Sys) int {
	af := cpu.Pop(sys)
	a := uint8(af >> 8)
	f := uint8(af & 0xf0)

	cpu.setFlags(f)
	cpu.a = a

	cpu.ip++
	return 12
}

func PUSH(sr ShortRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		cpu.Push(sys, cpu.rrs(sr))

		cpu.ip++
		return 16
	}
}

func RST(addr uint16, fromInterrupt bool) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		ra := cpu.ip
		if !fromInterrupt {
			ra += 1
		}
		cpu.Push(sys, ra)
		cpu.ip = addr

		return 16
	}
}

func PANIC(cpu *CPU, sys *Sys) int {
	panic("This should never get called!")
}

func DRAGONS(cpu *CPU, sys *Sys) int {
	log.Println("This op shouldn't exist! That means that either someone is relying on unspecified operations or the emulation is in a really bad state!")
	log.Println(cpu.State(sys))

	cpu.ip++
	return 4
}

func LDH(atoaddr bool) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := uint16(0xff00) | uint16(sys.Rb(cpu.ip+1))
		if atoaddr {
			sys.Wb(addr, cpu.a)
		} else {
			cpu.a = sys.Rb(addr)
		}

		cpu.ip += 2
		return 12
	}
}

func LDHC(atoaddr bool) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := uint16(0xff00) | uint16(cpu.c)
		if atoaddr {
			sys.Wb(addr, cpu.a)
		} else {
			cpu.a = sys.Rb(addr)
		}

		cpu.ip++
		return 8
	}
}

func ADDSPimm(cpu *CPU, sys *Sys) int {
	v := signExtend(sys.Rb(cpu.ip + 1))

	cpu.fz = false
	cpu.fn = false
	cpu.fh = halfCarry(uint8(cpu.sp), uint8(v))
	cpu.fc = carry(uint8(cpu.sp), uint8(v))

	cpu.sp += v

	cpu.ip += 2
	return 16
}

func LDSimmAddr(atoaddr bool) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := sys.Rs(cpu.ip + 1)
		if atoaddr {
			sys.Wb(addr, cpu.a)
		} else {
			cpu.a = sys.Rb(addr)
		}

		cpu.ip += 3
		return 16
	}
}

func DI(cpu *CPU, sys *Sys) int {
	cpu.interrupts = false

	cpu.ip++
	return 4
}

func EI(cpu *CPU, sys *Sys) int {
	cpu.interrupts = true

	cpu.ip++
	return 4
}

func LDHLSPimm(cpu *CPU, sys *Sys) int {
	v := signExtend(sys.Rb(cpu.ip + 1))

	cpu.fz = false
	cpu.fn = false
	cpu.fh = halfCarry(uint8(cpu.sp), uint8(v))
	cpu.fc = carry(uint8(cpu.sp), uint8(v))

	cpu.wrs(HL, cpu.sp+v)

	cpu.ip += 2
	return 12
}

func LDSPHL(cpu *CPU, sys *Sys) int {
	cpu.sp = cpu.rrs(HL)

	cpu.ip++
	return 8
}

type CBROp int

const (
	RLC = iota
	RRC
	RL
	RR
	SLA
	SRA
	SRL
)

/* CB Rotates */
func CBR(op CBROp, br ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		var v uint8
		if br == HLind {
			v = sys.Rb(cpu.rrs(HL))
		} else {
			v = cpu.rrb(br)
		}
		oldV := v
		ifc := cpu.fc
		switch op {
		case RLC:
			/* Left rotate */
			cpu.fc = 0x80&v != 0
			v <<= 1
			v |= oldV >> 7
		case RRC:
			/* Right rotate */
			cpu.fc = 0x01&v != 0
			v >>= 1
			v |= oldV << 7
		case RL:
			/* Left rotate through carry */
			cpu.fc = 0x80&v != 0
			v <<= 1
			if ifc {
				v |= 0x01
			}
		case RR:
			/* Right rotate through carry */
			cpu.fc = 0x01&v != 0
			v >>= 1
			if ifc {
				v |= 0x80
			}
		case SLA:
			/* Left shift */
			cpu.fc = 0x80&v != 0
			v <<= 1
		case SRA:
			/* Signed right shift */
			cpu.fc = 0x01&v != 0
			v >>= 1
			v |= oldV & 0x80
		case SRL:
			/* Unsigned right shift */
			cpu.fc = 0x01&v != 0
			v >>= 1
		default:
			panic("invalid op!")
		}
		if br == HLind {
			sys.Wb(cpu.rrs(HL), v)
		} else {
			cpu.wrb(br, v)
		}

		cpu.fz = v == 0
		cpu.fn = false
		cpu.fh = false

		cpu.ip += 2
		if br == HLind {
			return 16
		}
		return 8
	}
}

func SWAP(br ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		var v uint8
		if br == HLind {
			v = sys.Rb(cpu.rrs(HL))
		} else {
			v = cpu.rrb(br)
		}

		v = v>>4 | v<<4
		if br == HLind {
			sys.Wb(cpu.rrs(HL), v)
		} else {
			cpu.wrb(br, v)
		}

		cpu.fz = v == 0
		cpu.fn = false
		cpu.fh = false
		cpu.fc = false

		cpu.ip += 2
		if br == HLind {
			return 16
		}
		return 8
	}
}

func BIT(n uint, br ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		var v uint8
		if br == HLind {
			v = sys.Rb(cpu.rrs(HL))
		} else {
			v = cpu.rrb(br)
		}

		cpu.fz = (0x01<<n)&v == 0
		cpu.fn = false
		cpu.fh = true

		cpu.ip += 2
		if br == HLind {
			return 16
		}
		return 8
	}
}

func SETRES(set bool, n uint, br ByteRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		var v uint8
		if br == HLind {
			v = sys.Rb(cpu.rrs(HL))
		} else {
			v = cpu.rrb(br)
		}

		if set {
			v |= 0x01 << n
		} else {
			v &= ^(0x01 << n)
		}

		if br == HLind {
			sys.Wb(cpu.rrs(HL), v)
		} else {
			cpu.wrb(br, v)
		}

		cpu.ip += 2
		if br == HLind {
			return 16
		}
		return 8
	}
}

var ops [0x100]OpFunc = [0x100]OpFunc{
	/* 0x00 */
	NOP,              /* NOP */
	LDSImm(BC),       /* LD BC,d16 */
	LDARegInd(BC, 0), /* LD (BC),A */
	INCDECS(BC, 1),   /* INC BC */
	INCDECB(B, 1),    /* INC B */
	INCDECB(B, -1),   /* DEC B */
	LDBImm(B),        /* LD B,d8 */
	RLCA,             /* RLCA */
	LDSPImmInd,       /* LD (a16),SP */
	ADDS(BC),         /* ADD HL,BC */
	LDBInd(A, BC, 0), /* LD A,(BC) */
	INCDECS(BC, -1),  /* DEC BC */
	INCDECB(C, 1),    /* INC C */
	INCDECB(C, -1),   /* DEC C */
	LDBImm(C),        /* LD C,d8 */
	RRCA,             /* RRCA */
	/* 0x10 */
	STOP,             /* STOP 0 */
	LDSImm(DE),       /* LD DE,d16 */
	LDARegInd(DE, 0), /* LD (DE),A */
	INCDECS(DE, 1),   /* INC DE */
	INCDECB(D, 1),    /* INC D */
	INCDECB(D, -1),   /* DEC D */
	LDBImm(D),        /* LD D,d8 */
	RLA,              /* RLA */
	JR(condNone),     /* JR r8 */
	ADDS(DE),         /* ADD HL,DE */
	LDBInd(A, DE, 0), /* LD A,(DE) */
	INCDECS(DE, -1),  /* DEC DE */
	INCDECB(E, 1),    /* INC E */
	INCDECB(E, -1),   /* DEC E */
	LDBImm(E),        /* LD E,d8 */
	RRA,              /* RRA */
	/* 0x20 */
	JR(condNZ),       /* JR NZ,r8 */
	LDSImm(HL),       /* LD HL,d16 */
	LDARegInd(HL, 1), /* LD (HL+),A */
	INCDECS(HL, 1),   /* INC HL */
	INCDECB(H, 1),    /* INC H */
	INCDECB(H, -1),   /* DEC H */
	LDBImm(H),        /* LD H,d8 */
	DAA,              /* DAA */
	JR(condZ),        /* JR Z,r8 */
	ADDS(HL),         /* ADD HL,HL */
	LDBInd(A, HL, 1), /* LD A,(HL+) */
	INCDECS(HL, -1),  /* DEC HL */
	INCDECB(L, 1),    /* INC L */
	INCDECB(L, -1),   /* DEC L */
	LDBImm(L),        /* LD L,d8 */
	CPL,              /* CPL */
	/* 0x30 */
	JR(condNC),         /* JR NC,r8 */
	LDSImm(SP),         /* LD SP,d16 */
	LDARegInd(HL, -1),  /* LD (HL-),A */
	INCDECS(SP, 1),     /* INC SP */
	INCDECB(HLind, 1),  /* INC (HL) */
	INCDECB(HLind, -1), /* DEC (HL) */
	LDBImm(HLind),      /* LD (HL),d8 */
	SCF,                /* SCF */
	JR(condC),          /* JR C,r8 */
	ADDS(SP),           /* ADD HL,SP */
	LDBInd(A, HL, -1),  /* LD A,(HL-) */
	INCDECS(SP, -1),    /* DEC SP */
	INCDECB(A, 1),      /* INC A */
	INCDECB(A, -1),     /* DEC A */
	LDBImm(A),          /* LD A,d8 */
	CCF,                /* CCF */
	/* 0x40 */
	LDB(B, B),        /* LD B,B */
	LDB(B, C),        /* LD B,C */
	LDB(B, D),        /* LD B,D */
	LDB(B, E),        /* LD B,E */
	LDB(B, H),        /* LD B,H */
	LDB(B, L),        /* LD B,L */
	LDBInd(B, HL, 0), /* LD B,(HL) */
	LDB(B, A),        /* LD B,A */
	LDB(C, B),        /* LD C,B */
	LDB(C, C),        /* LD C,C */
	LDB(C, D),        /* LD C,D */
	LDB(C, E),        /* LD C,E */
	LDB(C, H),        /* LD C,H */
	LDB(C, L),        /* LD C,L */
	LDBInd(C, HL, 0), /* LD C,(HL) */
	LDB(C, A),        /* LD C,A */
	/* 0x50 */
	LDB(D, B),        /* LD D,B */
	LDB(D, C),        /* LD D,C */
	LDB(D, D),        /* LD D,D */
	LDB(D, E),        /* LD D,E */
	LDB(D, H),        /* LD D,H */
	LDB(D, L),        /* LD D,L */
	LDBInd(D, HL, 0), /* LD D,(HL) */
	LDB(D, A),        /* LD D,A */
	LDB(E, B),        /* LD E,B */
	LDB(E, C),        /* LD E,C */
	LDB(E, D),        /* LD E,D */
	LDB(E, E),        /* LD E,E */
	LDB(E, H),        /* LD E,H */
	LDB(E, L),        /* LD E,L */
	LDBInd(E, HL, 0), /* LD E,(HL) */
	LDB(E, A),        /* LD E,A */
	/* 0x60 */
	LDB(H, B),        /* LD H,B */
	LDB(H, C),        /* LD H,C */
	LDB(H, D),        /* LD H,D */
	LDB(H, E),        /* LD H,E */
	LDB(H, H),        /* LD H,H */
	LDB(H, L),        /* LD H,L */
	LDBInd(H, HL, 0), /* LD H,(HL) */
	LDB(H, A),        /* LD H,A */
	LDB(L, B),        /* LD L,B */
	LDB(L, C),        /* LD L,C */
	LDB(L, D),        /* LD L,D */
	LDB(L, E),        /* LD L,E */
	LDB(L, H),        /* LD L,H */
	LDB(L, L),        /* LD L,L */
	LDBInd(L, HL, 0), /* LD L,(HL) */
	LDB(L, A),        /* LD L,A */
	/* 0x70 */
	LDHLBindir(B),    /* LD (HL),B */
	LDHLBindir(C),    /* LD (HL),C */
	LDHLBindir(D),    /* LD (HL),D */
	LDHLBindir(E),    /* LD (HL),E */
	LDHLBindir(H),    /* LD (HL),H */
	LDHLBindir(L),    /* LD (HL),L */
	HALT,             /* HALT */
	LDHLBindir(A),    /* LD (HL),A */
	LDB(A, B),        /* LD A,B */
	LDB(A, C),        /* LD A,C */
	LDB(A, D),        /* LD A,D */
	LDB(A, E),        /* LD A,E */
	LDB(A, H),        /* LD A,H */
	LDB(A, L),        /* LD A,L */
	LDBInd(A, HL, 0), /* LD A,(HL) */
	LDB(A, A),        /* LD A,A */
	/* 0x80 */
	ALU(ADD, B),     /* ADD A,B */
	ALU(ADD, C),     /* ADD A,C */
	ALU(ADD, D),     /* ADD A,D */
	ALU(ADD, E),     /* ADD A,E */
	ALU(ADD, H),     /* ADD A,H */
	ALU(ADD, L),     /* ADD A,L */
	ALU(ADD, HLind), /* ADD A,(HL) */
	ALU(ADD, A),     /* ADD A,A */
	ALU(ADC, B),     /* ADC A,B */
	ALU(ADC, C),     /* ADC A,C */
	ALU(ADC, D),     /* ADC A,D */
	ALU(ADC, E),     /* ADC A,E */
	ALU(ADC, H),     /* ADC A,H */
	ALU(ADC, L),     /* ADC A,L */
	ALU(ADC, HLind), /* ADC A,(HL) */
	ALU(ADC, A),     /* ADC A,A */
	/* 0x90 */
	ALU(SUB, B),     /* SUB A,B */
	ALU(SUB, C),     /* SUB A,C */
	ALU(SUB, D),     /* SUB A,D */
	ALU(SUB, E),     /* SUB A,E */
	ALU(SUB, H),     /* SUB A,H */
	ALU(SUB, L),     /* SUB A,L */
	ALU(SUB, HLind), /* SUB A,(HL) */
	ALU(SUB, A),     /* SUB A,A */
	ALU(SBC, B),     /* SBC A,B */
	ALU(SBC, C),     /* SBC A,C */
	ALU(SBC, D),     /* SBC A,D */
	ALU(SBC, E),     /* SBC A,E */
	ALU(SBC, H),     /* SBC A,H */
	ALU(SBC, L),     /* SBC A,L */
	ALU(SBC, HLind), /* SBC A,(HL) */
	ALU(SBC, A),     /* SBC A,A */
	/* 0xa0 */
	ALU(AND, B),     /* AND A,B */
	ALU(AND, C),     /* AND A,C */
	ALU(AND, D),     /* AND A,D */
	ALU(AND, E),     /* AND A,E */
	ALU(AND, H),     /* AND A,H */
	ALU(AND, L),     /* AND A,L */
	ALU(AND, HLind), /* AND A,(HL) */
	ALU(AND, A),     /* AND A,A */
	ALU(XOR, B),     /* XOR A,B */
	ALU(XOR, C),     /* XOR A,C */
	ALU(XOR, D),     /* XOR A,D */
	ALU(XOR, E),     /* XOR A,E */
	ALU(XOR, H),     /* XOR A,H */
	ALU(XOR, L),     /* XOR A,L */
	ALU(XOR, HLind), /* XOR A,(HL) */
	ALU(XOR, A),     /* XOR A,A */
	/* 0xb0 */
	ALU(OR, B),     /* OR A,B */
	ALU(OR, C),     /* OR A,C */
	ALU(OR, D),     /* OR A,D */
	ALU(OR, E),     /* OR A,E */
	ALU(OR, H),     /* OR A,H */
	ALU(OR, L),     /* OR A,L */
	ALU(OR, HLind), /* OR A,(HL) */
	ALU(OR, A),     /* OR A,A */
	ALU(CP, B),     /* CP A,B */
	ALU(CP, C),     /* CP A,C */
	ALU(CP, D),     /* CP A,D */
	ALU(CP, E),     /* CP A,E */
	ALU(CP, H),     /* CP A,H */
	ALU(CP, L),     /* CP A,L */
	ALU(CP, HLind), /* CP A,(HL) */
	ALU(CP, A),     /* CP A,A */
	/* 0xc0 */
	RET(condNZ, false),   /* RET NZ */
	POP(BC),              /* POP BC */
	JP(condNZ),           /* JP NZ,a16 */
	JP(condNone),         /* JP a16 */
	CALL(condNZ),         /* CALL NZ,a16 */
	PUSH(BC),             /* PUSH BC */
	ALU(ADD, Imm),        /* ADD A,d8 */
	RST(0x00, false),     /* RST 00H */
	RET(condZ, false),    /* RET Z */
	RET(condNone, false), /* RET */
	JP(condZ),            /* JP Z,a16 */
	PANIC,                /* PREFIX CB */
	CALL(condZ),          /* CALL Z,a16 */
	CALL(condNone),       /* CALL a16 */
	ALU(ADC, Imm),        /* ADC A,d8 */
	RST(0x08, false),     /* RST 08H */
	/* 0xd0 */
	RET(condNC, false),  /* RET NC */
	POP(DE),             /* POP DE */
	JP(condNC),          /* JP NC,a16 */
	DRAGONS,             /* XXX */
	CALL(condNC),        /* CALL NC,a16 */
	PUSH(DE),            /* PUSH DE */
	ALU(SUB, Imm),       /* SUB A,d8 */
	RST(0x10, false),    /* RST 10H */
	RET(condC, false),   /* RET C */
	RET(condNone, true), /* RETI */
	JP(condC),           /* JP C,a16 */
	DRAGONS,             /* XXX */
	CALL(condC),         /* CALL C,a16 */
	DRAGONS,             /* XXX */
	ALU(SBC, Imm),       /* SBC A,d8 */
	RST(0x18, false),    /* RST 18H */
	/* 0xe0 */
	LDH(true),        /* LDH (a8),A */
	POP(HL),          /* POP HL */
	LDHC(true),       /* LD (C),A */
	DRAGONS,          /* XXX */
	DRAGONS,          /* XXX */
	PUSH(HL),         /* PUSH HL */
	ALU(AND, Imm),    /* AND A,d8 */
	RST(0x20, false), /* RST 20H */
	ADDSPimm,         /* ADD SP,r8 */
	JPHLind,          /* JP (HL) */
	LDSimmAddr(true), /* LD (a16),A */
	DRAGONS,          /* XXX */
	DRAGONS,          /* XXX */
	DRAGONS,          /* XXX */
	ALU(XOR, Imm),    /* XOR A,d8 */
	RST(0x28, false), /* RST 28H */
	/* 0xf0 */
	LDH(false),        /* LDH A,(a8) */
	POPAF,             /* POP AF */
	LDHC(false),       /* LD A,(C) */
	DI,                /* DI */
	DRAGONS,           /* XXX */
	PUSH(AF),          /* PUSH AF */
	ALU(OR, Imm),      /* OR A,d8 */
	RST(0x30, false),  /* RST 30H */
	LDHLSPimm,         /* LD HL,SP+r8 */
	LDSPHL,            /* LD SP,HL */
	LDSimmAddr(false), /* LD A,(a16) */
	EI,                /* EI */
	DRAGONS,           /* XXX */
	DRAGONS,           /* XXX */
	ALU(CP, Imm),      /* CP A,d8 */
	RST(0x38, false),  /* RST 38H */
}

var cbops [0x100]OpFunc = [0x100]OpFunc{
	/* 0x00 */
	CBR(RLC, B),     /* RLC B */
	CBR(RLC, C),     /* RLC C */
	CBR(RLC, D),     /* RLC D */
	CBR(RLC, E),     /* RLC E */
	CBR(RLC, H),     /* RLC H */
	CBR(RLC, L),     /* RLC L */
	CBR(RLC, HLind), /* RLC (HL) */
	CBR(RLC, A),     /* RLC A */
	CBR(RRC, B),     /* RRC B */
	CBR(RRC, C),     /* RRC C */
	CBR(RRC, D),     /* RRC D */
	CBR(RRC, E),     /* RRC E */
	CBR(RRC, H),     /* RRC H */
	CBR(RRC, L),     /* RRC L */
	CBR(RRC, HLind), /* RRC (HL) */
	CBR(RRC, A),     /* RRC A */
	/* 0x10 */
	CBR(RL, B),     /* RL B */
	CBR(RL, C),     /* RL C */
	CBR(RL, D),     /* RL D */
	CBR(RL, E),     /* RL E */
	CBR(RL, H),     /* RL H */
	CBR(RL, L),     /* RL L */
	CBR(RL, HLind), /* RL (HL) */
	CBR(RL, A),     /* RL A */
	CBR(RR, B),     /* RR B */
	CBR(RR, C),     /* RR C */
	CBR(RR, D),     /* RR D */
	CBR(RR, E),     /* RR E */
	CBR(RR, H),     /* RR H */
	CBR(RR, L),     /* RR L */
	CBR(RR, HLind), /* RR (HL) */
	CBR(RR, A),     /* RR A */
	/* 0x20 */
	CBR(SLA, B),     /* SLA B */
	CBR(SLA, C),     /* SLA C */
	CBR(SLA, D),     /* SLA D */
	CBR(SLA, E),     /* SLA E */
	CBR(SLA, H),     /* SLA H */
	CBR(SLA, L),     /* SLA L */
	CBR(SLA, HLind), /* SLA (HL) */
	CBR(SLA, A),     /* SLA A */
	CBR(SRA, B),     /* SRA B */
	CBR(SRA, C),     /* SRA C */
	CBR(SRA, D),     /* SRA D */
	CBR(SRA, E),     /* SRA E */
	CBR(SRA, H),     /* SRA H */
	CBR(SRA, L),     /* SRA L */
	CBR(SRA, HLind), /* SRA (HL) */
	CBR(SRA, A),     /* SRA A */
	/* 0x30 */
	SWAP(B),         /* SWAP B */
	SWAP(C),         /* SWAP C */
	SWAP(D),         /* SWAP D */
	SWAP(E),         /* SWAP E */
	SWAP(H),         /* SWAP H */
	SWAP(L),         /* SWAP L */
	SWAP(HLind),     /* SWAP (HL) */
	SWAP(A),         /* SWAP A */
	CBR(SRL, B),     /* SRL B */
	CBR(SRL, C),     /* SRL C */
	CBR(SRL, D),     /* SRL D */
	CBR(SRL, E),     /* SRL E */
	CBR(SRL, H),     /* SRL H */
	CBR(SRL, L),     /* SRL L */
	CBR(SRL, HLind), /* SRL (HL) */
	CBR(SRL, A),     /* SRL A */
	/* 0x40 */
	BIT(0, B),     /* BIT 0,B */
	BIT(0, C),     /* BIT 0,C */
	BIT(0, D),     /* BIT 0,D */
	BIT(0, E),     /* BIT 0,E */
	BIT(0, H),     /* BIT 0,H */
	BIT(0, L),     /* BIT 0,L */
	BIT(0, HLind), /* BIT 0,(HL) */
	BIT(0, A),     /* BIT 0,A */
	BIT(1, B),     /* BIT 1,B */
	BIT(1, C),     /* BIT 1,C */
	BIT(1, D),     /* BIT 1,D */
	BIT(1, E),     /* BIT 1,E */
	BIT(1, H),     /* BIT 1,H */
	BIT(1, L),     /* BIT 1,L */
	BIT(1, HLind), /* BIT 1,(HL) */
	BIT(1, A),     /* BIT 1,A */
	/* 0x50 */
	BIT(2, B),     /* BIT 2,B */
	BIT(2, C),     /* BIT 2,C */
	BIT(2, D),     /* BIT 2,D */
	BIT(2, E),     /* BIT 2,E */
	BIT(2, H),     /* BIT 2,H */
	BIT(2, L),     /* BIT 2,L */
	BIT(2, HLind), /* BIT 2,(HL) */
	BIT(2, A),     /* BIT 2,A */
	BIT(3, B),     /* BIT 3,B */
	BIT(3, C),     /* BIT 3,C */
	BIT(3, D),     /* BIT 3,D */
	BIT(3, E),     /* BIT 3,E */
	BIT(3, H),     /* BIT 3,H */
	BIT(3, L),     /* BIT 3,L */
	BIT(3, HLind), /* BIT 3,(HL) */
	BIT(3, A),     /* BIT 3,A */
	/* 0x60 */
	BIT(4, B),     /* BIT 4,B */
	BIT(4, C),     /* BIT 4,C */
	BIT(4, D),     /* BIT 4,D */
	BIT(4, E),     /* BIT 4,E */
	BIT(4, H),     /* BIT 4,H */
	BIT(4, L),     /* BIT 4,L */
	BIT(4, HLind), /* BIT 4,(HL) */
	BIT(4, A),     /* BIT 4,A */
	BIT(5, B),     /* BIT 5,B */
	BIT(5, C),     /* BIT 5,C */
	BIT(5, D),     /* BIT 5,D */
	BIT(5, E),     /* BIT 5,E */
	BIT(5, H),     /* BIT 5,H */
	BIT(5, L),     /* BIT 5,L */
	BIT(5, HLind), /* BIT 5,(HL) */
	BIT(5, A),     /* BIT 5,A */
	/* 0x70 */
	BIT(6, B),     /* BIT 6,B */
	BIT(6, C),     /* BIT 6,C */
	BIT(6, D),     /* BIT 6,D */
	BIT(6, E),     /* BIT 6,E */
	BIT(6, H),     /* BIT 6,H */
	BIT(6, L),     /* BIT 6,L */
	BIT(6, HLind), /* BIT 6,(HL) */
	BIT(6, A),     /* BIT 6,A */
	BIT(7, B),     /* BIT 7,B */
	BIT(7, C),     /* BIT 7,C */
	BIT(7, D),     /* BIT 7,D */
	BIT(7, E),     /* BIT 7,E */
	BIT(7, H),     /* BIT 7,H */
	BIT(7, L),     /* BIT 7,L */
	BIT(7, HLind), /* BIT 7,(HL) */
	BIT(7, A),     /* BIT 7,A */
	/* 0x80 */
	SETRES(false, 0, B),     /* RES 0,B */
	SETRES(false, 0, C),     /* RES 0,C */
	SETRES(false, 0, D),     /* RES 0,D */
	SETRES(false, 0, E),     /* RES 0,E */
	SETRES(false, 0, H),     /* RES 0,H */
	SETRES(false, 0, L),     /* RES 0,L */
	SETRES(false, 0, HLind), /* RES 0,(HL) */
	SETRES(false, 0, A),     /* RES 0,A */
	SETRES(false, 1, B),     /* RES 1,B */
	SETRES(false, 1, C),     /* RES 1,C */
	SETRES(false, 1, D),     /* RES 1,D */
	SETRES(false, 1, E),     /* RES 7,E */
	SETRES(false, 1, H),     /* RES 1,H */
	SETRES(false, 1, L),     /* RES 1,L */
	SETRES(false, 1, HLind), /* RES 1,(HL) */
	SETRES(false, 1, A),     /* RES 1,A */
	/* 0x90 */
	SETRES(false, 2, B),     /* RES 2,B */
	SETRES(false, 2, C),     /* RES 2,C */
	SETRES(false, 2, D),     /* RES 2,D */
	SETRES(false, 2, E),     /* RES 2,E */
	SETRES(false, 2, H),     /* RES 2,H */
	SETRES(false, 2, L),     /* RES 2,L */
	SETRES(false, 2, HLind), /* RES 2,(HL) */
	SETRES(false, 2, A),     /* RES 2,A */
	SETRES(false, 3, B),     /* RES 3,B */
	SETRES(false, 3, C),     /* RES 3,C */
	SETRES(false, 3, D),     /* RES 3,D */
	SETRES(false, 3, E),     /* RES 3,E */
	SETRES(false, 3, H),     /* RES 3,H */
	SETRES(false, 3, L),     /* RES 3,L */
	SETRES(false, 3, HLind), /* RES 3,(HL) */
	SETRES(false, 3, A),     /* RES 3,A */
	/* 0xa0 */
	SETRES(false, 4, B),     /* RES 4,B */
	SETRES(false, 4, C),     /* RES 4,C */
	SETRES(false, 4, D),     /* RES 4,D */
	SETRES(false, 4, E),     /* RES 4,E */
	SETRES(false, 4, H),     /* RES 4,H */
	SETRES(false, 4, L),     /* RES 4,L */
	SETRES(false, 4, HLind), /* RES 4,(HL) */
	SETRES(false, 4, A),     /* RES 4,A */
	SETRES(false, 5, B),     /* RES 5,B */
	SETRES(false, 5, C),     /* RES 5,C */
	SETRES(false, 5, D),     /* RES 5,D */
	SETRES(false, 5, E),     /* RES 5,E */
	SETRES(false, 5, H),     /* RES 5,H */
	SETRES(false, 5, L),     /* RES 5,L */
	SETRES(false, 5, HLind), /* RES 5,(HL) */
	SETRES(false, 5, A),     /* RES 5,A */
	/* 0xb0 */
	SETRES(false, 6, B),     /* RES 6,B */
	SETRES(false, 6, C),     /* RES 6,C */
	SETRES(false, 6, D),     /* RES 6,D */
	SETRES(false, 6, E),     /* RES 6,E */
	SETRES(false, 6, H),     /* RES 6,H */
	SETRES(false, 6, L),     /* RES 6,L */
	SETRES(false, 6, HLind), /* RES 6,(HL) */
	SETRES(false, 6, A),     /* RES 6,A */
	SETRES(false, 7, B),     /* RES 7,B */
	SETRES(false, 7, C),     /* RES 7,C */
	SETRES(false, 7, D),     /* RES 7,D */
	SETRES(false, 7, E),     /* RES 7,E */
	SETRES(false, 7, H),     /* RES 7,H */
	SETRES(false, 7, L),     /* RES 7,L */
	SETRES(false, 7, HLind), /* RES 7,(HL) */
	SETRES(false, 7, A),     /* RES 7,A */
	/* 0xc0 */
	SETRES(true, 0, B),     /* SET 0,B */
	SETRES(true, 0, C),     /* SET 0,C */
	SETRES(true, 0, D),     /* SET 0,D */
	SETRES(true, 0, E),     /* SET 0,E */
	SETRES(true, 0, H),     /* SET 0,H */
	SETRES(true, 0, L),     /* SET 0,L */
	SETRES(true, 0, HLind), /* SET 0,(HL) */
	SETRES(true, 0, A),     /* SET 0,A */
	SETRES(true, 1, B),     /* SET 1,B */
	SETRES(true, 1, C),     /* SET 1,C */
	SETRES(true, 1, D),     /* SET 1,D */
	SETRES(true, 1, E),     /* SET 1,E */
	SETRES(true, 1, H),     /* SET 1,H */
	SETRES(true, 1, L),     /* SET 1,L */
	SETRES(true, 1, HLind), /* SET 1,(HL) */
	SETRES(true, 1, A),     /* SET 1,A */
	/* 0xd0 */
	SETRES(true, 2, B),     /* SET 2,B */
	SETRES(true, 2, C),     /* SET 2,C */
	SETRES(true, 2, D),     /* SET 2,D */
	SETRES(true, 2, E),     /* SET 2,E */
	SETRES(true, 2, H),     /* SET 2,H */
	SETRES(true, 2, L),     /* SET 2,L */
	SETRES(true, 2, HLind), /* SET 2,(HL) */
	SETRES(true, 2, A),     /* SET 2,A */
	SETRES(true, 3, B),     /* SET 3,B */
	SETRES(true, 3, C),     /* SET 3,C */
	SETRES(true, 3, D),     /* SET 3,D */
	SETRES(true, 3, E),     /* SET 3,E */
	SETRES(true, 3, H),     /* SET 3,H */
	SETRES(true, 3, L),     /* SET 3,L */
	SETRES(true, 3, HLind), /* SET 3,(HL) */
	SETRES(true, 3, A),     /* SET 3,A */
	/* 0xe0 */
	SETRES(true, 4, B),     /* SET 4,B */
	SETRES(true, 4, C),     /* SET 4,C */
	SETRES(true, 4, D),     /* SET 4,D */
	SETRES(true, 4, E),     /* SET 4,E */
	SETRES(true, 4, H),     /* SET 4,H */
	SETRES(true, 4, L),     /* SET 4,L */
	SETRES(true, 4, HLind), /* SET 4,(HL) */
	SETRES(true, 4, A),     /* SET 4,A */
	SETRES(true, 5, B),     /* SET 5,B */
	SETRES(true, 5, C),     /* SET 5,C */
	SETRES(true, 5, D),     /* SET 5,D */
	SETRES(true, 5, E),     /* SET 5,E */
	SETRES(true, 5, H),     /* SET 5,H */
	SETRES(true, 5, L),     /* SET 5,L */
	SETRES(true, 5, HLind), /* SET 5,(HL) */
	SETRES(true, 5, A),     /* SET 5,A */
	/* 0xf0 */
	SETRES(true, 6, B),     /* SET 6,B */
	SETRES(true, 6, C),     /* SET 6,C */
	SETRES(true, 6, D),     /* SET 6,D */
	SETRES(true, 6, E),     /* SET 6,E */
	SETRES(true, 6, H),     /* SET 6,H */
	SETRES(true, 6, L),     /* SET 6,L */
	SETRES(true, 6, HLind), /* SET 6,(HL) */
	SETRES(true, 6, A),     /* SET 6,A */
	SETRES(true, 7, B),     /* SET 7,B */
	SETRES(true, 7, C),     /* SET 7,C */
	SETRES(true, 7, D),     /* SET 7,D */
	SETRES(true, 7, E),     /* SET 7,E */
	SETRES(true, 7, H),     /* SET 7,H */
	SETRES(true, 7, L),     /* SET 7,L */
	SETRES(true, 7, HLind), /* SET 7,(HL) */
	SETRES(true, 7, A),     /* SET 7,A */
}
