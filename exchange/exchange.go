package exchange

type Void = interface{}

const (
	DataTypeAudio uint8 = iota
	DataTypeAudioConfig
	DataTypeVideo
	DataTypeVideoNonKeyFrame
	DataTypeVideoKeyFrame
	DataTypeVideoConfig
	DataTypeData
	DataTypeDataAMF0
	DataTypeDataAMF3
)

const (
	AvFormatAVC uint8 = iota
	AvFormatHEVC
	AvFormatAAC
	AvFormatData
)

const (
	ProtocolNONE uint8 = iota
	ProtocolRTMP
	ProtocolFLVLIVE
)

type ExData struct {
	Timestamp      uint64
	DataType       uint8
	AvFormat       uint8
	OriginProtocol uint8

	//flv tag的data部分，以这个为标准来交换，这样flv和rtmp就不用转换了
	Payload []byte
}
