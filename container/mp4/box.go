package mp4

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"

	"github.com/chinasarft/golive/utils/byteio"
)

const (
	FULLBOX_ANY_FLAG    uint32 = 0xffffffff
	FULLBOX_ANY_VERSION bool   = true
	BOX_SIZE            int    = 8
	FULL_BOX_SIZE       int    = 12
)

type IBox interface {
	Parse(r io.Reader) (prasedLen int, err error)
	GetBoxType() uint32
	Serialize(w io.Writer) (writedLen int, err error)
}

//ISO_IEC_14496-12_2015 应该是最新box的说明
/*
aligned(8) class Box (unsigned int(32) boxtype,
         optional unsigned int(8)[16] extended_type) {
   unsigned int(32) size;
   unsigned int(32) type = boxtype;
   if (size==1) {
      unsigned int(64) largesize;
   } else if (size==0) {
      // box extends to end of file
   }
   if (boxtype==‘uuid’) {
      unsigned int(8)[16] usertype = extended_type;
} }
*/

type Box struct {
	Size         uint64  `json:"size"`
	BoxType      uint32  `json:"-`
	TypeName     string  `json"type"`
	ExtendedType []uint8 `json:"-`
}

/*
aligned(8) class FullBox(unsigned int(32) boxtype, unsigned int(8) v, bit(24) f)
   extends Box(boxtype) {
   unsigned int(8)   version = v;
   bit(24)           flags = f;
}
*/

type FullBox struct {
	*Box
	version    uint8  `json:"version"`
	flags24Bit uint32 `json:"flags"`
}

type UnsupporttedBox struct {
	*Box
	RawData []byte
}

func NewUnsupporttedBox(b *Box) *UnsupporttedBox {
	return &UnsupporttedBox{
		Box: b,
	}
}
func (b *UnsupporttedBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if len(b.RawData) > 0 {
		if curWriteLen, err = w.Write(b.RawData); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *UnsupporttedBox) Parse(r io.Reader) (totalReadLen int, err error) {

	var arr [4]byte
	byteio.PutU32BE(arr[0:4], b.BoxType)
	log.Printf("unknown box:%s %d\n", string(arr[0:4]), b.Size)

	remainLen := int(b.Size) - BOX_SIZE
	b.RawData = make([]byte, remainLen)
	return io.ReadFull(r, b.RawData)
}

func NewFullBox(b *Box) *FullBox {
	return &FullBox{
		Box: b,
	}
}

func (b *FullBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}
	vf := uint32(b.version)<<24 | b.flags24Bit
	if _, err = byteio.WriteU32BE(w, vf); err != nil {
		return
	}
	writedLen += 4

	return
}

func (b *FullBox) Parse(r io.Reader, expectVer uint8, isAnyVer bool, expectFlag uint32) (totalReadLen int, err error) {

	var arr [4]byte
	buf := arr[0:4]
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	version := buf[0]
	if !isAnyVer && version != expectVer {
		err = fmt.Errorf("fullbox expectVer:%d but:%d", expectVer, version)
		return
	}

	flags24Bit := byteio.U24BE(buf[1:4])
	if expectFlag != FULLBOX_ANY_FLAG && flags24Bit != expectFlag {
		err = fmt.Errorf("fullbox expectFlags:%d but:%d", expectFlag, flags24Bit)
		return
	}

	b.version = version
	b.flags24Bit = flags24Bit
	return
}

func NewBox() *Box {
	return &Box{}
}

func ParseBox(r io.Reader) (b *Box, readLen int, err error) {
	var arr [8]byte
	buf := arr[0:8]

	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}
	b = &Box{
		Size:    uint64(byteio.U32BE(buf)),
		BoxType: byteio.U32BE(buf[4:8]),
	}
	readLen = 8

	b.TypeName = string(buf[4:8])
	return
}

func (b *Box) GetBoxType() uint32 {
	return b.BoxType
}

func (b *Box) Serialize(w io.Writer) (writedLen int, err error) {

	if b.Size > uint64(math.MaxUint32) {
		if writedLen, err = byteio.WriteU32BE(w, 1); err != nil {
			return
		}
	} else {
		if writedLen, err = byteio.WriteU32BE(w, uint32(b.Size)); err != nil {
			return
		}
	}

	curWriteLen := 0
	if curWriteLen, err = byteio.WriteU32BE(w, b.BoxType); err != nil {
		return
	}
	writedLen += curWriteLen

	if b.Size > uint64(math.MaxUint32) {
		if curWriteLen, err = byteio.WriteU64BE(w, b.Size); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *Box) GetBoxTypeName() string {
	var arr [4]byte
	byteio.PutU32BE(arr[0:4], b.BoxType)
	return string(arr[0:4])
}

func (b *Box) Parse(r io.Reader) (res IBox, parsedLen int, err error) {
	var arr [8]byte
	buf := arr[0:8]

	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}

	b.Size = uint64(byteio.U32BE(buf))
	b.BoxType = byteio.U32BE(buf[4:8])
	b.TypeName = string(buf[4:8])

	switch b.BoxType {
	case BoxTypeFTYP:
		ftypBox := NewFtypBox(b)
		parsedLen, err = ftypBox.Parse(r)
		res = ftypBox
		return
	case BoxTypeSTYP:
		stypBox := NewStypBox(b)
		parsedLen, err = stypBox.Parse(r)
		res = stypBox
		return
	case BoxTypeSIDX:
		sidxBox := NewSidxBox(b)
		parsedLen, err = sidxBox.Parse(r)
		res = sidxBox
		return
	case BoxTypeMOOF:
		moofBox := NewMoofBox(b)
		parsedLen, err = moofBox.Parse(r)
		res = moofBox
		return
	case BoxTypeMDAT:
		mdatBox := NewMdatBox(b)
		parsedLen, err = mdatBox.Parse(r)
		res = mdatBox
		return
	case BoxTypeMOOV:
		moovBox := NewMoovBox(b)
		parsedLen, err = moovBox.Parse(r)
		res = moovBox
		return
	}

	return
}

func PrintBox(b IBox) {
	box, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(box))
}

func uint32Serialize(w io.Writer, nums []uint32) (writedLen int, err error) {
	curWriteLen := 0
	for i := 0; i < len(nums); i++ {
		if curWriteLen, err = byteio.WriteU32BE(w, nums[i]); err != nil {
			return
		}
		writedLen += curWriteLen
	}
	return
}
