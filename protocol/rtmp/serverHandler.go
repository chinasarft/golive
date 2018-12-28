package rtmp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/chinasarft/golive/utils/amf"
	"github.com/chinasarft/golive/utils/byteio"
)

type PutAVDMessage func(m *Message) error
type Pad interface {
	OnSourceDetermined(h *RtmpHandler, ctx context.Context) (PutAVDMessage, error)
	OnSinkDetermined(h *RtmpHandler, ctx context.Context) error
	OnDestroySource(h *RtmpHandler)
	OnDestroySink(h *RtmpHandler)
}

type ConnectCmdParam struct {
	PublishName string //stream name
	PublishType string
}

type PlayCmdParam struct {
	StreamName string
	Start      int
	Duration   int
	Reset      bool
}

var (
	rtmp_codec_h264 int = 7
	rtmp_codec_h265 int = 0x1c
	rtmp_codec_aac  int = 0x0a
)

type RtmpHandler struct {
	chunkUnpacker *ChunkUnpacker
	chunkPacker   *ChunkPacker
	rw            io.ReadWriter

	connetCmdObj  map[string]interface{}
	publishCmdObj ConnectCmdParam
	playCmdObj    PlayCmdParam
	appStreamKey  string

	avInfo             amf.Object
	videoCodecID       int
	audioCodecID       int
	hasReceivedAvMeta  bool   // librtmp就不会发送@setDataFrame
	functionalStreamId uint32 /* streambegin 里面的参数，应该没啥用
	                             目前的实现是，该streamid作为user control message的消息msid
		 	 	 	 	 	 	   并且该streamid的值为play这个消息的msid */

	ctx    context.Context
	cancel context.CancelFunc
	role   string
	pad    Pad
	putMsg PutAVDMessage

	status int // 做一个状态机？
}

func NewRtmpHandler(rw io.ReadWriter, pad Pad) *RtmpHandler {
	return &RtmpHandler{
		chunkUnpacker: NewChunkUnpacker(),
		chunkPacker:   NewChunkPacker(),
		rw:            rw,
		pad:           pad,
	}
}

func (h *RtmpHandler) Start() error {
	err := handshakeServer(h.rw)
	if err != nil {
		log.Println("rtmp HandshakeServer err:", err)
		return err
	}
	h.status = rtmp_state_hand_success

	h.ctx, h.cancel = context.WithCancel(context.Background())
	for {
		msg, err := h.chunkUnpacker.getRtmpMessage(h.rw)
		if err != nil {
			h.stop()
			return err
		}
		switch msg.MessageType {
		case 1, 2, 3, 5, 6:
			err = h.handleProtocolControlMessaage((*ProtocolControlMessaage)(msg))
		case 4:
			err = h.handleUserControlMessage((*UserControlMessage)(msg))
		case 8:
			err = h.handleAudioMessage((*AudioMessage)(msg))
		case 9:
			err = h.handleVideoMessage((*VideoMessage)(msg))
		case 15, 18:
			err = h.handleDataMessage((*DataMessage)(msg))
		case 17, 20:
			err = h.handleCommandMessage((*CommandMessage)(msg))
		case 16, 19:
			err = h.handleSharedObjectMessage((*SharedObjectMessage)(msg))
		case 22:
			err = h.handleAggregateMessage((*AggregateMessage)(msg))
		}
		if err != nil {
			h.stop()
			return err
		}

	}

}

func (h *RtmpHandler) GetAppStreamKey() string {
	return h.appStreamKey
}

func (h *RtmpHandler) GetFunctionalStreamId() uint32 {
	return h.functionalStreamId
}

func (h *RtmpHandler) stop() {
	if h.role == "source" {
		h.pad.OnDestroySource(h)
	} else if h.role == "sink" {
		h.pad.OnDestroySink(h)
	}
	h.Cancel()

	h.status = rtmp_state_stop
}

func (h *RtmpHandler) Cancel() {
	h.cancel()
}

func (h *RtmpHandler) handleProtocolControlMessaage(m *ProtocolControlMessaage) error {

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

func (h *RtmpHandler) handleUserControlMessage(m *UserControlMessage) error {
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

func (h *RtmpHandler) handleCommandMessage(m *CommandMessage) (err error) {
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
					log.Println("receive connect command")
					err = h.handleConnectCommand(r)
					if err == nil {
						h.status = rtmp_state_connect_success
					} else {
						h.status = rtmp_state_connect_fail
					}
				case "createStream":
					log.Println("receive createStream command")
					err = h.handleCreateStreamCommand(r)
					if err == nil {
						h.status = rtmp_state_crtstrm_success
					} else {
						h.status = rtmp_state_crtstrm_fail
					}
					return

				//NetStream command
				case "publish":
					log.Println("receive publish command")
					err = h.handlePublishCommand(r)
					if err == nil {
						h.putMsg, err = h.pad.OnSourceDetermined(h, h.ctx)
						if err == nil {
							h.role = "source"
						}
					}
					if err == nil {
						h.status = rtmp_state_publish_success
					} else {
						h.status = rtmp_state_publish_fail
					}
					return
				case "deleteStream":
					// 7.2.2.3 The server does not send any response.
					log.Println("receive deleteStream command")
				case "play":
					log.Println("receive play command")
					err = h.handlePlayCommand(r, m)
					if err == nil {
						err = h.pad.OnSinkDetermined(h, h.ctx)
						if err == nil {
							h.role = "sink"
						}
					}
					if err == nil {
						h.status = rtmp_state_play_start
					} else {
						h.status = rtmp_state_play_fail
					}
					return

				// TODO 以下命令文档里都没有找到
				case "releaseStream":
					log.Println("receive releaseStream command")
				case "FCPublish":
					log.Println("receive FCPublish command")
				case "FCUnpublish":
					log.Println("receive FCUnpublish command")
				}
			}
		} else {
			return e
		}
	}
	return nil
}

// TODO 其它data message需要透传么?
func (h *RtmpHandler) handleDataMessage(m *DataMessage) (err error) {
	switch m.MessageType {
	case 15: //AFM3
	case 18: //AFM0
		r := bytes.NewReader(m.Payload)
		v, e := amf.ReadValue(r)
		if e == nil {
			switch v.(type) {
			case string:
				value := v.(string)
				switch value {
				case "@setDataFrame":
					err = h.handleSetDataFrame(r, m)
					// @setDataFrame固定长度是16字节
					if err == nil {
						if err = h.putMsg((*Message)(m)); err != nil {
							return
						}
					}
					return
				}
			}
		}
	}
	return
}

func (h *RtmpHandler) handleVideoMessage(m *VideoMessage) error {

	isKeyFrame := (m.Payload[0] >> 4)
	vCodecId := int(m.Payload[0] & 0x0F)
	if h.hasReceivedAvMeta && h.videoCodecID != vCodecId {
		return fmt.Errorf("video codec id not same:%d %d", h.videoCodecID, vCodecId)
	}
	if m.Payload[1] == 0 { // sequence header
		if isKeyFrame != 1 {
			return fmt.Errorf("wrong vsequence header")
		}
		log.Println("receive metavideo and put:", len(m.Payload), m.Payload[0])
	}
	log.Println("receive video and put:", len(m.Payload), m.Payload[0], m.Timestamp)
	if err := h.putMsg((*Message)(m)); err != nil {
		return err
	}

	return nil
}

func (h *RtmpHandler) handleAudioMessage(m *AudioMessage) error {

	aCodecId := int((m.Payload[0] & 0xF0) >> 4)
	if h.hasReceivedAvMeta && h.audioCodecID != aCodecId {
		return fmt.Errorf("video codec id not same:%d %d", h.audioCodecID, aCodecId)
	}

	log.Println("receive audio and put:", len(m.Payload), m.Timestamp)
	if err := h.putMsg((*Message)(m)); err != nil {
		return err
	}

	return nil
}

func (h *RtmpHandler) handleSharedObjectMessage(m *SharedObjectMessage) error {
	switch m.MessageType {
	case 16: //AFM3
	case 19: //AFM0
	}
	return nil
}

func (h *RtmpHandler) handleAggregateMessage(m *AggregateMessage) error {
	return nil
}

func (h *RtmpHandler) handleConnectCommand(r amf.Reader) error {
	if h.status != rtmp_state_hand_success {
		panic("handle connect in wrong state")
	}

	for i := 0; i < 2; i++ {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			return fmt.Errorf("handleconnect amf:%s", e.Error())
		}
		switch i {
		case 0:
			if transactionId, ok := v.(float64); ok {
				log.Println("handleconnect transactionId:", int(transactionId)) //7.2.11 always set to 1
			} else {
				panic("transactionId not number")
			}
		case 1:
			if cmdObj, ok := v.(amf.Object); ok {
				h.connetCmdObj = cmdObj
			} else {
				log.Println("-------=>", v, reflect.TypeOf(v))
				panic("cmd object not map")
			}
		}

	}

	w := &bytes.Buffer{}

	ackMsg := NewAckMessage(2500000)
	chunkArray, err := h.chunkPacker.MessageToChunk(ackMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	setPeerBandwidthMsg := NewSetPeerBandwidthMessage(2500000, 2)
	chunkArray, err = h.chunkPacker.MessageToChunk(setPeerBandwidthMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	bakChunkSize := h.chunkPacker.GetChunkSize()
	defer func() {
		if err != nil {
			h.chunkPacker.SetChunkSize(bakChunkSize)
		}
	}()

	h.chunkPacker.SetChunkSize(1024)

	setChunkMsg := NewSetChunkSizeMessage(h.chunkPacker.GetChunkSize())
	chunkArray, err = h.chunkPacker.MessageToChunk(setChunkMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	connectOkMsg, err := NewConnectSuccessMessage()
	if err != nil {
		return err
	}
	chunkArray, err = h.chunkPacker.MessageToChunk(connectOkMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())

	return err
}

func (h *RtmpHandler) handleCreateStreamCommand(r amf.Reader) error {
	if h.status != rtmp_state_connect_success {
		panic("handle createstream in wrong state")
	}
	transactionId := 0
	for {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			log.Println("create stream--->", e)
			return e
		}
		if transId, ok := v.(float64); ok {
			transactionId = int(transId)
			log.Println("createstream transactionId:", transactionId)
		} else {
			log.Println("createstream value:", v)
		}
	}

	w := &bytes.Buffer{}

	createStreamMsg, err := NewCreateStreamSuccessMessage(transactionId)
	if err != nil {
		return err
	}
	chunkArray, err := h.chunkPacker.MessageToChunk(createStreamMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())

	return err
}

/*
publish request:
+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | Name of the command, set to "publish". |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | Transaction ID set to 0.               |
| ID           |          |                                        |
+--------------+----------+----------------------------------------+
| Command      |  Null    | Command information object does not    |
| Object       |          | exist. Set to null type.               |
+--------------+----------+----------------------------------------+
| Publishing   |  String  | Name with which the stream is          |
| Name         |          | published.                             |
+--------------+----------+----------------------------------------+
| Publishing   |  String  | Type of publishing. Set to "live",     |
| Type         |          | "record", or "append".                 |
|              |          | record: The stream is published and the|
|              |          | data is recorded to a new file.The file|
|              |          | is stored on the server in a           |
|              |          | subdirectory within the directory that |
|              |          | contains the server application. If the|
|              |          | file already exists, it is overwritten.|
|              |          | append: The stream is published and the|
|              |          | data is appended to a file. If no file |
|              |          | is found, it is created.               |
|              |          | live: Live data is published without   |
|              |          | recording it in a file.                |
+--------------+----------+----------------------------------------+


NetStream command response not just for publish
+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | The command name "onStatus".           |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | Transaction ID set to 0.               |
| ID           |          |                                        |
+--------------+----------+----------------------------------------+
| Command      |  Null    | There is no command object for         |
| Object       |          | onStatus messages.                     |
+--------------+----------+----------------------------------------+
|  Info Object | Object   |An AMF object having at least the       |
|              |          | following three properties:            |
|              |          |"level" (String): the level for this    |
|              |          |     message,  one of "warning",        |
|              |          |  "status", or "error";                 |
|              |          |                                        |
|              |          | "code" (String): the message code, for |
|              |          | example "NetStream.Play.Start";        |
|              |          |                                        |
|              |          | "description" (String): a human-       |
|              |          | readable description of the message.   |
|              |          | The Info object MAY contain other      |
|              |          | properties as appropriate to the code. |
+--------------+----------+----------------------------------------+
*/

func (h *RtmpHandler) handlePublishCommand(r amf.Reader) error {

	transactionId := 0
	for i := 0; i < 4; i++ {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			return fmt.Errorf("handlePublish amf:%s", e.Error())
		}

		switch i {
		case 0:
			if transId, ok := v.(float64); ok {
				transactionId = int(transId)
				log.Println("publish transactionId:", transactionId) //7.2.11 always set to 1
			} else {
				panic("transactionId not number")
			}
		case 1:
		case 2:
			if str, ok := v.(string); ok {
				h.publishCmdObj.PublishName = str
				h.appStreamKey = h.connetCmdObj["app"].(string) + "-" + str
			} else {
				panic("publish name not strig")
			}
		case 3:
			if str, ok := v.(string); ok {
				h.publishCmdObj.PublishType = str
			} else {
				panic("publish type not strig")
			}
		}
	}

	w := &bytes.Buffer{}

	// NetConnection 需要回复transactionid, netstream tid都设置为0 7.2.2
	publishOkMsg, err := NewPublishSuccessMessage()
	if err != nil {
		return err
	}
	chunkArray, err := h.chunkPacker.MessageToChunk(publishOkMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())

	return err
}

func (h *RtmpHandler) handlePlayCommand(r amf.Reader, m *CommandMessage) error {
	transactionId := 0
	for i := 0; ; i++ {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			log.Println("play--->", e)
			return e
		}
		switch i {
		case 0:
			if transId, ok := v.(float64); ok {
				transactionId = int(transId)
				log.Println("play transactionId:", transactionId) //7.2.11 always set to 1
			} else {
				panic("transactionId not number")
			}
		case 1: //Command Object must be nil
		case 2: //stream name
			if str, ok := v.(string); ok {
				h.playCmdObj.StreamName = str
				h.appStreamKey = h.connetCmdObj["app"].(string) + "-" + str
			} else {
				panic("stream name not strig")
			}
		}
	}

	w := &bytes.Buffer{}
	h.functionalStreamId = m.StreamID

	streamIsRecordedMsg := NewUserControlCommandStreamIsRecorded(h.functionalStreamId) // 这个streamid 应该没啥用
	chunkArray, err := h.chunkPacker.MessageToChunk(streamIsRecordedMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	streamBeginMsg := NewUserControlCommandStreamBegin(m.StreamID) // 这个streamid 应该也没啥用
	chunkArray, err = h.chunkPacker.MessageToChunk(streamBeginMsg)
	if err != nil {
		return err
	}
	err = h.chunkPacker.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	playResetMsg, _ := NewPlayResetMessage(m.StreamID) // 这个streamid 应该也没啥用
	chunkArray, err = h.chunkPacker.MessageToChunk(playResetMsg)
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

	w.Reset()
	playStartMsg, _ := NewPlayStartMessage(m.StreamID)
	chunkArray, err = h.chunkPacker.MessageToChunk(playStartMsg)
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

	w.Reset()
	dataStartMsg, _ := NewDataStartMessage(m.StreamID)
	chunkArray, err = h.chunkPacker.MessageToChunk(dataStartMsg)
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

	w.Reset()
	playPublishNotifyStartMsg, _ := NewPlayPublishNotifyMessage(m.StreamID)
	chunkArray, err = h.chunkPacker.MessageToChunk(playPublishNotifyStartMsg)
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

func (h *RtmpHandler) WriteMessage(m *Message) error {

	return h.chunkPacker.WriteMessage(h.rw, m)
}

func (h *RtmpHandler) handleSetDataFrame(r amf.Reader, m *DataMessage) error {

	log.Println("@setDataFrame")

	for i := 0; i < 2; i++ {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			return fmt.Errorf("handleSetDataFrame amf:%s", e.Error())
		}

		switch i {
		case 0:
		case 1:
			avinfo, ok := v.(amf.Object)
			if !ok {
				panic("amf.object to amf.Object fail")
			}
			h.avInfo = avinfo
			codecid, ok := avinfo["videocodecid"].(float64)
			if !ok {
				panic("videocodecid not float64")
			}
			vCodecID := int(codecid)
			if vCodecID != rtmp_codec_h264 && vCodecID != rtmp_codec_h265 {
				return fmt.Errorf("video not support codecid:%d", vCodecID)
			}

			codecid, ok = avinfo["audiocodecid"].(float64)
			if !ok {
				panic("audiocodecid not float64")
			}
			aCodecID := int(codecid)
			if aCodecID != rtmp_codec_aac {
				return fmt.Errorf("audio not support codecid:%d", aCodecID)
			}

			h.audioCodecID = aCodecID
			h.videoCodecID = vCodecID
			h.hasReceivedAvMeta = true
		}
	}

	return nil
}
