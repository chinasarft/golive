package rtmp

import (
	"context"
	"fmt"
	"io"
	"log"
)

type RtmpMessageHandler interface {
	OnError()
	OnHandleShakeSuccess()
	OnHandleShakeFail()
	OnProtocolControlMessaage(m *ProtocolControlMessaage) error
	OnUserControlMessage(m *UserControlMessage) error
	OnCommandMessage(m *CommandMessage) error
	OnDataMessage(m *DataMessage) error
	OnVideoMessage(m *VideoMessage) error
	OnAudioMessage(m *AudioMessage) error
	OnSharedObjectMessage(m *SharedObjectMessage) error
	OnAggregateMessage(m *AggregateMessage) error
}

type RtmpUnpacker struct {
	rw               io.ReadWriter //timeout is depend on rw
	chunkStreamSet   *ChunkStreamSet
	messageCollector *MessageCollector
	messageHandler   RtmpMessageHandler
	chunkSerializer  *ChunkSerializer
}

func NewRtmpUnpacker(rw io.ReadWriter, msgHandler RtmpMessageHandler) *RtmpUnpacker {
	messageStreamSet := NewMessageCollector()

	return &RtmpUnpacker{
		rw:               rw,
		chunkStreamSet:   NewChunkStreamSet(128),
		messageCollector: messageStreamSet,
		messageHandler:   msgHandler,
		chunkSerializer:  NewChunkSerializer(128),
	}
}

func (h *RtmpUnpacker) Start(ctx context.Context) error {
	err := handshake(h.rw)
	if err != nil {
		h.messageHandler.OnHandleShakeFail()
		log.Println("rtmp HandshakeServer err:", err)
		return err
	}
	h.messageHandler.OnHandleShakeSuccess()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		chunk, err := h.chunkStreamSet.ReadChunk(h.rw)
		if err != nil {
			h.messageHandler.OnError()
			return err
		}
		log.Println("chunk timestamp:", chunk.timestamp)

		msg, err := h.messageCollector.HandleReceiveChunk(chunk)
		if err != nil {
			h.messageHandler.OnError()
			return err
		}

		if msg != nil {
			switch msg.MessageType {
			case 1, 2, 3, 5, 6:
				if chunk.chunkStreamID != 2 {
					err = fmt.Errorf("csid:%d for proto ctrl msg", chunk.chunkStreamID)
					break
				}
				if msg.StreamID != 0 {
					err = fmt.Errorf("msid:%d for proto ctrl msg", msg.StreamID)
					break
				}
				err = h.messageHandler.OnProtocolControlMessaage((*ProtocolControlMessaage)(msg))
			case 4:
				if chunk.chunkStreamID != 2 {
					err = fmt.Errorf("csid:%d for user ctrl msg", chunk.chunkStreamID)
					break
				}
				if msg.StreamID != 0 {
					err = fmt.Errorf("msid:%d for user ctrl msg", msg.StreamID)
					break
				}
				err = h.messageHandler.OnUserControlMessage((*UserControlMessage)(msg))
			case 8:
				err = h.messageHandler.OnAudioMessage((*AudioMessage)(msg))
			case 9:
				err = h.messageHandler.OnVideoMessage((*VideoMessage)(msg))
			case 15, 18:
				err = h.messageHandler.OnDataMessage((*DataMessage)(msg))
			case 17, 20:
				if chunk.chunkStreamID < 3 {
					err = fmt.Errorf("csid:%d for cmd msg", chunk.chunkStreamID)
					break
				}
				err = h.messageHandler.OnCommandMessage((*CommandMessage)(msg))
			case 16, 19:
				err = h.messageHandler.OnSharedObjectMessage((*SharedObjectMessage)(msg))
			case 22:
				err = h.messageHandler.OnAggregateMessage((*AggregateMessage)(msg))
			}
			if err != nil {
				h.messageHandler.OnError()
				return err
			}
		}
	}

	return err
}
