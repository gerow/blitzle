package gb

import (
	"log"
	"testing"
)

var roms = []string{
	"../third_party/gblargg/cpu_instrs/individual/01-special.gb",
	//"../third_party/gblargg/cpu_instrs/individual/02-interrupts.gb",
	"../third_party/gblargg/cpu_instrs/individual/03-op sp,hl.gb",
	"../third_party/gblargg/cpu_instrs/individual/04-op r,imm.gb",
	"../third_party/gblargg/cpu_instrs/individual/05-op rp.gb",
	"../third_party/gblargg/cpu_instrs/individual/06-ld r,r.gb",
	"../third_party/gblargg/cpu_instrs/individual/07-jr,jp,call,ret,rst.gb",
	"../third_party/gblargg/cpu_instrs/individual/08-misc instrs.gb",
	"../third_party/gblargg/cpu_instrs/individual/09-op r,r.gb",
	"../third_party/gblargg/cpu_instrs/individual/10-bit ops.gb",
	"../third_party/gblargg/cpu_instrs/individual/11-op a,(hl).gb",
}

const (
	testLength  = 100000000
	passKeyword = "Passed"
)

type BufferSerialSwapper struct {
	buf []uint8
}

func (b *BufferSerialSwapper) SerialSwap(v uint8) uint8 {
	b.buf = append(b.buf, v)
	return 0xff
}

func (b *BufferSerialSwapper) passed() bool {
	if len(b.buf) < len(passKeyword) {
		return false
	}
	checkSlice := b.buf[len(b.buf)-len(passKeyword):]
	for i, b := range []byte(passKeyword) {
		if b != checkSlice[i] {
			return false
		}
	}
	return true
}

func TestROMs(t *testing.T) {
	for _, rom := range roms {
		r, err := LoadROMFromFile(rom)
		if err != nil {
			log.Fatal(err)
		}
		sys := NewSys(r)
		serialBuffer := &BufferSerialSwapper{}
		sys.SetSerialSwapper(serialBuffer)
		passed := false
		for i := 0; i < testLength; i++ {
			sys.Step()
			if serialBuffer.passed() {
				passed = true
				break
			}
		}
		if !passed {
			t.Errorf("ROM %s failed\n", rom)
		}
	}
}
