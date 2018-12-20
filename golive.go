package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/chinasarft/golive/protocol/rtmp"
)

func printNumGoroutine() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Goroutine num: ", runtime.NumGoroutine())
		}
	}
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8808", nil))
	}()

	go printNumGoroutine()

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("rtmp server starting...")
	err := rtmp.ListenAndServe("", nil)
	if err != nil {
		log.Println("fail to start rtmp:", err)
	}
}
