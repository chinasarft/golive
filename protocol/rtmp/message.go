package rtmp

import (
	"github.com/chinasarft/golive/utils/amf"
	"github.com/chinasarft/golive/utils/byteio"
)

var (
	BANDWIDTH_LIMIT_HARD = 0 //The peer SHOULD limit its output bandwidth to the indicated window size
	BANDWIDTH_LIMIT_SOFT = 1 //The peer SHOULD limit its output bandwidth to the the
	//window indicated in this message or the limit already in effect,
	//whichever is smaller
	BANDWIDTH_LIMIT_DYNAMIC = 1 //If the previous Limit Type was Hard, treat this message
	//as though it was marked Hard, otherwise ignore this message

)

type Message struct {
	MessageType uint8  `type:int endian:big length:1`
	Timestamp   uint32 `type:int endian:big length:4`
	StreamID    uint32 `type:int endian:little length:3`
	Payload     []byte `type:byte`
}

// 消息按照文档里面来分的
type ProtocolControlMessaage Message //chapter 5.4
type CommandMessage Message          //chapter 6.2
type UserControlMessage Message      //chapter 7
type DataMessage Message
type VideoMessage Message
type AudioMessage Message
type SharedObjectMessage Message
type AggregateMessage Message

type MessageStream struct {
	messageStreamID uint32
	chunkStreamID   uint32

	Format         uint8
	Timestamp      uint32
	TimestampDelta uint32
	messageLength  uint32
	MessageTypeID  uint8
	remain         uint32
	isCollecting   bool   // 消息收集中，因为type!=0 忽略的头信息和上一个一样，区分一个type!=0消息的开头
	Data           []byte // 读取的缓冲
}

type MessageCollector struct {
	streams map[uint32]*MessageStream
}

func NewMessageCollector() *MessageCollector {
	return &MessageCollector{
		streams: make(map[uint32]*MessageStream),
	}
}

func (s *MessageCollector) HandleReceiveChunk(chunk *Chunk) (*Message, error) {

	messageStream, ok := s.streams[chunk.messageStreamID]
	if !ok {
		messageStream = &MessageStream{}
		s.streams[chunk.messageStreamID] = messageStream
		if chunk.format != 0 {
			panic("first msg in msg strean format not zero")
		}
	}

	messageStream.Format = chunk.format
	messageStream.chunkStreamID = chunk.chunkStreamID

	if messageStream.isCollecting {
		messageStream.Data = append(messageStream.Data, chunk.data...)
		messageStream.remain -= uint32(len(chunk.data))
		if messageStream.remain == 0 {
			messageStream.isCollecting = false
			return s.getMessage(messageStream), nil
		}
	} else {

		messageStream.Timestamp = chunk.timestamp
		messageStream.messageLength = chunk.messageLength
		messageStream.MessageTypeID = chunk.messageTypeID
		messageStream.messageStreamID = chunk.messageStreamID

		if messageStream.messageLength == uint32(len(chunk.data)) {
			messageStream.Data = chunk.data
			return s.getMessage(messageStream), nil
		}
		messageStream.Data = append(messageStream.Data, chunk.data...)
		messageStream.isCollecting = true
		messageStream.remain = messageStream.messageLength - uint32(len(messageStream.Data))
	}
	return nil, nil
}

func (s *MessageCollector) getMessage(ms *MessageStream) *Message {

	message := &Message{}
	message.MessageType = ms.MessageTypeID
	message.StreamID = ms.messageStreamID
	message.Timestamp = ms.Timestamp
	message.Payload = ms.Data

	ms.Data = []byte{}
	ms.isCollecting = false

	return message
}

// 这里只负责生成chunk，format全是0，更精细的format的拆分在发送时候决定
func (m *Message) ToType0Chunk(chunkStreamID, chunkSize uint32) ([]*Chunk, error) {

	payloadLen := len(m.Payload)
	payload := m.Payload

	chunkBasicHdr := &ChunkBasicHeader{format: 0, chunkStreamID: chunkStreamID}

	chunkMsgHdr := &ChunkMessageHeader{
		timestamp:       m.Timestamp,
		messageTypeID:   m.MessageType,
		timestampExted:  false,
		messageStreamID: m.StreamID,
		messageLength:   uint32(payloadLen),
	}
	if m.Timestamp > 0xffffff {
		chunkMsgHdr.timestampExted = true
	}
	var chunks []*Chunk

	appendChunk := func() {

		chunk := &Chunk{
			ChunkBasicHeader:   chunkBasicHdr,
			ChunkMessageHeader: chunkMsgHdr,
		}

		if payloadLen <= int(chunkSize) {
			chunk.data = payload
			payloadLen = 0
		} else {
			chunk.data = payload[0:chunkSize]
			payload = payload[chunkSize:payloadLen]
			payloadLen -= int(chunkSize)
		}
		chunks = append(chunks, chunk)
	}

	// 这里所有chunk都共用了chunkMsgHdr chunkBasicHdr
	for payloadLen > 0 {
		appendChunk()
	}

	return chunks, nil
}

func NewAckMessage(windowSize uint32) *Message {
	message := &Message{
		MessageType: 5,
		Timestamp:   0,
		StreamID:    0,
	}
	message.Payload = make([]byte, 4)
	byteio.PutU32BE(message.Payload, windowSize)
	return message
}

func NewSetPeerBandwidthMessage(bandWidth uint32, limitType byte) *Message {
	message := &Message{
		MessageType: 6,
		Timestamp:   0,
		StreamID:    0,
	}
	message.Payload = make([]byte, 5)
	byteio.PutU32BE(message.Payload, bandWidth)
	message.Payload[4] = limitType
	return message
}

func NewSetChunkSizeMessage(chunkSize uint32) *Message {
	message := &Message{
		MessageType: 1,
		Timestamp:   0,
		StreamID:    0,
	}
	message.Payload = make([]byte, 4)
	byteio.PutU32BE(message.Payload, chunkSize)
	return message
}

func NewConnectSuccessMessage() (*Message, error) {

	message := &Message{
		MessageType: 0x14,
		Timestamp:   0,
		StreamID:    0,
	}

	var values []interface{}
	values = append(values, "_result")
	values = append(values, 1) // 7.2.1.1 transactionid Always set to 1

	obj1 := []interface{}{
		"fmsVer", "FMS/3,0,1,123",
		"capabilities", 31,
	}
	values = append(values, obj1)

	obj2 := []interface{}{
		"level", "status",
		"code", "NetConnection.Connect.Success",
		"description", "Connection succeeded.",
		"objectEncoding", 0,
	}
	values = append(values, obj2)

	data, err := amf.WriteArrayAsSiblingButElemArrayAsObject(values)
	if err != nil {
		return nil, err
	}

	message.Payload = data
	return message, nil
}

func NewCreateStreamSuccessMessage(transactionId int) (*Message, error) {
	message := &Message{
		MessageType: 0x14,
		Timestamp:   0,
		StreamID:    0,
	}

	var values []interface{}
	values = append(values, "_result")
	values = append(values, transactionId)
	values = append(values, nil)
	values = append(values, 1) //1 streamID, TODO 表示之后发送的用户控制消息的streamID么？

	data, err := amf.WriteArrayAsSiblingButElemArrayAsArray(values)
	if err != nil {
		return nil, err
	}

	message.Payload = data
	return message, nil
}

func NewPublishSuccessMessage() (*Message, error) {

	message := &Message{
		MessageType: 0x14,
		Timestamp:   0,
		StreamID:    0,
	}

	var values []interface{}
	values = append(values, "onStatus")
	values = append(values, 0) //7.2.2 transactionid 必须设置为0
	values = append(values, nil)

	obj1 := []interface{}{
		"description", "publishing",
		"level", "status",
		"code", "NetStream.Publish.Start",
	}
	values = append(values, obj1)

	data, err := amf.WriteArrayAsSiblingButElemArrayAsObject(values)
	if err != nil {
		return nil, err
	}

	message.Payload = data
	return message, nil
}

//---------------user control command---------

// 抓包来看, c->s s->c, 的message stream id 为1， 这个参数也为1，所以这里
// 认为是相等的

func NewUserControlCommandStreamBegin(stremid uint32) *Message {
	message := &Message{
		MessageType: 4,
		Timestamp:   0,
		StreamID:    stremid,
	}
	message.Payload = make([]byte, 6)
	message.Payload[0] = 0
	message.Payload[1] = 0
	byteio.PutU32BE(message.Payload[2:6], stremid)
	return message
}

// 同StreamBegin
func NewUserControlCommandStreamIsRecorded(stremid uint32) *Message {
	message := &Message{
		MessageType: 4,
		Timestamp:   0,
		StreamID:    stremid,
	}
	message.Payload = make([]byte, 6)
	message.Payload[0] = 0
	message.Payload[1] = 4
	byteio.PutU32BE(message.Payload[2:6], stremid)
	return message
}

func NewPlayResetMessage(stremid uint32) (*Message, error) {

	return NewNetStreamOnStatusMessageWithCodeLevelDesc(stremid,
		"NetStream.Play.Reset", "status", "Playing and resetting stream.")
}

func NewPlayStartMessage(stremid uint32) (*Message, error) {

	return NewNetStreamOnStatusMessageWithCodeLevelDesc(stremid,
		"NetStream.Play.Start", "status", "Started playing stream.")
}

func NewDataStartMessage(stremid uint32) (*Message, error) {

	return NewNetStreamOnStatusMessageWithCodeLevelDesc(stremid,
		"NetStream.Data.Start", "status", "Started playing stream.")
}

func NewPlayPublishNotifyMessage(stremid uint32) (*Message, error) {

	return NewNetStreamOnStatusMessageWithCodeLevelDesc(stremid,
		"NetStream.Play.PublishNotify", "status", "Started playing notify.")
}

func NewNetStreamOnStatusMessageWithCodeLevelDesc(stremid uint32, code, level, desc string) (*Message, error) {

	message := &Message{
		MessageType: 0x14,
		Timestamp:   0,
		StreamID:    stremid,
	}

	var values []interface{}
	values = append(values, "onStatus")
	values = append(values, 0)   // must be 0
	values = append(values, nil) // must be nil

	obj1 := []interface{}{
		"code", code,
		"level", level,
		"description", desc,
	}
	values = append(values, obj1)

	data, err := amf.WriteArrayAsSiblingButElemArrayAsObject(values)
	if err != nil {
		return nil, err
	}

	message.Payload = data
	return message, nil
}
