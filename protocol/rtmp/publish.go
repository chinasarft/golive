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

const (
	rtmp_state_connect_send        = 100
	rtmp_state_crtstm_send         = 101
	rtmp_state_publish_send        = 102
	rtmp_state_play_send           = 103
	rtmp_state_stream_is_record    = 104
	rtmp_state_stream_begin        = 105
	rtmp_state_play_reset          = 106
	rtmp_state_play_start          = 107
	rtmp_state_data_start          = 108
	rtmp_state_play_publish_notify = 110
)

type RtmpPublishHandler struct {
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
}

func NewRtmpPublishHandler(rw io.ReadWriter, url string) (*RtmpPublishHandler, error) {
	if strings.Index(url, "rtmp://") != 0 {
		return nil, fmt.Errorf("not corrent rtmp url")
	}
	parts := strings.Split(url, "/")

	ctx, cancel := context.WithCancel(context.Background())
	ch := &RtmpPublishHandler{
		chunkUnpacker: NewChunkUnpacker(),
		chunkPacker:   NewChunkPacker(),
		rw:            rw,
		txMsgChan:     make(chan *Message),
		ctx:           ctx,
		cancel:        cancel,
		status:        rtmp_state_init,
		transactionId: 1,
	}
	ch.appName = parts[3]
	ch.streamName = parts[4]
	ch.tcurl = strings.Join(parts[0:4], "/")
	return ch, nil
}

func (h *RtmpPublishHandler) Start() {
	err := HandshakeClient(h.rw)
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
			h.sendAVDMessage(txmsg)
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

func (h *RtmpPublishHandler) sendConnectMessage() error {
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

	return h.writeMessage(msg)
}

func (h *RtmpPublishHandler) handleReceiveMessage(msg *Message) (err error) {

	switch msg.MessageType {
	case 1, 2, 3, 5, 6:
		err = h.handleProtocolControlMessaage((*ProtocolControlMessaage)(msg))
	case 4:
		err = h.handleUserControlMessage((*UserControlMessage)(msg))
	case 15, 18:
		err = h.handleDataMessage((*DataMessage)(msg))
	case 17, 20:
		err = h.handleCommandMessage((*CommandMessage)(msg))
	case 16, 19:
		err = h.handleSharedObjectMessage((*SharedObjectMessage)(msg))
	case 22:
		err = h.handleAggregateMessage((*AggregateMessage)(msg))
	case TYPE_AUDIO:
		fallthrough
	case TYPE_VIDEO:
		panic("receive av message")
	}
	return
}

func (h *RtmpPublishHandler) handleShakeSuccess() {
	h.status = rtmp_state_hand_success
}

func (h *RtmpPublishHandler) handleShakeFail() {
	h.status = rtmp_state_hand_fail
}

func (h *RtmpPublishHandler) Cancel() {

	h.cancel()
	h.status = rtmp_state_stop
}

func (h *RtmpPublishHandler) handleProtocolControlMessaage(m *ProtocolControlMessaage) error {

	switch m.MessageType {
	case TYPE_PRTCTRL_SET_CHUNK_SIZE:
		chunkSize := byteio.U32BE(m.Payload)
		h.chunkUnpacker.SetChunkSize(chunkSize)
	case 2:
	case 3:
	case TYPE_PRTCTRL_WINDOW_ACK:
	case TYPE_PRTCTRL_SET_PEER_BW:
	}
	return nil
}

func (h *RtmpPublishHandler) handleUserControlMessage(m *UserControlMessage) error {
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

func (h *RtmpPublishHandler) handleCommandMessage(m *CommandMessage) (err error) {
	switch m.MessageType {
	case 17: //AMF3
		fallthrough
	case 20: //AMF0
		switch h.status {
		case rtmp_state_connect_send:
			err = h.handleConnectResponse(m)
			if err != nil {
				return
			}
			err = h.sendSetChunkMessage()
			if err != nil {
				return
			}
			err = h.sendReleaseStreamMessage()
			if err != nil {
				return
			}
			err = h.sendFCPublishMessage()
			if err != nil {
				return
			}
			err = h.sendCreateStreamMessage()
			if err != nil {
				return
			}
			h.status = rtmp_state_crtstm_send
		case rtmp_state_crtstm_send:
			err = h.handleCreateStreamResponse(m)
			if err != nil {
				return
			}
			err = h.sendPublishMessage()
			if err != nil {
				return
			}
			h.status = rtmp_state_publish_send
		case rtmp_state_publish_send:
			err = h.handlePublishResponse(m)
			if err != nil {
				return
			}
			h.status = rtmp_state_publish_success
		case rtmp_state_publish_success:
			log.Println("<<<<<<<<<<<<receive msg after publish success>>>>>>>>>>>>>")
		}

	}
	return
}

func (h *RtmpPublishHandler) sendSetChunkMessage() error {
	msg := NewSetChunkSizeMessage(1024)
	return h.writeMessage(msg)
}

func (h *RtmpPublishHandler) sendReleaseStreamMessage() error {
	msg, err := NewReleaseStreamMessage(h.transactionId, h.streamName)
	if err != nil {
		return err
	}
	h.transactionId += 1
	return h.writeMessage(msg)
}

func (h *RtmpPublishHandler) sendFCPublishMessage() error {
	msg, err := NewFCPublishMessage(h.transactionId, h.streamName)
	if err != nil {
		return err
	}
	h.transactionId += 1
	return h.writeMessage(msg)
}

func (h *RtmpPublishHandler) sendCreateStreamMessage() error {
	msg, err := NewCreateStreamMessage(h.transactionId)
	if err != nil {
		return err
	}
	h.transactionId += 1
	return h.writeMessage(msg)
}

func (h *RtmpPublishHandler) sendPublishMessage() error {
	msg, err := NewPublishMessage(h.transactionId, h.appName, h.streamName)
	if err != nil {
		return err
	}
	h.transactionId += 1
	return h.writeMessage(msg)
}

// TODO 其它data message需要透传么?
func (h *RtmpPublishHandler) handleDataMessage(m *DataMessage) (err error) {
	return
}

func (h *RtmpPublishHandler) handleSharedObjectMessage(m *SharedObjectMessage) error {
	switch m.MessageType {
	case 16: //AFM3
	case 19: //AFM0
	}
	return nil
}

func (h *RtmpPublishHandler) handleAggregateMessage(m *AggregateMessage) error {
	return nil
}

func (h *RtmpPublishHandler) handleConnectResponse(m *CommandMessage) error {

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
				h.status = rtmp_state_connect_success
				break
			}
		}
	}
	if h.status != rtmp_state_connect_success {
		return fmt.Errorf("connect fail")
	}
	return nil
}

func (h *RtmpPublishHandler) handleCreateStreamResponse(m *CommandMessage) error {

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

func (h *RtmpPublishHandler) handlePublishResponse(m *CommandMessage) error {

	r := bytes.NewReader(m.Payload)
	v, e := amf.ReadValue(r)
	if e != nil {
		return e
	}
	str, ok := v.(string)
	if !ok || str != "onStatus" {
		return fmt.Errorf("publish wrong response:%s", str)
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	transId, ok := v.(float64)
	if !ok {
		return fmt.Errorf("publish wrong transid")
	}
	if uint32(transId) != 0 {
		return fmt.Errorf("publish resp tranid must be zero:%d", uint32(transId))
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	if v != nil {
		return fmt.Errorf("publish wrong response")
	}

	obj, e := amf.ReadValue(r)
	if e != nil {
		return e
	}
	objmap, ok := obj.(amf.Object)
	if !ok {
		return fmt.Errorf("connecting wrong response")
	}
	code, ok := objmap["code"]
	if ok {
		if code.(string) != "codeNetStream.Publish.Start" {
			return fmt.Errorf("publish fail:%s", code.(string))
		}
	}

	return nil
}

func (h *RtmpPublishHandler) sendAVDMessage(m *Message) error {
	switch m.MessageType {
	case TYPE_AUDIO, TYPE_VIDEO, TYPE_DATA_AMF0, TYPE_DATA_AMF3:
		m.StreamID = h.functionalStreamId

	}
	return h.writeMessage(m)
}

func (h *RtmpPublishHandler) writeMessage(m *Message) error {

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

func (h *RtmpPublishHandler) SendAudio(data []byte, timestamp uint32) error {
	message := &Message{
		MessageType: TYPE_AUDIO,
		Timestamp:   timestamp,
		StreamID:    0,
		Payload:     data,
	}

	h.txMsgChan <- message
	return nil
}

func (h *RtmpPublishHandler) SendVideo(data []byte, timestamp uint32) error {
	message := &Message{
		MessageType: TYPE_VIDEO,
		Timestamp:   timestamp,
		StreamID:    0,
		Payload:     data,
	}

	h.txMsgChan <- message
	return nil
}
