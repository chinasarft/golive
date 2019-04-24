package rtmp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/chinasarft/golive/utils/amf"
	"github.com/chinasarft/golive/utils/byteio"
)

const (
	ROLE_PUBLISH = "publish"
	ROLE_PLAY    = "play"
)

type Play interface {
	OnError(err error)
	OnAudioMessage(m *AudioMessage)
	OnVideoMessage(m *VideoMessage)
	OnDataMessage(m *DataMessage)
}

type RtmpClientHandler struct {
	chunkUnpacker      *ChunkUnpacker
	chunkPacker        *ChunkPacker
	rw                 io.ReadWriter
	txMsgChan          chan *Message
	connectResult      chan bool
	ctx                context.Context
	rtmpUrl            *RtmpUrl
	status             int
	transactionId      uint32
	functionalStreamId uint32
	player             Play
	role               string
}

func NewRtmpClientHandler(rw io.ReadWriter, url, role string, player Play) (*RtmpClientHandler, error) {
	if role != ROLE_PLAY && role != ROLE_PUBLISH {
		return nil, fmt.Errorf("no such role:%s", role)
	}
	if role == ROLE_PLAY && player == nil {
		return nil, fmt.Errorf("player need Play interface")
	}
	rtmpUrl, err := parseRtmpUrl(url)
	if err != nil {
		return nil, err
	}

	ch := &RtmpClientHandler{
		chunkUnpacker: NewChunkUnpacker(),
		chunkPacker:   NewChunkPacker(),
		rw:            rw,
		txMsgChan:     make(chan *Message),
		connectResult: make(chan bool),
		status:        rtmp_state_init,
		transactionId: 1,
		player:        player,
		rtmpUrl:       rtmpUrl,
		role:          role,
	}

	return ch, nil
}

func (h *RtmpClientHandler) Start(ctx context.Context) {
	var err error
	defer func() {
		if h.status != rtmp_state_publish_success {
			h.connectResult <- false
		}
		if err != nil {
			if h.role == ROLE_PLAY {
				h.player.OnError(err)
			}
		}
	}()
	err = HandshakeClient(h.rw)
	if err != nil {
		log.Println("rtmp HandshakeClient err:", err)
		return
	}

	err = sendConnectMessage(h.rw, h.chunkPacker, h.rtmpUrl.getConnectCmdObj())
	if err != nil {
		log.Println("sendConnectMessage err:", err)
		return
	}
	h.status = rtmp_state_connect_send

	h.ctx, _ = context.WithCancel(ctx)

CONNOK:
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			var rxmsg *Message
			rxmsg, err = h.chunkUnpacker.getRtmpMessage(h.rw)
			if err != nil {
				log.Println(err)
				return
			}
			if err = h.handleReceiveMessage(rxmsg); err != nil {
				log.Println(err)
				return
			}
			//log.Println("ststus:", h.status)
			if h.status == rtmp_state_publish_success {
				h.connectResult <- true
				break CONNOK
			}
		}
	}

	for {
		select {
		case <-h.ctx.Done():
			return
		case txmsg := <-h.txMsgChan:
			if err = h.sendMessage(txmsg); err != nil {
				h.status = rtmp_state_send_fail
				log.Println(err)
			CLEAR:
				for {
					select {
					case <-h.txMsgChan:
					default:
						break CLEAR
					}
				}
				return
			}
		}
	}
}

func (h *RtmpClientHandler) ConnectResult() <-chan bool {
	return h.connectResult
}

func (h *RtmpClientHandler) handleReceiveMessage(msg *Message) (err error) {

	switch msg.MessageType {
	case 1, 2, 3, 5, 6:
		err = h.handleProtocolControlMessaage((*ProtocolControlMessaage)(msg))
	case 4:
		err = h.handleUserControlMessage((*UserControlMessage)(msg))
	case 17, 20:
		err = h.handleCommandMessage((*CommandMessage)(msg))
	case 16, 19:
		err = h.handleSharedObjectMessage((*SharedObjectMessage)(msg))
	case 22:
		err = h.handleAggregateMessage((*AggregateMessage)(msg))
	case TYPE_AUDIO:
		h.player.OnAudioMessage((*AudioMessage)(msg))
	case TYPE_VIDEO:
		h.player.OnVideoMessage((*VideoMessage)(msg))
	case 15, 18:
		h.player.OnDataMessage((*DataMessage)(msg))
	}

	return
}

func (h *RtmpClientHandler) Cancel() {
	h.status = rtmp_state_stop
}

func (h *RtmpClientHandler) handleProtocolControlMessaage(m *ProtocolControlMessaage) error {

	switch m.MessageType {
	case TYPE_PRTCTRL_SET_CHUNK_SIZE:
		chunkSize := byteio.U32BE(m.Payload)
		log.Println("set chunk size:", chunkSize)
		h.chunkUnpacker.SetChunkSize(chunkSize)
	case 2:
	case 3:
	case TYPE_PRTCTRL_WINDOW_ACK:
	case TYPE_PRTCTRL_SET_PEER_BW:
	}
	return nil
}

func (h *RtmpClientHandler) handleUserControlMessage(m *UserControlMessage) error {
	eventType := int(m.Payload[0])*256 + int(m.Payload[1])
	switch eventType {
	case 0: // StreamBegin
		if h.status != rtmp_state_stream_is_record {
			return fmt.Errorf("not int stream_is_record state")
		}
		log.Println("OnUserControlMessage:StreamBegin")
		h.status = rtmp_state_stream_begin
	case 1: // StreamEOF
		log.Println("OnUserControlMessage:StreamEOF")
	case 2: // StreamDry
		log.Println("OnUserControlMessage:StreamDry")
	case 3: // SetBufferLength
		log.Println("OnUserControlMessage:SetBufferLength")
	case 4: // StreamIsRecorded
		if h.status != rtmp_state_play_send {
			return fmt.Errorf("not int play_send state")
		}
		log.Println("OnUserControlMessage:StreamIsRecorded")
		h.status = rtmp_state_stream_is_record
	case 6: // PingRequest
		log.Println("OnUserControlMessage:PingRequest")
	case 7: // PingResponse
		log.Println("OnUserControlMessage:PingResponse")
	}
	return nil
}

func (h *RtmpClientHandler) handleCommandMessage(m *CommandMessage) (err error) {
	switch m.MessageType {
	case 17: //AMF3
		fallthrough
	case 20: //AMF0
		switch h.status {
		case rtmp_state_connect_send:
			err = handleConnectResponse(m)
			if err != nil {
				log.Println(err)
				return
			}
			h.status = rtmp_state_connect_success
			log.Println("rtmp_state_connect_success")

			h.chunkPacker.SetChunkSize(1024)
			err = sendSetChunkMessage(h.rw, h.chunkPacker, 1024)
			if err != nil {
				return
			}

			if h.role == ROLE_PLAY {
				err = sendWidowAckMessage(h.rw, h.chunkPacker, 2500000)
				if err != nil {
					return
				}
			} else {
				err = sendReleaseStreamMessage(h.rw, h.chunkPacker, h.transactionId, h.rtmpUrl.streamName)
				if err != nil {
					return
				}
				h.transactionId++

				err = sendFCPublishMessage(h.rw, h.chunkPacker, h.transactionId, h.rtmpUrl.streamName)
				if err != nil {
					return
				}
				h.transactionId++
			}

			err = sendCreateStreamMessage(h.rw, h.chunkPacker, h.transactionId, nil)
			if err != nil {
				return
			}
			h.transactionId++
			h.status = rtmp_state_crtstm_send

		case rtmp_state_crtstm_send:
			log.Println("rtmp_state_crtstm_send")
			h.functionalStreamId, err = handleCreateStreamResponse(m, 0)
			if err != nil {
				if err == error_Undefined {
					log.Println("rtmp_state_crtstm_send undefined")
					err = nil
					return
				}
				return
			}
			h.status = rtmp_state_crtstrm_success

			if h.role == ROLE_PLAY {
				err = sendGetStreamLengthMessage(h.rw, h.chunkPacker, h.transactionId, h.rtmpUrl.streamName)
				if err != nil {
					return
				}
				h.transactionId++

				err = sendPlayMessage(h.rw, h.chunkPacker, h.transactionId, -2000, h.rtmpUrl.streamName)
				if err != nil {
					return
				}
				h.transactionId++

				err = sendSetBufferLengthMessage(h.rw, h.chunkPacker, h.transactionId, 1000)
				if err != nil {
					return
				}
				h.transactionId++

				h.status = rtmp_state_play_send
			} else {
				err = sendPublishMessage(h.rw, h.chunkPacker, h.transactionId, h.rtmpUrl.appName, h.rtmpUrl.streamName)
				if err != nil {
					return
				}
				h.transactionId++
				h.status = rtmp_state_publish_send
			}
		case rtmp_state_stream_begin:
			if h.role != ROLE_PLAY {
				err = fmt.Errorf("state stream_begin role not math:%s", h.role)
				return
			}
			err = h.handlePlayResponseReset(m)
			if err != nil {
				return
			}
			h.status = rtmp_state_play_reset

		case rtmp_state_play_reset:
			if h.role != ROLE_PLAY {
				err = fmt.Errorf("state play_reset role not math:%s", h.role)
				return
			}
			var status int
			err, status = h.handlePlayResponseStart(m)
			if err != nil {
				return
			}
			log.Println("status:", status)
			if status != 0 {
				h.status = status // should rtmp_state_play_success/start
			}
		case rtmp_state_play_start:
			log.Println("<<<<<<<<<<<<receive msg after publish success>>>>>>>>>>>>>")
		case rtmp_state_publish_send:
			log.Println("rtmp_state_publish_send")
			if h.role != ROLE_PUBLISH {
				err = fmt.Errorf("state publish_send role not math:%s", h.role)
				return
			}
			err = handlePublishResponse(m)
			if err != nil {
				return
			}
			h.status = rtmp_state_publish_success
		case rtmp_state_publish_success:
			log.Println("rtmp_state_publish_success")
			if h.role != ROLE_PUBLISH {
				err = fmt.Errorf("state publish_success role not math:%s", h.role)
				return
			}
			log.Println("<<<<<<<<<<<<receive msg after publish success>>>>>>>>>>>>>")
		}

	}
	return
}

func (h *RtmpClientHandler) handleSharedObjectMessage(m *SharedObjectMessage) error {
	switch m.MessageType {
	case 16: //AFM3
	case 19: //AFM0
	}
	return nil
}

func (h *RtmpClientHandler) handleAggregateMessage(m *AggregateMessage) error {
	return nil
}

func (h *RtmpClientHandler) handlePlayResponseReset(m *CommandMessage) error {
	r := bytes.NewReader(m.Payload)
	v, e := amf.ReadValue(r)
	if e != nil {
		return e
	}
	str, ok := v.(string)
	if !ok || str != "onStatus" {
		return fmt.Errorf("play wrong response:%s", str)
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	transId, ok := v.(float64)
	if !ok {
		return fmt.Errorf("play wrong transid")
	}
	if transId != 0 {
		return fmt.Errorf("play transid must be zero")
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	if v != nil {
		return fmt.Errorf("play must be nil")
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	objmap, ok := v.(amf.Object)
	if !ok {
		return fmt.Errorf("play wrong response")
	}
	code, ok := objmap["code"]
	if ok {
		if code.(string) != "NetStream.Play.Reset" {
			return fmt.Errorf("play fail:%s", v.(string))
		}
	}

	return nil
}

func (h *RtmpClientHandler) handlePlayResponseStart(m *CommandMessage) (error, int) {
	r := bytes.NewReader(m.Payload)
	v, e := amf.ReadValue(r)
	if e != nil {
		return e, 0
	}
	str, ok := v.(string)
	if !ok || str != "onStatus" {
		return fmt.Errorf("play wrong response:%s", str), 0
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e, 0
	}
	transId, ok := v.(float64)
	if !ok {
		return fmt.Errorf("play wrong transid"), 0
	}
	if transId != 0 {
		return fmt.Errorf("play transid must be zero"), 0
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e, 0
	}
	if v != nil {
		return fmt.Errorf("play must be nil"), 0
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e, 0
	}
	objmap, ok := v.(amf.Object)
	if !ok {
		return fmt.Errorf("play wrong response"), 0
	}
	code, ok := objmap["code"]
	if ok {
		if code.(string) != "NetStream.Play.Start" {
			return nil, 0
		} else {
			return nil, rtmp_state_play_start
		}
	}

	return nil, 0
}

func (h *RtmpClientHandler) sendMessage(m *Message) error {
	if h.status != rtmp_state_publish_success {
		return fmt.Errorf("not in publish start state")
	}
	switch m.MessageType {
	case TYPE_AUDIO, TYPE_VIDEO, TYPE_DATA_AMF0, TYPE_DATA_AMF3:
		m.StreamID = h.functionalStreamId

	}
	return h.WriteMessage(m)
}

func (h *RtmpClientHandler) WriteMessage(m *Message) error {

	w := &bytes.Buffer{}

	chunkArray, err := h.chunkPacker.MessageToChunk(m)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())
	if err != nil {
		return err
	}

	return err
}

func (h *RtmpClientHandler) SendAudioMessage(data []byte, timestamp uint32) error {
	if h.status != rtmp_state_publish_success {
		return fmt.Errorf("not in publish start state")
	}
	message := &Message{
		MessageType: TYPE_AUDIO,
		Timestamp:   timestamp,
		StreamID:    0,
		Payload:     data,
	}

	h.txMsgChan <- message
	return nil
}

func (h *RtmpClientHandler) SendVideoMessage(data []byte, timestamp uint32) error {
	if h.status != rtmp_state_publish_success {
		return fmt.Errorf("not in publish start state")
	}
	message := &Message{
		MessageType: TYPE_VIDEO,
		Timestamp:   timestamp,
		StreamID:    0,
		Payload:     data,
	}
	// TODO 分析封装flv tag body(only body) known as rtmp message

	h.txMsgChan <- message
	return nil
}
