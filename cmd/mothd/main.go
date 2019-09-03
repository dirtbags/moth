package main

import (
	"time"
	"log"
)

func main() {
	log.Print("Started")
	
	theme := NewTheme("../../theme")
	state := NewState("../../state")
	puzzles := NewMothballs("../../mothballs")
	
	
	interval := 2 * time.Second
	go theme.Run(interval)
	go state.Run(interval)
	go puzzles.Run(interval)
	
	time.Sleep(1 * time.Second)
	log.Print(state.Export(""))
	time.Sleep(19 * time.Second)
}