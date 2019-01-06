package mp4

import (
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

/*
aligned(8) class HEVCDecoderConfigurationRecord {
        unsigned int(8) configurationVersion = 1;
        unsigned int(2) general_profile_space;
        unsigned int(1) general_tier_flag;
        unsigned int(5) general_profile_idc;
        unsigned int(32) general_profile_compatibility_flags;
        unsigned int(48) general_constraint_indicator_flags;
        unsigned int(8) general_level_idc;
        bit(4) reserved = 1111b;
        unsigned int(12) min_spatial_segmentation_idc;
        bit(6) reserved = 111111b;
        unsigned int(2) parallelismType;
        bit(6) reserved = 111111b;
        unsigned int(2) chromaFormat;
        bit(5) reserved = 11111b;
        unsigned int(3) bitDepthLumaMinus8;
        bit(5) reserved = 11111b;
        unsigned int(3) bitDepthChromaMinus8;
        bit(16) avgFrameRate;
        bit(2) constantFrameRate;
        bit(3) numTemporalLayers;
        bit(1) temporalIdNested;
        unsigned int(2) lengthSizeMinusOne;
        unsigned int(8) numOfArrays;
        for (j=0; j < numOfArrays; j++) {
                bit(1) array_completeness;
                unsigned int(1) reserved = 0;
                unsigned int(6) NAL_unit_type;
                unsigned int(16) numNalus;
                for (i=0; i< numNalus; i++) {
                        unsigned int(16) nalUnitLength;
                        bit(8*nalUnitLength) nalUnit;
                }
        }
}
*/
type HevcConfigNalu struct {
	NaluLength uint16
	Nalu       []byte
}
type HevcArrayItem struct {
	ArrayCompleteness1Bit uint8
	Reserved1Bit          uint8
	NalType6Bit           uint8
	NumNalus              uint16
	Nalus                 []*HevcConfigNalu
}
type HevcDecoderConfigurationRecord struct {
	ConfigurationVersion                 uint8
	GeneralProfileSpace2Bit              uint8
	GeneralTierGlag1Bit                  uint8
	GeneralProfileIdc5Bit                uint8
	GeneralProfileCompatibilityFlags     uint32
	GeneralConstraintIndicatorFlags48Bit uint64
	GeneralLevelIdc                      uint8
	Reserve4Bit1                         uint8
	MinSpatialSegmentationIdc12Bit       uint16
	Reserve6Bit2                         uint8
	ParallelismType2Bit                  uint8
	Reserve6Bit3                         uint8
	ChromaFormat2Bit                     uint8
	Reserve5Bit4                         uint8
	BitDepthLumaMinus83Bit               uint8
	Reserve5Bit5                         uint8
	BitDepthChromaMinus83Bit             uint8
	AvgFrameRate                         uint16 // bit(16)?
	ConstantFrameRate2Bit                uint8
	NumTemporalLayers3Bit                uint8
	TemporalIdNested1Bit                 uint8
	LengthSizeMinusOne2Bit               uint8
	NumOfArrays                          uint8
	Items                                []*HevcArrayItem
}

/*
aligned(8) class AVCDecoderConfigurationRecord {
        unsigned int(8) configurationVersion = 1;
        unsigned int(8) AVCProfileIndication;
        unsigned int(8) profile_compatibility;
        unsigned int(8) AVCLevelIndication;
        bit(6) reserved = 111111b;
        unsigned int(2) lengthSizeMinusOne;
        bit(3) reserved = 111b;
        unsigned int(5) numOfSequenceParameterSets;
        for (i=0; i< numOfSequenceParameterSets;  i++) {
                unsigned int(16) sequenceParameterSetLength ;
                bit(8*sequenceParameterSetLength) sequenceParameterSetNALUnit;
        }
        unsigned int(8) numOfPictureParameterSets;
        for (i=0; i< numOfPictureParameterSets;  i++) {
                unsigned int(16) pictureParameterSetLength;
                bit(8*pictureParameterSetLength) pictureParameterSetNALUnit;
        }
		// prifile_idc应该就是AVCLevelIndication
        if( profile_idc  ==  100  ||  profile_idc  ==  110  ||
            profile_idc  ==  122  ||  profile_idc  ==  144 )
        {
                bit(6) reserved = 111111b;
                unsigned int(2) chroma_format;
                bit(5) reserved = 11111b;
                unsigned int(3) bit_depth_luma_minus8;
                bit(5) reserved = 11111b;
                unsigned int(3) bit_depth_chroma_minus8;
                unsigned int(8) numOfSequenceParameterSetExt;
                for (i=0; i< numOfSequenceParameterSetExt; i++) {
                        unsigned int(16) sequenceParameterSetExtLength;
                        bit(8*sequenceParameterSetExtLength) sequenceParameterSetExtNALUnit;
                }
        }
}
*/

type AVCSpsNalu struct {
	Length  uint16
	SpsNalu []byte
}

type AVCSpsExtNalu struct {
	Length     uint16
	SpsExtNalu []byte
}

type AVCPpsNalu struct {
	Length  uint16
	PpsNalu []byte
}

type AVCDecoderConfigurationRecord struct {
	ConfigurationVersion           uint8
	AVCProfileIndication           uint8
	ProfileCompatibility           uint8
	AVCLevelIndication             uint8
	Reserved6Bit1                  uint8 //111111b
	LengthSizeMinusOne2Bit         uint8
	Reserved3Bit2                  uint8 //111b
	NumOfSequenceParameterSets5Bit uint8
	Sps                            []*AVCSpsNalu
	NumOfPictureParameterSets      uint8
	Pps                            []*AVCPpsNalu
	//if AVCLevelIndication in (100, 110, 122, 144)
	Reserved6Bit3                uint8 // 111111b;
	ChromaFormat                 uint8
	Reserved5Bit4                uint8 // 11111b
	BitDepthLumaMinus83Bit       uint8
	Reserved5Bit5                uint8 // 11111b
	BitDepthChromaMinus83Bit     uint8
	NumOfSequenceParameterSetExt uint8
	SpsExt                       []*AVCSpsExtNalu
}

func NewAVCDecoderConfigurationRecord() *AVCDecoderConfigurationRecord {
	return &AVCDecoderConfigurationRecord{}
}

func (c *AVCDecoderConfigurationRecord) Serialize(w io.Writer) (writedLen int, err error) {

	nums := []uint8{c.ConfigurationVersion, c.AVCProfileIndication, c.ProfileCompatibility, c.AVCLevelIndication}
	if writedLen, err = w.Write(nums); err != nil {
		return
	}

	curWriteLen := 0
	nums[0] = c.Reserved6Bit1<<2 | c.LengthSizeMinusOne2Bit
	nums[1] = c.Reserved3Bit2<<5 | c.NumOfSequenceParameterSets5Bit

	if curWriteLen, err = w.Write(nums[0:2]); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(c.Sps); i++ {
		nums[0] = uint8(c.Sps[i].Length >> 8)
		nums[1] = uint8(c.Sps[i].Length)
		if curWriteLen, err = w.Write(nums[0:2]); err != nil {
			return
		}
		writedLen += curWriteLen
		if curWriteLen, err = w.Write(c.Sps[i].SpsNalu); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	nums[0] = c.NumOfPictureParameterSets
	if curWriteLen, err = w.Write(nums[0:1]); err != nil {
		return
	}
	writedLen += curWriteLen

	for i := 0; i < len(c.Sps); i++ {
		nums[0] = uint8(c.Pps[i].Length >> 8)
		nums[1] = uint8(c.Pps[i].Length)
		if curWriteLen, err = w.Write(nums[0:2]); err != nil {
			return
		}
		writedLen += curWriteLen
		if curWriteLen, err = w.Write(c.Pps[i].PpsNalu); err != nil {
			return
		}
		writedLen += curWriteLen
	}

	if c.AVCLevelIndication == 100 || c.AVCLevelIndication == 110 ||
		c.AVCLevelIndication == 122 || c.AVCLevelIndication == 144 {
		nums[0] = c.Reserved6Bit3<<2 | c.ChromaFormat
		nums[1] = c.Reserved5Bit4 | c.BitDepthLumaMinus83Bit
		nums[2] = c.Reserved5Bit5 | c.BitDepthChromaMinus83Bit
		nums[3] = c.NumOfSequenceParameterSetExt
		if curWriteLen, err = w.Write(nums); err != nil {
			return
		}
		writedLen += curWriteLen

		for i := 0; i < len(c.SpsExt); i++ {
			nums[0] = uint8(c.SpsExt[i].Length >> 8)
			nums[1] = uint8(c.SpsExt[i].Length)
			if curWriteLen, err = w.Write(nums[0:2]); err != nil {
				return
			}
			writedLen += curWriteLen
			if curWriteLen, err = w.Write(c.SpsExt[i].SpsExtNalu); err != nil {
				return
			}
			writedLen += curWriteLen
		}
	}

	return
}

func (c *AVCDecoderConfigurationRecord) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 6)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	if buf[0] != 0x01 {
		err = fmt.Errorf("ConfigurationVersion:%d", buf[0])
		return
	}
	c.ConfigurationVersion = 0x01

	c.AVCProfileIndication = buf[1]
	c.ProfileCompatibility = buf[2]
	c.AVCLevelIndication = buf[3]

	if buf[4]&0xFC != 0xFC {
		err = fmt.Errorf("resver1 :%d", buf[4]&0xFC)
		return
	}
	c.Reserved6Bit1 = 0x3F

	c.LengthSizeMinusOne2Bit = buf[4] & 0x03

	reserve2 := (buf[5] & 0xE0) >> 5
	if reserve2 != 7 {
		err = fmt.Errorf("reserve2 :%d", reserve2)
		return
	}
	c.Reserved3Bit2 = 7

	c.NumOfSequenceParameterSets5Bit = buf[5] & 0x1F
	curReadLen := 0
	for i := uint8(0); i < c.NumOfSequenceParameterSets5Bit; i++ {
		if curReadLen, err = io.ReadFull(r, buf[0:2]); err != nil {
			return
		}
		totalReadLen += curReadLen

		sps := &AVCSpsNalu{}
		sps.Length = byteio.U16BE(buf)
		sps.SpsNalu = make([]byte, sps.Length)
		if curReadLen, err = io.ReadFull(r, sps.SpsNalu); err != nil {
			return
		}
		totalReadLen += curReadLen

		c.Sps = append(c.Sps, sps)
	}

	if curReadLen, err = io.ReadFull(r, buf[0:1]); err != nil {
		return
	}
	totalReadLen += curReadLen

	c.NumOfPictureParameterSets = buf[0]
	for i := uint8(0); i < c.NumOfPictureParameterSets; i++ {
		if curReadLen, err = io.ReadFull(r, buf[0:2]); err != nil {
			return
		}
		totalReadLen += curReadLen

		pps := &AVCPpsNalu{}
		pps.Length = byteio.U16BE(buf)
		pps.PpsNalu = make([]byte, pps.Length)
		if curReadLen, err = io.ReadFull(r, pps.PpsNalu); err != nil {
			return
		}
		totalReadLen += curReadLen

		c.Pps = append(c.Pps, pps)
	}

	if c.AVCLevelIndication == 100 || c.AVCLevelIndication == 110 ||
		c.AVCLevelIndication == 122 || c.AVCLevelIndication == 144 {

		if curReadLen, err = io.ReadFull(r, buf[0:4]); err != nil {
			return
		}
		totalReadLen += curReadLen

		if buf[0]&0xFC != 0xFC {
			err = fmt.Errorf("resver1 :%d", buf[4]&0xFC)
			return
		}
		c.Reserved6Bit3 = 0x3F
		c.ChromaFormat = buf[0] & 0x03

		c.BitDepthLumaMinus83Bit = (buf[1] & 0xC0) >> 5
		if buf[1]&0x1F != 0x1F {
			err = fmt.Errorf("reserved:%d", buf[1]&0x1F)
			return
		}
		c.Reserved5Bit4 = 0x1F

		c.BitDepthChromaMinus83Bit = (buf[2] & 0xC0) >> 5
		if buf[2]&0x1F != 0x1F {
			err = fmt.Errorf("reserved:%d", buf[2]&0x1F)
			return
		}
		c.Reserved5Bit5 = 0x1F

		c.NumOfSequenceParameterSetExt = buf[3]
		for i := uint8(0); i < c.NumOfSequenceParameterSetExt; i++ {
			if curReadLen, err = io.ReadFull(r, buf[0:2]); err != nil {
				return
			}
			totalReadLen += curReadLen

			spsExt := &AVCSpsExtNalu{}
			spsExt.Length = byteio.U16BE(buf)
			spsExt.SpsExtNalu = make([]byte, spsExt.Length)
			if curReadLen, err = io.ReadFull(r, spsExt.SpsExtNalu); err != nil {
				return
			}
			totalReadLen += curReadLen

			c.SpsExt = append(c.SpsExt, spsExt)
		}
	}

	return
}

func NewHevcDecoderConfigurationRecord() *HevcDecoderConfigurationRecord {
	return &HevcDecoderConfigurationRecord{}
}

func (c *HevcDecoderConfigurationRecord) Parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 23)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}
	if buf[0] != 0x01 {
		err = fmt.Errorf("ConfigurationVersion:%d", buf[0])
		return
	}
	c.ConfigurationVersion = 0x01

	c.GeneralProfileSpace2Bit = (buf[1] & 0xC0) >> 6
	c.GeneralTierGlag1Bit = (buf[1] & 0x20) >> 5
	c.GeneralProfileIdc5Bit = buf[1] & 0x1F

	c.GeneralProfileCompatibilityFlags = byteio.U32BE(buf[2:6])
	c.GeneralConstraintIndicatorFlags48Bit = byteio.U48BE(buf[6:12])
	c.GeneralLevelIdc = buf[12]

	if buf[13]&0xF0 != 0xF0 {
		err = fmt.Errorf("reserved 4bit:%d", buf[13]&0xF0)
	}
	c.Reserve4Bit1 = 0x0F
	c.MinSpatialSegmentationIdc12Bit = uint16((buf[13]&0x0F))*256 + uint16(buf[14])

	if buf[15]&0xFC != 0xFC {
		err = fmt.Errorf("reserved 6bit:%d", buf[15]&0xFC)
	}
	c.Reserve6Bit2 = 0x3F
	c.ParallelismType2Bit = buf[15] & 0x03

	if buf[16]&0xFC != 0xFC {
		err = fmt.Errorf("reserved 6bit:%d", buf[16]&0xFC)
	}
	c.Reserve6Bit3 = 0x3F
	c.ChromaFormat2Bit = buf[16] & 0x03

	if buf[17]&0xF8 != 0xF8 {
		err = fmt.Errorf("reserved 5bit:%d", buf[17]&0xF8)
	}
	c.Reserve5Bit4 = 0x1F
	c.BitDepthLumaMinus83Bit = buf[17] & 0x07

	if buf[18]&0xF8 != 0xF8 {
		err = fmt.Errorf("reserved 5bit:%d", buf[18]&0xF8)
	}
	c.Reserve5Bit5 = 0x1F
	c.BitDepthChromaMinus83Bit = buf[18] & 0x07

	c.AvgFrameRate = byteio.U16BE(buf[19:21])

	c.ConstantFrameRate2Bit = (buf[21] & 0xC0) >> 6
	c.NumTemporalLayers3Bit = (buf[21] & 0x38) >> 3
	c.TemporalIdNested1Bit = (buf[21] & 0x04) >> 2
	c.LengthSizeMinusOne2Bit = buf[21] & 0x03

	c.NumOfArrays = buf[22]

	curReadLen := 0
	for i := uint8(0); i < c.NumOfArrays; i++ {
		item := &HevcArrayItem{}
		if curReadLen, err = item.parse(r); err != nil {
			return
		}
		totalReadLen += curReadLen

		c.Items = append(c.Items, item)
	}
	return
}

func (item *HevcArrayItem) parse(r io.Reader) (totalReadLen int, err error) {

	buf := make([]byte, 3)
	if totalReadLen, err = io.ReadFull(r, buf); err != nil {
		return
	}

	item.ArrayCompleteness1Bit = (buf[0] & 0x80) >> 7
	item.Reserved1Bit = (buf[0] & 0x40) >> 6
	item.NalType6Bit = buf[0] & 0x3F
	item.NumNalus = byteio.U16BE(buf[1:3])

	buf = buf[0:2]
	curReadLen := 0
	for i := uint16(0); i < item.NumNalus; i++ {
		if curReadLen, err = io.ReadFull(r, buf); err != nil {
			return
		}
		totalReadLen += curReadLen

		nalu := &HevcConfigNalu{}
		nalu.NaluLength = byteio.U16BE(buf)
		nalu.Nalu = make([]byte, nalu.NaluLength)

		if curReadLen, err = io.ReadFull(r, nalu.Nalu); err != nil {
			return
		}
		totalReadLen += curReadLen

		item.Nalus = append(item.Nalus, nalu)
	}

	return
}

func (c *HevcDecoderConfigurationRecord) Serialize(w io.Writer) (writedLen int, err error) {
	buf := make([]byte, 23)
	buf[0] = c.ConfigurationVersion
	buf[1] = c.GeneralProfileSpace2Bit<<6 | c.GeneralTierGlag1Bit<<5 | c.GeneralProfileIdc5Bit
	byteio.PutU32BE(buf[2:6], c.GeneralProfileCompatibilityFlags)
	byteio.PutU48BE(buf[6:12], c.GeneralConstraintIndicatorFlags48Bit)
	buf[12] = c.GeneralLevelIdc

	buf[13] = c.Reserve4Bit1<<4 | byte(c.MinSpatialSegmentationIdc12Bit>>8)
	buf[14] = byte(c.MinSpatialSegmentationIdc12Bit)

	buf[15] = c.Reserve6Bit2<<2 | c.ParallelismType2Bit
	buf[16] = c.Reserve6Bit3<<2 | c.ChromaFormat2Bit
	buf[17] = c.Reserve5Bit4<<3 | c.BitDepthLumaMinus83Bit
	buf[18] = c.Reserve5Bit5<<3 | c.BitDepthChromaMinus83Bit

	buf[19] = byte(c.AvgFrameRate >> 8)
	buf[20] = byte(c.AvgFrameRate)

	buf[21] = c.ConstantFrameRate2Bit<<6 | c.NumTemporalLayers3Bit<<3 | c.TemporalIdNested1Bit<<2 | c.LengthSizeMinusOne2Bit

	buf[22] = c.NumOfArrays

	if writedLen, err = w.Write(buf); err != nil {
		return
	}

	if int(c.NumOfArrays) != len(c.Items) {
		err = fmt.Errorf("not consistent item:%d %d", c.NumOfArrays, len(c.Items))
		return
	}

	curWriteLen := 0
	for i := 0; i < len(c.Items); i++ {

		if int(c.Items[i].NumNalus) != len(c.Items[i].Nalus) {
			err = fmt.Errorf("not consistent nalu:%d %d", c.Items[i].NumNalus, len(c.Items[i].Nalus))
			return
		}

		buf[0] = c.Items[i].ArrayCompleteness1Bit<<7 | c.Items[i].Reserved1Bit<<6 | c.Items[i].NalType6Bit
		buf[1] = byte(c.Items[i].NumNalus >> 8)
		buf[2] = byte(c.Items[i].NumNalus)

		if curWriteLen, err = w.Write(buf[0:3]); err != nil {
			return
		}
		writedLen += curWriteLen

		for j := 0; j < len(c.Items[i].Nalus); j++ {
			buf[0] = uint8(c.Items[i].Nalus[j].NaluLength >> 8)
			buf[1] = uint8(c.Items[i].Nalus[j].NaluLength)
			if curWriteLen, err = w.Write(buf[0:2]); err != nil {
				return
			}
			writedLen += curWriteLen
			if curWriteLen, err = w.Write(c.Items[i].Nalus[j].Nalu); err != nil {
				return
			}
			writedLen += curWriteLen
		}
	}

	return
}
