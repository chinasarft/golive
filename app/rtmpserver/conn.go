package rtmpserver

import (
	"bufio"
	"io"
	"log"
	"net"
	"time"

	"github.com/chinasarft/golive/exchange"
	"github.com/chinasarft/golive/protocol/flvlive"
	"github.com/chinasarft/golive/protocol/rtmp"
)

const (
	_ = iota
	typeSetChunkSize
	typeAbortMessage
	typeAck
	typeUserControlMessages
	typeWindowAckSize
	typeSetPeerBandwidth
	typeCommandMessageAMF0 = 17
	typeCommandMessageAMF3 = 20
)

var (
	readTimeout  time.Duration = 0
	writeTimeout time.Duration = 0
)

type NetConnWrapper struct {
	net.Conn
	bufRW *bufio.ReadWriter
}

type conn struct {
	*NetConnWrapper
	handler ServeStart
}

type ServeStart interface {
	Start() error
}

func NewNetConnWrapper(c net.Conn, bufSize int) *NetConnWrapper {
	return &NetConnWrapper{
		Conn:  c,
		bufRW: bufio.NewReadWriter(bufio.NewReaderSize(c, bufSize), bufio.NewWriterSize(c, bufSize)),
	}
}

func (nc *NetConnWrapper) Read(p []byte) (n int, err error) {
	timeout := 5 * time.Second
	if readTimeout != 0 {
		timeout = readTimeout
	}

	t := time.Now()
	nc.Conn.SetDeadline(t.Add(timeout))
	n, err = nc.bufRW.Read(p)
	return
}

func (nc *NetConnWrapper) Write(p []byte) (n int, err error) {
	timeout := 5 * time.Second
	if writeTimeout != 0 {
		timeout = writeTimeout
	}

	nc.Conn.SetDeadline(time.Now().Add(timeout))
	return nc.Conn.Write(p)
	//return nc.bufRW.Write(p) //不能用bufio，会缓冲下来
}

func NewConn(netConn net.Conn, bufSize int) (*conn, error) {
	cw := NewNetConnWrapper(netConn, bufSize)
	c := &conn{
		NetConnWrapper: cw,
	}

	var flag [1]byte
	if _, err := io.ReadFull(netConn, flag[0:1]); err != nil {
		return nil, err
	}

	if flag[0] == 0x66 {
		c.handler = flvlive.NewFlvLiveHandler(cw, exchange.GetExchanger())
	} else {
		c.handler = rtmp.NewRtmpHandler(cw, exchange.GetExchanger(), flag[0])
	}

	return c, nil
}

func (c *conn) serve() {
	err := c.handler.Start()
	if err != nil {
		c.NetConnWrapper.Close()
		log.Println("start return:", err)
	}
}
