package gb

type Sys struct {
	rom       Rom
	systemRam SystemRam
	hiRam     Ram
	devs      []BusDev
}

type BusDev interface {
	Rb(addr uint16) uint8
	Wb(addr uint16, val uint8)
	Rs(addr uint16) uint16
	Ws(addr uint16, val uint16)
	Asserts(addr uint16) bool
}

func NewSys(rom Rom) *Sys {
	systemRam := NewSystemRam()
	hiRam := NewHiRam()
	devs := []BusDev{&rom, systemRam, hiRam}
	return &Sys{rom, *systemRam, *hiRam, devs}
}

func (s *Sys) IER() uint8 {
	return s.Rb(0xffff)
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
