package flvlive

import (
	"fmt"
	"io"

	"github.com/chinasarft/golive/container/flv"
	"github.com/chinasarft/golive/exchange"
)

func GetNextExData(r io.Reader) (d *exchange.ExData, err error) {
	var tag *flv.FlvTag
	if tag, err = flv.ParseTag(r); err != nil {
		return
	}

	switch tag.TagType {
	case flv.FlvTagAudio:
		d, err = getAudioExData(tag)
	case flv.FlvTagVideo:
		d, err = getVideoExData(tag)
	case flv.FlvTagAMF0:
		d, err = getScriptExData(tag, exchange.DataTypeDataAMF0)
	case flv.FlvTagAMF3:
		d, err = getScriptExData(tag, exchange.DataTypeDataAMF3)
	default:
		err = fmt.Errorf("not supported flv type:%d", d.DataType)
	}

	return
}

func getAudioExData(tag *flv.FlvTag) (d *exchange.ExData, err error) {
	afmt := (uint8(tag.Data[0]) & 0xF0) >> 4
	if afmt != 10 {
		err = fmt.Errorf("just support aac:%d", afmt)
		return
	}

	aType := exchange.DataTypeAudio
	if uint8(tag.Data[1]) == 0 {
		aType = exchange.DataTypeAudioConfig
	}

	d = &exchange.ExData{
		Timestamp:      uint64(tag.Timestamp),
		DataType:       aType,
		AvFormat:       exchange.AvFormatAAC,
		OriginProtocol: exchange.ProtocolFLVLIVE,
		Payload:        tag.Data,
	}

	return
}

func getVideoExData(tag *flv.FlvTag) (d *exchange.ExData, err error) {
	vfmt := (uint8(tag.Data[0]) & 0xF0) >> 4
	vCodecID := uint8(tag.Data[0]) & 0x0F
	if vCodecID != 7 { // 7 for avc
		err = fmt.Errorf("just support avc:%d", vCodecID)
		return
	}

	vType := exchange.DataTypeVideoNonKeyFrame
	if uint8(tag.Data[1]) == 0 {
		vType = exchange.DataTypeVideoConfig
	} else if uint8(tag.Data[1]) == 1 && vfmt == 1 {
		vType = exchange.DataTypeVideoKeyFrame
	}

	d = &exchange.ExData{
		Timestamp:      uint64(tag.Timestamp),
		DataType:       vType,
		AvFormat:       exchange.AvFormatAVC,
		OriginProtocol: exchange.ProtocolFLVLIVE,
		Payload:        tag.Data,
	}

	return
}

func getScriptExData(tag *flv.FlvTag, dataType uint8) (d *exchange.ExData, err error) {

	d = &exchange.ExData{
		Timestamp:      uint64(tag.Timestamp),
		DataType:       dataType,
		AvFormat:       exchange.AvFormatData,
		OriginProtocol: exchange.ProtocolFLVLIVE,
		Payload:        tag.Data,
	}

	return
}
