package rtmpserver

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/chinasarft/golive/config"
	log "github.com/chinasarft/golive/mylog"
)

var ErrServerClosed = errors.New("rtmp: Server closed")

type Handler interface {
	ServeRTMP(net.Conn)
}

//TODO
//这边的结构有点仿造net/http中的Server结构
//所以很多成员具体的用法还带分析
type Server struct {
	Addr         string
	Handler      Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	mu           sync.Mutex
	listeners    map[net.Listener]struct{}
	activeConn   map[*conn]struct{}
	doneChan     chan struct{}
	onShutdown   []func()
	pubConfig    config.PublishConfig
}

var defaultServer Server

func ListenAndServe(addr string, handler Handler) error {
	if addr != "" {
		defaultServer.Addr = addr
	}
	defaultServer.Handler = handler
	return defaultServer.ListenAndServe()
}

func ListenAndServeTls(addr string, handler Handler) error {
	server := Server{Addr: addr, Handler: handler}

	cert, err := tls.LoadX509KeyPair("./cert.pem", "./prikey.pem")
	if err != nil {
		return err
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", ":1953", cfg)
	if err != nil {
		return err
	}

	return server.Serve(listener)
}

func (s *Server) ServeRTMP(conn net.Conn) {

}

func (s *Server) SetPublishConfig(conf *config.PublishConfig) {
	s.pubConfig = *conf
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = srv.pubConfig.RtmpFlv.Addr
	}
	if addr == "" {
		addr = ":1935"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return srv.Serve(ln)
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		netconn, e := l.Accept()
		if e != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Error().Err(e).Str("module", "rtmp").Dur("wait", tempDelay).Msg("accept error")
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c, err := NewConn(netconn, 4*1024)
		if err != nil {
			log.Error().Err(err).Msg("")
			c.Close()
		} else {
			log.Debug().Msg("accept a rtmp connection")
			go c.serve()
		}
	}

}

func (s *Server) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getDoneChanLocked()
}

func (s *Server) getDoneChanLocked() chan struct{} {
	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}
