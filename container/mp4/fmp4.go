package mp4

import (
	"fmt"
	"time"
)

/*
fmp4 moovbox很多多不需要了，分析ffmpeg的生成的fmp4文件，里面还有很多moov下很多子box都还在，但是都没有值
moof/traf/tfhd.base_data_offset 看起来就是moof的偏移
moof/traf/trun.data_offset 用来表示和该 moof 配套的 mdat 中实际数据内容距 moof 开头有多远,其实值就是: moof.Size+BOX_SIZE(8）
moof/traf/trun.sample.sample_size就可以定位帧的位置了。这个长度的起始位置就是mdat的数据起始位置，然后累计sample_size作为偏移就是
下一个sample的开始了
19040101到19700101的秒数为2082844800
*/

const (
	Diff1970To1904 = 2082844800
)

type Fmp4MoofMdat struct {
	Moof MoofBox
	Mdat MdatBox
}

type Fmp4 struct {
	Ftyp         FtypBox
	Moov         MoovBox
	MoofMdat     []*Fmp4MoofMdat
	Mfra         MfraBox
	nextTrackId  uint32
	audioTrackId uint32
	videoTrackId uint32
	cmTime       uint64
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

func NewFmp4() *Fmp4 {

	nowSec := uint64(time.Now().Unix())

	mvhdBox := &MvhdBox{
		FullBox:          NewTypeFullBox(BoxTypeMVHD, 0, 0),
		CreationTime:     nowSec + Diff1970To1904,
		ModificationTime: nowSec + Diff1970To1904,
		Timescale:        1000,
		TemplateRate:     0x00010000,
		TemplateVolume:   0x0100,
		NextTrackID:      2, // TODO
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
		nextTrackId: 1,
		cmTime:      uint64(time.Now().Unix()) + Diff1970To1904,
	}
	f.Moov.Size += mvhdBox.Size
	f.Ftyp.Size += 8

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

func newFmp4MinfBox(ibox IBox) *MinfBox {

	dinfBox := newFmp4DinfBox()

	minfBox := &MinfBox{
		Box: NewTypeBox(BoxTypeMINF),
		SubBoxes: []IBox{
			ibox,
			dinfBox,
		},
	}

	minfBox.Size += ibox.GetBoxSize()
	minfBox.Size += dinfBox.Size

	return minfBox
}

func newFmp4AudioMinfBox(ibox IBox) *MinfBox {
	smhdBox := &SmhdBox{
		FullBox: NewTypeFullBox(BoxTypeSMHD, 0, 0),
	}
	smhdBox.Size += SmhdBoxBodyLen

	return newFmp4MinfBox(smhdBox)
}

func newFmp4VideoMinfBox(ibox IBox) *MinfBox {
	vmhdBox := &VmhdBox{
		FullBox: NewTypeFullBox(BoxTypeVMHD, 0, 1),
	}
	vmhdBox.Size += VmhdBoxBodyLen

	return newFmp4MinfBox(vmhdBox)
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

func (f *Fmp4) AddAudioTrack() error {

	if f.audioTrackId != 0 {
		return fmt.Errorf("audio trackid already exists")
	}

	tkhdBox := &TkhdBox{
		FullBox:          NewTypeFullBox(BoxTypeTKHD, 0, 3),
		CreationTime:     f.cmTime,
		ModificationTime: f.cmTime,
		TrackID:          f.nextTrackId,
		TemplateVolume:   0x0100,
	}
	f.nextTrackId++

	trackBox := &TrakBox{
		Box: NewTypeBox(BoxTypeTRAK),
		SubBoxes: []IBox{
			tkhdBox,
			&MdiaBox{
				Box: NewTypeBox(BoxTypeMDIA),
				SubBoxes: []IBox{
					&MdhdBox{
						FullBox: NewTypeFullBox(BoxTypeMDHD, 0, 0),
					},
				},
			},
		},
	}

	f.Moov.SubBoxes = append(f.Moov.SubBoxes, trackBox)

	return nil
}
