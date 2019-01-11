package mp4

import (
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
aligned(8) class SegmentIndexBox extends FullBox(‘sidx’, version, 0) {
   unsigned int(32) reference_ID;
   unsigned int(32) timescale;
   if (version==0) {
         unsigned int(32) earliest_presentation_time;
         unsigned int(32) first_offset;
      }
      else {
         unsigned int(64) earliest_presentation_time;
         unsigned int(64) first_offset;
      }
   unsigned int(16) reserved = 0;
   unsigned int(16) reference_count;
   for(i=1; i <= reference_count; i++)
   {
      bit (1)           reference_type;
      unsigned int(31)  referenced_size;
      unsigned int(32)  subsegment_duration;
      bit(1)            starts_with_SAP;
      unsigned int(3)   SAP_type;
      unsigned int(28)  SAP_delta_time;
   }
}
*/

type SidxReference struct {
	ReferenceType      uint8
	ReferencedSize     uint32
	SubsegmentDuration uint32
	StartsWithSAP      uint8
	SAPType            uint8
	SAPDeltaTime       uint32
}

type SidxBox struct {
	*FullBox
	ReferenceID              uint32
	Timescale                uint32
	EarliestPresentationTime uint64
	FirstOffset              uint64
	Reserved                 uint16
	ReferenceCount           uint16
	Refs                     []*SidxReference
}

func NewSidxBox(b *Box) *SidxBox {
	return &SidxBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func ParseSidxBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewSidxBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func (b *SidxBox) Parse(r io.Reader) (totalReadLen int, err error) {

	curReadLen := 0
	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 28)

	expReadLen := 28
	if b.version == 0 {
		expReadLen = 20
	}
	if curReadLen, err = io.ReadFull(r, buf[0:expReadLen]); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.ReferenceID = byteio.U32BE(buf[0:4])
	b.Timescale = byteio.U32BE(buf[4:8])
	if b.version == 0 {
		b.EarliestPresentationTime = uint64(byteio.U32BE(buf[8:12]))
		b.FirstOffset = uint64(byteio.U32BE(buf[12:16]))
		b.Reserved = byteio.U16BE(buf[16:18])
		b.ReferenceCount = byteio.U16BE(buf[18:20])
	} else {
		b.EarliestPresentationTime = byteio.U64BE(buf[8:16])
		b.FirstOffset = byteio.U64BE(buf[16:24])
		b.Reserved = byteio.U16BE(buf[24:26])
		b.ReferenceCount = byteio.U16BE(buf[26:28])
	}

	for i := 0; i < int(b.ReferenceCount); i++ {
		if curReadLen, err = io.ReadFull(r, buf[0:12]); err != nil {
			return
		}
		ref := &SidxReference{
			ReferenceType:      buf[0] >> 7,
			ReferencedSize:     byteio.U32BE(buf[0:4]) & 0x7FFFFFFF,
			SubsegmentDuration: byteio.U32BE(buf[4:8]),
			StartsWithSAP:      buf[8] >> 7,
			SAPType:            (buf[8] & 0x70) >> 4,
			SAPDeltaTime:       byteio.U32BE(buf[8:12]) & 0xFFFFFFF,
		}
		b.Refs = append(b.Refs, ref)
		totalReadLen += curReadLen
	}

	return
}

func (b *SidxBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	nums := []uint32{
		b.ReferenceID,
		b.Timescale,
		uint32(b.EarliestPresentationTime),
		uint32(b.FirstOffset),
	}

	if b.version == 0 {
		if curWriteLen, err = uint32Serialize(w, nums); err != nil {
			return
		}
		writedLen += curWriteLen
	} else {
		if curWriteLen, err = uint32Serialize(w, nums[0:2]); err != nil {
			return
		}
		writedLen += curWriteLen

		if curWriteLen, err = byteio.WriteU64BE(w, b.EarliestPresentationTime); err != nil {
			return
		}
		writedLen += curWriteLen
		if curWriteLen, err = byteio.WriteU64BE(w, b.FirstOffset); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	if curWriteLen, err = byteio.WriteU16BE(w, b.Reserved); err != nil {
		return
	}
	writedLen += curWriteLen

	if curWriteLen, err = byteio.WriteU16BE(w, b.ReferenceCount); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(b.Refs); i++ {
		tmp := (uint32(b.Refs[i].ReferenceType) << 31) | b.Refs[i].ReferencedSize
		if curWriteLen, err = byteio.WriteU32BE(w, tmp); err != nil {
			return
		}
		writedLen += curWriteLen

		if curWriteLen, err = byteio.WriteU32BE(w, b.Refs[i].SubsegmentDuration); err != nil {
			return
		}
		writedLen += curWriteLen

		tmp = uint32(b.Refs[i].StartsWithSAP)<<31 | uint32(b.Refs[i].SAPType)<<28 | b.Refs[i].SAPDeltaTime
		if curWriteLen, err = byteio.WriteU32BE(w, tmp); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}
