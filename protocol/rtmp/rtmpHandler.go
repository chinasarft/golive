package rtmp

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/chinasarft/golive/utils/amf"
)

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

type RtmpHandler struct {
	*RtmpUnpacker
	rw io.ReadWriter

	connetCmdObj  map[string]interface{}
	publishCmdObj ConnectCmdParam
	playCmdObj    PlayCmdParam
	appStreamKey  string

	AVCDecoderConfigurationRecord []byte
	AACSequenceHeader             []byte
	avMetaData                    []byte
	functionalStreamId            uint32 /* streambegin 里面的参数，应该没啥用
	                                      目前的实现是，该streamid作为user control message的消息msid
		 	 	 	 	 	 	 	 	 	 并且该streamid的值为play这个消息的msid */
	putMsg PutAvMessage
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
	eventType := int(m.Payload[0])*256 + int(m.Payload[1])
	switch eventType {
	case 0: // StreamBegin
		fmt.Println("OnUserControlMessage:StreamBegin")
	case 1: // StreamEOF
		fmt.Println("OnUserControlMessage:StreamEOF")
	case 2: // StreamDry
		fmt.Println("OnUserControlMessage:StreamDry")
	case 3: // SetBufferLength
		fmt.Println("OnUserControlMessage:SetBufferLength")
	case 4: // StreamIsRecorded
		fmt.Println("OnUserControlMessage:StreamIsRecorded")
	case 6: // PingRequest
		fmt.Println("OnUserControlMessage:PingRequest")
	case 7: // PingResponse
		fmt.Println("OnUserControlMessage:PingResponse")
	}
	return nil
}

func (h *RtmpHandler) OnCommandMessage(m *CommandMessage) (err error) {
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
					err = h.handleCreateStreamCommand(r)
					return

				//NetStream command
				case "publish":
					fmt.Println("receive publish command")
					err = h.handlePublishCommand(r)
					if err == nil {
						h.putMsg, err = RegisterSource(h)
					}
					return
				case "deleteStream":
					fmt.Println("receive deleteStream command")
				case "play":
					fmt.Println("receive play command")
					err = h.handlePlayCommand(r, m)
					if err == nil {
						err = RegisterSink(h)
					}
					return

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
		r := bytes.NewReader(m.Payload)
		v, e := amf.ReadValue(r)
		if e == nil {
			switch v.(type) {
			case string:
				value := v.(string)
				switch value {
				case "@setDataFrame":
					fmt.Println("@setDataFrame")
					// @setDataFrame固定长度是16字节
					h.avMetaData = m.Payload[16:]
					break
				}
			}
		}
	}
	return nil
}

func (h *RtmpHandler) OnVideoMessage(m *VideoMessage) error {

	// isKeyFrame := (m.Payload[0] >> 4) == 1
	// isAVCCodec := (m.Payload[0] | 0x0F) == 7
	// if m.Payload[1] == 0 && isKeyFrame && isAVCCodec {
	if h.AVCDecoderConfigurationRecord == nil && m.Payload[0] == 0x17 {
		fmt.Println("receive metavideo and put:", len(m.Payload), m.Payload[0])
		h.AVCDecoderConfigurationRecord = m.Payload
	} else {
		fmt.Println("receive video and put:", len(m.Payload), m.Payload[0], m.Timestamp)
		h.putMsg((*Message)(m))
	}

	return nil
}

func (h *RtmpHandler) OnAudioMessage(m *AudioMessage) error {

	if m.Payload[1] == 0 {
		h.AACSequenceHeader = m.Payload
	} else {
		fmt.Println("receive audio and put:", len(m.Payload), m.Timestamp)
		h.putMsg((*Message)(m))
	}
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

/*

connect request:
    +----------------+---------+---------------------------------------+
    |  Field Name    |  Type   |           Description                 |
    +--------------- +---------+---------------------------------------+
    | Command Name   | String  | Name of the command. Set to "connect".|
    +----------------+---------+---------------------------------------+
    | Transaction ID | Number  | Always set to 1.                      |
    +----------------+---------+---------------------------------------+
    | Command Object | Object  | Command information object which has  |
    |                |         | the name-value pairs.                 |
    +----------------+---------+---------------------------------------+
    | Optional User  | Object  | Any optional information              |
    | Arguments      |         |                                       |
    +----------------+---------+---------------------------------------+

connect response:
+--------------+----------+----------------------------------------+
| Field Name   |     Type |           Description                  |
+--------------+----------+----------------------------------------+
| Command Name |  String  | _result or _error; indicates whether   |
|              |          | the response is result or error.       |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  |     Transaction ID is 1 for connect    |
|      ID      |          |       responses                        |
+--------------+----------+----------------------------------------+
|  Properties  |  Object  |    Name-value pairs that describe the  |
|              |          | properties(fmsver etc.) of the         |
|              |          |      connection.                       |
+--------------+----------+----------------------------------------+
| Information  |  Object  |    Name-value pairs that describe the  |
|              |          | response from|the server. ’code’,      |
|              |          | ’level’, ’description’ are names of few|
|              |          | among such information.                |
+--------------+----------+----------------------------------------+

	Command Name已经被处理掉了
*/
func (h *RtmpHandler) handleConnectCommand(r amf.Reader) error {

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
				fmt.Println("handleconnect transactionId:", int(transactionId)) //7.2.11 always set to 1
			} else {
				panic("transactionId not number")
			}
		case 1:
			if cmdObj, ok := v.(amf.Object); ok {
				h.connetCmdObj = cmdObj
			} else {
				fmt.Println("-------=>", v, reflect.TypeOf(v))
				panic("cmd object not map")
			}
		}

	}

	w := &bytes.Buffer{}

	ackMsg := NewAckMessage(2500000)
	chunkArray, err := h.MessageToChunk(ackMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	setPeerBandwidthMsg := NewSetPeerBandwidthMessage(2500000, 2)
	chunkArray, err = h.MessageToChunk(setPeerBandwidthMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	bakChunkSize := h.chunkSerializer.GetChunkSize()
	defer func() {
		if err != nil {
			h.chunkSerializer.SetChunkSize(bakChunkSize)
		}
	}()

	h.chunkSerializer.SetChunkSize(1024)

	setChunkMsg := NewSetChunkSizeMessage(h.chunkSerializer.GetChunkSize())
	chunkArray, err = h.MessageToChunk(setChunkMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	connectOkMsg, err := NewConnectSuccessMessage()
	if err != nil {
		return err
	}
	chunkArray, err = h.MessageToChunk(connectOkMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

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
	chunkArray, err := h.MessageToChunk(createStreamMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
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
				fmt.Println("publish transactionId:", transactionId) //7.2.11 always set to 1
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
	chunkArray, err := h.MessageToChunk(publishOkMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())

	return err
}

/*
+--------------+----------+-----------------------------------------+
| Field Name   |   Type   |             Description                 |
+--------------+----------+-----------------------------------------+
| Command Name |  String  | Name of the command. Set to "play".     |
+--------------+----------+-----------------------------------------+
| Transaction  |  Number  | Transaction ID set to 0.                |
| ID           |          |                                         |
+--------------+----------+-----------------------------------------+
| Command      |   Null   | Command information does not exist.     |
| Object       |          | Set to null type.                       |
+--------------+----------+-----------------------------------------+
| Stream Name  |  String  | Name of the stream to play.             |
|              |          | To play video (FLV) files, specify the  |
|              |          | name of the stream without a file       |
|              |          | extension (for example, "sample"). To   |
|              |          | play back MP3 or ID3 tags, you must     |
|              |          | precede the stream name with mp3:       |
|              |          | (for example, "mp3:sample". To play     |
|              |          | H.264/AAC files, you must precede the   |
|              |          | stream name with mp4: and specify the   |
|              |          | file extension. For example, to play the|
|              |          | file sample.m4v,specify "mp4:sample.m4v"|
|              |          |                                         |
+--------------+----------+-----------------------------------------+
|     Start    |  Number  |   An optional parameter that specifies  |
|              |          | the start time in seconds. The default  |
|              |          | value is -2, which means the subscriber |
|              |          | first tries to play the live stream     |
|              |          | specified in the Stream Name field. If a|
|              |          | live stream of that name is not found,it|
|              |          | plays the recorded stream of the same   |
|              |          | name. If there is no recorded stream    |
|              |          | with that name, the subscriber waits for|
|              |          | a new live stream with that name and    |
|              |          | plays it when available. If you pass -1 |
|              |          | in the Start field, only the live stream|
|              |          | specified in the Stream Name field is   |
|              |          | played. If you pass 0 or a positive     |
|              |          | number in the Start field, a recorded   |
|              |          | stream specified in the Stream Name     |
|              |          | field is played beginning from the time |
|              |          | specified in the Start field. If no     |
|              |          | recorded stream is found, the next item |
|              |          | in the playlist is played.              |
|              |          |                                         |
+--------------+----------+-----------------------------------------+
|   Duration   |  Number  | An optional parameter that specifies the|
|              |          | duration of playback in seconds. The    |
|              |          | default value is -1. The -1 value means |
|              |          | a live stream is played until it is no  |
|              |          | longer available or a recorded stream is|
|              |          | played until it ends. If you pass 0, it |
|              |          | plays the single frame since the time   |
|              |          | specified in the Start field from the   |
|              |          | beginning of a recorded stream. It is   |
|              |          | assumed that the value specified in     |
|              |          | the Start field is equal to or greater  |
|              |          | than 0. If you pass a positive number,  |
|              |          | it plays a live stream for              |
|              |          | the time period specified in the        |
|              |          | Duration field. After that it becomes   |
|              |          | available or plays a recorded stream    |
|              |          | for the time specified in the Duration  |
|              |          | field. (If a stream ends before the     |
|              |          | time specified in the Duration field,   |
|              |          | playback ends when the stream ends.)    |
|              |          | If you pass a negative number other     |
|              |          | than -1 in the Duration field, it       |
|              |          | interprets the value as if it were -1.  |
|              |          |                                         |
+--------------+----------+-----------------------------------------+
| Reset        | Boolean  | An optional Boolean value or number     |
|              |          | that specifies whether to flush any     |
|              |          | previous playlist.                      |
+--------------+----------+-----------------------------------------+
*/

func (h *RtmpHandler) handlePlayCommand(r amf.Reader, m *CommandMessage) error {
	transactionId := 0
	for i := 0; ; i++ {
		v, e := amf.ReadValue(r)
		if e != nil {
			if e == io.EOF {
				break
			}
			fmt.Println("play--->", e)
			return e
		}
		switch i {
		case 0:
			if transId, ok := v.(float64); ok {
				transactionId = int(transId)
				fmt.Println("publish transactionId:", transactionId) //7.2.11 always set to 1
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

	streamIsRecordedMsg := NewUserControlCommandStreamIsRecorded(m.StreamID) // 这个streamid 应该没啥用
	chunkArray, err := h.MessageToChunk(streamIsRecordedMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	streamBeginMsg := NewUserControlCommandStreamBegin(m.StreamID) // 这个streamid 应该也没啥用
	chunkArray, err = h.MessageToChunk(streamBeginMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	playResetMsg, _ := NewPlayResetMessage(m.StreamID) // 这个streamid 应该也没啥用
	chunkArray, err = h.MessageToChunk(playResetMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())
	if err != nil {
		return err
	}

	w.Reset()
	playStartMsg, _ := NewPlayStartMessage(m.StreamID)
	chunkArray, err = h.MessageToChunk(playStartMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())
	if err != nil {
		return err
	}

	w.Reset()
	dataStartMsg, _ := NewDataStartMessage(m.StreamID)
	chunkArray, err = h.MessageToChunk(dataStartMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())
	if err != nil {
		return err
	}

	w.Reset()
	playPublishNotifyStartMsg, _ := NewPlayPublishNotifyMessage(m.StreamID)
	chunkArray, err = h.MessageToChunk(playPublishNotifyStartMsg, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())
	if err != nil {
		return err
	}

	return err
}

func (h *RtmpHandler) writeMessage(m *Message) error {

	w := &bytes.Buffer{}

	chunkArray, err := h.MessageToChunk(m, h.chunkSerializer.sendChunkSize)
	if err != nil {
		return err
	}
	err = h.chunkSerializer.SerializerChunk(chunkArray, w)
	if err != nil {
		return err
	}

	_, err = h.rw.Write(w.Bytes())
	if err != nil {
		return err
	}

	return err
}

// m是一个完整的消息，这个函数会拆分成chunk
func (s *RtmpHandler) MessageToChunk(m *Message, chunkSize uint32) ([]*Chunk, error) {

	csid := 2
	switch m.MessageType {
	case 1, 2, 3, 5, 6:
		if m.StreamID != 0 {
			return nil, fmt.Errorf("send msg streamid:%d for prot ctrl msg", m.StreamID)
		}
		csid = 2
	case 17, 20:
		csid = 3
	case 9:
		csid = 6 //TODO
	case 8:
		csid = 4
	case 4:
		csid = 8
	}
	fmt.Println("message.go:175", csid)

	chunkArray, err := m.ToType0Chunk(uint32(csid), chunkSize)

	return chunkArray, err
}