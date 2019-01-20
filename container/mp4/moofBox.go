package mp4

import (
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
aligned(8) class MovieFragmentHeaderBox
 extends FullBox(‘mfhd’, 0, 0){
 unsigned int(32) sequence_number;
}
*/
type MfhdBox struct {
	*FullBox
	SequenceNumber uint32
}

/*
aligned(8) class MovieFragmentBox extends Box(‘moof’){
}
*/

type MoofBox = SimpleBoxContainer

/*
aligned(8) class TrackFragmentBox extends Box(‘traf’){
}
*/
type TrafBox = SimpleBoxContainer

/*
aligned(8) class TrackFragmentHeaderBox
         extends FullBox(‘tfhd’, 0, tf_flags){
   unsigned int(32)  track_ID;
   // all the following are optional fields
   unsigned int(64)  base_data_offset;
   unsigned int(32)  sample_description_index;
   unsigned int(32)  default_sample_duration;
   unsigned int(32)  default_sample_size;
   unsigned int(32)  default_sample_flags
}
*/
type TfhdBox struct {
	*FullBox
	TrackID                uint32
	BaseDataOffset         uint64 //exits when Fullbox.falg24Bits & 0x000001 = true
	SampleDescriptionIndex uint32 //exits when Fullbox.falg24Bits & 0x000002 = true
	DefaultSampleDuration  uint32 //exits when Fullbox.falg24Bits & 0x000008 = true
	DefaultSampleSize      uint32 //exits when Fullbox.falg24Bits & 0x000010 = true
	DefaultSampleFlags     uint32 //exits when Fullbox.falg24Bits & 0x000020 = true
}

/*
aligned(8) class TrackFragmentBaseMediaDecodeTimeBox
   extends FullBox(‘tfdt’, version, 0) {
   if (version==1) {
      unsigned int(64) baseMediaDecodeTime;
   } else { // version==0
      unsigned int(32) baseMediaDecodeTime;
   }
}
*/
type TfdtBox struct {
	*FullBox
	BaseMediaDecodeTime uint64 // 就是当前mdat的起始时间戳，单位是对应track的timescale
}

/*
aligned(8) class TrackRunBox
         extends FullBox(‘trun’, version, tr_flags) {
   unsigned int(32)  sample_count;
   // the following are optional fields
   signed int(32) data_offset;
   unsigned int(32)  first_sample_flags;
   // all fields in the following array are optional
   {
      unsigned int(32)  sample_duration;
      unsigned int(32)  sample_size;
      unsigned int(32)  sample_flags
      if (version == 0)
         { unsigned int(32) sample_composition_time_offset; }
      else
         { signed int(32) sample_composition_time_offset; }
   }[ sample_count ]
}
*/
type TrunBoxSample struct {
	SampleDuration               uint32 //exits when Fullbox.falg24Bits & 0x000100 = true
	SampleSize                   uint32 //exits when Fullbox.falg24Bits & 0x000200 = true
	SampleFlags                  uint32 //exits when Fullbox.falg24Bits & 0x000400 = true
	SampleCompositionTimeOffset  uint32 //exits when Fullbox.falg24Bits & 0x000800 = true
	SSampleCompositionTimeOffset int32
}
type TrunBox struct {
	*FullBox
	SampleCount      uint32
	DataOffset       uint32 //exits when Fullbox.falg24Bits & 0x000001 = true
	FirstSampleFlags uint32 //exits when Fullbox.falg24Bits & 0x000004 = true
	BoxSamples       []*TrunBoxSample
}

func NewMoofBox(b *Box) *MoofBox {
	return NewSimpleBoxContainer(b)
}

func ParseMoofBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewMoofBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func NewMfhdBox(b *Box) *MfhdBox {
	return &MfhdBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *MfhdBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0

	if curWriteLen, err = byteio.WriteU32BE(w, b.SequenceNumber); err != nil {
		return
	}
	writedLen += curWriteLen

	return
}

func ParseMfhdBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewMfhdBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func (b *MfhdBox) Parse(r io.Reader) (totalReadLen int, err error) {

	curReadLen := 0
	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 4)
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	b.SequenceNumber = byteio.U32BE(buf)

	return
}

func NewTrafBox(b *Box) *TrafBox {
	return NewSimpleBoxContainer(b)
}

func ParseTrafBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewTrafBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func NewTfhdBox(b *Box) *TfhdBox {
	return &TfhdBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *TfhdBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	nums := []uint32{b.TrackID}

	if b.isBaseDataOffsetExists() {
		half1 := uint32(b.BaseDataOffset >> 32)
		half2 := uint32(b.BaseDataOffset & 0x00000000FFFFFFFF)
		nums = append(nums, half1, half2)
	}

	if b.isSampleDescriptionIndexExists() {
		nums = append(nums, b.SampleDescriptionIndex)
	}

	if b.isDefaultSampleDurationExists() {
		nums = append(nums, b.DefaultSampleDuration)
	}

	if b.isDefaultSampleSizeExists() {
		nums = append(nums, b.DefaultSampleSize)
	}

	if b.isDefaultSampleFlagsExists() {
		nums = append(nums, b.DefaultSampleFlags)
	}

	curWriteLen := 0
	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}

	writedLen += curWriteLen
	return
}

func ParseTfhdBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewTfhdBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func (b *TfhdBox) Parse(r io.Reader) (totalReadLen int, err error) {

	curReadLen := 0
	if totalReadLen, err = b.FullBox.Parse(r, 0, !FULLBOX_ANY_VERSION, FULLBOX_ANY_FLAG); err != nil {
		return
	}

	buf := make([]byte, 8)
	if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
		return
	}
	totalReadLen += curReadLen
	b.TrackID = byteio.U32BE(buf)

	if b.isBaseDataOffsetExists() {
		if curReadLen, err = io.ReadFull(r, buf[0:8]); err != nil {
			return
		}
		totalReadLen += curReadLen
		b.BaseDataOffset = byteio.U64BE(buf)
	}

	if b.isSampleDescriptionIndexExists() {
		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		totalReadLen += curReadLen
		b.SampleDescriptionIndex = byteio.U32BE(buf)
	}

	if b.isDefaultSampleDurationExists() {
		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		totalReadLen += curReadLen
		b.DefaultSampleDuration = byteio.U32BE(buf)
	}

	if b.isDefaultSampleSizeExists() {
		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		totalReadLen += curReadLen
		b.DefaultSampleSize = byteio.U32BE(buf)
	}

	if b.isDefaultSampleFlagsExists() {
		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		totalReadLen += curReadLen
		b.DefaultSampleFlags = byteio.U32BE(buf)
	}
	return
}

func (b *TfhdBox) isBaseDataOffsetExists() bool {
	return b.FullBox.flags24Bit&0x000001 == 0x000001
}

func (b *TfhdBox) isSampleDescriptionIndexExists() bool {
	return b.FullBox.flags24Bit&0x000002 == 0x000002
}

func (b *TfhdBox) isDefaultSampleDurationExists() bool {
	return b.FullBox.flags24Bit&0x000008 == 0x000008
}

func (b *TfhdBox) isDefaultSampleSizeExists() bool {
	return b.FullBox.flags24Bit&0x000010 == 0x000010
}

func (b *TfhdBox) isDefaultSampleFlagsExists() bool {
	return b.FullBox.flags24Bit&0x000020 == 0x000020
}

func NewTfdtBox(b *Box) *TfdtBox {
	return &TfdtBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *TfdtBox) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}
	curWriteLen := 0
	if b.version == 0 {
		curWriteLen, err = byteio.WriteU32BE(w, uint32(b.BaseMediaDecodeTime))
	} else {
		curWriteLen, err = byteio.WriteU64BE(w, b.BaseMediaDecodeTime)
	}
	if err != nil {
		return
	}

	writedLen += curWriteLen
	return
}

func ParseTfdtBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewTfdtBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func (b *TfdtBox) Parse(r io.Reader) (totalReadLen int, err error) {

	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, 0); err != nil {
		return
	}

	buf := make([]byte, 8)
	curReadLen := 0
	if b.version == 1 {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		b.BaseMediaDecodeTime = byteio.U64BE(buf)
	} else {
		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		b.BaseMediaDecodeTime = uint64(byteio.U32BE(buf))
	}
	totalReadLen += curReadLen

	return
}

func NewTrunBox(b *Box) *TrunBox {
	return &TrunBox{
		FullBox: &FullBox{
			Box: b,
		},
	}
}

func (b *TrunBox) Serialize(w io.Writer) (writedLen int, err error) {
	if writedLen, err = b.FullBox.Serialize(w); err != nil {
		return
	}

	curWriteLen := 0
	nums := []uint32{b.SampleCount}
	if b.isDataOffsetExists() {
		nums = append(nums, b.DataOffset)
	}

	if b.isFirstSampleFlagsExists() {
		nums = append(nums, b.FirstSampleFlags)
	}

	isSmapleDuration := b.isSampleDurationExists()
	isSampleSize := b.isSampleSizeExists()
	isSmapleFlags := b.isSampleFlagsExists()
	isSampleComp := b.isSampleCompositionTimeOffsetExists()

	for i := 0; i < len(b.BoxSamples); i++ {
		if isSmapleDuration {
			nums = append(nums, b.BoxSamples[i].SampleDuration)
		}
		if isSampleSize {
			nums = append(nums, b.BoxSamples[i].SampleSize)
		}
		if isSmapleFlags {
			nums = append(nums, b.BoxSamples[i].SampleFlags)
		}
		if isSampleComp {
			nums = append(nums, b.BoxSamples[i].SampleCompositionTimeOffset)
		}

	}

	if curWriteLen, err = uint32Serialize(w, nums); err != nil {
		return
	}
	writedLen += curWriteLen
	return
}

func ParseTrunBox(r io.Reader, box *Box) (b IBox, totalReadLen int, err error) {
	b = NewTrunBox(box)
	totalReadLen, err = b.Parse(r)
	return
}

func (b *TrunBox) Parse(r io.Reader) (totalReadLen int, err error) {

	curReadLen := 0
	if totalReadLen, err = b.FullBox.Parse(r, 0, FULLBOX_ANY_VERSION, FULLBOX_ANY_FLAG); err != nil {
		return
	}

	buf := make([]byte, 16)
	if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
		return
	}
	totalReadLen += curReadLen
	b.SampleCount = byteio.U32BE(buf)

	if b.isDataOffsetExists() {
		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		totalReadLen += curReadLen
		b.DataOffset = byteio.U32BE(buf)
	}

	if b.isFirstSampleFlagsExists() {
		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		totalReadLen += curReadLen
		b.FirstSampleFlags = byteio.U32BE(buf)
	}

	isSmapleDuration := b.isSampleDurationExists()
	isSampleSize := b.isSampleSizeExists()
	isSmapleFlags := b.isSampleFlagsExists()
	isSampleComp := b.isSampleCompositionTimeOffsetExists()

	itemLen := 0
	if isSmapleDuration {
		itemLen += 4
	}
	if isSampleSize {
		itemLen += 4
	}
	if isSmapleFlags {
		itemLen += 4
	}
	if isSampleComp {
		itemLen += 4
	}

	buf = buf[0:itemLen]
	offset := 0
	for i := uint32(0); i < b.SampleCount; i++ {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		totalReadLen += curReadLen

		boxSample := &TrunBoxSample{}
		offset = 0
		if isSmapleDuration {
			boxSample.SampleDuration = byteio.U32BE(buf[offset : offset+4])
			offset += 4
		}
		if isSampleSize {
			boxSample.SampleSize = byteio.U32BE(buf[offset : offset+4])
			offset += 4
		}
		if isSmapleFlags {
			boxSample.SampleFlags = byteio.U32BE(buf[offset : offset+4])
			offset += 4
		}
		if isSampleComp {
			boxSample.SampleCompositionTimeOffset = byteio.U32BE(buf[offset : offset+4])
			boxSample.SSampleCompositionTimeOffset = int32(boxSample.SampleCompositionTimeOffset)
		}
		b.BoxSamples = append(b.BoxSamples, boxSample)
	}

	return
}

func (b *TrunBox) isDataOffsetExists() bool {
	return b.FullBox.flags24Bit&0x000001 == 0x000001
}

func (b *TrunBox) isFirstSampleFlagsExists() bool {
	return b.FullBox.flags24Bit&0x000004 == 0x000004
}

func (b *TrunBox) isSampleDurationExists() bool {
	return b.FullBox.flags24Bit&0x000100 == 0x000100
}

func (b *TrunBox) isSampleSizeExists() bool {
	return b.FullBox.flags24Bit&0x000200 == 0x000200
}

func (b *TrunBox) isSampleFlagsExists() bool {
	return b.FullBox.flags24Bit&0x000400 == 0x000400
}

func (b *TrunBox) isSampleCompositionTimeOffsetExists() bool {
	return b.FullBox.flags24Bit&0x000800 == 0x000800
}
