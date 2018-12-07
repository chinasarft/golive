package rtmp

import (
	"fmt"

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

type MessageStreamStatus struct {
	remain          uint32
	messageStreamID uint32 // 暂时没用到, chunkstream 记录了
}
type MessageStreamStatusGetter interface {
	GetMessageStreamStatus(streamID uint32) (*MessageStreamStatus, bool)
}

type PrevMessageStreamInfo struct {
	prevMessageStreamID uint32
	prevMessageLength   uint32
	prevTimestamp       uint32
	prevMessageTypeID   uint8
}
type PrevMessageStreamInfoGetter interface {
	GetPrevMessageStreamInfo(streamID uint32) (*PrevMessageStreamInfo, bool)
}

type Message struct {
	MessageType   uint8  `type:int endian:big length:1`
	PayloadLength uint32 `type:int endian:big length:3`
	Timestamp     uint32 `type:int endian:big length:4`
	StreamID      uint32 `type:int endian:big length:3`
	Payload       []byte `type:byte`
}

//这三种消息，按照文档里面来分的
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
	isCollecting   bool   //消息收集中，因为type!=0 忽略的头信息和上一个一样，区分一个type!=0消息的开头
	Data           []byte //读取的缓冲
}

type MessageStreamSet struct {
	streams map[uint32]*MessageStream
}

type SendMessageStream struct {
	lastMessageStreamID uint32
	lastChunkStreamID   uint32
	lastTimestamp       uint32
	lastMessageTypeID   uint8
}

type SendMessageStreamSet struct {
	streams map[uint32]*SendMessageStream
}

func NewMessageStreamSet() *MessageStreamSet {
	return &MessageStreamSet{streams: make(map[uint32]*MessageStream)}
}

func (s *MessageStreamSet) GetMessageStreamStatus(streamID uint32) (*MessageStreamStatus, bool) {
	messageStream, ok := s.streams[streamID]
	if !ok {
		return nil, false
	}
	if messageStream.isCollecting {
		return &MessageStreamStatus{
			remain:          messageStream.remain,
			messageStreamID: messageStream.messageStreamID,
		}, true
	} else {
		return &MessageStreamStatus{
			remain:          messageStream.messageLength,
			messageStreamID: messageStream.messageStreamID,
		}, true
	}
}

func (s *MessageStreamSet) HandleReceiveChunk(chunk *Chunk) (*Message, error) {

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
	switch chunk.format {
	case 0:
		messageStream.Timestamp = chunk.timestamp
		messageStream.messageLength = chunk.messageLength
		messageStream.MessageTypeID = chunk.messageTypeID
		messageStream.messageStreamID = chunk.messageStreamID
	case 1:
		messageStream.messageLength = chunk.messageLength
		messageStream.MessageTypeID = chunk.messageTypeID
		fallthrough
	case 2:
		messageStream.TimestampDelta = chunk.timestamp
		fallthrough
	case 3:
		messageStream.Timestamp += messageStream.TimestampDelta
	}

	if messageStream.isCollecting {
		messageStream.Data = append(messageStream.Data, chunk.data...)
		messageStream.remain -= uint32(len(chunk.data))
		if messageStream.remain == 0 {
			messageStream.isCollecting = false
			return s.getMessage(messageStream), nil
		}
	} else {
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

func (s *MessageStreamSet) getMessage(ms *MessageStream) *Message {

	message := &Message{}
	message.MessageType = ms.MessageTypeID
	message.PayloadLength = ms.messageLength
	message.StreamID = ms.messageStreamID
	message.Timestamp = ms.Timestamp
	message.Payload = ms.Data

	ms.Data = []byte{}
	ms.isCollecting = false

	return message
}

// m是一个完整的消息，这个函数会拆分成chunk
func (s *SendMessageStreamSet) MessageToChunk(m *Message, chunkSize uint32) ([]*Chunk, error) {

	messageStream, ok := s.streams[m.StreamID]
	if !ok {
		messageStream = &SendMessageStream{}
		s.streams[m.StreamID] = messageStream
	}

	csid := 2
	switch m.MessageType {
	case 1, 2, 3, 5, 6:
		if m.StreamID != 0 {
			return nil, fmt.Errorf("send msg streamid:%d for prot ctrl msg", m.StreamID)
		}
		csid = 2
	}
	fmt.Println(csid)

	chunkArray, err := m.ToChunk(uint32(csid), chunkSize)

	messageStream.lastMessageStreamID = m.StreamID
	messageStream.lastMessageTypeID = m.MessageType
	messageStream.lastTimestamp = m.Timestamp
	if !ok {
		messageStream.lastChunkStreamID = uint32(csid)
	} else {
		if messageStream.lastChunkStreamID != uint32(csid) {
			panic("message csid is not same as before")
		}
	}

	return chunkArray, err
}

//这里只负责生成chunk，format全是0或者3，更精细的format的拆分在发送时候决定
//都按照第一个是type 0来拆成chunk
func (m *Message) ToChunk(chunkStreamID, chunkSize uint32) ([]*Chunk, error) {

	payloadLen := len(m.Payload)

	chunkBasicHdr := &ChunkBasicHeader{format: 0, chunkStreamID: chunkStreamID}

	chunkMsgHdr := &ChunkMessageHeader{
		timestamp:        m.Timestamp,
		messageTypeID:    m.MessageType,
		isStreamIDExists: true,
		timestampExted:   false,
		messageStreamID:  m.StreamID,
	}
	if m.Timestamp > 0xffffff {
		chunkMsgHdr.timestampExted = true
	}
	var chunks []*Chunk

	appendChunk := func(basicHdr *ChunkBasicHeader) {

		chunk := &Chunk{
			ChunkBasicHeader:   basicHdr,
			ChunkMessageHeader: chunkMsgHdr,
		}

		if payloadLen <= int(chunkSize) {
			chunkMsgHdr.messageLength = uint32(payloadLen)
			chunk.data = m.Payload
			payloadLen = 0
		} else {
			chunk.data = m.Payload[0:chunkSize]
			m.Payload = m.Payload[chunkSize:payloadLen]
			payloadLen -= int(chunkSize)
		}
		chunks = append(chunks, chunk)
	}

	appendChunk(chunkBasicHdr)

	chunkBasicHdr3 := &ChunkBasicHeader{format: 3, chunkStreamID: chunkStreamID}
	//type 3
	//这里所有chunk都共用了chunkMsgHdr，type3 chunk都共用了chunkBasicHdr3
	for payloadLen > 0 {
		appendChunk(chunkBasicHdr3)
	}

	return chunks, nil
}

func NewAckMessage(windowSize uint32) *Message {
	message := &Message{
		MessageType:   5,
		PayloadLength: 4,
		Timestamp:     0,
		StreamID:      0,
	}
	message.Payload = make([]byte, 4)
	byteio.PutU32BE(message.Payload, windowSize)
	return message
}

func NewSetPeerBandwidthMessage(bandWidth uint32, limitType byte) *Message {
	message := &Message{
		MessageType:   6,
		PayloadLength: 5,
		Timestamp:     0,
		StreamID:      0,
	}
	message.Payload = make([]byte, 5)
	byteio.PutU32BE(message.Payload, bandWidth)
	message.Payload[4] = limitType
	return message
}

func NewSetChunkSizeMessage(chunkSize uint32) *Message {
	message := &Message{
		MessageType:   1,
		PayloadLength: 4,
		Timestamp:     0,
		StreamID:      0,
	}
	message.Payload = make([]byte, 4)
	byteio.PutU32BE(message.Payload, chunkSize)
	return message
}

func NewConnectSuccessMessage() (*Message, error) {

	message := &Message{
		MessageType:   0x14,
		PayloadLength: 4,
		Timestamp:     0,
		StreamID:      0,
	}

	var values []interface{}
	values = append(values, "_result")
	values = append(values, 1)

	obj1 := map[string]interface{}{
		"capabilities": 31,
		"fmsVer":       "FMS/3,0,1,123",
	}
	values = append(values, obj1)

	obj2 := map[string]interface{}{
		"level":          "status",
		"code":           "NetConnection.Connect.Success",
		"description":    "Connection succeeded.",
		"objectEncoding": 0,
	}
	values = append(values, obj2)

	data, err := amf.WriteArrayAsSiblingButElemArrayAsArray(values)
	if err != nil {
		return nil, err
	}

	message.Payload = data
	return message, nil
}
