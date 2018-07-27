package main

import (
	"fmt"

	"github.com/chinasarft/golive/protocol/rtmp"
)

func main() {
	fmt.Println("rtmp server starting...")
	err := rtmp.ListenAndServe("", nil)
	if err != nil {
		fmt.Println("fail to start rtmp:", err)
	}
}
