package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/chinasarft/golive/protocol/rtmp"
)

type ConnClient struct {
	rtmpUrl     string
	conn        net.Conn
	rtmpHandler *rtmp.RtmpClientHandler
	ctx         context.Context
	cancel      context.CancelFunc
	useTls      bool
	group       sync.WaitGroup
}

func (connClient *ConnClient) Start(rtmpUrl string) error {

	u, err := url.Parse(rtmpUrl)
	if err != nil {
		return err
	}
	connClient.rtmpUrl = rtmpUrl
	path := strings.TrimLeft(u.Path, "/")
	ps := strings.SplitN(path, "/", 2)
	if len(ps) != 2 {
		return fmt.Errorf("u path err: %s", path)
	}

	connClient.useTls = (strings.Index(rtmpUrl, "rtmps://") == 0)

	port := ":1935"
	host := u.Host
	localIP := ":0"
	var remoteIP string
	if strings.Index(host, ":") != -1 {
		host, port, err = net.SplitHostPort(host)
		if err != nil {
			return err
		}
		port = ":" + port
	}
	ips, err := net.LookupIP(host)
	log.Printf("ips: %v, host: %v", ips, host)
	if err != nil {
		log.Println(err)
		return err
	}
	remoteIP = ips[rand.Intn(len(ips))].String()
	if strings.Index(remoteIP, ":") == -1 {
		remoteIP += port
	}

	local, err := net.ResolveTCPAddr("tcp", localIP)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("remoteIP: ", remoteIP)
	remote, err := net.ResolveTCPAddr("tcp", remoteIP)
	if err != nil {
		log.Println(err)
		return err
	}

	var conn *net.TCPConn
	var tlsConn *tls.Conn
	if connClient.useTls {
		tlsConn, err = tls.Dial("tcp", u.Host, nil)
	} else {
		conn, err = net.DialTCP("tcp", local, remote)
	}
	if err != nil {
		log.Println(err)
		return err
	}

	var rtmpHandler *rtmp.RtmpClientHandler
	if connClient.useTls {
		connClient.conn = tlsConn
		rtmpHandler, err = rtmp.NewRtmpClientHandler(tlsConn, rtmpUrl, rtmp.ROLE_PUBLISH, nil)
		connClient.rtmpHandler = rtmpHandler
		if err != nil {
			tlsConn.Close()
			return err
		}
	} else {
		connClient.conn = conn
		log.Println("connection:", "local:", conn.LocalAddr(), "remote:", conn.RemoteAddr())
		rtmpHandler, err = rtmp.NewRtmpClientHandler(conn, rtmpUrl, rtmp.ROLE_PUBLISH, nil)
		connClient.rtmpHandler = rtmpHandler
		if err != nil {
			conn.Close()
			return err
		}
	}

	connClient.ctx, connClient.cancel = context.WithCancel(context.Background())
	go connClient.rtmpHandler.Start(connClient.ctx)

	return nil
}

func (connClient *ConnClient) OnError(err error) {
	return
}

func (connClient *ConnClient) OnAudioMessage(m *rtmp.AudioMessage) {
	log.Println("receive audio message:", m.Timestamp)
	return
}

func (connClient *ConnClient) OnVideoMessage(m *rtmp.VideoMessage) {
	log.Println("receive video message:", m.Timestamp)
	return
}

func (connClient *ConnClient) OnDataMessage(m *rtmp.DataMessage) {
	log.Println("receive data message:", m.Timestamp)
	return
}

func (connClient *ConnClient) readAndSend(ctx context.Context) error {
	var err error
	var flvBuf []byte
	if flvBuf, err = ioutil.ReadFile("a.flv"); err != nil {
		return err
	}

	go connClient.readAudioTag(flvBuf, ctx)
	go connClient.readVideoTag(flvBuf, ctx)

	return nil
}

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	pub := &ConnClient{
		ctx: context.Background(),
	}
	if len(os.Args) > 1 {
		log.Println(pub.Start(os.Args[1]))
	} else {
		log.Println(pub.Start("rtmp://127.0.0.1:8008/live/t1"))
	}

	connectOk := <-pub.rtmpHandler.ConnectResult()
	if !connectOk {
		log.Println("rtmp connect fail")
		return
	}
	pub.group.Add(2)
	log.Println("rtmp connect success")
	if err := pub.readAndSend(pub.ctx); err != nil {
		log.Println(err)
		return
	}

	pub.group.Wait()
}
