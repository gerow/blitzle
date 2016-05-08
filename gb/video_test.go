package gb

import "testing"

type NullVideoSwapper struct {
}

func (n *NullVideoSwapper) VideoSwap(_ [LCDSizeX * LCDSizeY]Pixel) {
}

func TestCorrectVideoThingAsserts(t *testing.T) {
	v := NewVideo(&NullVideoSwapper{})

	for addr := uint16(0xfe00); addr < 0xfea0; addr++ {
		if v.getHandler(addr) != v.oam {
			t.Errorf("expected %04Xh to be handled by oam\n", addr)
		}
	}
	for addr := uint16(0x8000); addr < 0xa000; addr++ {
		if v.getHandler(addr) != v.videoRAM {
			t.Errorf("expected %04Xh to be handled by videoRAM\n", addr)
		}
	}
	for addr := uint16(0xfe00); addr < 0xfea0; addr++ {
		if v.getHandler(addr) != v.oam {
			t.Errorf("expected %04Xh to be handled by oam\n", addr)
		}
	}

	if v.getHandler(0xff41) != v.stat {
		t.Errorf("expected %04Xh to be handled by stat\n", 0xff41)
	}
	if v.getHandler(0xff42) != v.scy {
		t.Errorf("expected %04Xh to be handled by scy\n", 0xff42)
	}
	if v.getHandler(0xff43) != v.scx {
		t.Errorf("expected %04Xh to be handled by scx\n", 0xff43)
	}
	if v.getHandler(0xff44) != v.ly {
		t.Errorf("expected %04Xh to be handled by ly\n", 0xff44)
	}
	if v.getHandler(0xff45) != v.lyc {
		t.Errorf("expected %04Xh to be handled by lyc\n", 0xff45)
	}
	if v.getHandler(0xff40) != v.lcdc {
		t.Errorf("expected %04Xh to be handled by lcdc\n", 0xff40)
	}
	if v.getHandler(0xff4a) != v.wy {
		t.Errorf("expected %04Xh to be handled by wy\n", 0xff4a)
	}
	if v.getHandler(0xff4b) != v.wx {
		t.Errorf("expected %04Xh to be handled by wx\n", 0xff4b)
	}
	if v.getHandler(0xff47) != v.bgp {
		t.Errorf("expected %04Xh to be handled by bgp\n", 0xff47)
	}
	if v.getHandler(0xff48) != v.obp0 {
		t.Errorf("expected %04Xh to be handled by obp0\n", 0xff48)
	}
	if v.getHandler(0xff49) != v.obp1 {
		t.Errorf("expected %04Xh to be handled by obp1\n", 0xff49)
	}
	if v.getHandler(0xff46) != v.dma {
		t.Errorf("expected %04Xh to be handled by dma\n", 0xff46)
	}
}
