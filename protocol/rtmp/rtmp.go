package rtmp

import (
	"fmt"
	"io"
	"log"
)

type RtmpProtocolParamGetter interface {
	GetChunkSize(c *Chunk) uint32
}

type RtmpMessageHandler interface {
	//OnError(w io.Writer)
	OnProtocolControlMessaage(m *ProtocolControlMessaage) error
	OnUserControlMessage(m *UserControlMessage) error
	OnCommandMessage(m *CommandMessage) error
}

type RtmpHandler struct {
	rw               io.ReadWriter //timeout is depend on rw
	chunkStreamSet   *ChunkStreamSet
	messageStreamSet *MessageStreamSet
	messageHandler   RtmpMessageHandler
}

func NewRtmpHandler(rw io.ReadWriter, msgHandler RtmpMessageHandler) *RtmpHandler {
	messageStreamSet := NewMessageStreamSet()
	return &RtmpHandler{
		rw:               rw,
		chunkStreamSet:   NewChunkStreamSet(messageStreamSet),
		messageStreamSet: messageStreamSet,
		messageHandler:   msgHandler,
	}
}

func (h *RtmpHandler) Start() error {
	err := handshake(h.rw)
	if err != nil {
		log.Println("rtmp HandshakeServer err:", err)
		return err
	}
	for {
		chunk, err := h.chunkStreamSet.ReadChunk(h.rw)
		if err != nil {
			return err
		}

		msg, err := h.messageStreamSet.HandleReceiveChunk(chunk)
		if err != nil {
			return err
		}

		if msg != nil {
			switch msg.MessageType {
			case 1, 2, 3, 5, 6:
				if chunk.chunkStreamID != 2 {
					return fmt.Errorf("csid:%d for proto ctrl msg", chunk.chunkStreamID)
				}
				if msg.StreamID != 0 {
					return fmt.Errorf("msid:%d for proto ctrl msg", msg.StreamID)
				}
				h.messageHandler.OnProtocolControlMessaage((*ProtocolControlMessaage)(msg))
			case 4:
				if chunk.chunkStreamID != 2 {
					return fmt.Errorf("csid:%d for user ctrl msg", chunk.chunkStreamID)
				}
				if msg.StreamID != 0 {
					return fmt.Errorf("msid:%d for user ctrl msg", msg.StreamID)
				}
				h.messageHandler.OnUserControlMessage((*UserControlMessage)(msg))
			case 8, 9, 15, 16, 17, 18, 19, 20, 22:
				h.messageHandler.OnCommandMessage((*CommandMessage)(msg))

			}
		}
	}

	return nil
}
