package gb

import "testing"

func FakeROM(codeBytes []byte) *ROM {
	/* Make a fake 32k ROM */
	romData := make([]byte, 0x8000)
	ip := 0x0100
	for _, b := range codeBytes {
		romData[ip] = b
		ip++
	}

	rom, err := LoadROM(romData)
	if err != nil {
		panic(err)
	}
	return rom
}

func S(codeBytes []byte) *Sys {
	r := FakeROM(codeBytes)
	s := NewSys(r, nil, nil)

	/* Initialize hl to point to some location in RAM */
	s.cpu.h = 0xc0
	s.cpu.l = 0x00

	/* And initialize SP to point to the top of normal memory */
	s.cpu.sp = 0xcfff

	/* Initialize all the gp registers to interesting values */
	s.cpu.a = 0x01
	s.cpu.b = 0x02
	s.cpu.c = 0x03
	s.cpu.d = 0x04
	s.cpu.e = 0x05

	return s
}

func checkStep(t *testing.T, s *Sys, expected int) {
	if c := s.cpu.Step(s); c != expected {
		t.Errorf("expected %d steps, got %d\n", expected, c)
	}
}

func checkIP(t *testing.T, s *Sys, expected uint16) {
	if s.cpu.ip != expected {
		t.Errorf("Expected IP=%04Xh, got %04Xh\n", expected, s.cpu.ip)
	}
}

func checkSP(t *testing.T, s *Sys, expected uint16) {
	if s.cpu.sp != expected {
		t.Errorf("Expected SP=%04Xh, got %04Xh\n", expected, s.cpu.sp)
	}
}

func checkBr(t *testing.T, s *Sys, br ByteRegister, expected uint8) {
	if v := s.cpu.rrb(br); v != expected {
		t.Errorf("Expected %s=%02Xh, got %02Xh\n", ByteRegisterNameMap[br], expected, v)
	}
}

func checkBus(t *testing.T, s *Sys, addr uint16, expected uint8) {
	if v := s.Rb(addr); v != expected {
		t.Errorf("Expected (%04Xh)=%02Xh, got %02Xh\n", addr, expected, v)
	}
}

func checkFlags(t *testing.T, s *Sys, expected uint8) {
	if v := s.cpu.flags(); v != expected {
		t.Errorf("Expected flags=%02Xh, got %02Xh\n", expected, v)
	}
}

func TestNOP(t *testing.T) {
	s := S([]byte{
		0x00, // NOP
	})

	checkStep(t, s, 4)
	checkIP(t, s, 0x101)
}

/* LD xx,$xxxx */
func TestLDBCd16(t *testing.T) {
	s := S([]byte{
		0x01, 0x34, 0x12, // LD BC,$1234
	})

	checkStep(t, s, 12)
	checkIP(t, s, 0x103)
	checkBr(t, s, B, 0x12)
	checkBr(t, s, C, 0x34)
}

func TestLDDEd16(t *testing.T) {
	s := S([]byte{
		0x11, 0x34, 0x12, // LD DE,$1234
	})

	checkStep(t, s, 12)
	checkIP(t, s, 0x103)
	checkBr(t, s, D, 0x12)
	checkBr(t, s, E, 0x34)
}

func TestLDHLd16(t *testing.T) {
	s := S([]byte{
		0x21, 0x34, 0x12, // LD HL,$1234
	})

	checkStep(t, s, 12)
	checkIP(t, s, 0x103)
	checkBr(t, s, H, 0x12)
	checkBr(t, s, L, 0x34)
}

func TestLDSPd16(t *testing.T) {
	s := S([]byte{
		0x31, 0x34, 0x12, // LD SP,$1234
	})

	checkStep(t, s, 12)
	checkIP(t, s, 0x103)
	checkSP(t, s, 0x1234)
}

/* LD (xx),A */
func TestLDBCindA(t *testing.T) {
	s := S([]byte{
		0x02, // LD (BC),A
	})
	s.cpu.b = 0xc0
	s.cpu.c = 0x00

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBus(t, s, 0xc000, s.cpu.a)
}

func TestLDDEindA(t *testing.T) {
	s := S([]byte{
		0x12, // LD (DE),A
	})
	s.cpu.d = 0xc0
	s.cpu.e = 0x00

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBus(t, s, 0xc000, s.cpu.a)
}

func TestLDHLIindA(t *testing.T) {
	s := S([]byte{
		0x22, // LD (HL+),A
	})
	s.cpu.h = 0xc0
	s.cpu.l = 0x00

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBus(t, s, 0xc000, s.cpu.a)
	checkBr(t, s, H, 0xc0)
	checkBr(t, s, L, 0x01)
}

func TestLDHLDindA(t *testing.T) {
	s := S([]byte{
		0x32, // LD (HL-),A
	})
	s.cpu.h = 0xc0
	s.cpu.l = 0x00

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBus(t, s, 0xc000, s.cpu.a)
	checkBr(t, s, H, 0xbf)
	checkBr(t, s, L, 0xff)
}

/* INC xx */
func TestINCBC(t *testing.T) {
	s := S([]byte{
		0x03, // INC BC
	})
	s.cpu.b = 0x10
	s.cpu.c = 0xff

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBr(t, s, B, 0x11)
	checkBr(t, s, C, 0x00)
}

func TestINCDE(t *testing.T) {
	s := S([]byte{
		0x13, // INC DE
	})
	s.cpu.d = 0x10
	s.cpu.e = 0xff

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBr(t, s, D, 0x11)
	checkBr(t, s, E, 0x00)
}

func TestINCHL(t *testing.T) {
	s := S([]byte{
		0x23, // INC HL
	})
	s.cpu.h = 0x10
	s.cpu.l = 0xff

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBr(t, s, H, 0x11)
	checkBr(t, s, L, 0x00)
}

func TestINCSP(t *testing.T) {
	s := S([]byte{
		0x33, // INC SP
	})
	s.cpu.sp = 0x10ff

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkSP(t, s, 0x1100)
}

/* DEC xx */
func TestDECBC(t *testing.T) {
	s := S([]byte{
		0x0b, // DEC BC
	})
	s.cpu.b = 0x10
	s.cpu.c = 0x00

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBr(t, s, B, 0x0f)
	checkBr(t, s, C, 0xff)
}

func TestDECDE(t *testing.T) {
	s := S([]byte{
		0x1b, // DEC DE
	})
	s.cpu.d = 0x10
	s.cpu.e = 0x00

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBr(t, s, D, 0x0f)
	checkBr(t, s, E, 0xff)
}

func TestDECHL(t *testing.T) {
	s := S([]byte{
		0x2b, // DEC HL
	})
	s.cpu.h = 0x10
	s.cpu.l = 0x00

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkBr(t, s, H, 0x0f)
	checkBr(t, s, L, 0xff)
}

func TestDECSP(t *testing.T) {
	s := S([]byte{
		0x3b, // DEC SP
	})
	s.cpu.sp = 0x1000

	checkStep(t, s, 8)
	checkIP(t, s, 0x101)
	checkSP(t, s, 0x0fff)
}

func TestJPHL(t *testing.T) {
	s := S([]byte{
		0xe9, // JP (HL)
	})
	s.cpu.h = 0xab
	s.cpu.l = 0xcd
	checkStep(t, s, 4)
	checkIP(t, s, 0xabcd)
}

func TestADD(t *testing.T) {
	s := S([]byte{
		0x80, // ADD A,B
	})

	// No carry
	s.cpu.setFlags(^uint8(0x00))
	s.cpu.a = 1
	s.cpu.b = 1
	checkStep(t, s, 4)
	if s.cpu.a != 0x02 {
		t.Errorf("Expected A=02h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x00)

	// Half carry
	s.cpu.ip = 0x100
	s.cpu.setFlags(^uint8(0x20))
	s.cpu.a = 0x0f
	s.cpu.b = 0x0f
	checkStep(t, s, 4)
	if s.cpu.a != 0x1e {
		t.Errorf("Expected A=1Eh, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x20)

	// Full carry, no half
	s.cpu.ip = 0x100
	s.cpu.setFlags(^uint8(0x10))
	s.cpu.a = 0xf0
	s.cpu.b = 0xf0
	checkStep(t, s, 4)
	if s.cpu.a != 0xe0 {
		t.Errorf("Expected A=E0h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x10)

	// Full and half carry
	s.cpu.ip = 0x100
	s.cpu.setFlags(^uint8(0x30))
	s.cpu.a = 0xff
	s.cpu.b = 0xff
	checkStep(t, s, 4)
	if s.cpu.a != 0xfe {
		t.Errorf("Expected A=FEh, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x30)

	// All carry to zero
	s.cpu.ip = 0x100
	s.cpu.setFlags(^uint8(0xb0))
	s.cpu.a = 0xff
	s.cpu.b = 0x01
	checkStep(t, s, 4)
	if s.cpu.a != 0x00 {
		t.Errorf("Expected A=00h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0xb0)
}

func TestDAAafterAdd(t *testing.T) {
	// Add, no carry
	s := S([]byte{
		0x80, // ADD B
		0x27, // DAA
	})
	s.cpu.a = 0x12
	s.cpu.b = 0x34
	s.cpu.setFlags(^uint8(0x00))
	checkStep(t, s, 4)
	checkStep(t, s, 4)
	if s.cpu.a != 0x46 {
		t.Errorf("Expected A=46h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x00)

	// Add, ones carry, no true half carry
	s.cpu.ip = 0x100

	s.cpu.a = 0x12
	s.cpu.b = 0x39
	s.cpu.setFlags(^uint8(0x00))
	checkStep(t, s, 4)
	checkStep(t, s, 4)
	if s.cpu.a != 0x51 {
		t.Errorf("Expected A=51h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x00)

	// Add, true half carry
	s.cpu.ip = 0x100

	s.cpu.a = 0x19
	s.cpu.b = 0x39
	s.cpu.setFlags(^uint8(0x00))
	checkStep(t, s, 4)
	checkFlags(t, s, 0x20)
	checkStep(t, s, 4)
	if s.cpu.a != 0x58 {
		t.Errorf("Expected A=58h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x00)

	// Add, tens carry, no true carry
	s.cpu.ip = 0x100

	s.cpu.a = 0x50
	s.cpu.b = 0x60
	s.cpu.setFlags(^uint8(0x10))
	checkStep(t, s, 4)
	checkFlags(t, s, 0x00)
	checkStep(t, s, 4)
	if s.cpu.a != 0x10 {
		t.Errorf("Expected A=10h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x10)

	// Add, tens carry, true carry
	s.cpu.ip = 0x100

	s.cpu.a = 0x90
	s.cpu.b = 0x90
	s.cpu.setFlags(^uint8(0x10))
	checkStep(t, s, 4)
	checkFlags(t, s, 0x10)
	checkStep(t, s, 4)
	if s.cpu.a != 0x80 {
		t.Errorf("Expected A=80h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x10)

	// Everything carries
	s.cpu.ip = 0x100

	s.cpu.a = 0x99
	s.cpu.b = 0x99
	s.cpu.setFlags(^uint8(0x10))
	checkStep(t, s, 4)
	checkFlags(t, s, 0x30)
	checkStep(t, s, 4)
	if s.cpu.a != 0x98 {
		t.Errorf("Expected A=98h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x10)

	// Carry to zero
	s.cpu.ip = 0x100

	s.cpu.a = 0x99
	s.cpu.b = 0x01
	s.cpu.setFlags(^uint8(0x40))
	checkStep(t, s, 4)
	checkFlags(t, s, 0x00)
	checkStep(t, s, 4)
	if s.cpu.a != 0x00 {
		t.Errorf("Expected A=00h, got %02Xh\n", s.cpu.a)
	}
	checkFlags(t, s, 0x90)
}

func TestDAAafterSub(t *testing.T) {
	s := S([]byte{
		0x90, // SUB B
		0x27, // DAA
	})

	// Half carry
	s.cpu.a = 0x30
	s.cpu.b = 0x04
	checkStep(t, s, 4)
	checkFlags(t, s, 0x60)
	checkStep(t, s, 4)
	checkFlags(t, s, 0x40)
	if s.cpu.a != 0x26 {
		t.Errorf("Expected A=26h, got %02Xh\n", s.cpu.a)
	}

	// Full carry
	s.cpu.ip = 0x100

	s.cpu.a = 0x30
	s.cpu.b = 0x40
	checkStep(t, s, 4)
	checkFlags(t, s, 0x50)
	checkStep(t, s, 4)
	checkFlags(t, s, 0x50)
	if s.cpu.a != 0x90 {
		t.Errorf("Expected A=90h, got %02Xh\n", s.cpu.a)
	}

	// Full and half carry
	s.cpu.ip = 0x100

	s.cpu.a = 0x30
	s.cpu.b = 0x41
	checkStep(t, s, 4)
	checkFlags(t, s, 0x70)
	checkStep(t, s, 4)
	checkFlags(t, s, 0x50)
	if s.cpu.a != 0x89 {
		t.Errorf("Expected A=89h, got %02Xh\n", s.cpu.a)
	}

	// Zero
	s.cpu.ip = 0x100

	s.cpu.a = 0x30
	s.cpu.b = 0x30
	checkStep(t, s, 4)
	checkFlags(t, s, 0xc0)
	checkStep(t, s, 4)
	checkFlags(t, s, 0xc0)
	if s.cpu.a != 0x00 {
		t.Errorf("Expected A=00h, got %02Xh\n", s.cpu.a)
	}
}

func TestCP(t *testing.T) {
	s := S([]byte{
		0xfe, 0x01, // CP 1
	})
	s.cpu.a = 0
	checkStep(t, s, 8)
	if s.cpu.ip != 0x102 {
		t.Errorf("Expected CP d8 to be two bytes wide")
	}
	checkFlags(t, s, 0x70)
	if s.cpu.a != 0x00 {
		t.Errorf("Expected A=00h, got %02Xh\n", s.cpu.a)
	}
}

// LD nn,nn
func TestLD8Bit(t *testing.T) {
	s := S([]byte{
		0x40, // LD B,B
		0x41, // LD B,C
		0x42, // LD B,D
		0x43, // LD B,E
		0x44, // LD B,H
		0x45, // LD B,L
		0x46, // LD B,(HL)
		0x47, // LD B,A

		0x48, // LD C,B
		0x49, // LD C,C
		0x4a, // LD C,D
		0x4b, // LD C,E
		0x4c, // LD C,H
		0x4d, // LD C,L
		0x4e, // LD C,(HL)
		0x4f, // LD C,A
	})
	s.cpu.b = 1
	s.cpu.c = 2
	s.cpu.d = 3
	s.cpu.e = 4
	s.cpu.h = 0xc0
	s.cpu.l = 0x05
	s.Wb(0xc005, 6)
	s.cpu.a = 7

	// Check target B
	checkStep(t, s, 4) // LD B,B
	checkIP(t, s, 0x101)
	checkBr(t, s, B, 1)

	checkStep(t, s, 4) // LD B,C
	checkIP(t, s, 0x102)
	checkBr(t, s, B, 2)

	checkStep(t, s, 4) // LD B,D
	checkIP(t, s, 0x103)
	checkBr(t, s, B, 3)

	checkStep(t, s, 4) // LD B,E
	checkIP(t, s, 0x104)
	checkBr(t, s, B, 4)

	checkStep(t, s, 4) // LD B,H
	checkIP(t, s, 0x105)
	checkBr(t, s, B, 0xc0)

	checkStep(t, s, 4) // LD B,L
	checkIP(t, s, 0x106)
	checkBr(t, s, B, 0x05)

	checkStep(t, s, 8) // LD B,(HL)
	checkIP(t, s, 0x107)
	checkBr(t, s, B, 6)

	checkStep(t, s, 4) // LD B,A
	checkIP(t, s, 0x108)
	checkBr(t, s, B, 7)

	// Reset b to 1
	s.cpu.b = 1

	// Check target C
	checkStep(t, s, 4) // LD C,B
	checkIP(t, s, 0x109)
	checkBr(t, s, C, 1)

	// Reset c to 2
	s.cpu.c = 2

	checkStep(t, s, 4) // LD C,C
	checkIP(t, s, 0x10a)
	checkBr(t, s, C, 2)

	checkStep(t, s, 4) // LD C,D
	checkIP(t, s, 0x10b)
	checkBr(t, s, C, 3)

	checkStep(t, s, 4) // LD C,E
	checkIP(t, s, 0x10c)
	checkBr(t, s, C, 4)

	checkStep(t, s, 4) // LD C,H
	checkIP(t, s, 0x10d)
	checkBr(t, s, C, 0xc0)

	checkStep(t, s, 4) // LD C,L
	checkIP(t, s, 0x10e)
	checkBr(t, s, C, 0x05)

	checkStep(t, s, 8) // LD C,(HL)
	checkIP(t, s, 0x10f)
	checkBr(t, s, C, 6)

	checkStep(t, s, 4) // LD C,A
	checkIP(t, s, 0x110)
	checkBr(t, s, C, 7)
}
