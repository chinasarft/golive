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
				//NetConnection command
				case "connect":
					fmt.Println("receive connect command")
					return recv.handleConnectCommand(r)
				case "createStream":
					fmt.Println("receive createStream command")
					return recv.handleCreateStreamCommand(r)

				//NetStream command
				case "publish":
					fmt.Println("receive publish command")
					return recv.handlePublishCommand(r)
				case "deleteStream":
					fmt.Println("receive deleteStream command")

				// TODO 以下命令文档里都没有找到
				case "releaseStream":
					fmt.Println("receive releaseStream command")
				case "FCPublish":
					fmt.Println("receive FCPublish command")
				case "FCUnpublish":
					fmt.Println("receive FCUnpublish command")
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
		if transactionId, ok := v.(float64); ok {
			fmt.Println("handleconnect transactionId:", int(transactionId)) //7.2.11 always set to 1
		} else {
			fmt.Println("handleconnect value:", v)
		}
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
	transactionId := 0
	for {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			fmt.Println("create stream--->", e)
			return e
		}
		if transId, ok := v.(float64); ok {
			transactionId = int(transId)
			fmt.Println("createstream transactionId:", transactionId)
		} else {
			fmt.Println("createstream value:", v)
		}
	}

	w := &bytes.Buffer{}

	createStreamMsg, err := NewCreateStreamSuccessMessage(transactionId)
	if err != nil {
		return err
	}
	chunkArray, err := recv.sendMessageStreamSet.MessageToChunk(createStreamMsg, recv.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = recv.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	recv.sendMessageStreamSet.upadateLastStreamInfo(createStreamMsg, chunkArray[0].chunkStreamID)

	_, err = recv.rw.Write(w.Bytes())

	return err
}

func (recv *RtmpReceiver) handlePublishCommand(r amf.Reader) error {

	transactionId := 0
	for {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			fmt.Println("publish stream--->", e)
			return e
		}
		if transId, ok := v.(float64); ok {
			transactionId = int(transId)
			fmt.Println("publish transactionId:", transactionId)
		} else {
			fmt.Println("publish value:", v)
		}
	}

	w := &bytes.Buffer{}

	// NetConnection 需要回复transactionid, netstream tid都设置为0 7.2.2
	publishOkMsg, err := NewPublishSuccessMessage()
	if err != nil {
		return err
	}
	chunkArray, err := recv.sendMessageStreamSet.MessageToChunk(publishOkMsg, recv.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = recv.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	recv.sendMessageStreamSet.upadateLastStreamInfo(publishOkMsg, chunkArray[0].chunkStreamID)

	_, err = recv.rw.Write(w.Bytes())

	return err
}
