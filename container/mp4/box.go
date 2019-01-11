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
	GetBoxSize() uint64
	GetSubBoxes() []IBox
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

type BoxParser func(r io.Reader, box *Box) (b IBox, readLen int, err error)

var (
	parseTable = map[uint32]BoxParser{
		BoxTypeFTYP: ParseFtypBox,
		BoxTypeSTYP: ParseFtypBox,
		BoxTypeSIDX: ParseSidxBox,
		BoxTypeMOOF: ParseMoofBox,
		BoxTypeMFHD: ParseMfhdBox,
		BoxTypeTRAF: ParseTrafBox,
		BoxTypeTFHD: ParseTfhdBox,
		BoxTypeTFDT: ParseTfdtBox,
		BoxTypeTRUN: ParseTrunBox,
		BoxTypeMVHD: ParseMvhdBox,
		BoxTypeTRAK: ParseTrakBox,
		BoxTypeTKHD: ParseTkhdBox,
		BoxTypeEDTS: ParseEdtsBox,
		BoxTypeMDIA: ParseMdiaBox,
		BoxTypeMDHD: ParseMdhdBox,
		BoxTypeHDLR: ParseHdlrBox,
		BoxTypeMINF: ParseMinfBox,
		BoxTypeVMHD: ParseVmhdBox,
		BoxTypeDINF: ParseDinfBox,
		BoxTypeDREF: ParseDrefBox,
		BoxTypeMVEX: ParseMvexBox,
		BoxTypeUDTA: ParseUdtaBox,
		BoxTypeSTBL: ParseStblBox,
		BoxTypeSTSD: ParseStsdBox,
		BoxTypeSTTS: ParseSttsBox,
		BoxTypeSTSC: ParseStscBox,
		BoxTypeSTSZ: ParseStszBox,
		BoxTypeSTZ2: ParseStszBox,
		BoxTypeSTCO: ParseStcoBox,
		BoxTypeELST: ParseElstBox,
		BoxTypeTREX: ParseTrexBox,
		BoxTypeMETA: ParseMetaBox,
		BoxTypeURL:  ParseUrlBox,
		BoxTypeURN:  ParseUrnBox,
		BoxTypeAVC1: ParseAvc1Box,
		BoxTypeMP4A: ParseMp4aBox,
		BoxTypeHEV1: ParseHev1Box,
		BoxTypePASP: ParsePaspBox,
		BoxTypeESDS: ParseEsdsBox,
		BoxTypeSMHD: ParseSmhdBox,
		BoxTypeMFRA: ParseMfraBox,
		BoxTypeTFRA: ParseTfraBox,
		BoxTypeMFRO: ParseMfroBox,
	}
)

func getParser(boxType uint32) BoxParser {
	if parser, ok := parseTable[boxType]; ok {
		return parser
	}

	return ParseUnsupporttedBox
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

func ParseUnsupporttedBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewUnsupporttedBox(box)
	totalReadLen, err = b.Parse(r)
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

func NewTypeBox(boxType uint32) *Box {
	b := &Box{
		BoxType: boxType,
		Size:    uint64(BOX_SIZE),
	}
	b.setTypeName()
	return b
}

func NewTypeFullBox(boxType uint32, verion uint8, flags uint32) *FullBox {

	b := &FullBox{
		Box:        NewTypeBox(boxType),
		version:    verion,
		flags24Bit: flags,
	}
	b.Size = uint64(FULL_BOX_SIZE)
	return b
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

	b.setTypeName()
	if b.Size == 1 && b.BoxType != BoxTypeMDAT {
		err = fmt.Errorf("large size in %s box", b.TypeName)
	}
	return
}

func (b *Box) setTypeName() {
	buf := []byte{0, 0, 0, 0}
	byteio.PutU32BE(buf, b.BoxType)
	b.TypeName = string(buf)
}

func (b *Box) GetBoxType() uint32 {
	return b.BoxType
}

func (b *Box) GetBoxSize() uint64 {
	return b.Size
}

func (b *Box) GetSubBoxes() []IBox {
	return nil
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

func (b *Box) Parse(r io.Reader) (res IBox, totalReadLen int, err error) {
	var arr [8]byte
	buf := arr[0:8]

	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	b.Size = uint64(byteio.U32BE(buf))
	b.BoxType = byteio.U32BE(buf[4:8])
	b.TypeName = string(buf[4:8])

	parsedLen := 0
	switch b.BoxType {
	case BoxTypeFTYP:
		ftypBox := NewFtypBox(b)
		parsedLen, err = ftypBox.Parse(r)
		res = ftypBox
	case BoxTypeSTYP:
		stypBox := NewStypBox(b)
		parsedLen, err = stypBox.Parse(r)
		res = stypBox
	case BoxTypeSIDX:
		sidxBox := NewSidxBox(b)
		parsedLen, err = sidxBox.Parse(r)
		res = sidxBox
	case BoxTypeMOOF:
		moofBox := NewMoofBox(b)
		parsedLen, err = moofBox.Parse(r)
		res = moofBox
	case BoxTypeMDAT:
		mdatBox := NewMdatBox(b)
		parsedLen, err = mdatBox.Parse(r)
		res = mdatBox
	case BoxTypeMOOV:
		moovBox := NewMoovBox(b)
		parsedLen, err = moovBox.Parse(r)
		res = moovBox
	case BoxTypeMFRA:
		mfraBox := NewMfraBox(b)
		parsedLen, err = mfraBox.Parse(r)
		res = mfraBox
	default:
		err = fmt.Errorf("no such box:%s", b.TypeName)
	}
	totalReadLen += parsedLen

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
