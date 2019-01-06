package mp4

import (
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
aligned(8) class MovieBox extends Box(‘moov’){
}
*/
type MoovBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class MovieHeaderBox extends FullBox(‘mvhd’, version, 0) {
   if (version==1) {
      unsigned int(64)  creation_time;
      unsigned int(64)  modification_time;
      unsigned int(32)  timescale;
      unsigned int(64)  duration;
   } else { // version==0
      unsigned int(32)  creation_time;
      unsigned int(32)  modification_time;
      unsigned int(32)  timescale;
      unsigned int(32)  duration;
   }
   template int(32)  rate = 0x00010000; // typically 1.0
   template int(16)  volume = 0x0100;   // typically, full volume
22 or 34
   const bit(16)  reserved = 0;
   const unsigned int(32)[2]  reserved = 0;
10
   template int(32)[9]  matrix =
      { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
      // Unity matrix
36
   bit(32)[6]  pre_defined = 0;
24
   unsigned int(32)  next_track_ID;
4
}
fix 96 byte if version == 0
fix 108 byte if version == 1
*/
type MvhdBox struct {
	*FullBox
	CreationTime     uint64
	ModificationTime uint64
	Timescale        uint32
	Duration         uint64
	TemplateRate     int32
	TemplateVolume   int16
	Reserved1        [2]byte
	Reserved2        [2]uint32
	TemplateMatrix   [9]int32
	PreDefined       [6][4]byte

	NextTrackID uint32
}

/*
aligned(8) class TrackBox extends Box(‘trak’) {
}
*/
type TrakBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class TrackHeaderBox
   extends FullBox(‘tkhd’, version, flags){
   if (version==1) {
      unsigned int(64)  creation_time;
      unsigned int(64)  modification_time;
      unsigned int(32)  track_ID;
      const unsigned int(32)  reserved = 0;
      unsigned int(64)  duration;
   } else { // version==0
      unsigned int(32)  creation_time;
      unsigned int(32)  modification_time;
      unsigned int(32)  track_ID;
      const unsigned int(32)  reserved = 0;
      unsigned int(32)  duration;
   }
   const unsigned int(32)[2]  reserved = 0;
   template int(16) layer = 0;
   template int(16) alternate_group = 0;
   template int(16)  volume = {if track_is_audio 0x0100 else 0};
   const unsigned int(16)  reserved = 0;
   template int(32)[9]  matrix=
      { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
      // unity matrix
   unsigned int(32) width;
   unsigned int(32) height;
}
fix 80 byte if version == 0
fix 92 byte if version == 1
*/
type TkhdBox struct {
	*FullBox
	CreationTime           uint64
	ModificationTime       uint64
	TrackID                uint32
	Reserved1              uint32
	Duration               uint64
	Reserved2              [2]uint32
	TemplateLayer          int16
	TemplatealTernateGroup int16
	TemplateVolume         int16
	Reserved3              int16
	TemplateMatrix         [9]int32
	Width                  uint32
	Height                 uint32
}

/*
aligned(8) class EditBox extends Box(‘edts’) {
}
*/
type EdtsBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class EditListBox extends FullBox(‘elst’, version, 0) {
   unsigned int(32)  entry_count;
   for (i=1; i <= entry_count; i++) {
     if (version==1) {
        unsigned int(64) segment_duration;
        int(64) media_time;
     } else { // version==0
        unsigned int(32) segment_duration;
        int(32)  media_time;
     }
     int(16) media_rate_integer;
     int(16) media_rate_fraction = 0;
    }
}
*/
type ElstEntry struct {
	SegmentDuration   uint64
	MediaFrame        uint64
	MediaRateInteger  int16
	MediaRateFraction int16
}

type ElstBox struct {
	*FullBox
	EntryCount uint32
	Entries    []*ElstEntry
}

/*
aligned(8) class MediaBox extends Box(‘mdia’) {
}
*/
type MdiaBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class MediaHeaderBox extends FullBox(‘mdhd’, version, 0) {
   if (version==1) {
      unsigned int(64)  creation_time;
      unsigned int(64)  modification_time;
      unsigned int(32)  timescale;
      unsigned int(64)  duration;
   } else { // version==0
      unsigned int(32)  creation_time;
      unsigned int(32)  modification_time;
      unsigned int(32)  timescale;
      unsigned int(32)  duration;
   }
   bit(1)   pad = 0;
   unsigned int(5)[3]   language;   // ISO-639-2/T language code
   unsigned int(16)  pre_defined = 0;
}
*/
type MdhdBox struct {
	*FullBox
	CreationTime     uint64
	ModificationTime uint64
	Timescale        uint32
	Duration         uint64
	Pad              uint8
	Language         [3]int8
	PreDefined       uint16
}

/*
aligned(8) class HandlerBox extends FullBox(‘hdlr’, version = 0, 0) {
   unsigned int(32)  pre_defined = 0;
   unsigned int(32)  handler_type;
   const unsigned int(32)[3]  reserved = 0;
   string   name;
}
*/
type HdlrBox struct {
	*FullBox
	PreDefined  uint32
	handlerType uint32
	Reserved    [3]uint32
	Name        []byte
}

/*
aligned(8) class MediaInformationBox extends Box(‘minf’) {
}
*/
type MinfBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class VideoMediaHeaderBox
   extends FullBox(‘vmhd’, version = 0, 1) {
   template unsigned int(16)  graphicsmode = 0;   // copy, see below
   template unsigned int(16)[3]  opcolor = {0, 0, 0};
}
*/
type VmhdBox struct {
	*FullBox
	TemplateGraphicsMode uint16
	TemplateOpcolor      [3]uint16
}

/*
aligned(8) class DataInformationBox extends Box(‘dinf’) {
}
*/
type DinfBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class DataEntryUrlBox (bit(24) flags)
   extends FullBox(‘url ’, version = 0, flags) {
   string   location;
}
aligned(8) class DataEntryUrnBox (bit(24) flags)
   extends FullBox(‘urn ’, version = 0, flags) {
   string   name;
   string   location;
}
aligned(8) class DataReferenceBox
   extends FullBox(‘dref’, version = 0, 0) {
   unsigned int(32)  entry_count;
   for (i=1; i <= entry_count; i++) {
	 DataEntryBox(entry_version, entry_flags) data_entry;
   }
}
*/
type UrnBox struct {
	*FullBox
	Name     []byte
	Location []byte
}
type UrlBox struct {
	*FullBox
	Location []byte
}
type DrefBox struct {
	*FullBox
	EntryCount uint32
	SubBoxes   []IBox
}

/*
aligned(8) class SampleTableBox extends Box(‘stbl’) {
}
*/
type StblBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class SampleDescriptionBox (unsigned int(32) handler_type)
   extends FullBox('stsd', version, 0){
   int i ;
   unsigned int(32) entry_count;
   for (i = 1 ; i <= entry_count ; i++){
      SampleEntry();    // an instance of a class derived from SampleEntry
  }
}
*/
type StsdBox struct {
	*FullBox
	EntryCount uint32
	SubBoxes   []IBox //entry还是一些box，只不过是定义在其它标准里面比如14496-15里面
}

/*
class PixelAspectRatioBox extends Box(‘pasp’){
   unsigned int(32) hSpacing;
   unsigned int(32) vSpacing;
}
*/
type PaspBox struct {
	*Box
	HSpacing uint32
	VSpacing uint32
}

/*
aligned(8) abstract class SampleEntry (unsigned int(32) format)
   extends Box(format){
   const unsigned int(8)[6] reserved = 0;
   unsigned int(16) data_reference_index;
}

//https://github.com/copiousfreetime/mp4parser/blob/master/isoparser/src/main/java/com/coremedia/iso/boxes/sampleentry/VisualSampleEntry.java
 VisualSampleEntry定义并没有在14496-15中找到
class VisualSampleEntry(codingname) extends SampleEntry (codingname){
  unsigned int(16) pre_defined = 0;
  const unsigned int(16) reserved = 0;
  unsigned int(32)[3] pre_defined = 0;
  unsigned int(16) width;
  unsigned int(16) height;
  template unsigned int(32) horizresolution = 0x00480000; // 72 dpi
  template unsigned int(32) vertresolution = 0x00480000; // 72 dpi
  const unsigned int(32) reserved = 0;
  template unsigned int(16) frame_count = 1;
  string[32] compressorname;
  template unsigned int(16) depth = 0x0018;
  int(16) pre_defined = -1;
}

class AVCConfigurationBox extends Box(‘avcC’) {
  AVCDecoderConfigurationRecord() AVCConfig;
}

class AVCConfigurationBox extends Box(‘avcC’) {
	AVCDecoderConfigurationRecord() AVCConfig;
}

class AVCSampleEntry() extends VisualSampleEntry(type) {// type is ‘avc1’ or 'avc3'
  AVCConfigurationBox	config;
  MPEG4BitRateBox (); 					// optional
  MPEG4ExtensionDescriptorsBox ();	// optional
  extra_boxes				boxes;				// optional
}
*/
type SampleEntry struct {
	Reserved           [6]uint8
	DataReferenceIndex uint16
}
type VisualSampleEntry struct {
	PreDefined1             uint16
	Reserved1               uint16
	PreDefined2             [3]uint32
	Width                   uint16
	Height                  uint16
	TemplateHorizResolution uint32 //0x00480000 72 dpi
	TemplateVertResolution  uint32 //0x00480000 72 dpi
	Reserved3               uint32
	TemplateFrameCount      uint16 // =1?
	CompressorName          [32]byte
	TemplateDepth           uint16 // 0x0018
	PreDefined3             int16
}
type AVCConfigurationBox struct {
	*Box
	AVCDecoderConfigurationRecord
}
type AVCSampleEntry struct {
	SampleEntry
	VisualSampleEntry
	AVCConfigurationBox
}
type Avc1Box struct {
	*Box
	AVCEntry AVCSampleEntry
	SubBoxes []IBox
}

type HVCCConfigurationBox struct {
	*Box
	HevcDecoderConfigurationRecord
}
type HevcSampleEntry struct {
	SampleEntry
	VisualSampleEntry
	AVCConfigurationBox
}
type Hev1Box struct {
	*Box
	AVCEntry AVCSampleEntry
	SubBoxes []IBox
}

/*
aligned(8) class TimeToSampleBox

   extends FullBox(’stts’, version = 0, 0) {
   unsigned int(32)  entry_count;
      int i;
   for (i=0; i < entry_count; i++) {
      unsigned int(32)  sample_count;
      unsigned int(32)  sample_delta;
   }
}
*/
type SttsEntry struct {
	SampleCount uint32
	SampleDelta uint32
}
type SttsBox struct {
	*FullBox
	EntryCount uint32
	Entries    []*SttsEntry
}

/*
aligned(8) class SampleToChunkBox
   extends FullBox(‘stsc’, version = 0, 0) {
   unsigned int(32)  entry_count;
   for (i=1; i <= entry_count; i++) {
	   unsigned int(32)  first_chunk;
       unsigned int(32)  samples_per_chunk;
       unsigned int(32)  sample_description_index;
    }
}
*/
type StscEntry struct {
	FirstChunk             uint32
	SamplePerChunk         uint32
	SampleDescriptionIndex uint32
}
type StscBox struct {
	*FullBox
	EntryCount uint32
	Entries    []*StscEntry
}

/*
aligned(8) class SampleSizeBox extends FullBox(‘stsz’, version = 0, 0) {
   unsigned int(32)  sample_size;
   unsigned int(32)  sample_count;
   if (sample_size==0) {
      for (i=1; i <= sample_count; i++) {
        unsigned int(32)  entry_size;
      }
    }
}
*/
type StszBox struct {
	*FullBox
	SampleSize  uint32
	SampleCount uint32
	EnriesSize  []uint32
}

/*
aligned(8) class ChunkOffsetBox
   extends FullBox(‘stco’, version = 0, 0) {
   unsigned int(32)  entry_count;
   for (i=1; i <= entry_count; i++) {
      unsigned int(32)  chunk_offset;
   }
}
*/
type StcoBox struct {
	*FullBox
	EntryCount  uint32
	ChunkOffset []uint32
}

/*
aligned(8) class MovieExtendsBox extends Box(‘mvex’){
}
*/
type MvexBox struct {
	*Box
	SubBoxes []IBox
}

/*
aligned(8) class TrackExtendsBox extends FullBox(‘trex’, 0, 0){
   unsigned int(32)  track_ID;
   unsigned int(32)  default_sample_description_index;
   unsigned int(32)  default_sample_duration;
   unsigned int(32)  default_sample_size;
   unsigned int(32)  default_sample_flags;
}
*/
type TrexBox struct {
	*FullBox
	TrackID                       uint32
	DefaultSampleDescriptionIndex uint32
	DefaultSampleDuration         uint32
	DefaultSampleSize             uint32
	DefaultSampleFlags            uint32
}

/*
aligned(8) class UserDataBox extends Box(‘udta’) {
}
*/
type UdtaBox struct {
	*Box
	SubBoxes []IBox
}

func NewMoovBox(b *Box) *MoovBox {
	return &MoovBox{
		Box: b,
	}
}

func (b *MoovBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *MoovBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in moov box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeMVHD:
			mvhdBox := NewMvhdBox(bb)
			if curReadLen, err = mvhdBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, mvhdBox)
		case BoxTypeTRAK:
			trakBox := NewTrakBox(bb)
			if curReadLen, err = trakBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, trakBox)
		case BoxTypeMVEX:
			mvexBox := NewMvexBox(bb)
			if curReadLen, err = mvexBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, mvexBox)

		case BoxTypeUDTA:
			udtaBox := NewUdtaBox(bb)
			if curReadLen, err = udtaBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, udtaBox)

		default:
			unsprtBox := NewUnsupporttedBox(bb)
			if curReadLen, err = unsprtBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, unsprtBox)
		}
		if curReadLen > 0 {
			totalReadLen += curReadLen
			remainSize -= curReadLen
		}

	}
	return
}

func NewMvhdBox(b *Box) *MvhdBox {
	return &MvhdBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *MvhdBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	bufLen := 96
	if b.version == 1 {
		bufLen += 12
	}
	buf := make([]byte, bufLen)

	curWriteLen := 0

	offset := 16
	if b.version == 1 {
		byteio.PutU64BE(buf[0:8], b.CreationTime)
		byteio.PutU64BE(buf[8:16], b.ModificationTime)
		byteio.PutU32BE(buf[16:20], b.Timescale)
		byteio.PutU64BE(buf[20:28], b.Duration)
		offset += 12
	} else {
		byteio.PutU32BE(buf[0:4], uint32(b.CreationTime))
		byteio.PutU32BE(buf[4:8], uint32(b.ModificationTime))
		byteio.PutU32BE(buf[8:12], b.Timescale)
		byteio.PutU32BE(buf[12:16], uint32(b.Duration))
	}
	byteio.PutU32BE(buf[offset:offset+4], uint32(b.TemplateRate))
	offset += 4
	byteio.PutU16BE(buf[offset:offset+2], uint16(b.TemplateVolume))
	offset += 2

	buf[offset] = b.Reserved1[0]
	offset++
	buf[offset] = b.Reserved1[1]
	offset++

	byteio.PutU32BE(buf[offset:offset+4], b.Reserved2[0])
	offset += 4
	byteio.PutU32BE(buf[offset:offset+4], b.Reserved2[1])
	offset += 4

	for i := 0; i < 9; i++ {
		byteio.PutU32BE(buf[offset:offset+4], uint32(b.TemplateMatrix[i]))
		offset += 4
	}

	for i := 0; i < 6; i++ {
		copy(buf[offset:offset+4], b.PreDefined[i][0:4])
		offset += 4
	}

	byteio.PutU32BE(buf[offset:offset+4], b.NextTrackID)
	offset += 4

	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *MvhdBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in mvhd box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	bufLen := 96
	if b.version == 1 {
		bufLen += 12
	}
	buf := make([]byte, bufLen)

	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	offset := 16
	if b.version == 1 {
		b.CreationTime = byteio.U64BE(buf[0:8])
		b.ModificationTime = byteio.U64BE(buf[8:16])
		b.Timescale = byteio.U32BE(buf[16:20])
		b.Duration = byteio.U64BE(buf[20:28])
		offset += 12
	} else {
		b.CreationTime = uint64(byteio.U32BE(buf[0:4]))
		b.ModificationTime = uint64(byteio.U32BE(buf[4:8]))
		b.Timescale = byteio.U32BE(buf[8:12])
		b.Duration = uint64(byteio.U32BE(buf[12:16]))
	}
	b.TemplateRate = byteio.I32BE(buf[offset : offset+4])
	offset += 4
	b.TemplateVolume = byteio.I16BE(buf[offset : offset+2])
	offset += 2

	b.Reserved1[0] = buf[offset]
	offset++
	b.Reserved1[1] = buf[offset]
	offset++

	b.Reserved2[0] = byteio.U32BE(buf[offset : offset+4])
	offset += 4
	b.Reserved2[1] = byteio.U32BE(buf[offset : offset+4])
	offset += 4

	for i := 0; i < 9; i++ {
		b.TemplateMatrix[i] = byteio.I32BE(buf[offset : offset+4])
		offset += 4
	}

	for i := 0; i < 6; i++ {
		copy(b.PreDefined[i][0:4], buf[offset:offset+4])
		offset += 4
	}
	b.NextTrackID = byteio.U32BE(buf[offset : offset+4])

	return
}

func NewTrakBox(b *Box) *TrakBox {
	return &TrakBox{
		Box: b,
	}
}

func (b *TrakBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *TrakBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in trak box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeTKHD:
			tkhdBox := NewTkhdBox(bb)
			if curReadLen, err = tkhdBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, tkhdBox)
		case BoxTypeEDTS:
			edtsBox := NewEdtsBox(bb)
			if curReadLen, err = edtsBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, edtsBox)
		case BoxTypeMDIA:
			mdiaBox := NewMdiaBox(bb)
			if curReadLen, err = mdiaBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, mdiaBox)
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

func NewTkhdBox(b *Box) *TkhdBox {
	return &TkhdBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *TkhdBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	bufLen := 80
	if b.version == 1 {
		bufLen += 12
	}
	buf := make([]byte, bufLen)

	offset := 20
	if b.version == 1 {
		byteio.PutU64BE(buf[0:8], b.CreationTime)
		byteio.PutU64BE(buf[8:16], b.ModificationTime)
		byteio.PutU32BE(buf[16:20], b.TrackID)
		byteio.PutU32BE(buf[20:24], b.Reserved1)
		byteio.PutU64BE(buf[24:32], b.Duration)
		offset += 12
	} else {
		byteio.PutU32BE(buf[0:4], uint32(b.CreationTime))
		byteio.PutU32BE(buf[4:8], uint32(b.ModificationTime))
		byteio.PutU32BE(buf[8:12], b.TrackID)
		byteio.PutU32BE(buf[12:16], b.Reserved1)
		byteio.PutU32BE(buf[16:20], uint32(b.Duration))
	}

	byteio.PutU32BE(buf[offset:offset+4], b.Reserved2[0])
	offset += 4
	byteio.PutU32BE(buf[offset:offset+4], b.Reserved2[1])
	offset += 4

	byteio.PutU16BE(buf[offset:offset+2], uint16(b.TemplateLayer))
	offset += 2
	byteio.PutU16BE(buf[offset:offset+2], uint16(b.TemplatealTernateGroup))
	offset += 2
	byteio.PutU16BE(buf[offset:offset+2], uint16(b.TemplateVolume))
	offset += 2
	byteio.PutU16BE(buf[offset:offset+2], uint16(b.Reserved3))
	offset += 2

	for i := 0; i < 9; i++ {
		byteio.PutU32BE(buf[offset:offset+4], uint32(b.TemplateMatrix[i]))
		offset += 4
	}
	byteio.PutU32BE(buf[offset:offset+4], b.Width)
	offset += 4
	byteio.PutU32BE(buf[offset:offset+4], b.Height)
	offset += 4

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *TkhdBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in tkhd box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, FULLBOX_ANY_FLAG); err != nil {
		return
	}

	bufLen := 80
	if b.version == 1 {
		bufLen += 12
	}
	buf := make([]byte, bufLen)

	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	offset := 20
	if b.version == 1 {
		b.CreationTime = byteio.U64BE(buf[0:8])
		b.ModificationTime = byteio.U64BE(buf[8:16])
		b.TrackID = byteio.U32BE(buf[16:20])
		b.Reserved1 = byteio.U32BE(buf[20:24])
		b.Duration = byteio.U64BE(buf[24:32])
		offset += 12
	} else {
		b.CreationTime = uint64(byteio.U32BE(buf[0:4]))
		b.ModificationTime = uint64(byteio.U32BE(buf[4:8]))
		b.TrackID = byteio.U32BE(buf[8:12])
		b.Reserved1 = byteio.U32BE(buf[12:16])
		b.Duration = uint64(byteio.U32BE(buf[16:20]))
	}

	b.Reserved2[0] = byteio.U32BE(buf[offset : offset+4])
	offset += 4
	b.Reserved2[1] = byteio.U32BE(buf[offset : offset+4])
	offset += 4

	b.TemplateLayer = byteio.I16BE(buf[offset : offset+2])
	offset += 2
	b.TemplatealTernateGroup = byteio.I16BE(buf[offset : offset+2])
	offset += 2
	b.TemplateVolume = byteio.I16BE(buf[offset : offset+2])
	offset += 2
	b.Reserved3 = byteio.I16BE(buf[offset : offset+2])
	offset += 2

	for i := 0; i < 9; i++ {
		b.TemplateMatrix[i] = byteio.I32BE(buf[offset : offset+4])
		offset += 4
	}
	b.Width = byteio.U32BE(buf[offset : offset+4])
	offset += 4
	b.Height = byteio.U32BE(buf[offset : offset+4])
	offset += 4

	return
}

func NewEdtsBox(b *Box) *EdtsBox {
	return &EdtsBox{
		Box: b,
	}
}

func (b *EdtsBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *EdtsBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in edts box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeELST:
			elstBox := NewElstBox(bb)
			if curReadLen, err = elstBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, elstBox)

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

func NewElstBox(b *Box) *ElstBox {
	return &ElstBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *ElstBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	if b.EntryCount != uint32(len(b.Entries)) {
		err = fmt.Errorf("elst not consistent:%d %d", b.EntryCount, uint32(len(b.Entries)))
		return
	}

	curWriteLen := 0
	if curWriteLen, err = byteio.WriteU32BE(w, b.EntryCount); err != nil {
		return
	}
	writedLen += curWriteLen

	buf := make([]byte, 20)
	if b.version == 0 {
		buf = buf[0:12]
	}
	for i := uint32(0); i < b.EntryCount; i++ {

		if b.version == 1 {
			byteio.PutU64BE(buf[0:8], b.Entries[i].SegmentDuration)
			byteio.PutU64BE(buf[8:16], b.Entries[i].MediaFrame)
			byteio.PutU16BE(buf[16:18], uint16(b.Entries[i].MediaRateInteger))
			byteio.PutU16BE(buf[18:20], uint16(b.Entries[i].MediaRateFraction))
		} else {
			byteio.PutU32BE(buf[0:4], uint32(b.Entries[i].SegmentDuration))
			byteio.PutU32BE(buf[4:8], uint32(b.Entries[i].MediaFrame))
			byteio.PutU16BE(buf[8:10], uint16(b.Entries[i].MediaRateInteger))
			byteio.PutU16BE(buf[10:12], uint16(b.Entries[i].MediaRateFraction))
		}

		if curWriteLen, err = w.Write(buf); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *ElstBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in elst box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 20)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.EntryCount = byteio.U32BE(buf)

	if b.version == 0 {
		buf = buf[0:12]
	}
	for i := uint32(0); i < b.EntryCount; i++ {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		totalReadLen += curReadLen

		entry := &ElstEntry{}
		if b.version == 1 {
			entry.SegmentDuration = byteio.U64BE(buf[0:8])
			entry.MediaFrame = byteio.U64BE(buf[8:16])
			entry.MediaRateInteger = byteio.I16BE(buf[16:18])
			entry.MediaRateFraction = byteio.I16BE(buf[18:20])
		} else {
			entry.SegmentDuration = uint64(byteio.U32BE(buf[0:4]))
			entry.MediaFrame = uint64(byteio.U32BE(buf[4:8]))
			entry.MediaRateInteger = byteio.I16BE(buf[8:10])
			entry.MediaRateFraction = byteio.I16BE(buf[10:12])
		}
		b.Entries = append(b.Entries, entry)
	}

	return
}

func NewMdiaBox(b *Box) *MdiaBox {
	return &MdiaBox{
		Box: b,
	}
}

func (b *MdiaBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *MdiaBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in mdia box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeMDHD:
			mdhdBox := NewMdhdBox(bb)
			if curReadLen, err = mdhdBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, mdhdBox)

		case BoxTypeHDLR:
			hdlrBox := NewHdlrBox(bb)
			if curReadLen, err = hdlrBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, hdlrBox)

		case BoxTypeMINF:
			minfBox := NewMinfBox(bb)
			if curReadLen, err = minfBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, minfBox)

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

func NewMdhdBox(b *Box) *MdhdBox {
	return &MdhdBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *MdhdBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 32)
	if b.version == 0 {
		buf = buf[0:20]
	}

	offset := 16
	if b.version == 1 {
		byteio.PutU64BE(buf[0:8], b.CreationTime)
		byteio.PutU64BE(buf[8:16], b.ModificationTime)
		byteio.PutU32BE(buf[16:20], b.Timescale)
		byteio.PutU64BE(buf[20:28], b.Duration)
		offset += 12
	} else {
		byteio.PutU32BE(buf[0:4], uint32(b.CreationTime))
		byteio.PutU32BE(buf[4:8], uint32(b.ModificationTime))
		byteio.PutU32BE(buf[8:12], b.Timescale)
		byteio.PutU32BE(buf[12:16], uint32(b.Duration))
	}

	pl := uint16(b.Pad)<<15 | uint16(b.Language[0])<<10 | uint16(b.Language[1])<<5 | uint16(b.Language[2])
	byteio.PutU16BE(buf[offset:offset+2], pl)
	offset += 2

	byteio.PutU16BE(buf[offset:offset+2], b.PreDefined)

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *MdhdBox) Parse(r io.Reader) (totalReadLen int, err error) {
	if b.Size == 1 {
		err = fmt.Errorf("large size in mdhd box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 32)
	if b.version == 0 {
		buf = buf[0:20]
	}

	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	offset := 16
	if b.version == 1 {
		b.CreationTime = byteio.U64BE(buf[0:8])
		b.ModificationTime = byteio.U64BE(buf[8:16])
		b.Timescale = byteio.U32BE(buf[16:20])
		b.Duration = byteio.U64BE(buf[20:28])
		offset += 12
	} else {
		b.CreationTime = uint64(byteio.U32BE(buf[0:4]))
		b.ModificationTime = uint64(byteio.U32BE(buf[4:8]))
		b.Timescale = byteio.U32BE(buf[8:12])
		b.Duration = uint64(byteio.U32BE(buf[12:16]))
	}

	b.Pad = buf[offset] >> 7
	b.Language[0] = int8((buf[offset] & 0x7c) >> 2)
	b.Language[1] = int8((buf[offset]&0x03)<<3 | (buf[offset+1]&0xE0)>>5)
	offset++
	b.Language[2] = int8(buf[offset] & 0x1F)
	offset++

	b.PreDefined = byteio.U16BE(buf[offset : offset+2])

	return
}

func NewHdlrBox(b *Box) *HdlrBox {

	return &HdlrBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *HdlrBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}
	curWriteLen := 0
	nums := []uint32{b.PreDefined, b.handlerType, b.Reserved[0], b.Reserved[1], b.Reserved[2]}
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen

	if curWriteLen, err = w.Write(b.Name); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *HdlrBox) Parse(r io.Reader) (totalReadLen int, err error) {
	if b.Size == 1 {
		err = fmt.Errorf("large size in hdlr box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 20)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.PreDefined = byteio.U32BE(buf[0:4])
	b.handlerType = byteio.U32BE(buf[4:8])
	for i := 0; i < 3; i++ {
		b.Reserved[i] = byteio.U32BE(buf[i*4+8 : i*4+12])
	}

	nameLen := int(b.Size) - totalReadLen - BOX_SIZE
	if nameLen <= 0 {
		fmt.Errorf("hdlrbox name:%d", nameLen)
	}

	b.Name = make([]byte, nameLen)
	if curReadLen, err = io.ReadFull(r, b.Name); err != nil {
		return
	}
	totalReadLen += curReadLen

	return
}

func NewMinfBox(b *Box) *MinfBox {
	return &MinfBox{
		Box: b,
	}
}

func (b *MinfBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *MinfBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in minf box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeVMHD:
			vmhdBox := NewVmhdBox(bb)
			if curReadLen, err = vmhdBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, vmhdBox)

		case BoxTypeDINF:
			dinfBox := NewDinfBox(bb)
			if curReadLen, err = dinfBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, dinfBox)

		case BoxTypeSTBL:
			stblBox := NewStblBox(bb)
			if curReadLen, err = stblBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, stblBox)

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

func NewVmhdBox(b *Box) *VmhdBox {
	return &VmhdBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *VmhdBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	buf := make([]byte, 8)
	byteio.PutU16BE(buf[0:2], b.TemplateGraphicsMode)
	for i := 0; i < 3; i++ {
		byteio.PutU16BE(buf[2*i+2:2*i+4], b.TemplateOpcolor[i])
	}

	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *VmhdBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in vmhd box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 1); err != nil {
		return
	}

	buf := make([]byte, 8)

	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.TemplateGraphicsMode = byteio.U16BE(buf[0:2])
	for i := 0; i < 3; i++ {
		b.TemplateOpcolor[i] = byteio.U16BE(buf[2*i+2 : 2*i+4])
	}

	return
}

func NewDinfBox(b *Box) *DinfBox {
	return &DinfBox{
		Box: b,
	}
}

func (b *DinfBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *DinfBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in dinf box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeDREF:
			drefBox := NewDrefBox(bb)
			if curReadLen, err = drefBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, drefBox)

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

func NewDrefBox(b *Box) *DrefBox {
	return &DrefBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *DrefBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	if b.EntryCount != uint32(len(b.SubBoxes)) {
		err = fmt.Errorf("dref not consistent:%d %d", b.EntryCount, len(b.SubBoxes))
		return
	}

	curWriteLen := 0

	if curWriteLen, err = byteio.WriteU32BE(w, b.EntryCount); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(b.SubBoxes); i++ {
		if curWriteLen, err = b.SubBoxes[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *DrefBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in dref box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 4)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.EntryCount = byteio.U32BE(buf)
	for i := uint32(0); i < b.EntryCount; i++ {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen

		switch bb.BoxType {
		case BoxTypeURL:
			urlBox := NewUrlBox(bb)
			if curReadLen, err = urlBox.Parse(r); err != nil {
				return
			}
			totalReadLen += curReadLen
			b.SubBoxes = append(b.SubBoxes, urlBox)
		case BoxTypeURN:
			urnBox := NewUrnBox(bb)
			if curReadLen, err = urnBox.Parse(r); err != nil {
				return
			}
			totalReadLen += curReadLen
			b.SubBoxes = append(b.SubBoxes, urnBox)
		}

	}
	return
}

func NewUrnBox(b *Box) *UrnBox {
	return &UrnBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *UrnBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if len(b.Name) > 0 {
		if curWriteLen, err = w.Write(b.Name); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	if len(b.Location) > 0 {
		if curWriteLen, err = w.Write(b.Location); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *UrnBox) Parse(r io.Reader) (totalReadLen int, err error) {
	if b.Size == 1 {
		err = fmt.Errorf("large size in urn box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, FULLBOX_ANY_FLAG); err != nil {
		return
	}

	if b.flags24Bit == 1 {
		return
	}

	// TODO 怎么区别name和locatoin的分隔(box的字符串应该都是以0结尾的)?
	remainSize := int(b.Size) - FULL_BOX_SIZE
	buf := make([]byte, remainSize)
	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}
	idx := 0
	for i := 0; i < len(buf) && i < remainSize; i++ {
		if buf[i] != 0 {
			idx = i
			break
		}
	}

	b.Name = buf[0 : idx+1]
	b.Location = buf[idx+1:]
	totalReadLen += remainSize

	return
}

func NewUrlBox(b *Box) *UrlBox {
	return &UrlBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *UrlBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if len(b.Location) > 0 {
		if curWriteLen, err = w.Write(b.Location); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *UrlBox) Parse(r io.Reader) (totalReadLen int, err error) {
	if b.Size == 1 {
		err = fmt.Errorf("large size in url box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, FULLBOX_ANY_FLAG); err != nil {
		return
	}

	if b.flags24Bit == 1 {
		return
	}

	locationLen := int(b.Size) - totalReadLen - BOX_SIZE
	if locationLen <= 0 {
		fmt.Errorf("urlbox location:%d", locationLen)
	}

	b.Location = make([]byte, locationLen)
	if totalReadLen, err = io.ReadFull(r, b.Location); err != nil {
		return
	}

	return
}

func NewStblBox(b *Box) *StblBox {
	return &StblBox{
		Box: b,
	}
}

func (b *StblBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *StblBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in stbl box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeSTSD:
			stsdBox := NewStsdBox(bb)
			if curReadLen, err = stsdBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, stsdBox)

		case BoxTypeSTTS:
			sttsBox := NewSttsBox(bb)
			if curReadLen, err = sttsBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, sttsBox)

		case BoxTypeSTSC:
			stscBox := NewStscBox(bb)
			if curReadLen, err = stscBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, stscBox)
		case BoxTypeSTSZ:
			fallthrough
		case BoxTypeSTZ2:
			stszBox := NewStszBox(bb)
			if curReadLen, err = stszBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, stszBox)

		case BoxTypeSTCO:
			stcoBox := NewStcoBox(bb)
			if curReadLen, err = stcoBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, stcoBox)

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

func NewStsdBox(b *Box) *StsdBox {
	return &StsdBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *StsdBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if curWriteLen, err = byteio.WriteU32BE(w, b.EntryCount); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(b.SubBoxes); i++ {
		if curWriteLen, err = b.SubBoxes[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *StsdBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in stsd box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}
	remainSize := int(b.Size) - FULL_BOX_SIZE

	buf := make([]byte, 4)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	remainSize -= curReadLen
	totalReadLen += curReadLen

	b.EntryCount = byteio.U32BE(buf)
	for remainSize > 0 {
		for i := uint32(0); i < b.EntryCount; i++ {
			var bb *Box
			if bb, curReadLen, err = ParseBox(r); err != nil {
				return
			}
			remainSize -= curReadLen
			totalReadLen += curReadLen

			switch bb.BoxType {
			case BoxTypeAVC1:
				avc1Box := NewAvc1Box(bb)
				if curReadLen, err = avc1Box.Parse(r); err != nil {
					return
				}
				b.SubBoxes = append(b.SubBoxes, avc1Box)

			case BoxTypeHEV1:
				hev1Box := NewHev1Box(bb)
				if curReadLen, err = hev1Box.Parse(r); err != nil {
					return
				}
				b.SubBoxes = append(b.SubBoxes, hev1Box)

			case BoxTypePASP:
				paspBox := NewPaspBox(bb)
				if curReadLen, err = paspBox.Parse(r); err != nil {
					return
				}
				b.SubBoxes = append(b.SubBoxes, paspBox)

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
	}

	return
}

func NewPaspBox(b *Box) *PaspBox {
	return &PaspBox{
		Box: b,
	}
}

func (b *PaspBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	nums := []uint32{
		b.HSpacing,
		b.VSpacing,
	}

	curWriteLen := 0
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *PaspBox) Parse(r io.Reader) (totalReadLen int, err error) {
	buf := make([]byte, 8)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	b.HSpacing = byteio.U32BE(buf[0:4])
	b.VSpacing = byteio.U32BE(buf[4:8])
	return
}

func NewAvc1Box(b *Box) *Avc1Box {
	return &Avc1Box{
		Box: b,
	}
}

func (b *Avc1Box) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 70) // 70 == VisualSampleEntry

	for i := 0; i < 6; i++ {
		buf[i] = b.AVCEntry.Reserved[i]
	}
	byteio.PutU16BE(buf[6:8], b.AVCEntry.DataReferenceIndex)

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf[0:8]); err != nil {
		return
	}
	writedLen += curWriteLen

	byteio.PutU16BE(buf[0:2], b.AVCEntry.PreDefined1)
	byteio.PutU16BE(buf[2:4], b.AVCEntry.Reserved1)
	byteio.PutU32BE(buf[4:8], b.AVCEntry.PreDefined2[0])
	byteio.PutU32BE(buf[8:12], b.AVCEntry.PreDefined2[1])
	byteio.PutU32BE(buf[12:16], b.AVCEntry.PreDefined2[2])
	byteio.PutU16BE(buf[16:18], b.AVCEntry.Width)
	byteio.PutU16BE(buf[18:20], b.AVCEntry.Height)
	byteio.PutU32BE(buf[20:24], b.AVCEntry.TemplateHorizResolution)
	byteio.PutU32BE(buf[24:28], b.AVCEntry.TemplateVertResolution)
	byteio.PutU32BE(buf[28:32], b.AVCEntry.Reserved3)
	byteio.PutU16BE(buf[32:34], b.AVCEntry.TemplateFrameCount)
	copy(buf[34:66], b.AVCEntry.CompressorName[0:32])
	byteio.PutU16BE(buf[66:68], b.AVCEntry.TemplateDepth)
	byteio.PutU16BE(buf[68:70], uint16(b.AVCEntry.PreDefined3))

	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(b.SubBoxes); i++ {
		if curWriteLen, err = b.SubBoxes[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *Avc1Box) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 70) // 70 == VisualSampleEntry
	if totalReadLen, err = io.ReadFull(r, buf[0:8]); err != nil {
		return
	}

	for i := 0; i < 6; i++ {
		b.AVCEntry.Reserved[i] = buf[i]
	}
	b.AVCEntry.DataReferenceIndex = byteio.U16BE(buf[6:8])

	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.AVCEntry.PreDefined1 = byteio.U16BE(buf[0:2])
	b.AVCEntry.Reserved1 = byteio.U16BE(buf[2:4])
	b.AVCEntry.PreDefined2[0] = byteio.U32BE(buf[4:8])
	b.AVCEntry.PreDefined2[1] = byteio.U32BE(buf[8:12])
	b.AVCEntry.PreDefined2[2] = byteio.U32BE(buf[12:16])
	b.AVCEntry.Width = byteio.U16BE(buf[16:18])
	b.AVCEntry.Height = byteio.U16BE(buf[18:20])
	b.AVCEntry.TemplateHorizResolution = byteio.U32BE(buf[20:24])
	b.AVCEntry.TemplateVertResolution = byteio.U32BE(buf[24:28])
	b.AVCEntry.Reserved3 = byteio.U32BE(buf[28:32])
	b.AVCEntry.TemplateFrameCount = byteio.U16BE(buf[32:34])
	copy(b.AVCEntry.CompressorName[0:32], buf[34:66])
	b.AVCEntry.TemplateDepth = byteio.U16BE(buf[66:68])
	b.AVCEntry.PreDefined3 = byteio.I16BE(buf[68:70])

	var bb *Box
	if bb, curReadLen, err = ParseBox(r); err != nil {
		return
	}
	totalReadLen += curReadLen

	switch bb.BoxType {
	case BoxTypeAVCC:
		avcCBox := NewAVCConfigurationBox(bb)
		if curReadLen, err = avcCBox.Parse(r); err != nil {
			return
		}
		b.SubBoxes = append(b.SubBoxes, avcCBox)

	default:
		unsprtBox := NewUnsupporttedBox(bb)
		if curReadLen, err = unsprtBox.Parse(r); err != nil {
			return
		}
		b.SubBoxes = append(b.SubBoxes, unsprtBox)
	}
	if curReadLen > 0 {
		totalReadLen += curReadLen
	}

	return
}

func NewHev1Box(b *Box) *Hev1Box {
	return &Hev1Box{
		Box: b,
	}
}

func (b *Hev1Box) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 70) // 70 == VisualSampleEntry

	for i := 0; i < 6; i++ {
		buf[i] = b.AVCEntry.Reserved[i]
	}
	byteio.PutU16BE(buf[6:8], b.AVCEntry.DataReferenceIndex)

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf[0:8]); err != nil {
		return
	}
	writedLen += curWriteLen

	byteio.PutU16BE(buf[0:2], b.AVCEntry.PreDefined1)
	byteio.PutU16BE(buf[2:4], b.AVCEntry.Reserved1)
	byteio.PutU32BE(buf[4:8], b.AVCEntry.PreDefined2[0])
	byteio.PutU32BE(buf[8:12], b.AVCEntry.PreDefined2[1])
	byteio.PutU32BE(buf[12:16], b.AVCEntry.PreDefined2[2])
	byteio.PutU16BE(buf[16:18], b.AVCEntry.Width)
	byteio.PutU16BE(buf[18:20], b.AVCEntry.Height)
	byteio.PutU32BE(buf[20:24], b.AVCEntry.TemplateHorizResolution)
	byteio.PutU32BE(buf[24:28], b.AVCEntry.TemplateVertResolution)
	byteio.PutU32BE(buf[28:32], b.AVCEntry.Reserved3)
	byteio.PutU16BE(buf[32:34], b.AVCEntry.TemplateFrameCount)
	copy(buf[34:66], b.AVCEntry.CompressorName[0:32])
	byteio.PutU16BE(buf[66:68], b.AVCEntry.TemplateDepth)
	byteio.PutU16BE(buf[68:70], uint16(b.AVCEntry.PreDefined3))

	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(b.SubBoxes); i++ {
		if curWriteLen, err = b.SubBoxes[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func (b *Hev1Box) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 70) // 70 == VisualSampleEntry
	if totalReadLen, err = io.ReadFull(r, buf[0:8]); err != nil {
		return
	}

	for i := 0; i < 6; i++ {
		b.AVCEntry.Reserved[i] = buf[i]
	}
	b.AVCEntry.DataReferenceIndex = byteio.U16BE(buf[6:8])

	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.AVCEntry.PreDefined1 = byteio.U16BE(buf[0:2])
	b.AVCEntry.Reserved1 = byteio.U16BE(buf[2:4])
	b.AVCEntry.PreDefined2[0] = byteio.U32BE(buf[4:8])
	b.AVCEntry.PreDefined2[1] = byteio.U32BE(buf[8:12])
	b.AVCEntry.PreDefined2[2] = byteio.U32BE(buf[12:16])
	b.AVCEntry.Width = byteio.U16BE(buf[16:18])
	b.AVCEntry.Height = byteio.U16BE(buf[18:20])
	b.AVCEntry.TemplateHorizResolution = byteio.U32BE(buf[20:24])
	b.AVCEntry.TemplateVertResolution = byteio.U32BE(buf[24:28])
	b.AVCEntry.Reserved3 = byteio.U32BE(buf[28:32])
	b.AVCEntry.TemplateFrameCount = byteio.U16BE(buf[32:34])
	copy(b.AVCEntry.CompressorName[0:32], buf[34:66])
	b.AVCEntry.TemplateDepth = byteio.U16BE(buf[66:68])
	b.AVCEntry.PreDefined3 = byteio.I16BE(buf[68:70])

	var bb *Box
	if bb, curReadLen, err = ParseBox(r); err != nil {
		return
	}
	totalReadLen += curReadLen

	switch bb.BoxType {
	case BoxTypeHVCC:
		hvcCBox := NewHVCCConfigurationBox(bb)
		if curReadLen, err = hvcCBox.Parse(r); err != nil {
			return
		}
		b.SubBoxes = append(b.SubBoxes, hvcCBox)

	default:
		unsprtBox := NewUnsupporttedBox(bb)
		if curReadLen, err = unsprtBox.Parse(r); err != nil {
			return
		}
		b.SubBoxes = append(b.SubBoxes, unsprtBox)
	}
	if curReadLen > 0 {
		totalReadLen += curReadLen
	}

	return
}

func NewAVCConfigurationBox(b *Box) *AVCConfigurationBox {
	return &AVCConfigurationBox{
		Box: b,
	}
}
func (b *AVCConfigurationBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if curWriteLen, err = b.AVCDecoderConfigurationRecord.Serialize(w); err != nil {
		return
	}
	writedLen += curWriteLen
	return
}

func (b *AVCConfigurationBox) Parse(r io.Reader) (totalReadLen int, err error) {
	totalReadLen, err = b.AVCDecoderConfigurationRecord.Parse(r)
	return
}

func NewHVCCConfigurationBox(b *Box) *HVCCConfigurationBox {
	return &HVCCConfigurationBox{
		Box: b,
	}
}
func (b *HVCCConfigurationBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.Box.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	if curWriteLen, err = b.HevcDecoderConfigurationRecord.Serialize(w); err != nil {
		return
	}
	writedLen += curWriteLen
	return
}

func (b *HVCCConfigurationBox) Parse(r io.Reader) (totalReadLen int, err error) {
	totalReadLen, err = b.HevcDecoderConfigurationRecord.Parse(r)
	return
}

func NewSttsBox(b *Box) *SttsBox {
	return &SttsBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *SttsBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	nums := []uint32{
		b.EntryCount,
	}

	for i := 0; i < len(b.Entries); i++ {
		nums = append(nums, b.Entries[i].SampleCount, b.Entries[i].SampleDelta)
	}

	curWriteLen := 0
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *SttsBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in stts box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}
	remainSize := int(b.Size) - FULL_BOX_SIZE

	buf := make([]byte, 8)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
		return
	}
	remainSize -= curReadLen
	totalReadLen += curReadLen

	b.EntryCount = byteio.U32BE(buf)

	for i := uint32(0); i < b.EntryCount && remainSize > 0; i++ {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		remainSize -= curReadLen
		totalReadLen += curReadLen

		entry := &SttsEntry{}
		entry.SampleCount = byteio.U32BE(buf[0:4])
		entry.SampleDelta = byteio.U32BE(buf[4:8])
		b.Entries = append(b.Entries, entry)
	}
	if remainSize > 0 {
		err = fmt.Errorf("sttsbox remainsize:%d", remainSize)
		return
	}

	return
}

func NewStscBox(b *Box) *StscBox {
	return &StscBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *StscBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	nums := []uint32{
		b.EntryCount,
	}

	for i := 0; i < len(b.Entries); i++ {
		nums = append(nums, b.Entries[i].FirstChunk, b.Entries[i].SamplePerChunk,
			b.Entries[i].SampleDescriptionIndex)
	}

	curWriteLen := 0
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *StscBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in stsc box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}
	remainSize := int(b.Size) - FULL_BOX_SIZE

	buf := make([]byte, 12)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
		return
	}
	remainSize -= curReadLen
	totalReadLen += curReadLen

	b.EntryCount = byteio.U32BE(buf)
	for i := uint32(0); i < b.EntryCount && remainSize > 0; i++ {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		remainSize -= curReadLen
		totalReadLen += curReadLen

		entry := &StscEntry{}
		entry.FirstChunk = byteio.U32BE(buf[0:4])
		entry.SamplePerChunk = byteio.U32BE(buf[4:8])
		entry.SampleDescriptionIndex = byteio.U32BE(buf[8:12])
		b.Entries = append(b.Entries, entry)
	}
	if remainSize > 0 {
		err = fmt.Errorf("stscbox remainsize:%d", remainSize)
		return
	}

	return
}

func NewStszBox(b *Box) *StszBox {
	return &StszBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *StszBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	nums := []uint32{
		b.SampleSize,
		b.SampleCount,
	}
	if b.SampleSize == 0 {
		nums = append(nums, b.EnriesSize...)
	}
	curWriteLen := 0
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *StszBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in stsz box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 8)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen
	b.SampleSize = byteio.U32BE(buf)
	b.SampleCount = byteio.U32BE(buf[4:8])
	if b.SampleSize == 0 {
		buf = buf[0:4]
		for i := uint32(0); i < b.SampleCount; i++ {
			if curReadLen, err = io.ReadFull(r, buf); err != nil {
				return
			}
			totalReadLen += curReadLen

			b.EnriesSize = append(b.EnriesSize, byteio.U32BE(buf))
		}
	}
	return
}

func NewStcoBox(b *Box) *StcoBox {
	return &StcoBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *StcoBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	nums := []uint32{
		b.EntryCount,
	}
	nums = append(nums, b.ChunkOffset...)
	curWriteLen := 0
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *StcoBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in stco box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}
	remainSize := int(b.Size) - FULL_BOX_SIZE

	buf := make([]byte, 4)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	remainSize -= curReadLen
	totalReadLen += curReadLen

	b.EntryCount = byteio.U32BE(buf)
	for i := uint32(0); i < b.EntryCount && remainSize > 0; i++ {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		remainSize -= curReadLen
		totalReadLen += curReadLen

		b.ChunkOffset = append(b.ChunkOffset, byteio.U32BE(buf))
	}
	if remainSize > 0 {
		err = fmt.Errorf("sctobox remainsize:%d", remainSize)
		return
	}

	return
}

func NewMvexBox(b *Box) *MvexBox {
	return &MvexBox{
		Box: b,
	}
}

func (b *MvexBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *MvexBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in mvex box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeTREX:
			trexBox := NewTrexBox(bb)
			if curReadLen, err = trexBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, trexBox)

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

func NewTrexBox(b *Box) *TrexBox {
	return &TrexBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *TrexBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	nums := []uint32{
		b.TrackID,
		b.DefaultSampleDescriptionIndex,
		b.DefaultSampleDuration,
		b.DefaultSampleSize,
		b.DefaultSampleFlags,
	}
	curWriteLen := 0
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func (b *TrexBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in trex box")
		return
	}

	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 20)

	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.TrackID = byteio.U32BE(buf[0:4])
	b.DefaultSampleDescriptionIndex = byteio.U32BE(buf[4:8])
	b.DefaultSampleDuration = byteio.U32BE(buf[8:12])
	b.DefaultSampleSize = byteio.U32BE(buf[12:16])
	b.DefaultSampleFlags = byteio.U32BE(buf[16:20])

	return
}

func NewUdtaBox(b *Box) *UdtaBox {
	return &UdtaBox{
		Box: b,
	}
}

func (b *UdtaBox) Serialize(w io.Writer) (writedLen int, err error) {

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

func (b *UdtaBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if b.Size == 1 {
		err = fmt.Errorf("large size in udta box")
		return
	}

	remainSize := int(b.Size) - BOX_SIZE
	curReadLen := 0
	for remainSize > 0 {

		var bb *Box
		if bb, curReadLen, err = ParseBox(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch bb.BoxType {
		case BoxTypeMETA:
			metaBox := NewMetaBox(bb)
			if curReadLen, err = metaBox.Parse(r); err != nil {
				return
			}
			b.SubBoxes = append(b.SubBoxes, metaBox)
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
