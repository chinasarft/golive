package rtmp

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"
)

var (
	connectMsg = "0300000000008b1400000000020007636f6e6e656374003ff00000000000000300036170700200046c697665000474797065" +
		"02000a6e6f6e707269766174650008666c617368566572020024464d4c452f332e302028636f6d70617469626c653b204c61" +
		"766635372e38332e313030290005746355726c02001a72746d703a2f2f3132372e302e302e313a313933352f6c6976650000" +
		"09"
)

type testRecv struct {
	r         io.Reader
	writeChan chan int
	writeBuf  []byte
}

func newTestRecv(msg []byte) *testRecv {
	return &testRecv{
		r:         bytes.NewReader(msg),
		writeBuf:  make([]byte, 0, 1024*10),
		writeChan: make(chan int),
	}
}

func (hs *testRecv) Read(p []byte) (n int, err error) {
	return hs.r.Read(p)
}

func (hs *testRecv) Write(p []byte) (n int, err error) {
	copy(hs.writeBuf[len(hs.writeBuf):], p)
	return len(p), nil
}

func TestRtmpReceiver(t *testing.T) {
	msg := testc0c1c2 + connectMsg
	msgByte := make([]byte, len(msg)/2)
	_, err := hex.Decode(msgByte, []byte(msg))
	if err != nil {
		t.Errorf("hex decode msg fail:%s", err)
	}

	rw := newTestRecv(msgByte)

	recv := NewRtmpReceiver(rw)

	err = recv.Start()
	if err != nil {
		t.Fatal("recv.Start", err)
	}
}
