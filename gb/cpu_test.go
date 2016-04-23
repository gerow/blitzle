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
	return NewSys(*r)
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

func checkBr(t *testing.T, s *Sys, br ByteRegister, expected uint8) {
	if v := s.cpu.rrb(br); v != expected {
		t.Errorf("Expected reg=%02Xh, got %02Xh\n", expected, v)
	}
}

func TestNOP(t *testing.T) {
	s := S([]byte{
		0x00, // NOP
	})

	checkStep(t, s, 4)
	checkIP(t, s, 0x101)
}

func TestLD(t *testing.T) {
	s := S([]byte{
		0x01, 0x34, 0x12, // LD BC,$1234
	})

	checkStep(t, s, 12)
	checkIP(t, s, 0x103)
	checkBr(t, s, B, 0x12)
	checkBr(t, s, C, 0x34)
}
