package flv

import (
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

const (
	FlvTagAudio = 8
	FlvTagVideo = 9
	FlvTagAMF0  = 18
	FlvTagAMF3  = 0xf
)

type FlvTag struct {
	TagType      uint8
	DataSize24   uint32
	Timestamp24  uint32
	TimestampExt uint8
	Timestamp    uint32
	StreamID     uint32
	Data         []byte
}

func ParseTag(r io.Reader) (tag *FlvTag, err error) {
	var bin [11]byte
	if _, err = io.ReadFull(r, bin[0:]); err != nil {
		return
	}
	tag = &FlvTag{
		TagType:      uint8(bin[0]),
		DataSize24:   byteio.U24BE(bin[1:4]),
		Timestamp24:  byteio.U24BE(bin[4:7]),
		TimestampExt: uint8(bin[7]),
		StreamID:     byteio.U24BE(bin[8:11]),
	}
	tag.Timestamp = tag.Timestamp24 | (uint32(tag.TimestampExt) << 24)
	tag.Data = make([]byte, tag.DataSize24)
	if _, err = io.ReadFull(r, tag.Data); err != nil {
		return
	}

	if _, err = io.ReadFull(r, bin[0:4]); err != nil {
		return
	}

	expectLen := byteio.U32BE(bin[0:4])
	if expectLen != tag.DataSize24+11 {
		err = fmt.Errorf("previous tag len:%d:%d", expectLen, tag.DataSize24)
		return
	}

	return
}
