package main

import (
	"fmt"
	"os"

	"github.com/chinasarft/golive/container/mp4"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage as:", os.Args[0], "filename")
		return
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		b := mp4.NewBox()
		targetBox, _, err := b.Parse(file)
		if err != nil {
			fmt.Println(err)
			break
		}
		mp4.PrintBox(targetBox)
	}
}
