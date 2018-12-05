package rtmp

import (
	"bufio"
	"log"
	"net"
	"time"
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
	*RtmpReceiver
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

func NewConn(netConn net.Conn, bufSize int) *conn {
	cw := NewNetConnWrapper(netConn, bufSize)
	c := &conn{
		NetConnWrapper: cw,
		RtmpReceiver:   NewRtmpReceiver(cw),
	}

	return c
}

func (c *conn) serve() {
	err := c.Start()
	if err != nil {
		c.NetConnWrapper.Close()
		log.Println("rtmp HandshakeServer err:", err)
	}

}
