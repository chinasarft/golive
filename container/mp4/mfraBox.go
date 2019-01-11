package mp4

import (
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
aligned(8) class MovieFragmentRandomAccessBox extends Box(‘mfra’)
{

}
*/
type MfraBox = SimpleBoxContainer

/*
aligned(8) class TrackFragmentRandomAccessBox extends FullBox(‘tfra’, version, 0) {
  unsigned int(32)  track_ID;
  const unsigned int(26)  reserved = 0;
  unsigned int(2) length_size_of_traf_num;
  unsigned int(2) length_size_of_trun_num;
  unsigned int(2) length_size_of_sample_num;
  unsigned int(32)  number_of_entry;
  for(i=1; i • number_of_entry; i++){
    if(version==1){
      unsigned int(64)  time;
      unsigned int(64)  moof_offset;
    }else{
      unsigned int(32)  time;
      unsigned int(32)  moof_offset;
    }
    unsigned int((length_size_of_traf_num+1)*8) traf_number;
    unsigned int((length_size_of_trun_num+1)*8) trun_number;
    unsigned int((length_size_of_sample_num+1) * 8)sample_number;
  }
}
*/
type TfraEntry struct {
	Time         uint64
	MoofOffset   uint64
	TrafNumber   uint32 //length 只有2bit，最大为3，所以最多是uint32
	TrunNumber   uint32
	SampleNumber uint32
}
type TfraBox struct {
	*FullBox
	TrackID               uint32
	Reserved26Bit         uint32
	LengthOfTrafNum2Bit   uint8
	LengthOfTrunNum2Bit   uint8
	LengthOfSampleNum2Bit uint8
	NumberOfEntry         uint32
	Entries               []*TfraEntry
}

/*
aligned(8) class MovieFragmentRandomAccessOffsetBox extends FullBox(‘mfro’, version, 0) {
   unsigned int(32)  size;
}
*/
type MfroBox struct {
	*FullBox
	Size uint32
}

func NewMfraBox(b *Box) *MfraBox {
	return NewSimpleBoxContainer(b)
}

func ParseMfraBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewMfraBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func ParseTfraBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewTfraBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func NewTfraBox(b *Box) *TfraBox {
	return &TfraBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *TfraBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 28)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf[0:12]); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.TrackID = byteio.U32BE(buf)
	num := byteio.U32BE(buf[4:8])
	b.Reserved26Bit = num & 0xffffffC0
	b.LengthOfTrafNum2Bit = (buf[7] & 0x30) >> 4
	b.LengthOfTrunNum2Bit = (buf[7] & 0x0C) >> 2
	b.LengthOfSampleNum2Bit = buf[7] & 0x03
	b.NumberOfEntry = byteio.U32BE(buf[8:12])

	entryLen := 8 + b.LengthOfTrafNum2Bit + b.LengthOfTrafNum2Bit + b.LengthOfSampleNum2Bit + 3
	if b.version == 1 {
		entryLen += 8
	}
	buf = buf[0:entryLen]

	for i := uint32(0); i < b.NumberOfEntry; i++ {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		totalReadLen += curReadLen
		entry := &TfraEntry{}
		offset := uint8(8)
		if b.version == 1 {
			entry.Time = byteio.U64BE(buf)
			entry.MoofOffset = byteio.U64BE(buf[8:16])
			offset += 8
		} else {
			entry.Time = uint64(byteio.U32BE(buf))
			entry.MoofOffset = uint64(byteio.U32BE(buf[4:8]))
		}
		entry.TrafNumber = getValueByByteLength(buf[offset:], b.LengthOfTrafNum2Bit+1)
		offset += (b.LengthOfTrafNum2Bit + 1)
		entry.TrunNumber = getValueByByteLength(buf[offset:], b.LengthOfTrunNum2Bit+1)
		offset += (b.LengthOfTrunNum2Bit + 1)
		entry.SampleNumber = getValueByByteLength(buf[offset:], b.LengthOfSampleNum2Bit+1)

		b.Entries = append(b.Entries, entry)
	}

	return
}

func getValueByByteLength(buf []byte, length uint8) uint32 {
	switch length {
	case 1:
		return uint32(buf[0])
	case 2:
		return uint32(byteio.U16BE(buf))
	case 3:
		return byteio.U24BE(buf)
	case 4:
		return byteio.U32BE(buf)
	}
	return 0
}

func (b *TfraBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}
	curWriteLen := 0

	buf := make([]byte, 28)

	byteio.PutU32BE(buf, b.TrackID)
	num := b.Reserved26Bit | uint32(b.LengthOfTrafNum2Bit<<4) |
		uint32(b.LengthOfTrunNum2Bit) | uint32(b.LengthOfSampleNum2Bit)
	byteio.PutU32BE(buf[4:8], num)
	byteio.PutU32BE(buf[8:12], b.NumberOfEntry)
	if curWriteLen, err = w.Write(buf[0:12]); err != nil {
		return
	}
	writedLen += curWriteLen

	entryLen := 8 + b.LengthOfTrafNum2Bit + b.LengthOfTrafNum2Bit + b.LengthOfSampleNum2Bit + 3
	if b.version == 1 {
		entryLen += 8
	}
	buf = buf[0:entryLen]

	for i := uint32(0); i < b.NumberOfEntry; i++ {
		offset := uint8(8)
		if b.version == 1 {
			byteio.PutU64BE(buf, b.Entries[i].Time)
			byteio.PutU64BE(buf[8:16], b.Entries[i].MoofOffset)
			offset += 8
		} else {
			byteio.PutU32BE(buf, uint32(b.Entries[i].Time))
			byteio.PutU32BE(buf[4:8], uint32(b.Entries[i].MoofOffset))
		}

		putValueByByteLength(buf[offset:], b.LengthOfTrafNum2Bit+1, b.Entries[i].TrafNumber)
		offset += (b.LengthOfTrafNum2Bit + 1)
		putValueByByteLength(buf[offset:], b.LengthOfTrunNum2Bit+1, b.Entries[i].TrunNumber)
		offset += (b.LengthOfTrunNum2Bit + 1)
		putValueByByteLength(buf[offset:], b.LengthOfSampleNum2Bit+1, b.Entries[i].SampleNumber)

		if curWriteLen, err = w.Write(buf); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func putValueByByteLength(buf []byte, length uint8, v uint32) {
	switch length {
	case 1:
		buf[0] = uint8(v)
	case 2:
		byteio.PutU16BE(buf, uint16(v))
	case 3:
		byteio.PutU24BE(buf, v)
	case 4:
		byteio.PutU32BE(buf, v)
	}
}

func ParseMfroBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewMfroBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func NewMfroBox(b *Box) *MfroBox {
	return &MfroBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *MfroBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 4)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	b.Size = byteio.U32BE(buf)
	totalReadLen += curReadLen

	return
}

func (b *MfroBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if curWriteLen, err = byteio.WriteU32BE(w, b.Size); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}
