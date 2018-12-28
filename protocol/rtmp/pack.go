package rtmp

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

type ChunkUnpacker struct {
	streams          map[uint32]*ChunkStream
	chunkSize        uint32
	messageCollector *MessageCollector
}

func NewChunkUnpacker() *ChunkUnpacker {
	return &ChunkUnpacker{
		streams:          make(map[uint32]*ChunkStream),
		chunkSize:        128,
		messageCollector: NewMessageCollector(),
	}
}

func (s *ChunkUnpacker) SetChunkSize(size uint32) {
	s.chunkSize = size
}

// chunk都带有完整的信息
func (s *ChunkUnpacker) ReadChunk(r io.Reader) (*Chunk, error) {

	chunkBasicHdr, err := getChunkBasicHeader(r)
	if err != nil {
		return nil, err
	}

	cs, ok := s.streams[chunkBasicHdr.chunkStreamID]
	if !ok {
		cs = &ChunkStream{}
		s.streams[chunkBasicHdr.chunkStreamID] = cs
		cs.ChunkStreamID = chunkBasicHdr.chunkStreamID
	}

	chunkMessageHdr, err := readChunkMessageHeader(r, chunkBasicHdr.format)
	if err != nil {
		return nil, err
	}

	if cs.remain == 0 {
		switch chunkBasicHdr.format {
		case 0:
			cs.remain = chunkMessageHdr.messageLength
			cs.ChunkBasicHeader = *chunkBasicHdr
			cs.ChunkMessageHeader = *chunkMessageHdr
		case 1:
			cs.remain = chunkMessageHdr.messageLength
			cs.messageLength = chunkMessageHdr.messageLength
			cs.messageTypeID = chunkMessageHdr.messageTypeID
			chunkMessageHdr.timestamp += cs.timestamp
			cs.timestamp = chunkMessageHdr.timestamp
			chunkMessageHdr.messageStreamID = cs.messageStreamID
		case 2:
			cs.remain = cs.messageLength
			chunkMessageHdr.timestamp += cs.timestamp
			cs.timestamp = chunkMessageHdr.timestamp
			chunkMessageHdr.messageLength = cs.messageLength
			chunkMessageHdr.messageTypeID = cs.messageTypeID
			chunkMessageHdr.messageStreamID = cs.messageStreamID
		case 3:
			cs.remain = cs.messageLength
			chunkMessageHdr.timestamp = cs.timestamp
			chunkMessageHdr.messageLength = cs.messageLength
			chunkMessageHdr.messageTypeID = cs.messageTypeID
			chunkMessageHdr.messageStreamID = cs.messageStreamID
		}
	}
	if !chunkBasicHdr.isStreamIDExists() {
		chunkMessageHdr.messageStreamID = cs.messageStreamID
	}

	data, err := readChunkData(r, cs.remain, s.chunkSize)
	if err != nil {
		return nil, err
	}

	cs.remain -= uint32(len(data))

	return &Chunk{chunkBasicHdr, chunkMessageHdr, data}, nil
}

func (h *ChunkUnpacker) getRtmpMessage(rw io.ReadWriter) (*Message, error) {

	for {
		chunk, err := h.ReadChunk(rw)
		if err != nil {
			return nil, err
		}
		log.Println("chunk timestamp:", chunk.timestamp)

		msg, err := h.messageCollector.HandleReceiveChunk(chunk)
		if err != nil {
			return nil, err
		}

		if chunk.chunkStreamID < 2 {
			return nil, fmt.Errorf("wrong csid:%d", chunk.chunkStreamID)
		}

		if msg == nil {
			continue
		}

		if msg.MessageType > 0 && msg.MessageType < 7 {
			if chunk.chunkStreamID != 2 {
				return nil, fmt.Errorf("csid:%d for msgtype:%d", chunk.chunkStreamID, msg.MessageType)
			}
			if msg.StreamID != 0 {
				return nil, fmt.Errorf("msid:%d for msgtype:%d", chunk.chunkStreamID, msg.MessageType)
			}
		}
		return msg, nil
	}

	panic("getRtmpMessage can't be here")
	return nil, fmt.Errorf("getRtmpMessage can't be here")
}

// ---------------------------

type ChunkPacker struct {
	sendChunkSize uint32
	sendStreams   map[uint32]*ChunkStream // 一个chunkStream可以发送多个messageStream，所以这里还是要需要该信息
}

func NewChunkPacker() *ChunkPacker {
	return &ChunkPacker{
		sendChunkSize: 128,
		sendStreams:   make(map[uint32]*ChunkStream),
	}
}

// 送入的chunk.timestamp都是type0的timestamp，需要自己判断
func (s *ChunkPacker) SerializerChunk(chunkArray []*Chunk, w *bytes.Buffer) (err error) {

	for _, chunk := range chunkArray {

		if err != nil {
			return err
		}

		cs, ok := s.sendStreams[chunk.chunkStreamID]
		if !ok {
			cs = &ChunkStream{}
			s.sendStreams[chunk.chunkStreamID] = cs
			cs.ChunkStreamID = chunk.chunkStreamID
			if cs.format != 0 {
				panic("chunkSerializer first chunk fmt not 0")
			}
		}

		isProtoCtrlMsg := false
		if chunk.chunkStreamID == 2 && chunk.messageStreamID == 0 &&
			chunk.messageTypeID != 4 && chunk.messageTypeID < 7 {
			isProtoCtrlMsg = true
		}

		if isProtoCtrlMsg || cs.messageStreamID != chunk.messageStreamID || !ok {
			cs.ChunkBasicHeader = *chunk.ChunkBasicHeader
			cs.ChunkMessageHeader = *chunk.ChunkMessageHeader
			cs.format = 0
			err = chunk.serializerType0(w)
		} else {
			if cs.messageStreamID != chunk.messageStreamID {
				panic("msid should equal")
			}
			timeDelta := chunk.timestamp - cs.timestamp
			if cs.messageLength == chunk.messageLength && cs.messageTypeID == chunk.messageTypeID {

				if cs.timestamp == chunk.timestamp || cs.timeDelta == timeDelta {
					chunk.serializerType3(w)
					cs.format = 3
				} else {
					cs.timeDelta = timeDelta
					chunk.timeDelta = timeDelta
					chunk.serializerType2(w, cs.timeDelta)
					cs.format = 2
					cs.timestamp = chunk.timestamp
				}

			} else {
				cs.timeDelta = timeDelta
				chunk.timeDelta = timeDelta
				chunk.serializerType1(w, cs.timeDelta)
				cs.format = 1
				cs.timestamp = chunk.timestamp
				cs.messageLength = chunk.messageLength
				cs.messageTypeID = chunk.messageTypeID
			}
		}

	}

	return nil
}

func (s *ChunkPacker) SetChunkSize(chunkSize uint32) {
	s.sendChunkSize = chunkSize
}

func (s *ChunkPacker) GetChunkSize() uint32 {
	return s.sendChunkSize
}

// m是一个完整的消息，这个函数会拆分成chunk
func (p *ChunkPacker) MessageToChunk(m *Message) ([]*Chunk, error) {

	csid := 2
	switch m.MessageType {
	case 1, 2, 3, 4, 5, 6:
		if m.StreamID != 0 {
			return nil, fmt.Errorf("send msg streamid:%d for prot ctrl msg", m.StreamID)
		}
		csid = 2
	case TYPE_CMDMSG_AMF0, TYPE_CMDMSG_AMF3:
		csid = 3 // TODO ffmpeg抓包,有时候是3有时候是8,应该是不通的命令
	case TYPE_DATA_AMF0, TYPE_DATA_AMF3: // data message
		csid = 4
	case TYPE_VIDEO:
		csid = 6 // TODO csid怎么选择?
	case TYPE_AUDIO:
		csid = 4
	}

	chunkArray, err := m.ToType0Chunk(uint32(csid), p.sendChunkSize)

	return chunkArray, err
}

func (p *ChunkPacker) WriteMessage(w io.Writer, m *Message) error {

	wBuf := &bytes.Buffer{}

	chunkArray, err := p.MessageToChunk(m)
	if err != nil {
		return err
	}
	err = p.SerializerChunk(chunkArray, wBuf)
	if err != nil {
		return err
	}

	data := wBuf.Bytes()
	_, err = w.Write(data)

	return err
}
