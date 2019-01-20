package mp4

const (
	BoxTypeForbidden = 0x00

	BoxTypeUUID = 0x75756964 // 'uuid'
	BoxTypeSTYP = 0x73747970 // 'styp'
	BoxTypeFTYP = 0x66747970 // 'ftyp'
	BoxTypeMDAT = 0x6d646174 // 'mdat'
	BoxTypeSIDX = 0x73696478 // 'sidx'

	BoxTypeMOOF = 0x6d6f6f66 // 'moof'
	BoxTypeMFHD = 0x6d666864 // '--mfhd'
	BoxTypeTRAF = 0x74726166 // '--traf'
	BoxTypeTFHD = 0x74666864 // '----tfhd'
	BoxTypeTFDT = 0x74666474 // '----tfdt'
	BoxTypeTRUN = 0x7472756e // '----trun'

	BoxTypeMFRA = 0x6d667261 // 'mfra'
	BoxTypeTFRA = 0x74667261 // '--tfra'
	BoxTypeMFRO = 0x6d66726f // '--mfro'

	BoxTypeFREE = 0x66726565 // 'free'
	BoxTypeSKIP = 0x736b6970 // 'skip'

	BoxTypeMOOV = 0x6d6f6f76 // 'moov'
	BoxTypeMVEX = 0x6d766578 // '--mvex'
	BoxTypeTREX = 0x74726578 // '----trex'
	BoxTypeMVHD = 0x6d766864 // '--mvhd'
	BoxTypeTRAK = 0x7472616b // '--trak'
	BoxTypeUDTA = 0x75647461 // '----udta'
	BoxTypeTKHD = 0x746b6864 // '----tkhd'
	BoxTypeEDTS = 0x65647473 // '----edts'
	BoxTypeELST = 0x656c7374 // '------elst'
	BoxTypeMDIA = 0x6d646961 // '----mdia'
	BoxTypeMDHD = 0x6d646864 // '------mdhd'
	BoxTypeHDLR = 0x68646c72 // '------hdlr'
	BoxTypeMINF = 0x6d696e66 // '------minf'
	BoxTypeVMHD = 0x766d6864 // '--------vmhd'
	BoxTypeSMHD = 0x736d6864 // '--------smhd'
	BoxTypeDINF = 0x64696e66 // '--------dinf'
	BoxTypeDREF = 0x64726566 // '----------dref'
	BoxTypeURL  = 0x75726c20 // '------------url '
	BoxTypeURN  = 0x75726e20 // '------------urn '
	BoxTypeSTBL = 0x7374626c // '--------stbl'
	BoxTypeSTSD = 0x73747364 // '----------stsd'
	BoxTypeAVC1 = 0x61766331 // '------------avc1'
	BoxTypeHEV1 = 0x68657631 // '------------hev1'
	BoxTypeSTTS = 0x73747473 // '----------stts'
	BoxTypeCTTS = 0x63747473 // '----------ctts'
	BoxTypeSTSS = 0x73747373 // '----------stss'
	BoxTypeSTSC = 0x73747363 // '----------stsc'
	BoxTypeSTCO = 0x7374636f // '----------stco'
	BoxTypeSTSZ = 0x7374737a // '----------stsz'
	BoxTypeSTZ2 = 0x73747a32 // '----------stz2'

	BoxTypePASP = 0x70617370

	BoxTypeMETA = 0x6d657461 // '-meta'
	BoxTypeILST = 0x696c7374 // '-ilst'

	BoxTypeCO64 = 0x636f3634 // 'co64'

	BoxTypeAVCC = 0x61766343 // 'avcC'
	BoxTypeHVCC = 0x68766343 // 'hvcC'
	BoxTypeMP4A = 0x6d703461 // 'mp4a'
	BoxTypeESDS = 0x65736473 // 'esds'

	Mp4BoxBrandForbidden = 0x00
	Mp4BoxBrandISOM      = 0x69736f6d // 'isom'
	Mp4BoxBrandISO2      = 0x69736f32 // 'iso2'
	Mp4BoxBrandISO6      = 0x69736f36 // 'iso6'
	Mp4BoxBrandAVC1      = 0x61766331 // 'avc1'
	Mp4BoxBrandMP41      = 0x6d703431 // 'mp41'

	VideoHandlerType = 0x76797065 //'vide'
	AudioHandlerType = 0x736f756e //'soun'

)

const (
	MvhdBoxBodyLenVer0   = 96
	MvhdBoxBodyLenVer1   = 108
	TkhdBoxBodyLenVer0   = 80
	TkhdBoxBodyLenVer1   = 92
	SmhdBoxBodyLen       = 4
	VmhdBoxBodyLen       = 8
	MdhdBoxBodyLenVer0   = 20
	MdhdBoxBodyLenVer1   = 32
	PaspBoxBodyLen       = 8
	SampleEntryLen       = 8
	VisualSampleEntryLen = 70
	AudioSampleEntryLen  = 20
	TrexBoxBodyLen       = 20
	MfhdBoxBodyLen       = 4
	TfdtBoxBodyLenVer0   = 4
	TfdtBoxBodyLenVer1   = 8
)

/**
 * The video codec id.
 * @doc video_file_format_spec_v10_1.pdf, page78, E.4.3.1 VIDEODATA
 * CodecID UB [4]
 * Codec Identifier. The following values are defined for FLV:
 *      2 = Sorenson H.263
 *      3 = Screen video
 *      4 = On2 VP6
 *      5 = On2 VP6 with alpha channel
 *      6 = Screen video version 2
 *      7 = AVC
 */
const (
	// set to the zero to reserved, for array map.
	VideoCodecIdReserved  = 0
	VideoCodecIdForbidden = 0
	VideoCodecIdReserved1 = 1
	VideoCodecIdReserved2 = 9

	// for user to disable video, for example, use pure audio hls.
	VideoCodecIdDisabled = 8

	VideoCodecIdSorensonH263           = 2
	VideoCodecIdScreenVideo            = 3
	VideoCodecIdOn2VP6                 = 4
	VideoCodecIdOn2VP6WithAlphaChannel = 5
	VideoCodecIdScreenVideoVersion2    = 6
	VideoCodecIdAVC                    = 7
)

/**
 * The audio codec id.
 * @doc video_file_format_spec_v10_1.pdf, page 76, E.4.2 Audio Tags
 * SoundFormat UB [4]
 * Format of SoundData. The following values are defined:
 *     0 = Linear PCM, platform endian
 *     1 = ADPCM
 *     2 = MP3
 *     3 = Linear PCM, little endian
 *     4 = Nellymoser 16 kHz mono
 *     5 = Nellymoser 8 kHz mono
 *     6 = Nellymoser
 *     7 = G.711 A-law logarithmic PCM
 *     8 = G.711 mu-law logarithmic PCM
 *     9 = reserved
 *     10 = AAC
 *     11 = Speex
 *     14 = MP3 8 kHz
 *     15 = Device-specific sound
 * Formats 7, 8, 14, and 15 are reserved.
 * AAC is supported in Flash Player 9,0,115,0 and higher.
 * Speex is supported in Flash Player 10 and higher.
 */
const (
	// set to the max value to reserved, for array map.
	AudioCodecIdReserved1 = 16
	AudioCodecIdForbidden = 16

	// for user to disable audio, for example, use pure video hls.
	AudioCodecIdDisabled = 17

	AudioCodecIdLinearPCMPlatformEndian         = 0
	AudioCodecIdADPCM                           = 1
	AudioCodecIdMP3                             = 2
	AudioCodecIdLinearPCMLittleEndian           = 3
	AudioCodecIdNellymoser16kHzMono             = 4
	AudioCodecIdNellymoser8kHzMono              = 5
	AudioCodecIdNellymoser                      = 6
	AudioCodecIdReservedG711AlawLogarithmicPCM  = 7
	AudioCodecIdReservedG711MuLawLogarithmicPCM = 8
	AudioCodecIdReserved                        = 9
	AudioCodecIdAAC                             = 10
	AudioCodecIdSpeex                           = 11
	AudioCodecIdReservedMP3_8kHz                = 14
	AudioCodecIdReservedDeviceSpecificSound     = 15
)

/**
 * The audio sample rate.
 * @see srs_flv_srates and srs_aac_srates.
 * @doc video_file_format_spec_v10_1.pdf, page 76, E.4.2 Audio Tags
 *      0 = 5.5 kHz = 5512 Hz
 *      1 = 11 kHz = 11025 Hz
 *      2 = 22 kHz = 22050 Hz
 *      3 = 44 kHz = 44100 Hz
 * However, we can extends this table.
 */
const (
	// set to the max value to reserved, for array map.
	AudioSampleRateReserved  = 4
	AudioSampleRateForbidden = 4

	AudioSampleRate5512  = 0
	AudioSampleRate11025 = 1
	AudioSampleRate22050 = 2
	AudioSampleRate44100 = 3
)

/**
 * The frame type, for example, audio, video or data.
 * @doc video_file_format_spec_v10_1.pdf, page 75, E.4.1 FLV Tag
 */
const (
	// set to the zero to reserved, for array map.
	FrameTypeReserved  = 0
	FrameTypeForbidden = 0

	// 8 = audio
	FrameTypeAudio = 8
	// 9 = video
	FrameTypeVideo = 9
	// 18 = script data
	FrameTypeScript = 18
)

/**
 * The audio sample size in bits.
 * @doc video_file_format_spec_v10_1.pdf, page 76, E.4.2 Audio Tags
 * Size of each audio sample. This parameter only pertains to
 * uncompressed formats. Compressed formats always decode
 * to 16 bits internally.
 *      0 = 8-bit samples
 *      1 = 16-bit samples
 */
const (
	// set to the max value to reserved, for array map.
	AudioSampleBitsReserved  = 2
	AudioSampleBitsForbidden = 2

	AudioSampleBits8bit  = 0
	AudioSampleBits16bit = 1
)

/**
 * The audio channels.
 * @doc video_file_format_spec_v10_1.pdf, page 77, E.4.2 Audio Tags
 * Mono or stereo sound
 *      0 = Mono sound
 *      1 = Stereo sound
 */
const (
	// set to the max value to reserved, for array map.
	AudioChannelsReserved  = 2
	AudioChannelsForbidden = 2

	AudioChannelsMono   = 0
	AudioChannelsStereo = 1
)

/**
 * Table 7-1 - NAL unit type codes, syntax element categories, and NAL unit type classes
 * ISO_IEC_14496-10-AVC-2012.pdf, page 83.
 */
const (
	// Unspecified
	AvcNaluTypeReserved  = 0
	AvcNaluTypeForbidden = 0

	// Coded slice of a non-IDR picture slice_layer_without_partitioning_rbsp( )
	AvcNaluTypeNonIDR = 1
	// Coded slice data partition A slice_data_partition_a_layer_rbsp( )
	AvcNaluTypeDataPartitionA = 2
	// Coded slice data partition B slice_data_partition_b_layer_rbsp( )
	AvcNaluTypeDataPartitionB = 3
	// Coded slice data partition C slice_data_partition_c_layer_rbsp( )
	AvcNaluTypeDataPartitionC = 4
	// Coded slice of an IDR picture slice_layer_without_partitioning_rbsp( )
	AvcNaluTypeIDR = 5
	// Supplemental enhancement information (SEI) sei_rbsp( )
	AvcNaluTypeSEI = 6
	// Sequence parameter set seq_parameter_set_rbsp( )
	AvcNaluTypeSPS = 7
	// Picture parameter set pic_parameter_set_rbsp( )
	AvcNaluTypePPS = 8
	// Access unit delimiter access_unit_delimiter_rbsp( )
	AvcNaluTypeAccessUnitDelimiter = 9
	// End of sequence end_of_seq_rbsp( )
	AvcNaluTypeEOSequence = 10
	// End of stream end_of_stream_rbsp( )
	AvcNaluTypeEOStream = 11
	// Filler data filler_data_rbsp( )
	AvcNaluTypeFilterData = 12
	// Sequence parameter set extension seq_parameter_set_extension_rbsp( )
	AvcNaluTypeSPSExt = 13
	// Prefix NAL unit prefix_nal_unit_rbsp( )
	AvcNaluTypePrefixNALU = 14
	// Subset sequence parameter set subset_seq_parameter_set_rbsp( )
	AvcNaluTypeSubsetSPS = 15
	// Coded slice of an auxiliary coded picture without partitioning slice_layer_without_partitioning_rbsp( )
	AvcNaluTypeLayerWithoutPartition = 19
	// Coded slice extension slice_layer_extension_rbsp( )
	AvcNaluTypeCodedSliceExt = 20
)

/**
 * the aac profile, for ADTS(HLS/TS)
 * @see https://github.com/ossrs/srs/issues/310
 */
const (
	AacProfileReserved = 3

	// @see 7.1 Profiles, aac-iso-13818-7.pdf, page 40
	AacProfileMain = 0
	AacProfileLC   = 1
	AacProfileSSR  = 2
)

/**
 * the level for avc/h.264.
 * @see Annex A Profiles and levels, ISO_IEC_14496-10-AVC-2003.pdf, page 207.
 */
const (
	AvcLevelReserved = 0

	AvcLevel_1  = 10
	AvcLevel_11 = 11
	AvcLevel_12 = 12
	AvcLevel_13 = 13
	AvcLevel_2  = 20
	AvcLevel_21 = 21
	AvcLevel_22 = 22
	AvcLevel_3  = 30
	AvcLevel_31 = 31
	AvcLevel_32 = 32
	AvcLevel_4  = 40
	AvcLevel_41 = 41
	AvcLevel_5  = 50
	AvcLevel_51 = 51
)

/**
 * 8.4.3.3 Semantics
 * ISO_IEC_14496-12-base-format-2012.pdf, page 37
 */
const (
	Mp4HandlerTypeForbidden = 0x00

	Mp4HandlerTypeVIDE = 0x76696465 // 'vide'
	Mp4HandlerTypeSOUN = 0x736f756e // 'soun'
)

// Table 5 — objectTypeIndication Values
// ISO_IEC_14496-1-System-2010.pdf, page 49
const (
	Mp4ObjectTypeForbidden = 0x00
	// Audio ISO/IEC 14496-3
	Mp4ObjectTypeAac = 0x40
)

// Table 6 — streamType Values
// ISO_IEC_14496-1-System-2010.pdf, page 51
const (
	Mp4StreamTypeForbidden   = 0x00
	Mp4StreamTypeAudioStream = 0x05
)
