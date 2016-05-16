package gb

import "testing"

func TestButtons(t *testing.T) {
	s := S([]byte{
		0xf0, 0x00, // LDH A,($00)
		0x3e, 0x20, // LD A,$20
		0xe0, 0x00, // LDH ($00),A
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0x3e, 0x10, // LD A,$10
		0xe0, 0x00, // LDH ($00),A
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
		0xf0, 0x00, // LDH A,($00)
	})
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x102)
	// Expecting everything to start high
	if s.cpu.a != 0xff {
		t.Errorf("Expected A=FFh, got %02Xh\n", s.cpu.a)
	}

	// Now load $20 into A  and push it to ($FF00) to select
	// Right/Left/Up/Down
	checkStep(t, s, 8) // LD A,$20
	checkIP(t, s, 0x104)
	if s.cpu.a != 0x20 {
		t.Errorf("Expected A=20h, got %02Xh\n", s.cpu.a)
	}
	checkStep(t, s, 12) // LDH ($00),A
	checkIP(t, s, 0x106)
	// Press all the buttons in the other side to make sure they don't
	// leak through.
	state := ButtonState{}
	state.A = true
	state.B = true
	state.SelectButton = true
	state.Start = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x108)
	// Buttons should still be high since we haven't pressed one.
	if s.cpu.a != 0xef {
		t.Errorf("Expected A=EFh, got %02Xh\n", s.cpu.a)
	}

	// Press down
	state = ButtonState{}
	state.Down = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x10a)
	// Bit three should be low
	if s.cpu.a != 0xe7 {
		t.Errorf("Expected A=E7h, got %02Xh\n", s.cpu.a)
	}

	// Press up
	state = ButtonState{}
	state.Up = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x10c)
	// Bit two should be low
	if s.cpu.a != 0xeb {
		t.Errorf("Expected A=EBh, got %02Xh\n", s.cpu.a)
	}

	// Press left
	state = ButtonState{}
	state.Left = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x10e)
	// Bit one should be low
	if s.cpu.a != 0xed {
		t.Errorf("Expected A=EDh, got %02Xh\n", s.cpu.a)
	}

	// Press right
	state = ButtonState{}
	state.Right = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x110)
	// Bit zero should be low
	if s.cpu.a != 0xee {
		t.Errorf("Expected A=EEh, got %02Xh\n", s.cpu.a)
	}

	// Press up and right, just for giggles
	state = ButtonState{}
	state.Up = true
	state.Right = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x112)
	// Bit zero and two should be low
	if s.cpu.a != 0xea {
		t.Errorf("Expected A=EAh, got %02Xh\n", s.cpu.a)
	}

	// And now for the other bank
	// Load $10 into A  and push it to ($FF00) to select
	// A/B/Select/Start
	checkStep(t, s, 8) // LD A,$10
	checkIP(t, s, 0x114)
	if s.cpu.a != 0x10 {
		t.Errorf("Expected A=10h, got %02Xh\n", s.cpu.a)
	}
	checkStep(t, s, 12) // LDH ($00),A
	checkIP(t, s, 0x116)
	// Press all the buttons in the other side to make sure they don't
	// leak through.
	state = ButtonState{}
	state.Right = true
	state.Left = true
	state.Up = true
	state.Down = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x118)
	// Buttons should still be high since we haven't pressed one.
	if s.cpu.a != 0xdf {
		t.Errorf("Expected A=DFh, got %02Xh\n", s.cpu.a)
	}

	// Press start
	state = ButtonState{}
	state.Start = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x11a)
	// Bit three should be low
	if s.cpu.a != 0xd7 {
		t.Errorf("Expected A=D7h, got %02Xh\n", s.cpu.a)
	}

	// Press select
	state = ButtonState{}
	state.SelectButton = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x11c)
	// Bit two should be low
	if s.cpu.a != 0xdb {
		t.Errorf("Expected A=DBh, got %02Xh\n", s.cpu.a)
	}

	// Press B
	state = ButtonState{}
	state.B = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x11e)
	// Bit one should be low
	if s.cpu.a != 0xdd {
		t.Errorf("Expected A=DDh, got %02Xh\n", s.cpu.a)
	}

	// Press A
	state = ButtonState{}
	state.A = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x120)
	// Bit zero should be low
	if s.cpu.a != 0xde {
		t.Errorf("Expected A=DEh, got %02Xh\n", s.cpu.a)
	}

	// Press A and B, just for giggles
	state = ButtonState{}
	state.A = true
	state.B = true
	s.UpdateButtons(state)
	checkStep(t, s, 12) // LDH A,($00)
	checkIP(t, s, 0x122)
	// Bit zero and two should be low
	if s.cpu.a != 0xdc {
		t.Errorf("Expected A=DAh, got %02Xh\n", s.cpu.a)
	}
}
