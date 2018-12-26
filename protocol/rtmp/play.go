package rtmp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chinasarft/golive/utils/amf"
	"github.com/chinasarft/golive/utils/byteio"
)

type Play interface {
	OnError(err error)
	OnAudioMessage(m *AudioMessage)
	OnVideoMessage(m *VideoMessage)
	OnDataMessage(m *DataMessage)
}

type RtmpPlayHandler struct {
	chunkUnpacker      *ChunkUnpacker
	chunkPacker        *ChunkPacker
	rw                 io.ReadWriter
	txMsgChan          chan *Message
	ctx                context.Context
	cancel             context.CancelFunc
	appName            string
	streamName         string
	tcurl              string
	status             int
	transactionId      int
	functionalStreamId uint32
	player             Play
}

func NewRtmpPlayHandler(rw io.ReadWriter, url string, player Play) (*RtmpPlayHandler, error) {
	if strings.Index(url, "rtmp://") != 0 {
		return nil, fmt.Errorf("not corrent rtmp url")
	}
	parts := strings.Split(url, "/")

	ctx, cancel := context.WithCancel(context.Background())
	ch := &RtmpPlayHandler{
		chunkUnpacker: NewChunkUnpacker(),
		chunkPacker:   NewChunkPacker(),
		rw:            rw,
		txMsgChan:     make(chan *Message),
		ctx:           ctx,
		cancel:        cancel,
		status:        rtmp_state_init,
		transactionId: 1,
		player:        player,
	}
	ch.appName = parts[3]
	ch.streamName = parts[4]
	ch.tcurl = strings.Join(parts[0:4], "/")
	return ch, nil
}

func (h *RtmpPlayHandler) Start() {
	var err error
	defer func() {
		if err != nil {
			log.Println(err)
			h.player.OnError(err)
		}
	}()
	err = HandshakeClient(h.rw)
	if err != nil {
		log.Println("rtmp HandshakeClient err:", err)
		return
	}

	err = h.sendConnectMessage()
	if err != nil {
		log.Println("sendConnectMessage err:", err)
		return
	}
	h.status = rtmp_state_connect_send

	for {
		select {
		case <-h.ctx.Done():
			return
		case txmsg := <-h.txMsgChan:
			if err = h.sendAVDMessage(txmsg); err != nil {
				log.Println(err)
				return
			}
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
		}
	}
}

func (h *RtmpPlayHandler) sendConnectMessage() error {
	event := make(amf.Object)
	event["app"] = h.appName
	event["type"] = "nonprivate"
	event["flashVer"] = "FMS.3.1"
	event["tcUrl"] = h.tcurl

	msg, err := NewConnectMessage(event, h.transactionId)
	if err != nil {
		return err
	}
	h.transactionId += 1

	return h.WriteMessage(msg)
}

func (h *RtmpPlayHandler) handleReceiveMessage(msg *Message) (err error) {

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

func (h *RtmpPlayHandler) handleShakeSuccess() {
	h.status = rtmp_state_hand_success
}

func (h *RtmpPlayHandler) handleShakeFail() {
	h.status = rtmp_state_hand_fail
}

func (h *RtmpPlayHandler) Cancel() {

	h.cancel()
	h.status = rtmp_state_stop
}

func (h *RtmpPlayHandler) handleProtocolControlMessaage(m *ProtocolControlMessaage) error {

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

func (h *RtmpPlayHandler) handleUserControlMessage(m *UserControlMessage) error {
	eventType := int(m.Payload[0])*256 + int(m.Payload[1])
	switch eventType {
	case 0: // StreamBegin
		log.Println("OnUserControlMessage:StreamBegin")
	case 1: // StreamEOF
		log.Println("OnUserControlMessage:StreamEOF")
	case 2: // StreamDry
		log.Println("OnUserControlMessage:StreamDry")
	case 3: // SetBufferLength
		log.Println("OnUserControlMessage:SetBufferLength")
	case 4: // StreamIsRecorded
		log.Println("OnUserControlMessage:StreamIsRecorded")
	case 6: // PingRequest
		log.Println("OnUserControlMessage:PingRequest")
	case 7: // PingResponse
		log.Println("OnUserControlMessage:PingResponse")
	}
	return nil
}

func (h *RtmpPlayHandler) handleCommandMessage(m *CommandMessage) (err error) {
	switch m.MessageType {
	case 17: //AMF3
		fallthrough
	case 20: //AMF0
		switch h.status {
		case rtmp_state_connect_send:
			err = h.handleConnectResponse(m)
			if err != nil {
				log.Println(err)
				return
			}
			h.status = rtmp_state_connect_success
			log.Println("rtmp_state_connect_success")
			err = h.sendWidowAckMessage()
			if err != nil {
				return
			}

			err = h.sendCreateStreamMessage()
			if err != nil {
				return
			}
			h.status = rtmp_state_crtstm_send
			log.Println("rtmp_state_crtstm_send")

		case rtmp_state_crtstm_send:
			log.Println("handleCreateStreamResponse")
			err = h.handleCreateStreamResponse(m)
			if err != nil {
				return
			}
			err = h.sendGetStreamLengthMessage()
			if err != nil {
				return
			}
			err = h.sendPlayMessage()
			if err != nil {
				return
			}
			err = h.sendSetBufferLengthMessage()
			if err != nil {
				return
			}
			h.status = rtmp_state_play_send
		case rtmp_state_play_send:
			err = h.handlePlayResponseReset(m)
			if err != nil {
				return
			}
			h.status = rtmp_state_play_reset
		case rtmp_state_play_reset:
			var status int
			err, status = h.handlePlayResponseStart(m)
			if err != nil {
				return
			}
			if status != 0 {
				h.status = status // should rtmp_state_play_success/start
			}
		case rtmp_state_play_start:
			log.Println("<<<<<<<<<<<<receive msg after publish success>>>>>>>>>>>>>")
		}

	}
	return
}

func (h *RtmpPlayHandler) sendWidowAckMessage() error {
	msg := NewAckMessage(2500000)
	return h.WriteMessage(msg)
}

func (h *RtmpPlayHandler) sendCreateStreamMessage() error {
	msg, err := NewCreateStreamMessage(h.transactionId)
	if err != nil {
		return err
	}
	h.transactionId += 1
	return h.WriteMessage(msg)
}

func (h *RtmpPlayHandler) sendGetStreamLengthMessage() error {
	msg, err := NewGetStreamLengthMessage(h.transactionId, h.streamName)
	if err != nil {
		return err
	}
	h.transactionId += 1
	return h.WriteMessage(msg)

}

func (h *RtmpPlayHandler) sendPlayMessage() error {
	msg, err := NewPlayMessage(h.transactionId, h.streamName, -2000)
	if err != nil {
		return err
	}
	h.transactionId += 1
	return h.WriteMessage(msg)
}

func (h *RtmpPlayHandler) sendSetBufferLengthMessage() error {
	msg := NewSetBufferLengthMessage(h.functionalStreamId, 1000)
	return h.WriteMessage(msg)
}

func (h *RtmpPlayHandler) handleSharedObjectMessage(m *SharedObjectMessage) error {
	switch m.MessageType {
	case 16: //AFM3
	case 19: //AFM0
	}
	return nil
}

func (h *RtmpPlayHandler) handleAggregateMessage(m *AggregateMessage) error {
	return nil
}

func (h *RtmpPlayHandler) handleConnectResponse(m *CommandMessage) error {
	log.Println("handleConnectResponse")
	r := bytes.NewReader(m.Payload)
	v, e := amf.ReadValue(r)
	if e != nil {
		return e
	}
	str, ok := v.(string)
	if !ok || str != "_result" {
		return fmt.Errorf("connecting wrong response:%s", str)
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	transId, ok := v.(float64)
	if !ok {
		return fmt.Errorf("connecting wrong transid")
	}
	log.Println("response for transId:", transId)

	status := h.status
	for {
		obj, e := amf.ReadValue(r)
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		objmap, ok := obj.(amf.Object)
		if !ok {
			return fmt.Errorf("connecting wrong response")
		}
		code, ok := objmap["code"]
		if ok {
			if code.(string) != "NetConnection.Connect.Success" {
				return fmt.Errorf("connect fail:%s", v.(string))
			} else {
				status = rtmp_state_connect_success
				break
			}
		}
	}
	if status != rtmp_state_connect_success {
		return fmt.Errorf("connect fail")
	}
	return nil
}

func (h *RtmpPlayHandler) handleCreateStreamResponse(m *CommandMessage) error {

	r := bytes.NewReader(m.Payload)
	v, e := amf.ReadValue(r)
	if e != nil {
		return e
	}
	str, ok := v.(string)
	if !ok || str != "_result" {
		return fmt.Errorf("createstream wrong response:%s", str)
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	transId, ok := v.(float64)
	if !ok {
		return fmt.Errorf("createstream wrong transid")
	}
	log.Println("response for transId:", transId)

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	if v != nil {
		return fmt.Errorf("createstream wrong response")
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	functionalStreamId, ok := v.(float64)
	if !ok {
		return fmt.Errorf("createstream wrong functionalStreamId")
	}
	h.functionalStreamId = uint32(functionalStreamId)
	h.status = rtmp_state_crtstrm_success
	return nil
}

func (h *RtmpPlayHandler) handlePlayResponseReset(m *CommandMessage) error {
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

func (h *RtmpPlayHandler) handlePlayResponseStart(m *CommandMessage) (error, int) {
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

func (h *RtmpPlayHandler) sendAVDMessage(m *Message) error {
	if h.status != rtmp_state_play_start {
		return fmt.Errorf("not in play start state")
	}
	switch m.MessageType {
	case TYPE_AUDIO, TYPE_VIDEO, TYPE_DATA_AMF0, TYPE_DATA_AMF3:
		m.StreamID = h.functionalStreamId

	}
	return h.WriteMessage(m)
}

func (h *RtmpPlayHandler) WriteMessage(m *Message) error {

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
