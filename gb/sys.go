package gb

type Sys struct {
	rom       ROM
	systemRAM SystemRAM
	hiRAM     RAM
	video     Video
	cpu       CPU
	devs      []BusDev
}

type BusDev interface {
	Rb(addr uint16) uint8
	Wb(addr uint16, val uint8)
	Rs(addr uint16) uint16
	Ws(addr uint16, val uint16)
	Asserts(addr uint16) bool
}

func NewSys(rom ROM) *Sys {
	systemRAM := NewSystemRAM()
	hiRAM := NewHiRAM()
	video := NewVideo()
	cpu := NewCPU()
	devs := []BusDev{&rom, systemRAM, hiRAM, video}
	return &Sys{rom, *systemRAM, *hiRAM, *video, *cpu, devs}
}

func (s *Sys) IER() uint8 {
	return s.Rb(0xffff)
}

func (s *Sys) Run() {
	return
}

func (s *Sys) getHandler(addr uint16) BusDev {
	for _, bd := range s.devs {
		if bd.Asserts(addr) {
			return bd
		}
	}
	return nil
}

func (s *Sys) Rb(addr uint16) uint8 {
	return s.getHandler(addr).Rb(addr)
}

func (s *Sys) Wb(addr uint16, val uint8) {
	s.getHandler(addr).Wb(addr, val)
}

func (s *Sys) Rs(addr uint16) uint16 {
	return s.getHandler(addr).Rs(addr)
}

func (s *Sys) Ws(addr uint16, val uint16) {
	s.getHandler(addr).Ws(addr, val)
}
