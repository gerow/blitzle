package main

import (
	"flag"
	"fmt"
	"github.com/gerow/blitzle/frontend"
	"github.com/gerow/blitzle/gb"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	//	"net/http"
	//	_ "net/http/pprof"
	"os"
	"runtime"
)

func main() {
	//	go func() {
	//		log.Println(http.ListenAndServe("localhost:6060", nil))
	//	}()
	// XXX(gerow): Hack for issues in go-sdl2
	runtime.LockOSThread()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s ROM_FILE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	fn := flag.Args()[0]

	r, err := gb.LoadROMFromFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(r.Info())
	sdl.Init(sdl.INIT_EVERYTHING)

	fe, err := frontend.NewFrontend()
	if err != nil {
		panic(err)
	}
	serialOut, err := os.Create("serial.log")
	if err != nil {
		panic(err)
	}
	defer serialOut.Close()
	sys := gb.NewSys(r, fe, &frontend.WriterSerialSwapper{serialOut})
	bs := gb.ButtonState{false, false, false, false, true, false, false, true}
	sys.UpdateButtons(bs)
	sys.Run()
}
