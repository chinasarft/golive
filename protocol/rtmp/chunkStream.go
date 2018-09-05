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

type ChunkStream struct {
	Format    uint32
	CSID      uint32
	Timestamp uint32
	Length    uint32
	TypeID    uint32
	StreamID  uint32
	timeDelta uint32
	exted     bool
	index     uint32
	remain    uint32
	got       bool
	tmpFromat uint32
	Data      []byte
}

func (cs *ChunkStream) IsGetFullMessage() bool {
	return cs.got
}

func (cs *ChunkStream) allocDataBuffer() {
	cs.got = false
	cs.index = 0
	cs.remain = cs.Length
	if cs.Data == nil {
		cs.Data = make([]byte, cs.Length)
	} else {
		if uint32(len(cs.Data)) < cs.Length {
			cs.Data = make([]byte, cs.Length)
		}
	}
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
func getChunkBasicHeader(r io.Reader) (fmt, csid uint32, err error) {

	var h uint32 = 0
	h, err = byteio.ReadUint8(r)
	if err != nil {
		log.Println("read basic header: ", err)
		return
	}

	csid = h & 0x3f
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

	fmt = h >> 6

	return
}

func (cs *ChunkStream) readChunkBasicHeader(r io.Reader) error {

	fmt, csid, err := getChunkBasicHeader(r)
	if err != nil {
		return err
	}

	cs.tmpFromat = fmt
	cs.CSID = csid

	return err
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
func (cs *ChunkStream) readChunkMessageHeader(r io.Reader) error {

	switch cs.tmpFromat {
	case 0:
		err := cs.readChunkMessageHeaderType0(r)
		if err != nil {
			return err
		}
		cs.allocDataBuffer()
	case 1:
		err := cs.readChunkMessageHeaderType1(r)
		if err != nil {
			return err
		}
		cs.allocDataBuffer()
	case 2:
		err := cs.readChunkMessageHeaderType2(r)
		if err != nil {
			return err
		}
		cs.allocDataBuffer()
	case 3:
		err := cs.readChunkMessageHeaderType3(r)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid format=%d", cs.Format)
	}

	return nil
}

func (cs *ChunkStream) readChunkMessageHeaderType0(r io.Reader) error {
	buf := [11]byte{}
	bulSlice := buf[0:len(buf)]
	rLen, err := r.Read(bulSlice)
	if err != nil {
		return err
	}
	if rLen != len(buf) {
		return fmt.Errorf("not read enough data")
	}
	bio := bytes.NewReader(bulSlice)

	cs.Format = cs.tmpFromat
	cs.Timestamp, _ = byteio.ReadUint24BE(bio)
	cs.Length, _ = byteio.ReadUint24BE(bio)
	cs.TypeID, _ = byteio.ReadUint8(bio)
	cs.StreamID, _ = byteio.ReadUint32LE(bio)

	cs.exted = false
	if cs.Timestamp == 0xffffff {
		err = cs.readChunkExtTimestamp(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *ChunkStream) readChunkMessageHeaderType1(r io.Reader) error {
	buf := [7]byte{}
	bulSlice := buf[0:len(buf)]
	rLen, err := r.Read(bulSlice)
	if err != nil {
		return err
	}
	if rLen != len(buf) {
		return fmt.Errorf("not read enough data")
	}
	bio := bytes.NewReader(bulSlice)

	cs.Format = cs.tmpFromat
	timeStamp, _ := byteio.ReadUint24BE(bio)
	cs.Length, _ = byteio.ReadUint24BE(bio)
	cs.TypeID, _ = byteio.ReadUint8(bio)
	cs.exted = false
	if timeStamp == 0xffffff {
		err = cs.readChunkExtTimestamp(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *ChunkStream) readChunkMessageHeaderType2(r io.Reader) error {
	timeStamp, err := byteio.ReadUint24BE(r)
	if err != nil {
		return err
	}
	cs.exted = false
	cs.Format = cs.tmpFromat

	if timeStamp == 0xffffff {
		err = cs.readChunkExtTimestamp(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *ChunkStream) readChunkMessageHeaderType3(r io.Reader) (err error) {

	if cs.remain == 0 {
		if cs.exted {
			err = cs.readChunkExtTimestamp(r)
			if err != nil {
				return err
			}
		}
		cs.allocDataBuffer()
	} else {
		//这种情况扩展时间戳是多余的，读出来就行了
		// 但是livego里面确检查时候和之前的时间戳相等，不等的话是不会读出来的？
		if cs.exted {
			timestamp, err := byteio.ReadUint32BE(r)
			if timestamp != cs.timeDelta && timestamp != cs.Timestamp {
				panic("remain chunk wrong ext timestamp")
			}
			if err != nil {
				return err
			}
		}
	}

	return
}

func (cs *ChunkStream) readChunkExtTimestamp(r io.Reader) error {
	timestamp, err := byteio.ReadUint32BE(r)
	if err != nil {
		return err
	}
	cs.exted = true
	//format来判断ext timestamp是否为delta
	if cs.Format == 0 {
		cs.Timestamp = timestamp
	} else {
		cs.Timestamp += timestamp
		cs.timeDelta = timestamp
	}
	return nil
}

func (cs *ChunkStream) readChunkData(r io.Reader, chunkSize uint32) (err error) {
	realReadLen := 0
	shouldReadLen := chunkSize
	if cs.remain < chunkSize {
		shouldReadLen = cs.remain
	}
	for {
		buf := cs.Data[cs.index : cs.index+shouldReadLen]
		if realReadLen, err = r.Read(buf); err != nil {
			return err
		}
		cs.index += uint32(realReadLen)
		cs.remain -= uint32(realReadLen)
		shouldReadLen -= uint32(realReadLen)
		if shouldReadLen == 0 {
			break
		}
	}
	if cs.remain == 0 {
		cs.got = true
	}

	return
}

func (chunkStream *ChunkStream) readChunkWithoutBasicHeader(r io.Reader, chunkSize uint32) error {

	//message超过chunksize，后面的chunk format一定是3
	if chunkStream.remain != 0 && chunkStream.tmpFromat != 3 {
		return fmt.Errorf("inlaid remin = %d", chunkStream.remain)
	}

	err := chunkStream.readChunkMessageHeader(r)
	if err != nil {
		return err
	}

	return chunkStream.readChunkData(r, chunkSize)
}
