package main

import (
	"os"
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	
	for _, dirname := range flag.Args() {
		f, _ := os.Open(dirname)
		defer f.Close()
		names, _ := f.Readdirnames(0)
		fmt.Print(names)
	}
}