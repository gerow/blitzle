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
	s := NewSys(r, nil)

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
		t.Errorf("Expected reg=%02Xh, got %02Xh\n", expected, v)
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

func TestDAA(t *testing.T) {
	// Add, no carry
	s := S([]byte{
		0x80, // ADD A,B
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
