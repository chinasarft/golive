package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/chinasarft/golive/admin"
	"github.com/chinasarft/golive/app/rtmpserver"
	"github.com/chinasarft/golive/config"
	log "github.com/chinasarft/golive/mylog"
	sig "github.com/chinasarft/golive/signaling"
)

func printNumGoroutine() {
	ticker := time.NewTicker(time.Second * 8)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Debug().Int("NumGoroutine", runtime.NumGoroutine()).Msg("")
		}
	}
}

func startRTMP() {
	err := rtmpserver.ListenAndServe("", nil)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("fail to start rtmp")
	}
}

func startRTMPS() {
	err := rtmpserver.ListenAndServeTls("", nil)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("fail to start rtmps")
	}
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("usage as:%s file.conf\n", os.Args[0])
		os.Exit(1)
	}

	conf, err := config.LoadConfig(os.Args[1])
	if err != nil {
		fmt.Printf("fail to load config file:%s(%s)\n", os.Args[1], err.Error())
		os.Exit(2)
	}

	if err = log.UpdateConfig(&conf.Log); err != nil {
		fmt.Println("fail to set log config:", conf.Log)
		os.Exit(3)
	}

	admin.StartAdmin(&conf.Api)
	sig.StartProtooSignaling(&conf.Signaling)

	if conf.Log.Level == "debug" {
		go printNumGoroutine()
	}

	log.Info().Msg("rtmp server starting...")

	go startRTMPS()
	startRTMP()
}
