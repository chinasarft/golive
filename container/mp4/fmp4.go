package mp4

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/chinasarft/golive/av"
	"github.com/chinasarft/golive/utils/byteio"
)

/*
fmp4 moovbox很多都不需要了，分析ffmpeg的生成的fmp4文件，里面还有很多moov下很多子box都还在，但是都没有值
moof/traf/tfhd.base_data_offset 看起来就是moof离文件最开头的偏移
moof/traf/trun.data_offset 用来表示和该 moof 配套的 mdat 中实际数据内容距 moof 开头有多远,第一个track其实值就是: moof.Size+BOX_SIZE(8）
第二个track开始要累加第一个track的数据的长度
moof/traf/trun.sample.sample_size就可以定位帧的位置了。这个长度的起始位置就是mdat的数据起始位置，然后累计sample_size作为偏移就是
下一个sample的开始了
19040101到19700101的秒数为2082844800
一个mdat可以放多个track的内容，不过应该是串行排列的，根据偏移计算
*/

const (
	Diff1970To1904 = 2082844800
)

var (
	aacArIdxMap []uint32 = []uint32{96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050, 16000, 12000, 11025, 8000, 7350}
)

type Fmp4MoofMdat struct {
	Moof *MoofBox
	Mdat *MdatBox
}

type MdatCache struct {
	trunBox           *TrunBox
	buf               bytes.Buffer
	lastTs            int64
	firstTs           int64
	accOffset         uint64
	defaultSampleSize uint32
	baseDataOffset    uint64
	trackID           uint32
	baseDecodeTime    uint64
}

type Fmp4 struct {
	Ftyp           FtypBox
	Moov           MoovBox
	mvexInMoov     *MvexBox
	mvexAppended   bool
	MoofMdat       []Fmp4MoofMdat
	Mfra           MfraBox
	audioTrackId   uint32
	videoTrackId   uint32
	cmTime         uint64
	curTrackOffset uint64
	keyFrameCount  uint32
	moofBox        *MoofBox
	mdatBuf        bytes.Buffer
	vCache         MdatCache
	aCache         MdatCache
	segDuration    time.Duration
	headerBox      bytes.Buffer
	mvhdBox        *MvhdBox
	vTrackMdhdBox  *MdhdBox // 视频的时间的timescale不好计算，音频可以设定为频率，
	// 所以这里单独放出来，后面在更正timescale
	moofMdatSeqNum     uint32
	fmp4BaseDataOffset uint64
}

func findBoxByType(boxes []IBox, boxTypes []uint32) IBox {
	subBoxes := boxes
	for idx, boxType := range boxTypes {
		for _, box := range subBoxes {
			if box.GetBoxType() == boxType {
				if idx == len(boxTypes)-1 {
					return box
				} else {
					subBoxes = box.GetSubBoxes()
				}
			}
		}
	}

	return nil
}

func NewFmp4(segDuration time.Duration) *Fmp4 {

	nowSec := uint64(time.Now().Unix())

	mvhdBox := &MvhdBox{
		FullBox:          NewTypeFullBox(BoxTypeMVHD, 0, 0),
		CreationTime:     nowSec + Diff1970To1904,
		ModificationTime: nowSec + Diff1970To1904,
		Timescale:        1000,
		TemplateRate:     0x00010000,
		TemplateMatrix:   [9]int32{0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000},
		NextTrackID:      1, // TODO, ffmpeg封装有音视频的fmp4这个值是2，但是应该是3啊？
	}
	mvhdBox.Size += MvhdBoxBodyLenVer0

	f := &Fmp4{
		Ftyp: FtypBox{
			Box:        NewTypeBox(BoxTypeFTYP),
			MajorBrand: Mp4BoxBrandISOM,
			MinorBrand: 0x0200, // ffmpeg do so
		},
		Moov: MoovBox{
			Box: NewTypeBox(BoxTypeMOOV),
			SubBoxes: []IBox{
				mvhdBox,
			},
		},
		mvexInMoov: &MvexBox{
			Box: NewTypeBox(BoxTypeMVEX),
		},
		cmTime:         uint64(time.Now().Unix()) + Diff1970To1904,
		segDuration:    segDuration,
		mvhdBox:        mvhdBox,
		moofMdatSeqNum: 1,
	}
	f.Moov.Size += mvhdBox.Size
	f.Ftyp.Size += 8 //8 for major and minor brand

	return f
}

func (f *Fmp4) SetMajorBrand(brand uint32) {
	f.Ftyp.MajorBrand = brand
}

func (f *Fmp4) SetMinorBrand(brand uint32) {
	f.Ftyp.MinorBrand = brand
}

func (f *Fmp4) AppendCompatibleBrand(brand uint32) {
	for _, v := range f.Ftyp.CompatibleBrands {
		if brand == v {
			return
		}
	}
	f.Ftyp.CompatibleBrands = append(f.Ftyp.CompatibleBrands, brand)
	f.Ftyp.Size += 4
}

func (f *Fmp4) AddVideoH264Track(avcSeqHdlr []byte) (err error) {

	if f.videoTrackId != 0 {
		return fmt.Errorf("video trackid already exists")
	}

	dc := NewAVCDecoderConfigurationRecord()
	r := bytes.NewReader(avcSeqHdlr)
	if _, err = dc.Parse(r); err != nil {
		return
	}

	var sps *av.SPS
	if sps, err = av.ParseVideoSPS(dc.Sps[0].SpsNalu[1:]); err != nil {
		return
	}
	w, h := sps.GetWithHeight()

	tkhdBox := &TkhdBox{
		FullBox:          NewTypeFullBox(BoxTypeTKHD, 0, 3),
		CreationTime:     f.cmTime,
		ModificationTime: f.cmTime,
		TrackID:          f.mvhdBox.NextTrackID,
		Width:            uint32(w) << 16,
		Height:           uint32(h) << 16,
		TemplateMatrix:   [9]int32{0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000},
	}
	tkhdBox.Size += TkhdBoxBodyLenVer0

	paspBox := &PaspBox{
		Box:      NewTypeBox(BoxTypePASP),
		HSpacing: 16, // TODO how to get
		VSpacing: 15, // TODO
	}
	paspBox.Size += PaspBoxBodyLen
	avccBox := &AVCCConfigurationBox{
		Box:                           NewTypeBox(BoxTypeAVCC),
		AVCDecoderConfigurationRecord: *dc,
	}
	avccBox.Size += uint64(len(avcSeqHdlr))

	avc1Box := &Avc1Box{
		Box: NewTypeBox(BoxTypeAVC1),
		AVCEntry: AVCSampleEntry{
			VisualSampleEntry: VisualSampleEntry{
				SampleEntry: SampleEntry{
					DataReferenceIndex: 1,
				},
				Width:                   w,
				Height:                  h,
				TemplateHorizResolution: 0x00480000,
				TemplateVertResolution:  0x00480000,
				TemplateFrameCount:      1,
				TemplateDepth:           0x18,
				PreDefined3:             -1,
			},
			AVCCConfigurationBox: avccBox,
		},
		SubBoxes: []IBox{
			paspBox,
		},
	}
	avc1Box.Size += (VisualSampleEntryLen + SampleEntryLen)
	avc1Box.Size += paspBox.Size
	avc1Box.Size += avccBox.Size

	stblBox := newVideoStblBox(avc1Box)

	mdhdBox := newFmp4MdhdBox(1000, f.cmTime)
	hdlrBox := newFmp4VideoHdlrBox()
	minfBox := newFmp4VideoMinfBox(stblBox)

	mdiaBox := &MdiaBox{
		Box: NewTypeBox(BoxTypeMDIA),
		SubBoxes: []IBox{
			mdhdBox,
			hdlrBox,
			minfBox,
		},
	}
	mdiaBox.Size += mdhdBox.Size
	mdiaBox.Size += hdlrBox.Size
	mdiaBox.Size += minfBox.Size

	trakBox := &TrakBox{
		Box: NewTypeBox(BoxTypeTRAK),
		SubBoxes: []IBox{
			tkhdBox,
			mdiaBox,
		},
	}

	trakBox.Size += tkhdBox.Size
	trakBox.Size += mdiaBox.Size

	f.Moov.SubBoxes = append(f.Moov.SubBoxes, trakBox)
	f.vTrackMdhdBox = mdhdBox
	f.appendTrexBox()
	f.vCache.trackID = f.mvhdBox.NextTrackID
	f.videoTrackId = f.mvhdBox.NextTrackID
	f.mvhdBox.NextTrackID++
	f.Moov.Size += trakBox.Size
	return nil
}

func (f *Fmp4) appendTrexBox() {
	trexBox := &TrexBox{
		FullBox:                       NewTypeFullBox(BoxTypeTREX, 0, 0),
		TrackID:                       f.mvhdBox.NextTrackID,
		DefaultSampleDescriptionIndex: 1, // TODO mean what?
	}
	trexBox.Size += TrexBoxBodyLen
	f.mvexInMoov.Size += trexBox.Size
	f.mvexInMoov.SubBoxes = append(f.mvexInMoov.SubBoxes, trexBox)
	return
}

func (f *Fmp4) AddAudioTrack(aacSeqHdlr []byte) error {

	if f.audioTrackId != 0 {
		return fmt.Errorf("audio trackid already exists")
	}

	tkhdBox := &TkhdBox{
		FullBox:                NewTypeFullBox(BoxTypeTKHD, 0, 3),
		CreationTime:           f.cmTime,
		ModificationTime:       f.cmTime,
		TrackID:                f.mvhdBox.NextTrackID,
		TemplatealTernateGroup: 1,
		TemplateVolume:         0x100,
		TemplateMatrix:         [9]int32{0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000},
	}
	tkhdBox.Size += TkhdBoxBodyLenVer0

	esBd := &BaseDescriptor{
		Tag:  3,
		Size: uint32(32 + len(aacSeqHdlr)),
	}
	dcBd := &BaseDescriptor{
		Tag:  4,
		Size: uint32(18 + len(aacSeqHdlr)),
	}

	esdsBox := &EsdsBox{
		FullBox: NewTypeFullBox(BoxTypeESDS, 0, 0),
		EsDescr: ES_Descriptor{
			BaseDescriptor: esBd,
			ES_ID:          1, //只需要唯一，是否可以用对应audiotrack的trackid?
			DecoderConfigDescriptor: DecoderConfigDescriptor{
				BaseDescriptor:       dcBd,
				ObjectTypeIndication: 0x40, // 应该是固定的，14996-1 7.2.6.6.2
				StreamType:           5,    //14996-1 7.2.6.6.2, streamType解释，5表示audioStream
				Reserved:             1,
				MaxBitrate:           96 * 1024, //TODO
				DecoderConfig: &DecoderSpecificInfo{
					BaseDescriptor: &BaseDescriptor{
						Tag:  5,
						Size: uint32(len(aacSeqHdlr)),
					},
				},
			},
			SlConfigDescr: &SLConfigDescriptor{
				BaseDescriptor: &BaseDescriptor{
					Tag:  6,
					Size: 1,
				},
				PreDefined: 2,
			},
		},
	}
	esdsBox.EsDescr.DecoderConfigDescriptor.DecoderConfig.RawData = make([]byte, len(aacSeqHdlr))
	copy(esdsBox.EsDescr.DecoderConfigDescriptor.DecoderConfig.RawData, aacSeqHdlr)
	esdsBox.Size += uint64(esdsBox.EsDescr.Size + 5)

	sampleRateIdx := int((aacSeqHdlr[0]&0x07)<<1 | (aacSeqHdlr[1]&0x80)>>7)
	if sampleRateIdx > 12 {
		return fmt.Errorf("wrong samplerate idx:%d", sampleRateIdx)
	}
	mp4aBox := &Mp4aBox{
		Box: NewTypeBox(BoxTypeMP4A),
		AudioEntry: AudioSampleEntry{
			SampleEntry: SampleEntry{
				DataReferenceIndex: 1,
			},
			TemplateChannelCount: 2, //template开头都是固定的，为什么？
			TemplateSampleSize:   16,
			TemplateSampleRate:   (aacArIdxMap[sampleRateIdx]) << 16,
		},
		SubBoxes: []IBox{
			esdsBox,
		},
	}
	mp4aBox.Size += (AudioSampleEntryLen + SampleEntryLen)
	mp4aBox.Size += esdsBox.Size

	stblBox := newVideoStblBox(mp4aBox)

	mdhdBox := newFmp4MdhdBox(aacArIdxMap[sampleRateIdx], f.cmTime)
	hdlrBox := newFmp4AudioHdlrBox()
	minfBox := newFmp4AudioMinfBox(stblBox)

	mdiaBox := &MdiaBox{
		Box: NewTypeBox(BoxTypeMDIA),
		SubBoxes: []IBox{
			mdhdBox,
			hdlrBox,
			minfBox,
		},
	}
	mdiaBox.Size += mdhdBox.Size
	mdiaBox.Size += hdlrBox.Size
	mdiaBox.Size += minfBox.Size

	trakBox := &TrakBox{
		Box: NewTypeBox(BoxTypeTRAK),
		SubBoxes: []IBox{
			tkhdBox,
			mdiaBox,
		},
	}

	trakBox.Size += tkhdBox.Size
	trakBox.Size += mdiaBox.Size

	f.appendTrexBox()
	f.aCache.trackID = f.mvhdBox.NextTrackID
	f.audioTrackId = f.mvhdBox.NextTrackID
	f.mvhdBox.NextTrackID++
	f.Moov.SubBoxes = append(f.Moov.SubBoxes, trakBox)
	f.mvhdBox.TemplateVolume = 0x0100
	f.Moov.Size += trakBox.Size
	return nil
}

func (f *Fmp4) generateHeaderBox() (err error) {

	if _, err = f.Ftyp.Serialize(&f.headerBox); err != nil {
		return
	}
	if f.mvexAppended == false {
		f.mvexAppended = true
		f.Moov.SubBoxes = append(f.Moov.SubBoxes, f.mvexInMoov)
		f.Moov.Size += f.mvexInMoov.Size
	}

	if _, err = f.Moov.Serialize(&f.headerBox); err != nil {
		return
	}
	f.fmp4BaseDataOffset = uint64(len(f.headerBox.Bytes()))

	return
}

func (f *Fmp4) resetFrag(isForce bool, ts int64) {

	if f.audioTrackId > 0 {
		f.aCache.reset(f.fmp4BaseDataOffset, uint64(ts))
	}
	if f.videoTrackId > 0 {
		f.vCache.reset(f.fmp4BaseDataOffset, uint64(ts))
	}

	if isForce || f.moofBox == nil {
		mfhdBox := &MfhdBox{
			FullBox:        NewTypeFullBox(BoxTypeMFHD, 0, 0),
			SequenceNumber: f.moofMdatSeqNum,
		}
		mfhdBox.Size += MfhdBoxBodyLen
		f.moofBox = &MoofBox{
			Box: NewTypeBox(BoxTypeMOOF),
			SubBoxes: []IBox{
				mfhdBox,
			},
		}
		f.moofBox.Size += mfhdBox.Size
		f.mdatBuf.Reset()
	}
}

func newFmp4DinfBox() *DinfBox {
	urlBox := &UrlBox{
		FullBox: NewTypeFullBox(BoxTypeURL, 0, 0),
	}

	drefBox := &DrefBox{
		FullBox:    NewTypeFullBox(BoxTypeDREF, 0, 0),
		EntryCount: 1,
		SubBoxes: []IBox{
			urlBox,
		},
	}
	drefBox.Size += 4 // for EntryCount
	drefBox.Size += urlBox.Size

	dinfBox := &DinfBox{
		Box: NewTypeBox(BoxTypeDINF),
		SubBoxes: []IBox{
			drefBox,
		},
	}
	dinfBox.Size += drefBox.Size

	return dinfBox
}

func newFmp4MinfBox(mhd IBox, stblBox *StblBox) *MinfBox {

	dinfBox := newFmp4DinfBox()

	minfBox := &MinfBox{
		Box: NewTypeBox(BoxTypeMINF),
		SubBoxes: []IBox{
			mhd,
			dinfBox,
			stblBox,
		},
	}
	minfBox.Size += mhd.GetBoxSize()
	minfBox.Size += dinfBox.Size
	minfBox.Size += stblBox.Size

	return minfBox
}

func newFmp4VideoMinfBox(stblBox *StblBox) *MinfBox {

	vmhdBox := &VmhdBox{
		FullBox: NewTypeFullBox(BoxTypeVMHD, 0, 1),
	}
	vmhdBox.Size += VmhdBoxBodyLen

	return newFmp4MinfBox(vmhdBox, stblBox)
}

func newFmp4AudioMinfBox(stblBox *StblBox) *MinfBox {
	smhdBox := &SmhdBox{
		FullBox: NewTypeFullBox(BoxTypeSMHD, 0, 0),
	}
	smhdBox.Size += SmhdBoxBodyLen

	return newFmp4MinfBox(smhdBox, stblBox)
}

func newFmp4AudioHdlrBox() *HdlrBox {
	hdlrBox := &HdlrBox{
		FullBox:     NewTypeFullBox(BoxTypeHDLR, 0, 0),
		handlerType: AudioHandlerType,
		Name:        []byte{'S', 'o', 'u', 'n', 'd', 'H', 'a', 'n', 'd', 'l', 'e', 'r', 0},
	}
	hdlrBox.Size += 33
	return hdlrBox
}

func newFmp4VideoHdlrBox() *HdlrBox {
	hdlrBox := &HdlrBox{
		FullBox:     NewTypeFullBox(BoxTypeHDLR, 0, 0),
		handlerType: VideoHandlerType,
		Name:        []byte{'V', 'i', 'd', 'e', 'o', 'H', 'a', 'n', 'd', 'l', 'e', 'r', 0},
	}
	hdlrBox.Size += 33
	return hdlrBox
}

func newFmp4MdhdBox(timesacle uint32, cmTime uint64) *MdhdBox {

	mdhdBox := &MdhdBox{
		FullBox:          NewTypeFullBox(BoxTypeMDHD, 0, 0),
		CreationTime:     cmTime,
		ModificationTime: cmTime,
		Timescale:        timesacle,
		Language:         [3]int8{0x15, 0x0E, 0x04}, //55c4 ->101010111000100->10101 01110 00100 und(undtermined)
	}
	mdhdBox.Size += MdhdBoxBodyLenVer0
	return mdhdBox
}

func newVideoStblBox(stsdSubBox IBox) (stblBox *StblBox) {

	stsdBox := &StsdBox{
		FullBox:    NewTypeFullBox(BoxTypeSTSD, 0, 0),
		EntryCount: 1,
		SubBoxes: []IBox{
			stsdSubBox,
		},
	}
	stsdBox.Size += stsdSubBox.GetBoxSize()
	stsdBox.Size += 4

	sttsBox := &SttsBox{
		FullBox: NewTypeFullBox(BoxTypeSTTS, 0, 0),
	}
	sttsBox.Size += 4

	stscBox := &StscBox{
		FullBox: NewTypeFullBox(BoxTypeSTSC, 0, 0),
	}
	stscBox.Size += 4

	stszBox := &StszBox{
		FullBox: NewTypeFullBox(BoxTypeSTSZ, 0, 0),
	}
	stszBox.Size += 8

	stcoBox := &StcoBox{
		FullBox: NewTypeFullBox(BoxTypeSTCO, 0, 0),
	}
	stcoBox.Size += 4

	stblBox = &StblBox{
		Box: NewTypeBox(BoxTypeSTBL),
		SubBoxes: []IBox{
			stsdBox,
			sttsBox,
			stscBox,
			stszBox,
			stcoBox,
		},
	}
	stblBox.Size += stsdBox.Size
	stblBox.Size += sttsBox.Size
	stblBox.Size += stscBox.Size
	stblBox.Size += stszBox.Size
	stblBox.Size += stcoBox.Size

	return
}

func (f *Fmp4) AddAudioFrameWithoutLen(frame []byte, ts int64) (err error) {
	if f.audioTrackId < 1 {
		return fmt.Errorf("audio track not exists")
	}
	if f.videoTrackId > 0 && f.keyFrameCount == 0 {
		err = fmt.Errorf("no key frame")
	}
	if err = f.aCache.addFrame(frame, ts, len(frame)); err != nil {
		return
	}

	return
}

func (f *Fmp4) AddVideoFrameWithLen(frame []byte, ts int64, isKeyFrame bool) (err error) {
	if f.videoTrackId < 1 {
		return fmt.Errorf("video track not exists")
	}

	if isKeyFrame {
		if f.vCache.accOffset > 0 {
			if len(f.headerBox.Bytes()) == 0 {
				if err = f.generateHeaderBox(); err != nil {
					return
				}
				if f.vCache.baseDataOffset < 1 {
					f.vCache.baseDataOffset = f.fmp4BaseDataOffset
				}
				if f.aCache.baseDataOffset < 1 {
					f.aCache.baseDataOffset = f.fmp4BaseDataOffset
				}
			}
			f.generateOneFrag()

			f.keyFrameCount = 0

			if f.vCache.lastTs-f.vCache.firstTs > int64(f.segDuration) {
				f.curTrackOffset = 0
				f.moofMdatSeqNum = 1
				// TODO gen fmp4 file
			}

			f.resetFrag(true, ts)
		} else {
			f.resetFrag(true, ts)
		}
		f.keyFrameCount++
	}

	if err = f.vCache.addFrame(frame, ts, 0); err != nil {
		return
	}

	return
}

func (f *Fmp4) generateOneFrag() (err error) {
	type pair struct {
		idx uint32
		f   func() error
	}
	genVideoPair := func() error {
		f.generateVideoMoofMdat()
		_, err := f.mdatBuf.Write(f.vCache.buf.Bytes())
		return err
	}
	genAudioPair := func() error {
		f.generateAudioMoofMdat()
		_, err := f.mdatBuf.Write(f.aCache.buf.Bytes())
		return err
	}
	pairs := []pair{
		pair{f.videoTrackId, genVideoPair},
		pair{f.audioTrackId, genAudioPair},
	}
	if f.audioTrackId <= f.videoTrackId {
		pairs[0], pairs[1] = pairs[1], pairs[0]
	}

	if pairs[0].idx > 0 {
		if err = pairs[0].f(); err != nil {
			return
		}
		if pairs[1].idx > 0 {
			if err = pairs[1].f(); err != nil {
				return
			}
		}
	}
	mdatBox := &MdatBox{
		Box: NewTypeBox(BoxTypeMDAT),
	}
	mdatBox.Data = f.mdatBuf.Bytes()
	mdatBox.Size += uint64(len(mdatBox.Data))

	moofMdat := Fmp4MoofMdat{
		Moof: f.moofBox,
		Mdat: mdatBox,
	}
	f.MoofMdat = append(f.MoofMdat, moofMdat)
	if f.audioTrackId > 0 {
		f.aCache.trunBox.DataOffset += uint32(f.moofBox.Size)
	}
	if f.videoTrackId > 0 {
		f.vCache.trunBox.DataOffset += uint32(f.moofBox.Size)
	}
	f.fmp4BaseDataOffset += (f.moofBox.Size + mdatBox.Size)
	return
}

func (f *Fmp4) generateVideoMoofMdat() {

	duration := f.vCache.lastTs - f.vCache.firstTs
	duration = (duration / int64(f.vCache.trunBox.SampleCount))
	tfhdBox := &TfhdBox{
		FullBox:               NewTypeFullBox(BoxTypeTFHD, 0, 0x39),
		TrackID:               f.vCache.trackID,
		BaseDataOffset:        f.vCache.baseDataOffset,
		DefaultSampleSize:     f.vCache.defaultSampleSize,
		DefaultSampleDuration: uint32(duration), // TODO
		DefaultSampleFlags:    0x01010000,       // TODO 没有查到什么意思，跟着ffmpeg生成的fmp4文件来的
	}
	tfhdBox.Size += 24 //flag 39

	tfdtBox := &TfdtBox{
		FullBox:             NewTypeFullBox(BoxTypeTFDT, 1, 0),
		BaseMediaDecodeTime: f.vCache.baseDecodeTime,
	}
	tfdtBox.Size += TfdtBoxBodyLenVer1

	trafBox := &TrafBox{
		Box: NewTypeBox(BoxTypeTRAF),
	}
	trafBox.SubBoxes = append(trafBox.SubBoxes, tfhdBox)
	trafBox.Size += tfhdBox.Size
	trafBox.SubBoxes = append(trafBox.SubBoxes, tfdtBox)
	trafBox.Size += tfdtBox.Size

	f.vCache.trunBox.flags24Bit = 0x205
	f.vCache.trunBox.Size += 8
	f.vCache.trunBox.FirstSampleFlags = 0x02000000
	f.vCache.trunBox.DataOffset = uint32(BOX_SIZE) + uint32(f.curTrackOffset) //uint32(f.moofBox.Size)

	trafBox.SubBoxes = append(trafBox.SubBoxes, f.vCache.trunBox)
	trafBox.Size += f.vCache.trunBox.Size

	f.moofBox.SubBoxes = append(f.moofBox.SubBoxes, trafBox)
	f.moofBox.Size += trafBox.Size

	f.curTrackOffset += f.vCache.accOffset
	return
}

func (f *Fmp4) generateAudioMoofMdat() {

	tfhdBox := &TfhdBox{
		FullBox:               NewTypeFullBox(BoxTypeTFHD, 0, 0x39),
		TrackID:               f.aCache.trackID,
		BaseDataOffset:        f.aCache.baseDataOffset,
		DefaultSampleSize:     f.aCache.defaultSampleSize,
		DefaultSampleDuration: 1024,       // aac编码规则决定的固定值
		DefaultSampleFlags:    0x02000000, // TODO 没有查到什么意思，跟着ffmpeg生成的fmp4文件来的
	}
	tfhdBox.Size += 24 //flag 39

	tfdtBox := &TfdtBox{
		FullBox:             NewTypeFullBox(BoxTypeTFDT, 1, 0),
		BaseMediaDecodeTime: f.aCache.baseDecodeTime,
	}
	tfdtBox.Size += TfdtBoxBodyLenVer1

	trafBox := &TrafBox{
		Box: NewTypeBox(BoxTypeTRAF),
	}
	trafBox.SubBoxes = append(trafBox.SubBoxes, tfhdBox)
	trafBox.Size += tfhdBox.Size
	trafBox.SubBoxes = append(trafBox.SubBoxes, tfdtBox)
	trafBox.Size += tfdtBox.Size

	f.aCache.trunBox.flags24Bit = 0x201
	f.aCache.trunBox.Size += 4
	f.aCache.trunBox.DataOffset = uint32(BOX_SIZE) + uint32(f.curTrackOffset) //uint32(f.moofBox.Size)
	trafBox.SubBoxes = append(trafBox.SubBoxes, f.aCache.trunBox)
	trafBox.Size += f.aCache.trunBox.Size

	f.moofBox.SubBoxes = append(f.moofBox.SubBoxes, trafBox)
	f.moofBox.Size += trafBox.Size

	f.curTrackOffset += f.aCache.accOffset
	return
}

func (c *MdatCache) reset(baseDataOffset uint64, baseDecodeTime uint64) {

	c.trunBox = nil
	c.buf.Reset()
	c.lastTs = 0
	c.firstTs = 0
	c.accOffset = 0
	c.defaultSampleSize = 0
	c.baseDataOffset = baseDataOffset
	c.baseDecodeTime = baseDecodeTime
	return
}

func (c *MdatCache) addFrame(frame []byte, ts int64, frameLen int) (err error) {

	if c.trunBox == nil {
		c.firstTs = ts
		c.buf.Reset()

		c.trunBox = &TrunBox{
			FullBox: NewTypeFullBox(BoxTypeTRUN, 0, 0),
		}
		c.trunBox.Size += 4 // 4 for SampleCount
	}
	c.lastTs = ts

	curWriteLen := 0
	if frameLen > 0 {
		if curWriteLen, err = byteio.WriteU32BE(&c.buf, uint32(frameLen)); err != nil {
			return
		}
		c.accOffset += uint64(curWriteLen)
	} else {
		frameLen = len(frame)
	}

	if curWriteLen, err = c.buf.Write(frame); err != nil {
		return
	}
	c.accOffset += uint64(curWriteLen)
	c.trunBox.BoxSamples = append(c.trunBox.BoxSamples, &TrunBoxSample{
		SampleSize: uint32(frameLen),
	})
	c.trunBox.Size += 4 // fullbox的flag决定
	c.trunBox.SampleCount++

	return
}

func (f *Fmp4) serialize(w io.Writer) (writedLen int, err error) {

	if writedLen, err = w.Write(f.headerBox.Bytes()); err != nil {
		return
	}

	curWriteLen := 0
	for _, mm := range f.MoofMdat {
		if curWriteLen, err = mm.Moof.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen

		if curWriteLen, err = mm.Mdat.Serialize(w); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	return
}
