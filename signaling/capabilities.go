package signaling

import (
	"encoding/json"
	"errors"

	"github.com/chinasarft/golive/config"
)

var WRONGCODEC = errors.New("wrongcodec")

var capabilitiesStr = `
{
	"codecs":
	[
		{
			"kind": "audio",
			"mimeType": "audio/opus",
			"clockRate": 48000,
			"channels": 2
		},
		{
			"kind": "audio",
			"mimeType": "audio/PCMU",
			"preferredPayloadType": 0,
			"clockRate": 8000
		},
		{
			"kind": "audio",
			"mimeType": "audio/PCMA",
			"preferredPayloadType": 8,
			"clockRate": 8000
		},
		{
			"kind": "audio",
			"mimeType": "audio/ISAC",
			"clockRate": 32000
		},
		{
			"kind": "audio",
			"mimeType": "audio/ISAC",
			"clockRate": 16000
		},
		{
			"kind": "audio",
			"mimeType": "audio/G722",
			"preferredPayloadType": 9,
			"clockRate": 8000
		},
		{
			"kind": "audio",
			"mimeType": "audio/iLBC",
			"clockRate": 8000
		},
		{
			"kind": "audio",
			"mimeType": "audio/SILK",
			"clockRate": 24000
		},
		{
			"kind": "audio",
			"mimeType": "audio/SILK",
			"clockRate": 16000
		},
		{
			"kind": "audio",
			"mimeType": "audio/SILK",
			"clockRate": 12000
		},
		{
			"kind": "audio",
			"mimeType": "audio/SILK",
			"clockRate": 8000
		},
		{
			"kind": "audio",
			"mimeType": "audio/CN",
			"preferredPayloadType": 13,
			"clockRate": 32000
		},
		{
			"kind": "audio",
			"mimeType": "audio/CN",
			"preferredPayloadType": 13,
			"clockRate": 16000
		},
		{
			"kind": "audio",
			"mimeType": "audio/CN",
			"preferredPayloadType": 13,
			"clockRate": 8000
		},
		{
			"kind": "audio",
			"mimeType": "audio/telephone-event",
			"clockRate": 48000
		},
		{
			"kind": "audio",
			"mimeType": "audio/telephone-event",
			"clockRate": 32000
		},

		{
			"kind": "audio",
			"mimeType": "audio/telephone-event",
			"clockRate": 16000
		},
		{
			"kind": "audio",
			"mimeType": "audio/telephone-event",
			"clockRate": 8000
		},
		{
			"kind": "video",
			"mimeType": "video/VP8",
			"clockRate": 90000,
			"rtcpFeedback":
			[
				{ "type": "nack" },
				{ "type": "nack", "parameter": "pli" },
				{ "type": "ccm", "parameter": "fir" },
				{ "type": "goog-remb" },
				{ "type": "transport-cc" }
			]
		},
		{
			"kind": "video",
			"mimeType": "video/VP9",
			"clockRate": 90000,
			"rtcpFeedback":
			[
				{ "type": "nack" },
				{ "type": "nack", "parameter": "pli" },
				{ "type": "ccm", "parameter": "fir" },
				{ "type": "goog-remb" },
				{ "type": "transport-cc" }
			]
		},
		{
			"kind": "video",
			"mimeType": "video/H264",
			"clockRate": 90000,
			"parameters":
			{
				"packetization-mode"      : 1,
				"level-asymmetry-allowed" : 1
			},
			"rtcpFeedback":
			[
				{ "type": "nack" },
				{ "type": "nack", "parameter": "pli" },
				{ "type": "ccm", "parameter": "fir" },
				{ "type": "goog-remb" },
				{ "type": "transport-cc" }
			]
		},
		{
			"kind": "video",
			"mimeType": "video/H264",
			"clockRate": 90000,
			"parameters":
			{
				"packetization-mode"      : 0,
				"level-asymmetry-allowed" : 1
			},
			"rtcpFeedback":
			[
				{ "type": "nack" },
				{ "type": "nack", "parameter": "pli" },
				{ "type": "ccm", "parameter": "fir" },
				{ "type": "goog-remb" },
				{ "type": "transport-cc" }
			]
		},
		{
			"kind": "video",
			"mimeType": "video/H265",
			"clockRate": 90000,
			"parameters":
			{
				"packetization-mode"      : 1,
				"level-asymmetry-allowed" : 1
			},
			"rtcpFeedback":
			[
				{ "type": "nack" },
				{ "type": "nack", "parameter": "pli" },
				{ "type": "ccm", "parameter": "fir" },
				{ "type": "goog-remb" },
				{ "type": "transport-cc" }
			]
		},
		{
			"kind": "video",
			"mimeType": "video/H265",
			"clockRate": 90000,
			"parameters":
			{
				"packetization-mode"      : 0,
				"level-asymmetry-allowed" : 1
			},
			"rtcpFeedback":
			[
				{ "type": "nack" },
				{ "type": "nack", "parameter": "pli" },
				{ "type": "ccm", "parameter": "fir" },
				{ "type": "goog-remb" },
				{ "type": "transport-cc" }
			]
		}
	],
	"headerExtensions":
	[
		{
			"kind": "audio",
			"uri": "urn:ietf:params:rtp-hdrext:sdes:mid",
			"preferredId": 1,
			"preferredEncrypt": false,
			"direction": "recvonly"
		},
		{
			"kind": "video",
			"uri": "urn:ietf:params:rtp-hdrext:sdes:mid",
			"preferredId": 1,
			"preferredEncrypt": false,
			"direction": "recvonly"
		},
		{
			"kind": "video",
			"uri": "urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
			"preferredId": 2,
			"preferredEncrypt": false,
			"direction": "recvonly"
		},
		{
			"kind": "video",
			"uri": "urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
			"preferredId": 3,
			"preferredEncrypt": false,
			"direction": "recvonly"
		},
		{
			"kind": "audio",
			"uri": "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
			"preferredId": 4,
			"preferredEncrypt": false,
			"direction": "sendrecv"
		},
		{
			"kind": "video",
			"uri": "http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
			"preferredId": 4,
			"preferredEncrypt": false,
			"direction": "sendrecv"
		},
		{
			"kind": "audio",
			"uri": "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01",
			"preferredId": 5,
			"preferredEncrypt": false,
			"direction": "inactive"
		},
		{
			"kind": "video",
			"uri": "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01",
			"preferredId": 5,
			"preferredEncrypt": false,
			"direction": "inactive"
		},
		{
			"kind": "video",
			"uri": "http://tools.ietf.org/html/draft-ietf-avtext-framemarking-07",
			"preferredId": 6,
			"preferredEncrypt": false,
			"direction": "sendrecv"
		},
		{
			"kind": "video",
			"uri": "urn:ietf:params:rtp-hdrext:framemarking",
			"preferredId": 7,
			"preferredEncrypt": false,
			"direction": "sendrecv"
		},
		{
			"kind": "audio",
			"uri": "urn:ietf:params:rtp-hdrext:ssrc-audio-level",
			"preferredId": 10,
			"preferredEncrypt": false,
			"direction": "sendrecv"
		},
		{
			"kind": "video",
			"uri": "urn:3gpp:video-orientation",
			"preferredId": 11,
			"preferredEncrypt": false,
			"direction": "sendrecv"
		},
		{
			"kind": "video",
			"uri": "urn:ietf:params:rtp-hdrext:toffset",
			"preferredId": 12,
			"preferredEncrypt": false,
			"direction": "sendrecv"
		}
	],
	"fecMechanisms": []
}`

type HeaderExtension struct {
	Kind             string `json:"kind"`
	Uri              string `json:"uri"`
	PreferredId      int    `json:"preferredId"`
	PreferredEncrypt bool   `json:"PreferredEncrypt"`
	Direction        string `json:"Direction"`
}

type RtpCapabilities struct {
	HeaderExts    []HeaderExtension             `json:"headerExtensions"`
	Codecs        []config.MediasoupCodecConfig `json:"codecs"`
	FecMechanisms []string                      `json:"fecMechanisms"`
}

var supportedRtpCaps = &RtpCapabilities{}
var expectCaps = &RtpCapabilities{} // from config and supportedRtpCaps
var expectCapsStr = ""

var DynamicPayloadTypes = []uint8{
	100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
	111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121,
	122, 123, 124, 125, 126, 127, 96, 97, 98, 99,
}

func init() {
	if err := json.Unmarshal([]byte(capabilitiesStr), supportedRtpCaps); err != nil {
		panic(err)
	}
}

// 1. 观察了下ortc.js merge，可以简化很多，比如格式检查这些可以先忽略
// 2. 遇到video需要加上video/rtx, 应该是rtp的重传用的
func setExpectCapsFromConfigAndSupportedCaps(configCodecs []*config.MediasoupCodecConfig) error {
	typeIdx := 0
	var resultCodecs []config.MediasoupCodecConfig
	for _, codec := range configCodecs {
		for _, supportedCodec := range supportedRtpCaps.Codecs {
			if supportedCodec.MimeType == codec.MimeType {
				if supportedCodec.PayloadType == 0 {
					supportedCodec.PayloadType = DynamicPayloadTypes[typeIdx]
					typeIdx++
				}

				// merge parameters
				if codec.Parameters != nil {
					if supportedCodec.Parameters == nil {
						supportedCodec.Parameters = codec.Parameters
					} else {
						for k, p := range codec.Parameters {
							if _, ok := supportedCodec.Parameters[k]; !ok {
								supportedCodec.Parameters[k] = p
							}
						}
					}
				}

				if supportedCodec.Kind == "audio" {
					if supportedCodec.Channels == 0 && codec.Channels != 0 {
						supportedCodec.Channels = codec.Channels
					}
					if supportedCodec.Channels == 0 {
						supportedCodec.Channels = 1
					}
				}
				resultCodecs = append(resultCodecs, supportedCodec)
				if supportedCodec.Kind == "video" {
					resultCodecs = append(resultCodecs, config.MediasoupCodecConfig{
						Kind:        codec.Kind,
						MimeType:    codec.Kind + "/rtx",
						PayloadType: DynamicPayloadTypes[typeIdx],
						ClockRate:   codec.ClockRate,
						Parameters: map[string]interface{}{
							"apt": supportedCodec.PayloadType,
						},
					})
					typeIdx++
				}
				// 只要匹配了一个，就跳出
				break
			}
		}
	}

	if len(resultCodecs) < 1 {
		return WRONGCODEC
	}
	expectCaps.Codecs = resultCodecs
	expectCaps.FecMechanisms = []string{}
	expectCaps.HeaderExts = supportedRtpCaps.HeaderExts
	for i := 0; i < len(expectCaps.Codecs); i++ {
		if expectCaps.Codecs[i].RtcpFeedback == nil {
			expectCaps.Codecs[i].RtcpFeedback = []config.RtcpFeedback{}
		}
		if expectCaps.Codecs[i].Parameters == nil {
			expectCaps.Codecs[i].Parameters = make(map[string]interface{})
		}
	}

	b, _ := json.Marshal(&expectCaps)
	expectCapsStr = string(b)
	return nil
}
