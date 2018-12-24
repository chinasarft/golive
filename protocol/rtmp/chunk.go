package rtmp

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
+--------------+----------------+--------------------+--------------+
 | Basic Header | Message Header | Extended Timestamp |  Chunk Data  |
 +--------------+----------------+--------------------+--------------+
 |                                                    |
 |<------------------- Chunk Header ----------------->|
                            Chunk Format
*/

type ChunkBasicHeader struct {
	format        uint8
	chunkStreamID uint32
}

type ChunkMessageHeader struct {
	timestamp       uint32
	messageLength   uint32
	messageTypeID   uint8
	timestampExted  bool   // 对于接收都没啥用，都解析成了完整的timestamp
	timeDelta       uint32 // 对于接收可以是解析成完整的timestamp，只在发送时候使用
	messageStreamID uint32
}

type Chunk struct {
	*ChunkBasicHeader
	*ChunkMessageHeader
	data []byte
}

//Different message streams multiplexed onto the same chunk stream
//      are demultiplexed based on their message stream IDs
type ChunkStream struct {
	ChunkBasicHeader
	ChunkMessageHeader
	ChunkStreamID uint32
	remain        uint32
}

func (cbh ChunkBasicHeader) isStreamIDExists() bool {
	return cbh.format == 0
}

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
	_, err := io.ReadFull(r, bulSlice)
	if err != nil {
		return nil, err
	}
	bio := bytes.NewReader(bulSlice)

	chunk := &ChunkMessageHeader{}

	chunk.timestamp, _ = byteio.ReadUint24BE(bio)
	chunk.messageLength, _ = byteio.ReadUint24BE(bio)
	messageTypeID, _ := byteio.ReadUint8(bio)
	chunk.messageTypeID = uint8(messageTypeID)
	chunk.messageStreamID, _ = byteio.ReadUint32LE(bio)

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
	_, err := io.ReadFull(r, bulSlice)
	if err != nil {
		return nil, err
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

	data := make([]byte, readLen)

	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return data, nil
}

func serializerChunkBasicHeader(fmt uint8, csid uint32, w *bytes.Buffer) {

	h := fmt << 6
	if csid < 64 {
		w.WriteByte(byte(h | uint8(csid)))
	} else if csid < 256+64 {
		w.WriteByte(byte(h))
		w.WriteByte(byte(csid - 63))
	} else {
		w.WriteByte(byte(h + 1))
		csid -= 63
		w.WriteByte(byte(csid - 63))
		w.WriteByte(byte(csid % 256))
		w.WriteByte(byte(csid / 256))
	}

	return
}

func serializerChunkDeltaTime(w *bytes.Buffer, deltaTime uint32) {
	if deltaTime < 0xffffff {
		byteio.WriteU24BE(w, deltaTime)
	} else {
		w.WriteByte(0xff)
		w.WriteByte(0xff)
		w.WriteByte(0xff)
		byteio.WriteU32BE(w, deltaTime)
	}
}

func serializerChunkMessageHeaderNoStreamID(msgHdr *ChunkMessageHeader, w *bytes.Buffer) {
	byteio.WriteU24BE(w, msgHdr.timeDelta)
	byteio.WriteU24BE(w, msgHdr.messageLength)
	w.WriteByte(msgHdr.messageTypeID)
}

func serializerChunkMessageHeaderType1(msgHdr *ChunkMessageHeader, w *bytes.Buffer) {
	serializerChunkMessageHeaderNoStreamID(msgHdr, w)
	if msgHdr.timestamp > 0xffffff {
		byteio.WriteU32BE(w, msgHdr.timeDelta)
	}
}

func serializerChunkMessageHeader(msgHdr *ChunkMessageHeader, w *bytes.Buffer) {
	serializerChunkMessageHeaderNoStreamID(msgHdr, w)
	byteio.WriteU32LE(w, msgHdr.messageStreamID)
	if msgHdr.timestamp > 0xffffff {
		byteio.WriteU32BE(w, msgHdr.timestamp)
	}
}

func (c *Chunk) serializerType0(w *bytes.Buffer) error {
	serializerChunkBasicHeader(0, c.ChunkBasicHeader.chunkStreamID, w)
	serializerChunkMessageHeader(c.ChunkMessageHeader, w)
	w.Write(c.data)
	return nil
}

func (c *Chunk) serializerType1(w *bytes.Buffer, deltaTime uint32) error {
	serializerChunkBasicHeader(1, c.ChunkBasicHeader.chunkStreamID, w)
	serializerChunkMessageHeaderType1(c.ChunkMessageHeader, w)
	w.Write(c.data)
	return nil
}

func (c *Chunk) serializerType2(w *bytes.Buffer, deltaTime uint32) error {
	serializerChunkBasicHeader(2, c.ChunkBasicHeader.chunkStreamID, w)
	serializerChunkDeltaTime(w, deltaTime)
	w.Write(c.data)
	return nil
}

func (c *Chunk) serializerType3(w *bytes.Buffer) error {
	serializerChunkBasicHeader(3, c.ChunkBasicHeader.chunkStreamID, w)
	w.Write(c.data)
	return nil
}
