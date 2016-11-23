package main

import (
	"log"
	"strings"
)

func main() {
	top := NewTop(5)
	log.Println("Waiting...")
	for {
		select {
		case <-top.NextTick:
			for _, p := range top.ProcessList() {
				if strings.HasPrefix(p.Command, "Spotify") {
					if p.Command == "Spotify" {
						log.Printf("** %#v\n", p)
					} else {
						log.Printf("%#v\n", p)
					}
				}
			}
			log.Println("---")
		}
	}
}
