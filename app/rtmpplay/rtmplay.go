package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"strings"

	"github.com/chinasarft/golive/protocol/rtmp"
)

type ConnClient struct {
	rtmpUrl     string
	conn        net.Conn
	rtmpHandler *rtmp.RtmpPlayHandler
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
	rtmpHandler, err := rtmp.NewRtmpPlayHandler(conn, rtmpUrl, connClient)
	if err != nil {
		conn.Close()
		return err
	}
	connClient.rtmpHandler = rtmpHandler

	go connClient.rtmpHandler.Start()

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

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	player := &ConnClient{}
	log.Println(player.Start("rtmp://127.0.0.1/live/t1"))
	select {}
}
