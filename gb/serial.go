package gb

import (
	"fmt"
)

const sbAddr uint16 = 0xff01
const scAddr uint16 = 0xff02

type SerialSwapper interface {
	SerialSwap(out uint8) uint8
}

type Serial struct {
	swapper            SerialSwapper
	sb                 uint8
	sc                 uint8
	transferDone       chan bool
	transferInProgress bool
	newSb              uint8
}

func NewSerial() *Serial {
	return &Serial{
		nil,
		0,
		0,
		make(chan bool),
		false,
		0}
}

func (s *Serial) Step(sys *Sys) {
	if s.transferInProgress {
		select {
		case <-s.transferDone:
			s.transferInProgress = false
			s.sb = s.newSb
			sys.RaiseInterrupt(SerialInterrupt)
		default:
		}
	}
}

func (s *Serial) R(addr uint16) uint8 {
	switch addr {
	case sbAddr:
		return s.sb
	case scAddr:
		sc := s.sc
		if s.transferInProgress {
			sc |= 0x80
		}
		return sc
	default:
		panic("read on serial from bad address")
	}
}

func (s *Serial) doSwap(concurrent bool) {
	if !concurrent {
		if s.swapper != nil {
			s.sb = s.swapper.SerialSwap(s.sb)
		}
		return
	}
	s.transferInProgress = true
	oldSb := s.sb
	go func() {
		if s.swapper != nil {
			s.newSb = s.swapper.SerialSwap(oldSb)
		}
		s.transferDone <- true
	}()
}

func (s *Serial) W(addr uint16, val uint8) {
	switch addr {
	case sbAddr:
		s.sb = val
	case scAddr:
		s.sc = val & 0x03
		// Only do the transfer if we're using internal clock
		// XXX(gerow): This is incompatible with communication
		// between gameboys, need to find a better solution.
		if val&0x80 != 0 && s.sc&0x01 != 0 {
			if s.transferInProgress {
				// Just print a message and ignore.
				fmt.Printf("!!! attempt to start new transfer when transfer already in progress\n")
			} else {
				// XXX(gerow): This should probably be done
				// concurrently.
				s.doSwap(false)
			}
		}
		// We just remember the lower two bits.
	default:
		panic("write on serial from bad address")
	}
}

func (s *Serial) Asserts(addr uint16) bool {
	return addr == sbAddr || addr == scAddr
}

// Just acts like nothing is on the other side of the connection
type NullSerialSwapper struct {
}

func (n *NullSerialSwapper) SerialSwap(_ uint8) uint8 {
	return 0xff
}
