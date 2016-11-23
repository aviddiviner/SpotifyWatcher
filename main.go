package main

import (
	"fmt"
	"strings"
)

func main() {
	top := NewTop(1)
	for {
		select {
		case <-top.NextTick:
			for _, p := range top.ProcessList() {
				if strings.HasPrefix(p.Command, "Spotify") {
					if p.Command == "Spotify" {
						fmt.Printf("** %#v\n", p)
					} else {
						fmt.Printf("%#v\n", p)
					}
				}
			}
		}
	}
}
