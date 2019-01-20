package mp4

import (
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
aligned(8) class MediaDataBox extends Box(‘mdat’) {
   bit(8) data[];
}
*/

type MdatBox struct {
	*Box
	Data []byte `json:"-"`
}

func NewMdatBox(b *Box) *MdatBox {
	return &MdatBox{
		Box: b,
	}
}

func (b *MdatBox) Parse(r io.Reader) (totalReadLen int, err error) {

	var boxLen uint64 = 8
	var curReadLen = 0
	buf := make([]byte, 8)
	if b.Size == 1 {
		if totalReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		b.Size = byteio.U64BE(buf)
		boxLen += 8
	}

	b.Data = make([]byte, b.Size-boxLen)
	if curReadLen, err = io.ReadFull(r, b.Data); err != nil {
		return
	}
	totalReadLen += curReadLen
	return
}

func (b *MdatBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if curWriteLen, err = w.Write(b.Data); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}
