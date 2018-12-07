package rtmp

import (
	"bytes"
	"fmt"
	"io"
	"log"

	//"github.com/chinasarft/golive/av"
	"github.com/chinasarft/golive/utils/byteio"
)

type ChunkBasicHeader struct {
	format        uint8
	chunkStreamID uint32
}

type ChunkMessageHeader struct {
	timestamp        uint32
	messageLength    uint32
	messageTypeID    uint8
	isStreamIDExists bool // 表示messageStreamID是否存在，可以靠format判断，所以这里是否多余？
	timestampExted   bool
	messageStreamID  uint32
}

type Chunk struct {
	*ChunkBasicHeader
	*ChunkMessageHeader
	data []byte
}

//TODO chunkStream可以发送多个messageStream, 所以还是要记录chunk的信息，至少要记录上一次发送的messageStreamID
type ChunkSerializer struct {
	chunkSize    uint32
	lastMSIDInfo map[uint32]uint32 // map[csid]msid
}

/*
+--------------+----------------+--------------------+--------------+
 | Basic Header | Message Header | Extended Timestamp |  Chunk Data  |
 +--------------+----------------+--------------------+--------------+
 |                                                    |
 |<------------------- Chunk Header ----------------->|
                            Chunk Format
*/

//Different message streams multiplexed onto the same chunk stream
//      are demultiplexed based on their message stream IDs
//一个chunkstream上可以跑多个messagestream。 但是目前抓包没有看到过这样做的
//但是以message为中心应该可以的
type ChunkStream struct {
	LastMessageStreamID uint32
	LastChunkFormat     uint8
	ChunkStreamID       uint32
}

type ChunkStreamSet struct {
	streams      map[uint32]*ChunkStream
	statusGetter MessageStreamStatusGetter
	chunkSize    uint32
}

func NewChunkStreamSet(getStatus MessageStreamStatusGetter) *ChunkStreamSet {
	return &ChunkStreamSet{
		streams:      make(map[uint32]*ChunkStream),
		statusGetter: getStatus,
		chunkSize:    128,
	}
}

func (s *ChunkStreamSet) SetChunkSize(size uint32) {
	s.chunkSize = size
}

func (s *ChunkStreamSet) ReadChunk(r io.Reader) (*Chunk, error) {

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

	remain := chunkMessageHdr.messageLength
	if chunkBasicHdr.format > 0 {
		stat, ok := s.statusGetter.GetMessageStreamStatus(cs.LastMessageStreamID)
		if !ok {
			return nil, fmt.Errorf("unknown status")
		}
		chunkMessageHdr.messageStreamID = cs.LastMessageStreamID
		if chunkBasicHdr.format > 1 {
			remain = stat.remain
		}
	}
	if chunkMessageHdr.isStreamIDExists {
		cs.LastMessageStreamID = chunkMessageHdr.messageStreamID
	}
	cs.LastChunkFormat = chunkBasicHdr.format

	data, err := readChunkData(r, remain, s.chunkSize)
	if err != nil {
		return nil, err
	}

	return &Chunk{chunkBasicHdr, chunkMessageHdr, data}, nil
}

//------------------

/*
 0 1 2 3 4 5 6 7
  +-+-+-+-+-+-+-+-+
  |fmt|   cs id   |
  +-+-+-+-+-+-+-+-+
 Chunk basic header 1

  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |fmt|      0    |  cs id - 64   |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
      Chunk basic header 2

0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |fmt|         1 |          cs id - 64           |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
             Chunk basic header 3
*/
func getChunkBasicHeader(r io.Reader) (basicHeader *ChunkBasicHeader, err error) {

	var h uint32 = 0
	h, err = byteio.ReadUint8(r)
	if err != nil {
		log.Println("read basic header: ", err)
		return
	}

	csid := h & 0x3f
	if csid == 0 {
		csid, err = byteio.ReadUint8(r)
		if err != nil {
			return
		}
		csid += 64
	} else if csid == 1 {
		csid, err = byteio.ReadUint16BE(r)
		if err != nil {
			return
		}
		csid += 64
	}

	format := h >> 6

	basicHeader = &ChunkBasicHeader{format: uint8(format), chunkStreamID: csid}

	return
}

/*
注意：message header后面可能有Extended Timestamp
Formt == 0
  0                   1                   2                   3
  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                          timestamp            | message length|
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |      message length (cont)    |message type id| msg stream id |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |          message stream id (cont)             |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
                Chunk Message Header - Type 0

Format == 1
  0                   1                   2                   3
  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                     timestamp delta           | message length|
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |      message length (cont)    |message type id|
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
                Chunk Message Header - Type 1

Formt == 2
  0                   1                   2
  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                    timestamp delta            |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
          Chunk Message Header - Type 2

Format == 3
0字节，它表示这个chunk的Message Header和上一个是完全相同的，无需再次传送
*/
func readChunkMessageHeader(r io.Reader, chunkFmt uint8) (*ChunkMessageHeader, error) {

	switch chunkFmt {
	case 0:
		return readChunkMessageHeaderType0(r)
	case 1:
		return readChunkMessageHeaderType1(r)
	case 2:
		return readChunkMessageHeaderType2(r)
	case 3:
		return &ChunkMessageHeader{}, nil
	}

	return nil, fmt.Errorf("unknown format:%d", chunkFmt)
}

func readChunkMessageHeaderType0(r io.Reader) (*ChunkMessageHeader, error) {
	buf := [11]byte{}
	bulSlice := buf[0:len(buf)]
	rLen, err := r.Read(bulSlice)
	if err != nil {
		return nil, err
	}
	if rLen != len(buf) {
		return nil, fmt.Errorf("not read enough data")
	}
	bio := bytes.NewReader(bulSlice)

	chunk := &ChunkMessageHeader{}

	chunk.timestamp, _ = byteio.ReadUint24BE(bio)
	chunk.messageLength, _ = byteio.ReadUint24BE(bio)
	messageTypeID, _ := byteio.ReadUint8(bio)
	chunk.messageTypeID = uint8(messageTypeID)
	chunk.messageStreamID, _ = byteio.ReadUint32LE(bio)
	chunk.isStreamIDExists = true

	chunk.timestampExted = false
	if chunk.timestamp == 0xffffff {
		chunk.timestamp, err = readChunkExtTimestamp(r)
		if err != nil {
			return nil, err
		}
		chunk.timestampExted = true
	}

	return chunk, nil
}

func readChunkMessageHeaderType1(r io.Reader) (*ChunkMessageHeader, error) {
	buf := [7]byte{}
	bulSlice := buf[0:len(buf)]
	rLen, err := r.Read(bulSlice)
	if err != nil {
		return nil, err
	}
	if rLen != len(buf) {
		return nil, fmt.Errorf("not read enough data")
	}
	bio := bytes.NewReader(bulSlice)

	chunk := &ChunkMessageHeader{}

	chunk.timestamp, _ = byteio.ReadUint24BE(bio)
	chunk.messageLength, _ = byteio.ReadUint24BE(bio)
	messageTypeID, _ := byteio.ReadUint8(bio)
	chunk.messageTypeID = uint8(messageTypeID)
	chunk.timestampExted = false

	if chunk.timestamp == 0xffffff {
		chunk.timestamp, err = readChunkExtTimestamp(r)
		if err != nil {
			return nil, err
		}
		chunk.timestampExted = true
	}

	return chunk, nil
}

func readChunkMessageHeaderType2(r io.Reader) (*ChunkMessageHeader, error) {
	timeStamp, err := byteio.ReadUint24BE(r)
	if err != nil {
		return nil, err
	}

	chunk := &ChunkMessageHeader{}
	chunk.timestamp = timeStamp
	chunk.timestampExted = false

	if timeStamp == 0xffffff {
		chunk.timestamp, err = readChunkExtTimestamp(r)
		if err != nil {
			return nil, err
		}
		chunk.timestampExted = true
	}

	return chunk, nil
}

func readChunkExtTimestamp(r io.Reader) (uint32, error) {
	return byteio.ReadUint32BE(r)
}

func readChunkData(r io.Reader, remain, chunkSize uint32) ([]byte, error) {
	readLen := remain
	if readLen > chunkSize {
		readLen = chunkSize
	}
	totalRead := readLen

	data := make([]byte, readLen)
	var retLen int = 0
	var err error

	for {
		retLen, err = r.Read(data[totalRead-readLen : totalRead])
		if err != nil {
			return nil, err
		}
		readLen -= uint32(retLen)
		if readLen == 0 {
			return data, nil
		}
	}

	return nil, fmt.Errorf("can't be here")
}

func (s *ChunkSerializer) SerializerChunk(chunkArray []*Chunk) ([]byte, error) {
	/*
		for idx, chunk := range chunkArray {
			lastMSID, ok := s.lastMSIDInfo[chunk.chunkStreamID]
			if !ok {
				lastMSID = chunk.messageStreamID
				s.lastMSIDInfo[chunk.chunkStreamID] = lastMSID
			}

		}
	*/
	return nil, nil
}

func (s *ChunkSerializer) SetChunkSize(chunkSize uint32) {
	s.chunkSize = chunkSize
}

func (s *ChunkSerializer) GetChunkSize() uint32 {
	return s.chunkSize
}
