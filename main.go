package main

import (
	"flag"
	"fmt"
	"github.com/gerow/blitzle/gb"
	"log"
	"os"
)

func main() {
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

	r, err := gb.LoadRom(fn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(r.Info())
}
