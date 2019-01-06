package mp4

import (
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
aligned(8) class FileTypeBox
extends Box(‘ftyp’) {
unsigned int(32) major_brand;
unsigned int(32) minor_version;
unsigned int(32) compatible_brands[]; // to end of the box
}
*/

type FtypBox struct {
	*Box
	MajorBrand       uint32
	MinorBrand       uint32
	CompatibleBrands []uint32
}

func NewFtypBox(b *Box) *FtypBox {
	return &FtypBox{
		Box: b,
	}
}

func (b *FtypBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in ftyp box")
		return
	}
	var curReadLen int = 0
	buf := make([]byte, 8)
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.MajorBrand = byteio.U32BE(buf[0:4])
	b.MinorBrand = byteio.U32BE(buf[4:8])

	buf = buf[0:4]
	for i := uint32(16); i < uint32(b.Size); i += 4 {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		totalReadLen += curReadLen
		compatibleBrand := byteio.U32BE(buf)
		b.CompatibleBrands = append(b.CompatibleBrands, compatibleBrand)
	}

	return
}

func (b *FtypBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if curWriteLen, err = byteio.WriteU32BE(w, b.MajorBrand); err != nil {
		return
	}
	writedLen += curWriteLen

	if curWriteLen, err = byteio.WriteU32BE(w, b.MinorBrand); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(b.CompatibleBrands); i++ {
		if curWriteLen, err = byteio.WriteU32BE(w, b.CompatibleBrands[i]); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}
