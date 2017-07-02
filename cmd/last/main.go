package main

import (
	"fmt"
	"log"

	"github.com/chennqqi/last"
)

func main() {
	l, err := last.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(l)
}
