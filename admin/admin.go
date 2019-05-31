package admin

import (
	"context"
	"fmt"
	"net/http"
	"runtime"

	"github.com/chinasarft/golive/config"
	log "github.com/chinasarft/golive/mylog"
)

type adminServer struct {
	server    http.Server
	mux       *http.ServeMux
	isStarted bool
}

var server adminServer

func StartAdmin(conf *config.ApiConfig) error {
	if server.isStarted {
		return fmt.Errorf("already started")
	}
	server.isStarted = true
	server.server.Addr = conf.Addr

	server.installHander()

	go server.start()
	return nil
}

func (s *adminServer) start() {
	if err := s.server.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("admin finished")
	}
}

func getNumGoroutine(w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf("{\"goroutine\":%d}", runtime.NumGoroutine())
	w.Write([]byte(str))
}

func (s *adminServer) installHander() {

	mux := http.NewServeMux()
	server.mux = mux
	s.server.Handler = mux

	mux.Handle("/admin/count/goroutine", http.HandlerFunc(getNumGoroutine))
}

func StopAdmin() error {

	if server.isStarted {
		ctx, _ := context.WithCancel(context.Background())
		if err := server.server.Shutdown(ctx); err != nil {
			return err
		}
		server.isStarted = false
	}
	return nil
}
