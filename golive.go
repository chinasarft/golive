package main

import (
	"log"

	"github.com/chinasarft/golive/protocol/rtmp"
)

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("rtmp server starting...")
	err := rtmp.ListenAndServe("", nil)
	if err != nil {
		log.Println("fail to start rtmp:", err)
	}
}
