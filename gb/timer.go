package gb

type Timer struct {
	divReg  MemRegister
	timaReg MemRegister
	tmaReg  MemRegister
	tacReg  MemRegister

	devs []BusDev
}

const divRegAddr uint16 = 0xff04
const divFreq uint = 16384

var tacValToFreq map[uint8]uint = map[uint8]uint{
	0: 4096,
	1: 262144,
	2: 65536,
	3: 16384,
}

func NewTimer() *Timer {
	divReg := NewMemRegister(divRegAddr)
	timaReg := NewMemRegister(0xff05)
	tmaReg := NewMemRegister(0xff06)
	tacReg := NewMemRegister(0xff07)

	devs := []BusDev{divReg, timaReg, tmaReg, tacReg}

	return &Timer{*divReg, *timaReg, *tmaReg, *tacReg, devs}
}

func (t *Timer) getHandler(addr uint16) BusDev {
	for _, bd := range t.devs {
		if bd.Asserts(addr) {
			return bd
		}
	}
	return nil
}

func (t *Timer) Step(sys *Sys) {
	// Handle div register first
	if sys.FreqStep(divFreq) {
		t.divReg.set(t.divReg.val() + 1)
	}

	// Now handle timer
	tac := t.tacReg.val()
	// Timer is disabled, so don't do anything more
	if tac&0x04 == 0 {
		return
	}
	if sys.FreqStep(tacValToFreq[tac&0x03]) {
		if t.timaReg.val() == 0xff {
			// Overflow!
			t.timaReg.set(t.tmaReg.val())
			sys.RaiseInterrupt(TimerInterrupt)
		} else {
			// Not overflow!
			t.timaReg.set(t.timaReg.val() + 1)
		}
	}
}

func (t *Timer) R(addr uint16) uint8 {
	return t.getHandler(addr).R(addr)
}

func (t *Timer) W(addr uint16, val uint8) {
	// Have a special case for any kind of a write to divReg, which should set
	// it to 0
	if addr == divRegAddr {
		t.divReg.set(0)
		return
	}
	t.getHandler(addr).W(addr, val)
}

func (t *Timer) Asserts(addr uint16) bool {
	return t.getHandler(addr) != nil
}
