package rtmp

import (
	"log"
	"net"
	"time"
)

type conn struct {
	netConn net.Conn
	server  *Server

	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	chunks              map[uint32]*ChunkStream
}

func (c *conn) serve() {
	err := handshake(c)
	if err != nil {
		c.netConn.Close()
		log.Println("rtmp HandshakeServer err:", err)
	}
	err = c.DealMessage()
}

func (c *conn) Read(p []byte) (n int, err error) {
	timeout := 5 * time.Second
	if c.server.ReadTimeout != 0 {
		timeout = c.server.ReadTimeout
	}

	c.netConn.SetDeadline(time.Now().Add(timeout))
	return c.netConn.Read(p)
}

func (c *conn) Write(p []byte) (n int, err error) {
	timeout := 5 * time.Second
	if c.server.WriteTimeout != 0 {
		timeout = c.server.WriteTimeout
	}

	c.netConn.SetDeadline(time.Now().Add(timeout))
	return c.netConn.Write(p)
}

/**
 * connect publish等消息
 **/
func (c *conn) DealMessage() error {
	cs, err := c.ReadChunk()
	if err != nil {
		return err
	}

	switch cs.TypeID {
	case 20, 17:
	case 1:
	default:
		panic("unkown message type id")
	}

	return nil
}

func (c *conn) ReadChunk() (cs *ChunkStream, err error) {

	var ok bool
	var fmt, csid uint32
	for {
		fmt, csid, err = getChunkBasicHeader(c)
		if err != nil {
			return
		}

		cs, ok = c.chunks[csid]
		if !ok {
			cs = &ChunkStream{}
			c.chunks[csid] = cs
			cs.CSID = csid
		}
		cs.tmpFromat = fmt

		err = cs.readChunkWithoutBasicHeader(c, c.remoteChunkSize)
		if err != nil {
			return
		}

		if cs.IsGetFullMessage() {
			break
		}
	}

	return
}
