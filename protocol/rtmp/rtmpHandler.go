package rtmp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/amf"
)

type RtmpHandler struct {
	*RtmpUnpacker
	rw io.ReadWriter
}

func NewRtmpHandler(rw io.ReadWriter) *RtmpHandler {
	handler := &RtmpHandler{}
	handler.RtmpUnpacker = NewRtmpUnpacker(rw, handler)
	handler.rw = rw
	return handler
}

func (h *RtmpHandler) OnError(w io.Writer) {

}

func (h *RtmpHandler) OnProtocolControlMessaage(m *ProtocolControlMessaage) error {
	switch m.MessageType {
	case 1:
		h.chunkStreamSet.SetChunkSize(1024) // TODO 先设置成1024
	case 2:
	case 3:
	case 5:
	case 6:
	}
	return nil
}

func (h *RtmpHandler) OnUserControlMessage(m *UserControlMessage) error {
	return nil
}

func (h *RtmpHandler) OnCommandMessage(m *CommandMessage) error {
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
					return h.handleConnectCommand(r)
				case "createStream":
					fmt.Println("receive createStream command")
					return h.handleCreateStreamCommand(r)

				//NetStream command
				case "publish":
					fmt.Println("receive publish command")
					return h.handlePublishCommand(r)
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

func (h *RtmpHandler) OnDataMessage(m *DataMessage) error {
	switch m.MessageType {
	case 15: //AFM3
	case 18: //AFM0
	}
	return nil
}

func (h *RtmpHandler) OnVideoMessage(m *VideoMessage) error {
	fmt.Println("receive video:", m.PayloadLength, len(m.Payload))
	return nil
}

func (h *RtmpHandler) OnAudioMessage(m *AudioMessage) error {
	fmt.Println("receive audio:", m.PayloadLength, len(m.Payload))
	return nil
}

func (h *RtmpHandler) OnSharedObjectMessage(m *SharedObjectMessage) error {
	switch m.MessageType {
	case 16: //AFM3
	case 19: //AFM0
	}
	return nil
}

func (h *RtmpHandler) OnAggregateMessage(m *AggregateMessage) error {
	return nil
}

func (h *RtmpHandler) handleConnectCommand(r amf.Reader) error {
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
	chunkArray, err := h.sendMessageStreamSet.MessageToChunk(ackMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	h.sendMessageStreamSet.upadateLastStreamInfo(ackMsg, chunkArray[0].chunkStreamID)

	setPeerBandwidthMsg := NewSetPeerBandwidthMessage(2500000, 2)
	chunkArray, err = h.sendMessageStreamSet.MessageToChunk(setPeerBandwidthMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	h.sendMessageStreamSet.upadateLastStreamInfo(setPeerBandwidthMsg, chunkArray[0].chunkStreamID)

	bakChunkSize := h.chunkSerializer.GetChunkSize()
	defer func() {
		if err != nil {
			h.chunkSerializer.SetChunkSize(bakChunkSize)
		}
	}()

	h.chunkSerializer.SetChunkSize(1024)

	setChunkMsg := NewSetChunkSizeMessage(h.chunkSerializer.GetChunkSize())
	chunkArray, err = h.sendMessageStreamSet.MessageToChunk(setChunkMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	h.sendMessageStreamSet.upadateLastStreamInfo(setChunkMsg, chunkArray[0].chunkStreamID)

	connectOkMsg, err := NewConnectSuccessMessage()
	if err != nil {
		return err
	}
	chunkArray, err = h.sendMessageStreamSet.MessageToChunk(connectOkMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	h.sendMessageStreamSet.upadateLastStreamInfo(connectOkMsg, chunkArray[0].chunkStreamID)

	_, err = h.rw.Write(w.Bytes())

	return err
}

func (h *RtmpHandler) handleCreateStreamCommand(r amf.Reader) error {
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
	chunkArray, err := h.sendMessageStreamSet.MessageToChunk(createStreamMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	h.sendMessageStreamSet.upadateLastStreamInfo(createStreamMsg, chunkArray[0].chunkStreamID)

	_, err = h.rw.Write(w.Bytes())

	return err
}

func (h *RtmpHandler) handlePublishCommand(r amf.Reader) error {

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
	chunkArray, err := h.sendMessageStreamSet.MessageToChunk(publishOkMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}
	h.sendMessageStreamSet.upadateLastStreamInfo(publishOkMsg, chunkArray[0].chunkStreamID)

	_, err = h.rw.Write(w.Bytes())

	return err
}
