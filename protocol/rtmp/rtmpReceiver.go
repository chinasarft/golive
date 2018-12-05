package rtmp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/amf"
)

type RtmpReceiver struct {
	*RtmpHandler
}

func NewRtmpReceiver(rw io.ReadWriter) *RtmpReceiver {
	recv := &RtmpReceiver{}
	recv.RtmpHandler = NewRtmpHandler(rw, recv)
	return recv
}

func (recv *RtmpReceiver) OnError(w io.Writer) {

}

func (recv *RtmpReceiver) OnProtocolControlMessaage(m *ProtocolControlMessaage) error {
	return nil
}

func (recv *RtmpReceiver) OnUserControlMessage(m *UserControlMessage) error {
	return nil
}

func (recv *RtmpReceiver) OnCommandMessage(m *CommandMessage) error {
	switch m.MessageType {
	case 8:
	case 9:
	case 17: //AMF3
	case 20: //AMF0
		r := bytes.NewReader(m.Payload)
		v, e := amf.ReadValue(r)
		if e == nil {
			switch v.(type) {
			case string:
				value := v.(string)
				if value == "connect" {
					recv.handleConnectCommand(r)
				}
			}
		} else {
			return e
		}

	case 16:
	case 19:
	case 22:
	}
	return nil
}

func (recv *RtmpReceiver) handleConnectCommand(r amf.Reader) error {
	for {
		v, e := amf.ReadValue(r)
		if e != nil && e != io.EOF {
			fmt.Println("--->", e)
			if e == nil {
				return nil
			}
			return e
		}
		fmt.Println("value:", v)
	}

	//TODO write response

	return nil
}
