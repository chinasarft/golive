package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/chinasarft/golive/protocol/rtmp"
)

type ConnClient struct {
	rtmpUrl     string
	conn        net.Conn
	rtmpHandler *rtmp.RtmpClientHandler
	ctx         context.Context
	cancel      context.CancelFunc
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
	conn, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("connection:", "local:", conn.LocalAddr(), "remote:", conn.RemoteAddr())

	connClient.conn = conn
	rtmpHandler, err := rtmp.NewRtmpClientHandler(conn, rtmpUrl, rtmp.ROLE_PUBLISH, nil)
	if err != nil {
		conn.Close()
		return err
	}
	connClient.rtmpHandler = rtmpHandler

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

func (connClient *ConnClient) readAndSend(ctx context.Context) {
	var err error

	defer func() {
		if err != nil {
			log.Println(err)
			connClient.cancel()
			return
		}
	}()

	var pVideoData []byte
	var videoBuf []byte
	if audioBuf, err = ioutil.ReadFile("a.aac"); err != nil {
		return
	}

	if pVideoData, err = ioutil.ReadFile("a.h264"); err != nil {
		return
	}

	ctx, _ = context.WithCancel(ctx)
	bAudioOk := true
	bVideoOk := true

	nSysTimeBase := time.Now().UnixNano() / 1e6
	nNextAudioTime := nSysTimeBase
	nNextVideoTime := nSysTimeBase
	nNow := nSysTimeBase

	for {
		select {
		case <-ctx.Done():
		default:
		}
	}

	return
}

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	player := &ConnClient{}
	log.Println(player.Start("rtmp://127.0.0.1/live/t1"))

	go readAndSend(player.ctx)
	select {
	case <-player.ctx.Done():
	}
}
