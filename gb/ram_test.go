package gb

import "testing"

func TestRAMAsserts(t *testing.T) {
	r := NewRAM(0x800, 0xfff)
	for addr := uint(0); addr < 0x800; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected RAM not to assert %04Xh\n", addr)
		}
	}
	for addr := uint(0x800); addr < 0x1000; addr++ {
		if !r.Asserts(uint16(addr)) {
			t.Errorf("expected RAM to assert %04Xh\n", addr)
		}
	}
	for addr := uint(0x1000); addr < 0x10000; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected RAM not to assert %04Xh\n", addr)
		}
	}
}

func TestRAMStorageAndRetrieval(t *testing.T) {
	r := NewRAM(0x800, 0xfff)
	val := uint8(0)

	for addr := uint(0); addr < 0x1000; addr++ {
		r.W(uint16(addr), val)
		val++
	}
	val = 0
	for addr := uint(0); addr < 0x1000; addr++ {
		if got := r.R(uint16(addr)); got != val {
			t.Errorf("expected r.R(%04Xh) to be %d, got %d\n", val, got)
		}
		val++
	}
}

func TestMemRegisterAsserts(t *testing.T) {
	r := NewMemRegister(0xff10)
	for addr := uint(0); addr < 0xff10; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected emRegister not to assert %04Xh\n", addr)
		}
	}
	if !r.Asserts(uint16(0xff10)) {
		t.Errorf("expected MemRegister to assert %04Xh\n", 0xff10)
	}
	for addr := uint(0xff11); addr < 0x10000; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected MemRegister not to assert %04Xh\n", addr)
		}
	}
}

func TestMemRegisterStorageAndRetrieval(t *testing.T) {
	r := NewMemRegister(0xff10)
	r.W(0xff10, 0x33)
	if got := r.R(0xff10); got != 0x33 {
		t.Errorf("expected r.R() to be %d, got %d\n", 0x33, got)
	}
}

func TestReadOnlyRegisterAsserts(t *testing.T) {
	r := ReadOnlyRegister{0xff10, func() uint8 { return 0x33 }}
	for addr := uint(0); addr < 0xff10; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected ReadOnlyRegister not to assert %04Xh\n", addr)
		}
	}
	if !r.Asserts(uint16(0xff10)) {
		t.Errorf("expected ReadOnlyRegister to assert %04Xh\n", 0xff10)
	}
	for addr := uint(0xff11); addr < 0x10000; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected ReadOnlyRegister not to assert %04Xh\n", addr)
		}
	}
}

func TestReadOnlyRegisterRetrieval(t *testing.T) {
	r := ReadOnlyRegister{0xff10, func() uint8 { return 0x33 }}
	r.W(0xff10, 0xfe)
	if got := r.R(0xff10); got != 0x33 {
		t.Errorf("expected r.R() to be %d, got %d\n", 0x33, got)
	}
}

func TestWriteOnlyRegisterAsserts(t *testing.T) {
	r := WriteOnlyRegister{0xff10, func(_ uint8) {}}
	for addr := uint(0); addr < 0xff10; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected WriteOnlyRegister not to assert %04Xh\n", addr)
		}
	}
	if !r.Asserts(uint16(0xff10)) {
		t.Errorf("expected WriteOnlyRegister to assert %04Xh\n", 0xff10)
	}
	for addr := uint(0xff11); addr < 0x10000; addr++ {
		if r.Asserts(uint16(addr)) {
			t.Errorf("expected WriteOnlyRegister not to assert %04Xh\n", addr)
		}
	}
}

func TestWriteOnlyRegisterWriteAndRead(t *testing.T) {
	called := false
	calledPtr := &called
	val := uint8(0)
	valPtr := &val
	r := WriteOnlyRegister{0xff10, func(v uint8) {
		*calledPtr = true
		*valPtr = v
	}}

	r.W(0xff10, 0x33)
	if !called {
		t.Errorf("expected func to be called\n")
	}
	if val != 0x33 {
		t.Errorf("expected %d, got %d", 0x33, val)
	}

	if got := r.R(0xff10); got != 0xff {
		t.Errorf("Expected r.R() to be %d, got %d", 0xff, got)
	}
}
