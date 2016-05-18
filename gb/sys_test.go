package gb

import "testing"

func TestEverythingAsserts(t *testing.T) {
	r := FakeROM([]byte{})
	s := NewSys(r)
	for addr := uint(0); addr < 0x10000; addr++ {
		if s.getHandler(uint16(addr)) == nil {
			t.Errorf("no handler for %04Xh\n", addr)
		}
	}
}

func TestCorrectThingAsserts(t *testing.T) {
	r := FakeROM([]byte{})
	s := NewSys(r)

	for addr := uint(0); addr < 0x8000; addr++ {
		if s.getHandler(uint16(addr)) != s.rom {
			t.Errorf("expected %04Xh to be handled by rom\n", addr)
		}
	}
	for addr := uint(0x8000); addr < 0xa000; addr++ {
		if s.getHandler(uint16(addr)) != s.video {
			t.Errorf("expected %04Xh to be handled by video\n", addr)
		}
	}
	for addr := uint(0xc000); addr < 0xfe00; addr++ {
		if s.getHandler(uint16(addr)) != s.systemRAM {
			t.Errorf("expected %04Xh to be handled by systemRAM\n", addr)
		}
	}
	for addr := uint(0xfe00); addr < 0xfea0; addr++ {
		if s.getHandler(uint16(addr)) != s.video {
			t.Errorf("expected %04Xh to be handled by video\n", addr)
		}
	}
	for addr := uint(0xff80); addr < 0xffff; addr++ {
		if s.getHandler(uint16(addr)) != s.hiRAM {
			t.Errorf("expected %04Xh to be handled by hiRAM\n", addr)
		}
	}

	// And now for specific registers
	if s.getHandler(0xffff) != s.ieReg {
		t.Errorf("expected %04Xh to be handled by ieReg\n", 0xffff)
	}
	if s.getHandler(0xff0f) != s.ifReg {
		t.Errorf("expected %04Xh to be handled by ifReg\n", 0xff0f)
	}

	vidRegs := []uint16{0xff41, 0xff40, 0xff48, 0xff42, 0xff43, 0xff44, 0xff45,
		0xff4a, 0xff4b, 0xff47, 0xff48, 0xff49, 0xff46}
	for _, addr := range vidRegs {
		if s.getHandler(addr) != s.video {
			t.Errorf("expected %04Xh to be handled by video, got %+v\n", addr, s.getHandler(addr))
		}
	}
	timerRegs := []uint16{0xff04, 0xff05, 0xff06, 0xff07}
	for _, addr := range timerRegs {
		if s.getHandler(addr) != s.timer {
			t.Errorf("expected %04Xh to be handled by timer, got %+v\n", addr, s.getHandler(addr))
		}
	}
}
