package mp4

import (
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
	tag 在iso 14496-1这个文档里面
*/

type iTag interface {
	Parse(r io.Reader) (prasedLen int, err error)
	GetTag() uint8
	Serialize(w io.Writer) (writedLen int, err error)
}

/*
abstract aligned(8) expandable(2^28-1) class BaseDescriptor : bit(8) tag=0 {
// empty. To be filled by classes extending this class.
}
expandable(2^28-1)的格式如下，在14496-1的8.3章
int sizeOfInstance = 0;
bit(1) nextByte;
bit(7) sizeOfInstance;
while(nextByte) {
bit(1) nextByte;
bit(7) sizeByte;
sizeOfInstance = sizeOfInstance<<7 | sizeByte;
}
*/

type BaseDescriptor struct {
	Tag  uint8 // 上面结构里面并没有，但是实际上都有一个tag
	Size uint32
}

/*
abstract class DecoderSpecificInfo extends BaseDescriptor : bit(8) tag=DecSpecificInfoTag
{
// empty. To be filled by classes extending this class.
}
*/
type DecoderSpecificInfo struct {
	*BaseDescriptor
	RawData []byte
}

/*
class ProfileLevelIndicationIndexDescriptor () extends BaseDescriptor : bit(8) ProfileLevelIndicationIndexDescrTag {
bit(8) profileLevelIndicationIndex;
}
*/
type ProfileLevelIndicationIndexDescriptor struct {
	*BaseDescriptor
	ProfileLevelIndicationIndex uint8
}

/*
class DecoderConfigDescriptor extends BaseDescriptor : bit(8)
tag=DecoderConfigDescrTag {
 bit(8) objectTypeIndication;
 bit(6) streamType;
 bit(1) upStream;
 const bit(1) reserved=1;
 bit(24) bufferSizeDB;
 bit(32) maxBitrate;
 bit(32) avgBitrate;
 DecoderSpecificInfo decSpecificInfo[0 .. 1];
 profileLevelIndicationIndexDescriptor profileLevelIndicationIndexDescr[0..255];
 }
*/
type DecoderConfigDescriptor struct {
	*BaseDescriptor
	ObjectTypeIndication    uint8
	StreamType              uint8
	UpStream                uint8
	Reserved                uint8
	BufferSizeDB            uint32
	MaxBitrate              uint32
	AvgBitrate              uint32
	DecoderConfig           *DecoderSpecificInfo //这个至少音频就是audiosequence
	ProfileLevelIdcIdxDescr []*ProfileLevelIndicationIndexDescriptor
}

/*
class IPI_DescrPointer extends BaseDescriptor : bit(8) tag=IPI_DescrPointerTag {
bit(16) IPI_ES_Id;
}
*/
type IPI_DescrPointer struct {
	*BaseDescriptor
	IPI_ES_Id uint16
}

/*
class IPMP_DescriptorPointer extends BaseDescriptor : bit(8) tag = IPMP_DescrPtrTag
{
   bit(8) IPMP_DescriptorID;
   f (IPMP_DescriptorID == 0xff){
     bit(16) IPMP_DescriptorIDEx;
     bit(16) IPMP_ES_ID;
   }
}
*/
type IPMP_DescriptorPointer struct {
	*BaseDescriptor
	IPMP_DescriptorID   uint8
	IPMP_DescriptorIDEx uint16
	IPMP_ES_ID          uint16
}

/*
abstract class IP_IdentificationDataSet extends BaseDescriptor
: bit(8) tag=ContentIdentDescrTag..SupplContentIdentDescrTag
{
// empty. To be filled by classes extending this class.
}
*/
type IP_IdentificationDataSet = DecoderSpecificInfo

/*
abstract class OCI_Descriptor extends BaseDescriptor
: bit(8) tag= OCIDescrTagStartRange .. OCIDescrTagEndRange
{
// empty. To be filled by classes extending this class.
}

class LanguageDescriptor extends OCI_Descriptor : bit(8) tag=LanguageDescrTag {
bit(24) languageCode;
}
*/
type LanguageDescriptor struct {
	*BaseDescriptor
	LanguageCode24Bit uint32
}

/*
class QoS_Descriptor extends BaseDescriptor : bit(8) tag=QoS_DescrTag {
  bit(8) predefined;
  if (predefined==0) {
     QoS_Qualifier qualifiers[];
  }
}

abstract aligned(8) expandable(2^28-1) class QoS_Qualifier : bit(8) tag=0x01..0xff {
// empty. To be filled by classes extending this class.
}
class QoS_Qualifier_MAX_DELAY extends QoS_Qualifier : bit(8) tag=0x01 {
  unsigned int(32) MAX_DELAY;
}
class QoS_Qualifier_PREF_MAX_DELAY extends QoS_Qualifier : bit(8) tag=0x02 {
  unsigned int(32) PREF_MAX_DELAY;
}
class QoS_Qualifier_LOSS_PROB extends QoS_Qualifier : bit(8) tag=0x03 {
  double(32) LOSS_PROB;
}
class QoS_Qualifier_MAX_GAP_LOSS extends QoS_Qualifier : bit(8) tag=0x04 {
  unsigned int(32) MAX_GAP_LOSS;
}
class QoS_Qualifier_MAX_AU_SIZE extends QoS_Qualifier : bit(8) tag=0x41 {
  unsigned int(32) MAX_AU_SIZE;
}
class QoS_Qualifier_AVG_AU_SIZE extends QoS_Qualifier : bit(8) tag=0x42 {
  unsigned int(32) AVG_AU_SIZE;
}
class QoS_Qualifier_MAX_AU_RATE extends QoS_Qualifier : bit(8) tag=0x43 {
  unsigned int(32) MAX_AU_RATE;
}
class QoS_Qualifier_REBUFFERING_RATIO extends QoS_Qualifier : bit(8) tag=0x44 {
  bit(8) REBUFFERING_RATIO;
}
*/
type QoS_Qualifier struct {
	*BaseDescriptor //虽然不是这个含义，但是结构是一样的 //TODO 考虑做个typedef
	SpecificValue   uint32
}
type QoS_Descriptor struct {
	*BaseDescriptor
	Predefined uint8
	Qualifiers []*QoS_Qualifier //只有根据BaseDescriptor.Size - 1
}

/*
class RegistrationDescriptor extends BaseDescriptor : bit(8) tag=RegistrationDescrTag {
  bit(32) formatIdentifier;
  bit(8) additionalIdentificationInfo[sizeOfInstance-4];
}
*/
type RegistrationDescriptor struct {
	*BaseDescriptor
	FormatIdentifier             uint32
	AdditionalIdentificationInfo []byte
}

/*
abstract class ExtensionDescriptor extends BaseDescriptor
: bit(8) tag = ExtensionProfileLevelDescrTag, ExtDescrTagStartRange .. ExtDescrTagEndRange {
  // empty. To be filled by classes extending this class.
}
*/
type ExtensionDescriptor = DecoderSpecificInfo

/*
class SLConfigDescriptor extends BaseDescriptor : bit(8) tag=SLConfigDescrTag {
  bit(8) predefined;
  if (predefined==0) {
    bit(1) useAccessUnitStartFlag;
    bit(1) useAccessUnitEndFlag;
    bit(1) useRandomAccessPointFlag;
    bit(1) hasRandomAccessUnitsOnlyFlag;
    bit(1) usePaddingFlag;
    bit(1) useTimeStampsFlag;
    bit(1) useIdleFlag;
    bit(1) durationFlag;
    bit(32) timeStampResolution;
    bit(32) OCRResolution;
    bit(8) timeStampLength; // must be ≤ 64
    bit(8) OCRLength; // must be ≤ 64
    bit(8) AU_Length; // must be ≤ 32
    bit(8) instantBitrateLength;
    bit(4) degradationPriorityLength;
    bit(5) AU_seqNumLength; // must be ≤ 16
    bit(5) packetSeqNumLength; // must be ≤ 16
    bit(2) reserved=0b11;
  }
  if (durationFlag) {
     bit(32) timeScale;
     bit(16) accessUnitDuration;
     bit(16) compositionUnitDuration;
  }
  if (!useTimeStampsFlag) {
    bit(timeStampLength) startDecodingTimeStamp;
    bit(timeStampLength) startCompositionTimeStamp;
  }
}
*/
type SLConfigDescriptor struct {
	*BaseDescriptor
	PreDefined                       uint8
	UseAccessUnitStartFlag1Bit       uint8
	UseAccessUnitEndFlag1Bit         uint8
	UseRandomAccessPointFlag1Bit     uint8
	HasRandomAccessUnitsOnlyFlag1Bit uint8
	UsePaddingFlag1Bit               uint8
	UseTimeStampsFlag1Bit            uint8
	UseIdleFlag1Bit                  uint8
	DurationFlag1Bit                 uint8
	TimeStampResolution              uint32
	OCRResolution                    uint32
	TimeStampLength                  uint8 // must be ≤ 64
	OCRLength                        uint8 // must be ≤ 64
	AU_Length                        uint8 // must be ≤ 32
	InstantBitrateLength             uint8
	DegradationPriorityLength4Bit    uint8
	AU_seqNumLength5Bit              uint8 // must be ≤ 16
	PacketSeqNumLength5Bit           uint8 // must be ≤ 16
	Reserved2Bit                     uint8 //0b11;
	TimeScale                        uint32
	AccessUnitDuration               uint16
	CompositionUnitDuration          uint16
	StartDecodingTimeStamp           uint64
	StartCompositionTimeStamp        uint64
}

/*
class ES_Descriptor extends BaseDescriptor : bit(8) tag=ES_DescrTag {
bit(16) ES_ID;
bit(1) streamDependenceFlag;
bit(1) URL_Flag;
bit(1) OCRstreamFlag;
bit(5) streamPriority;
if (streamDependenceFlag)
  bit(16) dependsOn_ES_ID;
if (URL_Flag) {
  bit(8) URLlength;
  bit(8) URLstring[URLlength];
}
if (OCRstreamFlag)
  bit(16) OCR_ES_Id;
DecoderConfigDescriptor decConfigDescr;
if (ODProfileLevelIndication==0x01)
{
  SLConfigDescriptor slConfigDescr;
}
else {
  SLConfigDescriptor slConfigDescr;
}
//no SL extension.
// SL extension is possible.
IPI_DescrPointer ipiPtr[0 .. 1];
IP_IdentificationDataSet ipIDS[0 .. 255];
IPMP_DescriptorPointer ipmpDescrPtr[0 .. 255];
LanguageDescriptor langDescr[0 .. 255];
QoS_Descriptor qosDescr[0 .. 1];
RegistrationDescriptor regDescr[0 .. 1];
ExtensionDescriptor extDescr[0 .. 255];
}
*/
type ES_Descriptor struct {
	*BaseDescriptor
	ES_ID                uint16
	StreamDependenceFlag uint8
	URLFlag              uint8
	OCRstreamFlag        uint8
	StreamPriority       uint8
	DecoderConfigDescriptor
	SlConfigDescr *SLConfigDescriptor
	IpiPtr        *IPI_DescrPointer
	IpIDS         []*IP_IdentificationDataSet
	IpmpDescrPtr  []*IPMP_DescriptorPointer
	LangDescr     []*LanguageDescriptor
	QosDescr      *QoS_Descriptor
	RegDescr      *RegistrationDescriptor
	ExtDescr      []*ExtensionDescriptor
}

// class tag values of 14496-1
const (
	ForbiddenTag0                       = 0x00
	ObjectDescrTag                      = 0x01
	InitialObjectDescrTag               = 0x02
	ES_DescrTag                         = 0x03
	DecoderConfigDescrTag               = 0x04
	DecSpecificInfoTag                  = 0x05
	SLConfigDescrTag                    = 0x06
	ContentIdentDescrTag                = 0x07
	SupplContentIdentDescrTag           = 0x08
	IPI_DescrPointerTag                 = 0x09
	IPMP_DescrPointerTag                = 0x0A
	IPMP_DescrTag                       = 0x0B
	QoS_DescrTag                        = 0x0C
	RegistrationDescrTag                = 0x0D
	ES_ID_IncTag                        = 0x0E
	ES_ID_RefTag                        = 0x0F
	MP4_IOD_Tag                         = 0x10
	MP4_OD_Tag                          = 0x11
	IPL_DescrPointerRefTag              = 0x12
	ExtensionProfileLevelDescrTag       = 0x13
	ProfileLevelIndicationIndexDescrTag = 0x14
	Reserved1Start                      = 0x15 // for ISO use
	Reserved1End                        = 0x3F
	ContentClassificationDescrTag       = 0x40
	KeyWordDescrTag                     = 0x41
	RatingDescrTag                      = 0x42
	LanguageDescrTag                    = 0x43
	ShortTextualDescrTag                = 0x44
	ExpandedTextualDescrTag             = 0x45
	ContentCreatorNameDescrTag          = 0x46
	ContentCreationDateDescrTag         = 0x47
	OCICreatorNameDescrTag              = 0x48
	OCICreationDateDescrTag             = 0x49
	SmpteCameraPositionDescrTag         = 0x4A
	SegmentDescrTag                     = 0x4B
	MediaTimeDescrTag                   = 0x4C
	Reserved2Start                      = 0x4D // for ISO use (OCI extensions)
	Reserved2End                        = 0x5F
	IPMP_ToolsListDescrTag              = 0x60
	IPMP_ToolTag                        = 0x61
	M4MuxTimingDescrTag                 = 0x62
	M4MuxCodeTableDescrTag              = 0x63
	ExtSLConfigDescrTag                 = 0x64
	M4MuxBufferSizeDescrTag             = 0x65
	M4MuxIdentDescrTag                  = 0x66
	DependencyPointerTag                = 0x67
	DependencyMarkerTag                 = 0x68
	M4MuxChannelDescrTag                = 0x69
	Reserved3Start                      = 0x6A // for ISO use
	Reserved3End                        = 0xBF
	UserPrivateStart                    = 0xC0
	UserPrivateEnd                      = 0xFE
	ExtDescrTagStartRange               = 0x6A // 这里重叠了？
	ExtDescrTagEndRange                 = 0xFE
	ForbiddenTagF                       = 0xFF
)

/* objectTypeIndication as of 14496-1
0x00 Forbidden
0x01 Systems ISO/IEC 14496-1 a
0x02 Systems ISO/IEC 14496-1 b
0x03 Interaction Stream
0x04 Systems ISO/IEC 14496-1 Extended BIFS Configuration c
0x05 Systems ISO/IEC 14496-1 AFX d
0x06 Font Data Stream
0x07 Synthesized Texture Stream
0x08 Streaming Text Stream
0x09-0x1F reserved for ISO use
0x20 Visual ISO/IEC 14496-2 e
0x21 Visual ITU-T Recommendation H.264 | ISO/IEC 14496-10 f
0x22 Parameter Sets for ITU-T Recommendation H.264 | ISO/IEC 14496-10 f
0x23-0x3F reserved for ISO use
0x40 Audio ISO/IEC 14496-3 g
0x41-0x5F reserved for ISO use
0x60 Visual ISO/IEC 13818-2 Simple Profile
0x61 Visual ISO/IEC 13818-2 Main Profile
0x62 Visual ISO/IEC 13818-2 SNR Profile
0x63 Visual ISO/IEC 13818-2 Spatial Profile
0x64 Visual ISO/IEC 13818-2 High Profile
0x65 Visual ISO/IEC 13818-2 422 Profile
0x66 Audio ISO/IEC 13818-7 Main Profile
0x67 Audio ISO/IEC 13818-7 LowComplexity Profile
0x68 Audio ISO/IEC 13818-7 Scaleable Sampling Rate Profile
0x69 Audio ISO/IEC 13818-3
0x6A Visual ISO/IEC 11172-2
0x6B Audio ISO/IEC 11172-3
0x6C Visual ISO/IEC 10918-1
0x6D reserved for registration authority
0x6E Visual ISO/IEC 15444-1
0x6F - 0x9F reserved for ISO use
0xA0 - 0xBF reserved for registration authority i
0xC0 - 0xE0 user private
0xE1 reserved for registration authority i
0xE2 - 0xFE user private
0xFF no object type specified h
*/

type UnsupporttedTag = DecoderSpecificInfo

type tagParser func(r io.Reader, d *BaseDescriptor) (b iTag, readLen int, err error)

var (
	tagParseTable = map[uint8]tagParser{
		ES_DescrTag:                         parseEsDescrTag,
		DecoderConfigDescrTag:               parseDecoderConfigDescrTag,
		ContentIdentDescrTag:                parseDecoderConfigDescrTag,
		SupplContentIdentDescrTag:           parseDecoderConfigDescrTag,
		ExtensionProfileLevelDescrTag:       parseDecoderConfigDescrTag,
		DecSpecificInfoTag:                  parseDecSpecificInfoTag,
		ProfileLevelIndicationIndexDescrTag: parseProfileLevelIndicationIndexDescrTag,
		IPI_DescrPointerTag:                 parseIPI_DescrPointerTag,
		IPMP_DescrPointerTag:                parseIPMP_DescrPointerTag,
		LanguageDescrTag:                    parseLanguageDescrTag,
		RegistrationDescrTag:                parseRegistrationDescrTag,
		QoS_DescrTag:                        parseQoS_DescrTag,
		SLConfigDescrTag:                    parseSLConfigDescrTag,
	}
)

func parseUnsupporttedTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	tt := &UnsupporttedTag{
		BaseDescriptor: d,
	}
	tt.RawData = make([]byte, d.Size)
	if readLen, err = io.ReadFull(r, tt.RawData); err != nil {
		return
	}
	t = tt
	return
}

func getTagParser(tag uint8) tagParser {
	if parser, ok := tagParseTable[tag]; ok {
		return parser
	}

	return parseUnsupporttedTag
}

func parseOneTag(r io.Reader) (t iTag, totalReadLen int, err error) {

	tt := &BaseDescriptor{}
	if totalReadLen, err = tt.Parse(r); err != nil {
		return
	}

	curReadLen := 0
	parse := getTagParser(tt.GetTag())

	if t, curReadLen, err = parse(r, tt); err != nil {
		return
	}
	totalReadLen += curReadLen

	return
}

func getTagLength(buf []byte) int {
	len := 0
	for i := 0; i < 4; i++ {
		if (buf[i] & 0x80) > 0 {
			len |= int(buf[i]&0x7F) << 7
		} else {
			len |= int(buf[i] & 0x7F)
		}
	}
	return len
}

func (d *BaseDescriptor) Serialize(w io.Writer) (writedLen int, err error) {

	buf := make([]byte, 5)
	buf[0] = d.Tag
	len := ((d.Size << 3) | 0x80000000) & 0xFF000000
	len |= ((d.Size << 2) | 0x800000) & 0xFF0000
	len |= ((d.Size << 1) | 0x8000) & 0xFF00
	len |= (d.Size & 0x7F)
	byteio.PutU32BE(buf[1:5], len)

	writedLen, err = w.Write(buf)
	return
}

func (d *BaseDescriptor) GetTag() uint8 {
	return d.Tag
}

func (d *BaseDescriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 5)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	d.Tag = buf[0]
	d.Size = uint32(getTagLength(buf[1:5]))
	return
}

func parseDecSpecificInfoTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &DecoderSpecificInfo{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *DecoderSpecificInfo) Parse(r io.Reader) (totalReadLen int, err error) {

	d.RawData = make([]byte, d.Size)
	if totalReadLen, err = io.ReadFull(r, d.RawData); err != nil {
		return
	}
	return
}

func (d *DecoderSpecificInfo) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}
	if _, err = w.Write(d.RawData); err != nil {
		return
	}

	writedLen += len(d.RawData)
	return
}

func parseProfileLevelIndicationIndexDescrTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &ProfileLevelIndicationIndexDescriptor{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *ProfileLevelIndicationIndexDescriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 1)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	d.ProfileLevelIndicationIndex = buf[0]

	return
}

func (d *ProfileLevelIndicationIndexDescriptor) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}
	buf := []byte{d.ProfileLevelIndicationIndex}
	if _, err = w.Write(buf); err != nil {
		return
	}

	writedLen += 1
	return
}

func parseDecoderConfigDescrTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &DecoderConfigDescriptor{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *DecoderConfigDescriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	curReadLen := 0
	buf := make([]byte, 13)
	if curReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += curReadLen

	d.ObjectTypeIndication = buf[0]
	d.StreamType = buf[1] >> 2
	d.UpStream = (buf[1] & 0x02) >> 1
	d.Reserved = buf[1] & 0x01
	d.BufferSizeDB = byteio.U24BE(buf[2:5])
	d.MaxBitrate = byteio.U32BE(buf[5:9])
	d.AvgBitrate = byteio.U32BE(buf[9:13])

	remainSize := int(d.Size) - totalReadLen
	var itag iTag
	for remainSize > 0 {

		if itag, curReadLen, err = parseOneTag(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch itag.GetTag() {
		case ProfileLevelIndicationIndexDescrTag:
			d.ProfileLevelIdcIdxDescr = append(d.ProfileLevelIdcIdxDescr, itag.(*ProfileLevelIndicationIndexDescriptor))
		case DecSpecificInfoTag:
			d.DecoderConfig = itag.(*DecoderSpecificInfo)
		}
	}

	return
}

func (d *DecoderConfigDescriptor) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}
	buf := make([]byte, 13)
	buf[0] = d.ObjectTypeIndication
	buf[1] = d.StreamType<<2 | d.UpStream<<1 | d.Reserved
	byteio.PutU24BE(buf[2:5], d.BufferSizeDB)
	byteio.PutU32BE(buf[5:9], d.MaxBitrate)
	byteio.PutU32BE(buf[9:13], d.AvgBitrate)

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	if d.DecoderConfig != nil {
		if curWriteLen, err = d.DecoderConfig.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	for _, v := range d.ProfileLevelIdcIdxDescr {
		if curWriteLen, err = v.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}

func parseIPI_DescrPointerTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &IPI_DescrPointer{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *IPI_DescrPointer) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 2)
	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}
	totalReadLen += len(buf)
	d.IPI_ES_Id = byteio.U16BE(buf)

	return
}

func (d *IPI_DescrPointer) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 2)
	byteio.PutU16BE(buf, d.IPI_ES_Id)

	if _, err = w.Write(buf); err != nil {
		return
	}

	writedLen += len(buf)
	return
}

func parseIPMP_DescrPointerTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &IPMP_DescriptorPointer{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *IPMP_DescriptorPointer) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 4)
	if totalReadLen, err = io.ReadFull(r, buf[0:1]); err != nil {
		return
	}

	d.IPMP_DescriptorID = buf[0]
	if d.IPMP_DescriptorID == 0 {
		if _, err = io.ReadFull(r, buf); err != nil {
			return
		}
		totalReadLen += 4
		d.IPMP_DescriptorIDEx = byteio.U16BE(buf[1:3])
		d.IPMP_ES_ID = byteio.U16BE(buf[3:5])
	}

	return
}

func (d *IPMP_DescriptorPointer) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}

	buf := []byte{d.IPMP_DescriptorID, 0, 0, 0, 0}
	if d.IPMP_DescriptorID == 0 {
		byteio.PutU16BE(buf[1:3], d.IPMP_DescriptorIDEx)
		byteio.PutU16BE(buf[3:5], d.IPMP_ES_ID)
	} else {
		buf = buf[0:1]
	}

	if _, err = w.Write(buf); err != nil {
		return
	}

	writedLen += len(buf)
	return
}

func parseLanguageDescrTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &LanguageDescriptor{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *LanguageDescriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 3)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	d.LanguageCode24Bit = byteio.U24BE(buf)
	return
}

func (d *LanguageDescriptor) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 3)
	byteio.PutU24BE(buf, d.LanguageCode24Bit)

	if _, err = w.Write(buf); err != nil {
		return
	}

	writedLen += len(buf)
	return
}

func parseRegistrationDescrTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &RegistrationDescriptor{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *RegistrationDescriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 4)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	d.FormatIdentifier = byteio.U32BE(buf)
	d.AdditionalIdentificationInfo = make([]byte, d.Size-4)
	curReadLen := 0
	if curReadLen, err = io.ReadFull(r, d.AdditionalIdentificationInfo); err != nil {
		return
	}
	totalReadLen += curReadLen
	return
}

func (d *RegistrationDescriptor) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 4)
	byteio.PutU32BE(buf, d.FormatIdentifier)

	if _, err = w.Write(buf); err != nil {
		return
	}
	writedLen += len(buf)

	curWriteLen := 0
	if len(d.AdditionalIdentificationInfo) > 0 {
		if curWriteLen, err = w.Write(d.AdditionalIdentificationInfo); err != nil {
			return
		}
		writedLen += curWriteLen
	}
	return
}

func parseQoS_DescrTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &QoS_Descriptor{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *QoS_Descriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 4)
	if totalReadLen, err = io.ReadFull(r, buf[0:1]); err != nil {
		return
	}

	d.Predefined = buf[0]
	remainSize := int(d.Size) - 1
	curReadLen := 0
	if d.Predefined == 0 {
		for remainSize > 0 {
			dd := &BaseDescriptor{}
			if curReadLen, err = dd.Parse(r); err != nil {
				return
			}
			totalReadLen += curReadLen
			remainSize -= curReadLen

			f := &QoS_Qualifier{BaseDescriptor: dd}
			if dd.GetTag() == 0x44 {
				if curReadLen, err = io.ReadFull(r, buf[0:1]); err != nil {
					return
				}
				f.SpecificValue = uint32(buf[0])
			} else {
				if curReadLen, err = io.ReadFull(r, buf); err != nil {
					return
				}
				f.SpecificValue = byteio.U32BE(buf)
			}
			totalReadLen += curReadLen
			remainSize -= curReadLen
			d.Qualifiers = append(d.Qualifiers, f)
		}
	}

	return
}

func (d *QoS_Descriptor) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 4)
	buf[0] = d.Predefined

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf[0:1]); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(d.Qualifiers); i++ {

		if curWriteLen, err = d.Qualifiers[0].BaseDescriptor.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen

		if d.GetTag() == 0x44 {
			buf[0] = byte(d.Qualifiers[0].SpecificValue)
			if curWriteLen, err = w.Write(buf[0:1]); err != nil {
				return
			}
			writedLen += curWriteLen
		} else {
			byteio.PutU32BE(buf, d.Qualifiers[0].SpecificValue)
			if curWriteLen, err = w.Write(buf[0:1]); err != nil {
				return
			}
			writedLen += curWriteLen
		}
	}
	return
}

func parseSLConfigDescrTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &SLConfigDescriptor{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *SLConfigDescriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 15)
	if totalReadLen, err = io.ReadFull(r, buf[0:1]); err != nil {
		return
	}
	d.PreDefined = buf[0]

	curReadLen := 0
	if d.PreDefined == 0 {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		totalReadLen += curReadLen

		d.UseAccessUnitStartFlag1Bit = buf[0] & 0x80
		d.UseAccessUnitEndFlag1Bit = buf[0] & 0x40
		d.UseRandomAccessPointFlag1Bit = buf[0] & 0x20
		d.HasRandomAccessUnitsOnlyFlag1Bit = buf[0] & 0x10
		d.UsePaddingFlag1Bit = buf[0] & 0x08
		d.UseTimeStampsFlag1Bit = buf[0] & 0x04
		d.UseIdleFlag1Bit = buf[0] & 0x02
		d.DurationFlag1Bit = buf[0] & 0x01

		d.TimeStampResolution = byteio.U32BE(buf[1:5])
		d.OCRResolution = byteio.U32BE(buf[5:9])

		d.TimeStampLength = buf[9]
		d.OCRLength = buf[10]
		d.AU_Length = buf[11]
		d.InstantBitrateLength = buf[12]

		d.DegradationPriorityLength4Bit = (buf[13] & 0xF0) >> 4
		d.AU_seqNumLength5Bit = (buf[13]&0x0F)<<1 | (buf[14]&0x80)>>7
		d.PacketSeqNumLength5Bit = (buf[14] & 0x7C) >> 2
		d.Reserved2Bit = buf[14] & 0x03

		if d.DurationFlag1Bit > 0 {
			if curReadLen, err = io.ReadFull(r, buf[0:8]); err != nil {
				return
			}
			totalReadLen += curReadLen

			d.TimeScale = byteio.U32BE(buf[0:4])
			d.AccessUnitDuration = byteio.U16BE(buf[4:6])
			d.CompositionUnitDuration = byteio.U16BE(buf[6:8])
		}

		if d.UseTimeStampsFlag1Bit == 0 {
			buf := buf[0:d.TimeStampLength]
			if curReadLen, err = io.ReadFull(r, buf); err != nil {
				return
			}
			totalReadLen += curReadLen
			d.StartDecodingTimeStamp = byteio.UVarBE(buf)

			buf = buf[0:d.TimeStampLength]
			if curReadLen, err = io.ReadFull(r, buf); err != nil {
				return
			}
			totalReadLen += curReadLen
			d.StartCompositionTimeStamp = byteio.UVarBE(buf)
		}
	}

	return
}

func (d *SLConfigDescriptor) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}

	buf := make([]byte, 15)
	buf[0] = d.PreDefined

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf[0:1]); err != nil {
		return
	}
	writedLen += curWriteLen

	if d.PreDefined == 0 {
		buf[0] = d.UseAccessUnitStartFlag1Bit | d.UseAccessUnitEndFlag1Bit |
			d.UseRandomAccessPointFlag1Bit | d.HasRandomAccessUnitsOnlyFlag1Bit |
			d.UsePaddingFlag1Bit | d.UseTimeStampsFlag1Bit |
			d.UseIdleFlag1Bit | d.DurationFlag1Bit

		byteio.PutU32BE(buf[1:5], d.TimeStampResolution)
		byteio.PutU32BE(buf[5:9], d.OCRResolution)

		buf[9] = d.TimeStampLength
		buf[10] = d.OCRLength
		buf[11] = d.AU_Length
		buf[12] = d.InstantBitrateLength

		buf[13] = d.DegradationPriorityLength4Bit<<4 | d.AU_seqNumLength5Bit>>1
		buf[14] = d.AU_seqNumLength5Bit<<7 | d.PacketSeqNumLength5Bit<<2 | d.Reserved2Bit

		if curWriteLen, err = w.Write(buf); err != nil {
			return
		}
		writedLen += curWriteLen

		if d.DurationFlag1Bit > 0 {
			byteio.PutU32BE(buf[0:4], d.TimeScale)
			byteio.PutU16BE(buf[4:6], d.AccessUnitDuration)
			byteio.PutU16BE(buf[6:8], d.CompositionUnitDuration)

			if curWriteLen, err = w.Write(buf[0:8]); err != nil {
				return
			}
			writedLen += curWriteLen
		}

		if d.UseTimeStampsFlag1Bit == 0 {
			d.StartDecodingTimeStamp = byteio.UVarBE(buf)
			d.StartCompositionTimeStamp = byteio.UVarBE(buf)
			ts := d.StartDecodingTimeStamp
			for i := uint8(0); i < 2; i++ {
				if i == 1 {
					ts = d.StartCompositionTimeStamp
				}
				switch d.TimeStampLength {
				case 1:
					buf[i*d.TimeStampLength] = uint8(ts)
				case 2:
					byteio.PutU16BE(buf[i*d.TimeStampLength:i*d.TimeStampLength+2], uint16(ts))
				case 3:
					byteio.PutU24BE(buf[i*d.TimeStampLength:i*d.TimeStampLength+3], uint32(ts))
				case 4:
					byteio.PutU32BE(buf[i*d.TimeStampLength:i*d.TimeStampLength+4], uint32(ts))
				case 5:
					byteio.PutU40BE(buf[i*d.TimeStampLength:i*d.TimeStampLength+5], uint64(ts))
				case 6:
					byteio.PutU48BE(buf[i*d.TimeStampLength:i*d.TimeStampLength+6], uint64(ts))
				case 7:
					byteio.PutU56BE(buf[i*d.TimeStampLength:i*d.TimeStampLength+7], uint64(ts))
				case 8:
					byteio.PutU64BE(buf[i*d.TimeStampLength:i*d.TimeStampLength+8], uint64(ts))
				}
			}
			if curWriteLen, err = w.Write(buf[0 : d.TimeStampLength*2]); err != nil {
				return
			}
			writedLen += curWriteLen
		}
	}

	return
}

func parseEsDescrTag(r io.Reader, d *BaseDescriptor) (t iTag, readLen int, err error) {

	t = &ES_Descriptor{
		BaseDescriptor: d,
	}

	readLen, err = t.Parse(r)
	return
}

func (d *ES_Descriptor) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 3)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	if d.Tag != ES_DescrTag {
		err = fmt.Errorf("expect ES_DescrTag:%d but:%d", ES_DescrTag, d.Tag)
		return
	}

	curReadLen := 0

	d.ES_ID = byteio.U16BE(buf)
	d.StreamDependenceFlag = buf[2] & 0x80
	d.URLFlag = buf[2] & 0x40
	d.OCRstreamFlag = buf[2] & 0x20
	d.StreamPriority = buf[2] & 0x1F

	dd := &BaseDescriptor{}
	if curReadLen, err = dd.Parse(r); err != nil {
		return
	}
	totalReadLen += curReadLen

	d.DecoderConfigDescriptor.BaseDescriptor = dd
	if curReadLen, err = d.DecoderConfigDescriptor.Parse(r); err != nil {
		return
	}
	totalReadLen += curReadLen

	remainSize := int(d.Size) - totalReadLen
	var itag iTag
	for remainSize > 0 {

		if itag, curReadLen, err = parseOneTag(r); err != nil {
			return
		}
		totalReadLen += curReadLen
		remainSize -= curReadLen

		switch itag.GetTag() {
		case SLConfigDescrTag:
			d.SlConfigDescr = itag.(*SLConfigDescriptor)
		case IPI_DescrPointerTag:
			d.IpiPtr = itag.(*IPI_DescrPointer)
		case ContentIdentDescrTag:
			fallthrough
		case SupplContentIdentDescrTag:
			d.IpIDS = append(d.IpIDS, itag.(*IP_IdentificationDataSet))
		case IPMP_DescrPointerTag:
			d.IpmpDescrPtr = append(d.IpmpDescrPtr, itag.(*IPMP_DescriptorPointer))
		case LanguageDescrTag:
			d.LangDescr = append(d.LangDescr, itag.(*LanguageDescriptor))
		case QoS_DescrTag:
			d.QosDescr = itag.(*QoS_Descriptor)
		case RegistrationDescrTag:
			d.RegDescr = itag.(*RegistrationDescriptor)
		case ExtensionProfileLevelDescrTag:
			d.ExtDescr = append(d.ExtDescr, itag.(*ExtensionDescriptor))
		}
	}

	return
}

func (d *ES_Descriptor) Serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = d.BaseDescriptor.Serialize(w); err != nil {
		return
	}
	buf := make([]byte, 3)
	byteio.PutU16BE(buf, d.ES_ID)
	buf[2] = d.StreamDependenceFlag | d.URLFlag | d.OCRstreamFlag | d.StreamPriority

	curWriteLen := 0
	if curWriteLen, err = w.Write(buf); err != nil {
		return
	}
	writedLen += curWriteLen

	if curWriteLen, err = d.DecoderConfigDescriptor.Serialize(w); err != nil {
		return
	}
	writedLen += curWriteLen

	if d.SlConfigDescr != nil {
		if curWriteLen, err = d.SlConfigDescr.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	if d.IpiPtr != nil {
		if curWriteLen, err = d.IpiPtr.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	for i := 0; i < len(d.IpIDS); i++ {
		if curWriteLen, err = d.IpIDS[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	for i := 0; i < len(d.IpmpDescrPtr); i++ {
		if curWriteLen, err = d.IpmpDescrPtr[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	for i := 0; i < len(d.LangDescr); i++ {
		if curWriteLen, err = d.LangDescr[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	if d.QosDescr != nil {
		if curWriteLen, err = d.QosDescr.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	if d.RegDescr != nil {
		if curWriteLen, err = d.RegDescr.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	for i := 0; i < len(d.ExtDescr); i++ {
		if curWriteLen, err = d.ExtDescr[i].Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}
