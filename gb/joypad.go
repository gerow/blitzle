package gb

import (
	"fmt"
)

type ButtonState struct {
	down         bool
	up           bool
	left         bool
	right        bool
	start        bool
	selectButton bool
	b            bool
	a            bool
}

type Joypad struct {
	val   uint8
	state ButtonState
}

func (j *Joypad) UpdateButtons(sys *Sys, state ButtonState) {
	if j.state != state {
		sys.RaiseInterrupt(JoypadInterrupt)
	}
	j.state = state
}

func (j *Joypad) R(_ uint16) uint8 {
	v := j.val | 0xc0

	if v&0x30 != 0 {
		fmt.Printf("!!! both directions and buttons selected\n")
	}

	if v&0x10 != 0 {
		// Down/Up/Left/Right
		if j.state.down {
			v |= 0x80
		}
		if j.state.up {
			v |= 0x40
		}
		if j.state.left {
			v |= 0x20
		}
		if j.state.right {
			v |= 0x10
		}
	} else if v&0x20 != 0 {
		// Start/Select/B/A
		if j.state.start {
			v |= 0x80
		}
		if j.state.selectButton {
			v |= 0x40
		}
		if j.state.b {
			v |= 0x20
		}
		if j.state.a {
			v |= 0x10
		}
	} else {
		fmt.Printf("!!! neither buttons nor directions selected\n")
		return v | 0x0f
	}
	return v
}

func (j *Joypad) W(_ uint16, v uint8) {
	j.val = v & 0x03
}

func (j *Joypad) Asserts(addr uint16) bool {
	return addr == 0xff00
}
