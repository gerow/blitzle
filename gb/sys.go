package gb

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
)

// Main clock frequency (in Hz)
const clkFreq uint = 4194304

type Sys struct {
	rom       *ROM
	systemRAM *SystemRAM
	hiRAM     *RAM
	video     *Video
	cpu       *CPU
	/* Interrupt controller registers */
	ieReg  *MemRegister
	ifReg  *MemRegister
	timer  *Timer
	joypad *Joypad
	serial *Serial

	devs []BusDev
	Stop bool

	Wall    int
	cpuWait int

	Debug bool
}

type BusDev interface {
	R(addr uint16) uint8
	W(addr uint16, val uint8)
	Asserts(addr uint16) bool
}

type BusHole struct {
	startAddr uint16
	endAddr   uint16
}

func NewBusHole(startAddr uint16, endAddr uint16) *BusHole {
	return &BusHole{startAddr, endAddr}
}

func (b *BusHole) R(addr uint16) uint8 {
	return 0xff
}

func (b *BusHole) W(addr uint16, val uint8) {
}

func (b *BusHole) Asserts(addr uint16) bool {
	return addr >= b.startAddr && addr <= b.endAddr
}

func NewSys(rom *ROM) *Sys {
	systemRAM := NewSystemRAM()
	hiRAM := NewHiRAM()
	video := NewVideo()
	cpu := NewCPU()
	ieReg := NewMemRegister(0xffff)
	ifReg := NewMemRegister(0xff0f)
	timer := NewTimer()
	joypad := NewJoypad()
	serial := NewSerial()
	bh2 := NewBusHole(0xfea0, 0xff7f)
	devs := []BusDev{
		rom,
		systemRAM,
		hiRAM,
		video,
		ieReg,
		ifReg,
		timer,
		joypad,
		serial,
		bh2}

	s := &Sys{
		rom,
		systemRAM,
		hiRAM,
		video,
		cpu,
		ieReg,
		ifReg,
		timer,
		joypad,
		serial,
		devs,
		false,
		0,
		0,
		false}
	s.SetPostBootloaderState()

	return s
}

func (s *Sys) IER() uint8 {
	return s.Rb(0xffff)
}

func (s *Sys) Run() {
	for {
		s.Step()
	}
}

// Step four clock cycls.
func (s *Sys) Step() {
	// XXX(gerow): sys should NOT need to know ANYTHING about sdl
	if s.Wall%(vblankCycles*144) == 0 {
		sdl.PumpEvents()
	}
	s.timer.Step(s)
	s.serial.Step(s)
	s.video.Step(s)
	if s.cpuWait == 0 {
		s.cpuWait = s.cpu.Step(s)
		if s.Debug {
			fmt.Print(s.cpu.State(s))
		}
		//fmt.Printf(s.timer.State())
		//fmt.Print(s.video.State(s))
	} else {
		s.cpuWait -= 4
	}
	s.Wall += 4
}

/*
 * This only really works for values that divide evenly with the main clock,
 * but luckily those are all the values we need!
 */
func (s *Sys) FreqStep(desiredFreq uint) bool {
	divAmt := clkFreq / desiredFreq
	return uint(s.Wall)%divAmt == 0
}

func (s *Sys) getHandler(addr uint16) BusDev {
	for _, bd := range s.devs {
		if bd.Asserts(addr) {
			return bd
		}
	}
	log.Fatalf("Couldn't find handler for addr %04Xh\n", addr)
	return nil
}

func (s *Sys) RbLog(addr uint16, l bool) uint8 {
	rv := s.getHandler(addr).R(addr)
	if l && s.Debug {
		log.Printf("R1 (%04Xh) => %02Xh\n", addr, rv)
	}
	return rv
}

func (s *Sys) WbLog(addr uint16, val uint8, l bool) {
	if l && s.Debug {
		log.Printf("W1 %02Xh => (%04Xh)\n", val, addr)
	}
	s.getHandler(addr).W(addr, val)
}

func (s *Sys) RsLog(addr uint16, l bool) uint16 {
	rv := uint16(s.RbLog(addr, false)) | uint16(s.RbLog(addr+1, false))<<8
	if l && s.Debug {
		log.Printf("R2 (%04Xh) => %04Xh\n", addr, rv)
	}
	return rv
}

func (s *Sys) WsLog(addr uint16, val uint16, l bool) {
	if l && s.Debug {
		log.Printf("W2 %04Xh => (%04Xh)\n", val, addr)
	}
	s.WbLog(addr, uint8(val), false)
	s.WbLog(addr+1, uint8(val>>8), false)
}

func (s *Sys) Rb(addr uint16) uint8 {
	return s.RbLog(addr, true)
}

func (s *Sys) Wb(addr uint16, val uint8) {
	s.WbLog(addr, val, true)
}

func (s *Sys) Rs(addr uint16) uint16 {
	return s.RsLog(addr, true)
}

func (s *Sys) Ws(addr uint16, val uint16) {
	s.WsLog(addr, val, true)
}

func (s *Sys) SetPostBootloaderState() {
	s.cpu.SetPostBootloaderState(s)
}

func (s *Sys) WriteBytes(bytes []byte, addr uint16) {
	for _, b := range bytes {
		s.WbLog(addr, b, false)
		addr++
	}
}

func (s *Sys) ReadBytes(addr uint16, len uint16) []byte {
	o := make([]byte, len)
	for i := uint16(0); i < len; i++ {
		o[i] = s.RbLog(addr+i, false)
	}
	return o
}

type Interrupt uint

const (
	VBlankInterrupt = iota
	LCDStatInterrupt
	TimerInterrupt
	SerialInterrupt
	JoypadInterrupt
	nInterrupts
)

var interruptNameMap map[Interrupt]string = map[Interrupt]string{
	VBlankInterrupt:  "VBlank",
	LCDStatInterrupt: "LCD stat",
	TimerInterrupt:   "Timer",
	SerialInterrupt:  "Serial",
	JoypadInterrupt:  "Joypad",
}

func interruptName(intr Interrupt) string {
	return interruptNameMap[intr]
}

func (s *Sys) RaiseInterrupt(inter Interrupt) {
	//fmt.Printf("Interrupt %s raised!\n", interruptName(inter))
	s.ifReg.set(s.ifReg.val() | (1 << inter))
}

func (s *Sys) HandleInterrupt() *Interrupt {
	// Mask the current interrupts with the entabled mask
	firingInterrupts := s.ifReg.val() & s.ieReg.val()
	for i := Interrupt(0); i < nInterrupts; i++ {
		// The interrupts are in priority order, so just pick the first one
		if firingInterrupts&(1<<i) != 0 {
			// Reset the bit and return the val
			//fmt.Printf("Interrupt %s getting handled!\n", interruptName(i))
			//fmt.Printf("ifreg was %02Xh\n", s.ifReg.val())
			s.ifReg.set(s.ifReg.val() & ^(1 << i))
			//fmt.Printf("now is %02Xh\n", s.ifReg.val())
			return &i
		}
	}

	// If there aren't any interrupts to handle just return nil
	return nil
}

func (s *Sys) SetVideoSwapper(videoSwapper VideoSwapper) {
	s.video.swapper = videoSwapper
}

func (s *Sys) SetSerialSwapper(serialSwapper SerialSwapper) {
	s.serial.swapper = serialSwapper
}

type UpdateButtonser interface {
	UpdateButtons(state ButtonState)
}

func (s *Sys) UpdateButtons(state ButtonState) {
	s.joypad.UpdateButtons(s, state)
}
