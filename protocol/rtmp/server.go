package rtmp

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
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
}

func ListenAndServe(addr string, handler Handler) error {
	server := Server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}

func (s *Server) ServeRTMP(conn net.Conn) {

}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
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
		rw, e := l.Accept()
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
				log.Printf("rtmp: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c := srv.newConn(rw)
		log.Println("accept a rtmp connection")
		go c.serve()
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

func (srv *Server) newConn(netConn net.Conn) *conn {
	c := &conn{
		netConn: netConn,
		server:  srv,

		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		chunks:              make(map[uint32]*ChunkStream),
	}

	return c
}
