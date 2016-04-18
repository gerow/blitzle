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

func NewCPU() *CPU {
	cpu := &CPU{}
	cpu.ip = 0x100
	return cpu
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

var ops [0x100]OpFunc = [0x100]OpFunc{
	/* 0x00 */
	NOP,              /* NOP */
	LDSImm(BC),       /* LD BC,d16 */
	LDARegInd(BC, 0), /* LD (BC),A */
	INCDECS(BC, 1),   /* INC BC */
	INCDECB(B, 1),    /* INC B */
	INCDECB(B, -1),   /* DEC B */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	INCDECS(BC, -1), /* DEC BC */
	INCDECB(C, 1),   /* INC C */
	INCDECB(C, -1),  /* DEC C */
	NOP,
	NOP,
	/* 0x10 */
	NOP,
	LDSImm(DE),       /* LD DE,d16 */
	LDARegInd(DE, 0), /* LD (DE),A */
	INCDECS(DE, -1),  /* INC DE */
	INCDECB(D, 1),    /* INC D */
	INCDECB(D, -1),   /* DEC D */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	INCDECS(DE, -1), /* DEC DE */
	INCDECB(E, 1),   /* INC E */
	INCDECB(E, -1),  /* DEC E */
	NOP,
	NOP,
	/* 0x20 */
	NOP,
	LDSImm(HL),       /* LD HL,d16 */
	LDARegInd(HL, 1), /* LD (HL+),A */
	INCDECS(HL, 1),   /* INC HL */
	INCDECB(H, 1),    /* INC H */
	INCDECB(H, -1),   /* DEC H */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	INCDECS(HL, -1), /* DEC HL */
	INCDECB(L, 1),   /* INC L */
	INCDECB(L, -1),  /* DEC L */
	NOP,
	NOP,
	/* 0x30 */
	NOP,
	LDSImm(SP),         /* LD SP,d16 */
	LDARegInd(HL, -1),  /* LD (HL-),A */
	INCDECS(SP, 1),     /* INC SP */
	INCDECB(HLind, 1),  /* INC (HL) */
	INCDECB(HLind, -1), /* DEC (HL) */
	NOP,
	NOP,
	NOP,
	NOP,
	NOP,
	INCDECS(SP, -1), /* DEC SP */
	INCDECB(A, 1),   /* INC A */
	INCDECB(A, -1),  /* DEC A */
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
