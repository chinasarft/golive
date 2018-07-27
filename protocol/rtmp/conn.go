package rtmp

import (
	"net"
	"time"
)

type conn struct {
	netConn net.Conn
	server  *Server
}

func (c *conn) serve() {
	err := handshake(c)
	if err != nil {
		c.netConn.Close()
		log.Println("rtmp HandshakeServer err:", err)
	}

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
