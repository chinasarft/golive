package mp4

import (
	"io"
	//"github.com/chinasarft/golive/utils/byteio"
)

/*
关于meta的定义没有找到，按照实际ffmpeg切出来的fmp4的样子来的
*/

type MetaBox struct {
	*FullBox
	SubBoxes []IBox
}

func NewMetaBox(b *Box) *MetaBox {
	return &MetaBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *MetaBox) Serialize(w io.Writer) (writedLen int, err error) {
	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	for i := 0; i < len(b.SubBoxes); i++ {
		if curWriteLen, err = b.SubBoxes[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func ParseMetaBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewMetaBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func (b *MetaBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	remainSize := int(b.Size) - FULL_BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeHDLR:
			hdlrBox := NewHdlrBox(bb)
			if curReadLen, err = hdlrBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, hdlrBox)

		case BoxTypeILST:
			fallthrough
		default:
			unsprtBox := NewUnsupporttedBox(bb)
			if curReadLen, err = unsprtBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, unsprtBox)
		}
		if curReadLen > 0 {
			remainSize -= curReadLen
			totalReadLen += curReadLen
		}

	}
	return
}
