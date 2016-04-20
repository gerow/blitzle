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
	a16 := uint16(a) + carryMod
	return (a16&0xf+uint16(b)&0xf)&0x10 != 0
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
	return uint16(a16)+uint16(b)&0x100 != 0
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
	b16 := uint16(a) + carryMod
	if b16&0xf > uint16(a)&0xf {
		return true
	}
	return false
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
	b16 := uint16(a) + carryMod
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
	cpu.ip = 0x100
	cpu.interrupts = true
	cpu.sp = 0xfffe
	return cpu
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
	o.WriteString(fmt.Sprintf("  IP: %04Xh\n", c.ip))
	o.WriteString(fmt.Sprintf("  SP: %04Xh\n", c.sp))
	o.WriteString(fmt.Sprintf("Area around IP\n"))

	for addr := c.ip - 10; addr < c.ip+10; addr++ {
		ipChar := "*"
		if addr != c.ip {
			ipChar = " "
		}
		o.WriteString(fmt.Sprintf("%s%04Xh: %02Xh\n", ipChar, addr, sys.Rb(addr)))
	}

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
		if mod != 0 {
			cpu.wrs(br, uint16(int(addr)+mod))
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
		if br == HLind {
			v = sys.Rb(cpu.rrs(HL))
			sys.Wb(cpu.rrs(HL), uint8(int(v)+mod))
		} else {
			v = cpu.rrb(br)
			cpu.wrb(br, uint8(int(v)+mod))
		}
		cpu.fz = v == 0
		cpu.fn = mod == -1
		cpu.fh = halfCarry(v, uint8(mod))

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
		if cpu.cond(con) {
			cpu.ip += j
			duration = 12
		} else {
			cpu.ip += 2
			duration = 8
		}
		/*
		 * Apparently the unconditional JR takes 12 reguardless?
		 * That doesn't sound right...
		 */
		if con == condNone {
			duration = 12
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
	a <<= 1
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
		h := uint8(hl >> 8)
		valHigh := uint8(val >> 8)

		cpu.fn = false
		cpu.fh = halfCarry(h, valHigh)
		cpu.fc = carry(h, valHigh)

		cpu.ip++
		return 8
	}
}

func LDBInd(destReg ByteRegister, srcAddrReg ShortRegister, mod int) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		addr := cpu.rrs(srcAddrReg)
		cpu.wrb(A, sys.Rb(addr))
		if mod != 0 {
			cpu.wrs(srcAddrReg, addr+1)
		}

		cpu.ip++
		return 8
	}
}

func RRCA(cpu *CPU, sys *Sys) int {
	a := cpu.rrb(A)
	carry := a&1 != 0
	a >>= 1
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
	a := cpu.rrb(A)
	carry := a > 99
	tens := a / 10
	a -= tens * 10
	ones := a

	/* I don't actually know how the CPU handles bad cases */
	newA := tens<<4 | ones
	cpu.wrb(A, newA)

	cpu.fz = newA == 0
	cpu.fh = false
	cpu.fc = carry

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
			cpu.fh = !halfBorrow(cpu.a, val)
			cpu.fc = !borrow(cpu.a, val)

			cpu.a -= val
		case SBC:
			cpu.fn = true
			cpu.fh = !halfBorrowWithC(cpu.a, val, cpu.fc)
			cpu.fc = !borrowWithC(cpu.a, val, cpu.fc)

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
			cpu.fh = !halfBorrowWithC(cpu.a, val, cpu.fc)
			cpu.fc = !borrowWithC(cpu.a, val, cpu.fc)

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
			cpu.sp += 2
			ra := sys.Rs(cpu.sp)
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
			sys.Ws(cpu.sp, cpu.ip+3)
			cpu.sp -= 2
			cpu.ip = addr
			return 24
		} else {
			cpu.ip += 3
			return 12
		}
	}
}

func JPHLind(cpu *CPU, sys *Sys) int {
	addr := sys.Rs(cpu.rrs(HL))
	cpu.ip = addr

	return 4
}

func POP(sr ShortRegister) OpFunc {
	if sr == AF {
		panic("no, that won't work!")
	}

	return func(cpu *CPU, sys *Sys) int {
		cpu.sp -= 2
		cpu.wrs(sr, sys.Rs(cpu.sp))

		cpu.ip++
		return 12
	}
}

func POPAF(cpu *CPU, sys *Sys) int {
	cpu.sp -= 2
	af := sys.Rs(cpu.sp)
	a := uint8(af >> 8)
	f := uint8(af & 0xf)

	cpu.setFlags(f)
	cpu.a = a

	cpu.ip++
	return 12
}

func PUSH(sr ShortRegister) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		sys.Ws(cpu.sp, cpu.rrs(sr))
		cpu.sp += 2

		cpu.ip++
		return 16
	}
}

func RST(addr uint16) OpFunc {
	return func(cpu *CPU, sys *Sys) int {
		sys.Ws(cpu.sp, cpu.ip+1)
		cpu.sp -= 2
		cpu.ip = addr

		return 16
	}
}

func PANIC(cpu *CPU, sys *Sys) int {
	panic("This should never get called!")
}

func DRAGONS(cpu *CPU, sys *Sys) int {
	log.Printf("This op shouldn't exist!")

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
	cpu.fh = halfCarry(uint8(cpu.sp>>8), uint8(v>>8))
	cpu.fc = carry(uint8(cpu.sp>>8), uint8(v>>8))

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
	INCDECS(DE, -1),  /* INC DE */
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
	RST(0x00),            /* RST 00H */
	RET(condZ, false),    /* RET Z */
	RET(condNone, false), /* RET */
	JP(condZ),            /* JP Z,a16 */
	PANIC,                /* PREFIX CB */
	CALL(condZ),          /* CALL Z,a16 */
	CALL(condNone),       /* CALL a16 */
	ALU(ADC, Imm),        /* ADC A,d8 */
	RST(0x08),            /* RST 08H */
	/* 0xd0 */
	RET(condNC, false),  /* RET NC */
	POP(DE),             /* POP DE */
	JP(condNC),          /* JP NC,a16 */
	DRAGONS,             /* XXX */
	CALL(condNC),        /* CALL NC,a16 */
	PUSH(DE),            /* PUSH DE */
	ALU(SUB, Imm),       /* SUB A,d8 */
	RST(0x10),           /* RST 10H */
	RET(condC, false),   /* RET C */
	RET(condNone, true), /* RETI */
	JP(condC),           /* JP C,a16 */
	DRAGONS,             /* XXX */
	CALL(condC),         /* CALL C,a16 */
	DRAGONS,             /* XXX */
	ALU(SBC, Imm),       /* SBC A,d8 */
	RST(0x18),           /* RST 18H */
	/* 0xe0 */
	LDH(true),        /* LDH (a8),A */
	POP(HL),          /* POP HL */
	LDHC(true),       /* LD (C),A */
	DRAGONS,          /* XXX */
	DRAGONS,          /* XXX */
	PUSH(HL),         /* PUSH HL */
	ALU(AND, Imm),    /* AND A,d8 */
	RST(0x20),        /* RST 20H */
	ADDSPimm,         /* ADD SP,r8 */
	JPHLind,          /* JP (HL) */
	LDSimmAddr(true), /* LD (a16),A */
	DRAGONS,          /* XXX */
	DRAGONS,          /* XXX */
	DRAGONS,          /* XXX */
	ALU(XOR, Imm),    /* XOR A,d8 */
	RST(0x28),        /* RST 28H */
	/* 0xf0 */
	LDH(false),   /* LDH A,(a8) */
	POPAF,        /* POP AF */
	LDHC(false),  /* LD A,(C) */
	DI,           /* DI */
	DRAGONS,      /* XXX */
	PUSH(AF),     /* PUSH AF */
	ALU(OR, Imm), /* OR A,d8 */
	RST(0x30),    /* RST 30H */
	NOP,
	NOP,
	LDSimmAddr(true), /* LD (a16),A */
	EI,               /* EI */
	DRAGONS,          /* XXX */
	DRAGONS,          /* XXX */
	ALU(CP, Imm),     /* CP A,d8 */
	RST(0x38),        /* RST 38H */
}

var cbops [0x100]OpFunc = [0x100]OpFunc{
	/* 0x00 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x10 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x20 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x30 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x40 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x50 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x60 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x70 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x80 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0x90 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0xa0 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0xb0 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0xc0 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0xd0 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0xe0 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	/* 0xf0 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
}
