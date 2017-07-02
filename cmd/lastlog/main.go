package main

import (
	"fmt"

	"github.com/chennqqi/last"
)

func main() {
	for i := 0; i < 65535; i++ {
		l, err := last.ByUID(i)
		if err == nil {
			fmt.Println(i, l)
		}
	}
}
