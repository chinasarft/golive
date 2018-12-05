package rtmp

type MessageStreamStatus struct {
	remain uint32
}
type MessageStreamStatusGetter interface {
	GetMessageStreamStatus(streamID uint32) (*MessageStreamStatus, bool)
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

type MessageStream struct {
	messageStreamID uint32
	chunkStreamID   uint32

	Format        uint8
	Timestamp     uint32
	messageLength uint32
	MessageTypeID uint8
	remain        uint32
	isCollecting  bool   //消息收集中，因为type!=0 忽略的头信息和上一个一样，区分一个type!=0消息的开头
	Data          []byte //读取的缓冲
}

type MessageStreamSet struct {
	streams map[uint32]*MessageStream
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
		return &MessageStreamStatus{remain: messageStream.remain}, true
	} else {
		return &MessageStreamStatus{remain: messageStream.messageLength}, true
	}
}

func (s *MessageStreamSet) HandleReceiveChunk(chunk *Chunk) (*Message, error) {
	messageStream, ok := s.streams[chunk.messageStreamID]
	if !ok {
		messageStream = &MessageStream{}
		messageStream.Format = chunk.format
		messageStream.chunkStreamID = chunk.chunkStreamID
		messageStream.messageStreamID = chunk.messageStreamID
		messageStream.messageLength = chunk.messaageLength
		messageStream.Timestamp = chunk.timestamp
		messageStream.MessageTypeID = chunk.messageTypeID

		s.streams[chunk.messageStreamID] = messageStream

		if chunk.messaageLength == uint32(len(chunk.data)) {
			messageStream.Data = chunk.data
			return s.getMessage(messageStream), nil
		}
		messageStream.Data = append(messageStream.Data, chunk.data...)
		messageStream.isCollecting = true
		messageStream.remain = chunk.messaageLength - uint32(len(chunk.data))
		return nil, nil
	}

	if chunk.format == 0 {
		messageStream.Timestamp = chunk.timestamp
	} else if chunk.format == 1 || chunk.format == 2 {
		messageStream.Timestamp += chunk.timestamp
	}

	if messageStream.isCollecting {
		messageStream.Data = append(messageStream.Data, chunk.data...)
		messageStream.remain -= uint32(len(chunk.data))
		if messageStream.remain == 0 {
			messageStream.isCollecting = false
			return s.getMessage(messageStream), nil
		}
	} else {
		//TODO 与上面代码一样
		if chunk.messaageLength == uint32(len(chunk.data)) {
			messageStream.Data = chunk.data
			return s.getMessage(messageStream), nil
		}
		messageStream.Data = append(messageStream.Data, chunk.data...)
		messageStream.isCollecting = true
		messageStream.remain = chunk.messaageLength - uint32(len(chunk.data))
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

func (ms *MessageStream) SendMessage(m *Message) (int, error) {
	//TODO message to chunk and sendchunk by chunkStream
	return 0, nil
}

func (ms *MessageStream) HandleReceiveChunk(c *Chunk) error {
	//
	return nil
}
