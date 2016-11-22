package main

import (
	"fmt"
	"strings"
)

func main() {
	for _, p := range ProcessList() {
		if strings.HasPrefix(p.Command, "Spotify") {
			if p.Command == "Spotify" {
				fmt.Printf("** %#v\n", p)
			} else {
				fmt.Printf("%#v\n", p)
			}
		}
	}
}
