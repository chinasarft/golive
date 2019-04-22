package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/chinasarft/golive/app/flvserver"
	"github.com/chinasarft/golive/app/rtmpserver"
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

func startRTMP() {
	err := rtmpserver.ListenAndServe("", nil)
	if err != nil {
		log.Println("fail to start rtmp:", err)
	}
}

func startFlvLive() {
	err := flvserver.ListenAndServe("", nil)
	if err != nil {
		log.Println("fail to start flvlive:", err)
	}
}

func main() {
	//目前这个http服务只是为了观察运行时情况
	// 打算是启动一个内部http端口做一些控制
	go func() {
		log.Println(http.ListenAndServe("localhost:8808", nil))
	}()

	go printNumGoroutine()

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("rtmp server starting...")

	go startRTMP()
	startFlvLive()
}
