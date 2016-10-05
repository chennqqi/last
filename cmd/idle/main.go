package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dim13/last"
)

func main() {
	uname := flag.String("u", "", "username")
	flag.Parse()
	l, err := last.Username(*uname)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(time.Since(l))
}
