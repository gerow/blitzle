package gb

import (
	"bytes"
	"fmt"
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
}

func halfCarry(a uint8, b uint8) bool {
	return (a&0xf+b&0xf)&0x10 != 0
}

func carry(a uint8, b uint8) bool {
	return uint16(a)+uint16(b)&0x100 != 0
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
	default:
		panic("received invalid sr")
	}
}

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
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	LDBInd(B, HL, 0), /* LD B,(HL) */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	LDBInd(C, HL, 0), /* LD C,(HL) */
	NOP,
	/* 0x50 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	LDBInd(D, HL, 0), /* LD D,(HL) */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	LDBInd(E, HL, 0), /* LD E,(HL) */
	NOP,
	/* 0x60 */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	LDBInd(H, HL, 0), /* LD H,(HL) */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	LDBInd(L, HL, 0), /* LD L,(HL) */
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
	LDBInd(A, HL, 0), /* LD A,(HL) */
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
