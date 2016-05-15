package gb

type ButtonState struct {
	Down         bool
	Up           bool
	Left         bool
	Right        bool
	Start        bool
	SelectButton bool
	B            bool
	A            bool
}

type Joypad struct {
	val   uint8
	state ButtonState
}

func NewJoypad() *Joypad {
	return &Joypad{0x30, ButtonState{}}
}

func (j *Joypad) UpdateButtons(sys *Sys, state ButtonState) {
	initialVal := j.value()
	j.state = state
	newVal := j.value()
	// Me trying to be too clever, we only raise in interrupt if one of the
	// output bits goes from high to low. The xor will get us a 1
	// for every bit that changed while the and makes sure only the
	// ones that started high make it through.
	//
	// This could probably be reduced to a simpler form, but eh.
	if (initialVal^newVal)&initialVal != 0 {
		sys.RaiseInterrupt(JoypadInterrupt)
	}
}

func (j *Joypad) value() uint8 {
	// Set initial input select values (only bits 4,5 should be set/reset)
	v := j.val
	if v&^0x30 != 0 {
		panic("non-output-setlect values in joypad set")
	}
	// Now set the high two bits since they're unused
	v |= 0xc0
	// And finally set the default values of the outputs to high,
	// they will be pulled low if the corresponding button is set.
	v |= 0x0f
	if ^v&0x10 != 0 {
		// Down/Up/Left/Right
		if j.state.Down {
			v &= ^uint8(0x08)
		}
		if j.state.Up {
			v &= ^uint8(0x04)
		}
		if j.state.Left {
			v &= ^uint8(0x02)
		}
		if j.state.Right {
			v &= ^uint8(0x01)
		}
	}
	// These aren't mutually exclusive, if both the input bits are low
	// the output signifies that one or both of the buttons is pressed.
	if ^v&0x20 != 0 {
		// Start/Select/B/A
		if j.state.Start {
			v &= ^uint8(0x08)
		}
		if j.state.SelectButton {
			v &= ^uint8(0x04)
		}
		if j.state.B {
			v &= ^uint8(0x02)
		}
		if j.state.A {
			v &= ^uint8(0x01)
		}
	}
	return v
}

func (j *Joypad) R(_ uint16) uint8 {
	return j.value()
}

func (j *Joypad) W(_ uint16, v uint8) {
	//fmt.Printf("joypad W=%02Xh\n", v)
	j.val = v & 0x30
}

func (j *Joypad) Asserts(addr uint16) bool {
	return addr == 0xff00
}
