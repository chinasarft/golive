package rtmp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/amf"
)

type RtmpReceiver struct {
	*RtmpHandler
	rw io.ReadWriter
}

func NewRtmpReceiver(rw io.ReadWriter) *RtmpReceiver {
	recv := &RtmpReceiver{}
	recv.RtmpHandler = NewRtmpHandler(rw, recv)
	recv.rw = rw
	return recv
}

func (recv *RtmpReceiver) OnError(w io.Writer) {

}

func (recv *RtmpReceiver) OnProtocolControlMessaage(m *ProtocolControlMessaage) error {
	switch m.MessageType {
	case 1:
		recv.chunkStreamSet.SetChunkSize(1024) // TODO 先设置成1024
	case 2:
	case 3:
	case 5:
	case 6:
	}
	return nil
}

func (recv *RtmpReceiver) OnUserControlMessage(m *UserControlMessage) error {
	return nil
}

func (recv *RtmpReceiver) OnCommandMessage(m *CommandMessage) error {
	switch m.MessageType {
	case 17: //AMF3
	case 20: //AMF0
		r := bytes.NewReader(m.Payload)
		v, e := amf.ReadValue(r)
		if e == nil {
			switch v.(type) {
			case string:
				value := v.(string)
				switch value {
				case "connect":
					fmt.Println("receive connect command")
					return recv.handleConnectCommand(r)
				case "releaseStream":
					fmt.Println("receive releaseStream command")
				case "FCPublish":
					fmt.Println("receive FCPublish command")
				case "createStream":
					fmt.Println("receive createStream command")
					return recv.handleCreateStreamCommand(r)
				case "publish":
					fmt.Println("receive publish command")
				}
			}
		} else {
			return e
		}
	}
	return nil
}

func (recv *RtmpReceiver) OnDataMessage(m *DataMessage) error {
	switch m.MessageType {
	case 15: //AFM3
	case 18: //AFM0
	}
	return nil
}

func (recv *RtmpReceiver) OnVideoMessage(m *VideoMessage) error {
	fmt.Println("receive video:", m.PayloadLength, len(m.Payload))
	return nil
}

func (recv *RtmpReceiver) OnAudioMessage(m *AudioMessage) error {
	fmt.Println("receive audio:", m.PayloadLength, len(m.Payload))
	return nil
}

func (recv *RtmpReceiver) OnSharedObjectMessage(m *SharedObjectMessage) error {
	switch m.MessageType {
	case 16: //AFM3
	case 19: //AFM0
	}
	return nil
}

func (recv *RtmpReceiver) OnAggregateMessage(m *AggregateMessage) error {
	return nil
}

func (recv *RtmpReceiver) handleConnectCommand(r amf.Reader) error {
	for {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			fmt.Println("handleconnect--->", e)
			return e
		}
		fmt.Println("handleconnect value:", v)
	}

	w := &bytes.Buffer{}

	ackMsg := NewAckMessage(2500000)
	chunkArray, err := recv.sendMessageStreamSet.MessageToChunk(ackMsg, recv.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = recv.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	recv.sendMessageStreamSet.upadateLastStreamInfo(ackMsg, chunkArray[0].chunkStreamID)

	setPeerBandwidthMsg := NewSetPeerBandwidthMessage(2500000, 2)
	chunkArray, err = recv.sendMessageStreamSet.MessageToChunk(setPeerBandwidthMsg, recv.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = recv.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	recv.sendMessageStreamSet.upadateLastStreamInfo(setPeerBandwidthMsg, chunkArray[0].chunkStreamID)

	bakChunkSize := recv.chunkSerializer.GetChunkSize()
	defer func() {
		if err != nil {
			recv.chunkSerializer.SetChunkSize(bakChunkSize)
		}
	}()

	recv.chunkSerializer.SetChunkSize(1024)

	setChunkMsg := NewSetChunkSizeMessage(recv.chunkSerializer.GetChunkSize())
	chunkArray, err = recv.sendMessageStreamSet.MessageToChunk(setChunkMsg, recv.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = recv.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	recv.sendMessageStreamSet.upadateLastStreamInfo(setChunkMsg, chunkArray[0].chunkStreamID)

	connectOkMsg, err := NewConnectSuccessMessage()
	if err != nil {
		return err
	}
	chunkArray, err = recv.sendMessageStreamSet.MessageToChunk(connectOkMsg, recv.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = recv.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	recv.sendMessageStreamSet.upadateLastStreamInfo(connectOkMsg, chunkArray[0].chunkStreamID)

	_, err = recv.rw.Write(w.Bytes())

	return err
}
func (recv *RtmpReceiver) handleCreateStreamCommand(r amf.Reader) error {
	for {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			fmt.Println("create stream--->", e)
			return e
		}
		fmt.Println("createstream value:", v)
	}
	return nil
}
