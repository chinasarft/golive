package mp4

import (
	"io"
)

type SimpleBoxContainer struct {
	*Box
	SubBoxes []IBox
}

type SimpleFullBoxContainer struct {
	*FullBox
	EntryCount uint32
	SubBoxes   []IBox
}

func NewSimpleBoxContainer(b *Box) *SimpleBoxContainer {
	return &SimpleBoxContainer{
		Box: b,
	}
}

func (b *SimpleBoxContainer) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
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

func ParseSimpleBoxContainerBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewSimpleBoxContainer(box)
	totalReadLen, err = b.Parse(r)
	return
}

func parseChildBox(r io.Reader) (ibox IBox, totalReadLen int, err error) {

	var bb *Box
	if bb, totalReadLen, err = ParseBox(r); err != nil {
		return
	}

	curReadLen := 0
	parse := getParser(bb.BoxType)

	if ibox, curReadLen, err = parse(r, bb); err != nil {
		return
	}
	totalReadLen += curReadLen

	return
}

func (b *SimpleBoxContainer) Parse(r io.Reader) (totalReadLen int, err error) {

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	var ibox IBox
	for remainSize > 0 {
		if ibox, curReadLen, err = parseChildBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen
		b.SubBoxes = append(b.SubBoxes, ibox)
	}
	return
}

func (b *SimpleBoxContainer) GetSubBoxes() []IBox {
	return b.SubBoxes
}
