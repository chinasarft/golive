package signaling

import (
	"github.com/chinasarft/golive/config"
)

/*{
	"response": true,
	"id": 1744378,
	"ok": true,
	"data": {
		"id": "fe433124-0b60-4cd1-a7ba-e49b9b6939d1",
		"iceParameters": {
			"iceLite": true,
			"password": "6h7ypz8momd2bo3nni0jik9bpf2r8ps0",
			"usernameFragment": "t0ul0yr1rxoxebqm"
		},
		"iceCandidates": [{
			"foundation": "udpcandidate",
			"ip": "127.0.0.1",
			"port": 46190,
			"priority": 1076558079,
			"protocol": "udp",
			"type": "host"
		}, {
			"foundation": "tcpcandidate",
			"ip": "127.0.0.1",
			"port": 42902,
			"priority": 1076302079,
			"protocol": "tcp",
			"tcpType": "passive",
			"type": "host"
		}],
		"dtlsParameters": {
			"fingerprints": [{
				"algorithm": "sha-1",
				"value": "CB:8D:6D:6E:87:28:8E:3A:1F:09:27:64:65:BA:2F:EB:3C:F6:49:43"
			}, {
				"algorithm": "sha-224",
				"value": "D6:CC:90:A8:31:B4:7E:F5:FC:80:F1:D3:DF:0E:CF:3B:82:2A:74:B5:62:A9:EE:9B:AD:7D:D3:5D"
			}, {
				"algorithm": "sha-256",
				"value": "60:A6:43:C3:59:04:78:A3:BE:9C:82:0C:49:83:AB:C4:72:54:F3:5D:49:E5:C3:0D:63:9F:F5:D2:72:83:E7:BE"
			}, {
				"algorithm": "sha-384",
				"value": "ED:FC:E7:E1:5C:22:83:57:16:37:09:F7:F6:5D:44:4B:CB:C9:01:75:D3:4B:C5:A0:FB:80:B1:3E:E9:B7:87:72:7A:F4:0A:06:F2:A1:37:74:2A:A7:1D:10:F2:AE:15:51"
			}, {
				"algorithm": "sha-512",
				"value": "62:F7:4D:18:60:A7:43:44:4F:8A:37:83:1C:93:13:24:3B:39:18:48:65:D3:0B:C9:9D:7C:4B:72:BD:19:0A:92:F6:E1:22:DD:3B:1D:97:CD:D0:55:F9:EF:71:DC:E9:E2:C1:36:6E:68:9A:DD:D3:5D:6A:A3:9F:32:D2:3B:82:E9"
			}],
			"role": "auto"
		}
	}
}*/

type msIceParameters struct {
	IceLite          bool   `json:"iceLite"`
	Password         string `json:"password"`
	UsernameFragment string `json:"usernameFragment"`
}

type msIceCandidates struct {
	Foundation string `json:"foundation"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Priority   uint64 `json:"priority"`
	Protocol   string `json:"protocol"`
	Type       string `json:"type"`
	TcpType    string `json:"tcpType,omitempty"`
}

type msDtlsParameters struct {
	Role         string `json:"role"`
	Fingerprints []struct {
		Algorithm string `json:"algorithm"`
		Value     string `json:"value"`
	} `json:"fingerprints"`
}

type msWebrtcTransportResp struct {
	Id             string            `json:"id"`
	IceParameters  msIceParameters   `json:"iceParameters"`
	IceCandidates  []msIceCandidates `json:"iceCandidates"`
	DtlsParameters msDtlsParameters  `json:"dtlsParameters"`
	// TODO 其它字段
}

/*
{
	"accepted": true,
	"data": {
		"consumerIds": [],
		"dtlsParameters": {
			"fingerprints": [{
				"algorithm": "sha-1",
				"value": "B5:3C:4C:36:AD:0D:D6:42:AC:BD:D3:47:14:0A:DA:D3:12:15:F1:4D"
			}, {
				"algorithm": "sha-224",
				"value": "56:7B:15:37:05:94:9E:A5:F6:5F:97:22:7C:53:62:40:66:2E:55:8D:3C:67:FF:A8:54:1E:2B:38"
			}, {
				"algorithm": "sha-256",
				"value": "C4:82:1D:70:EB:1A:04:3F:B3:8F:FC:EF:46:7F:70:99:DA:1E:F4:E7:4A:1F:90:BF:E0:CF:94:AF:B3:1F:BC:E9"
			}, {
				"algorithm": "sha-384",
				"value": "CF:6E:42:32:AA:AE:A1:C8:F5:32:43:BF:8B:EF:D7:25:7B:95:3E:77:90:CD:9D:1C:E1:94:59:5A:4D:06:4B:EC:25:12:EF:E9:FA:C8:18:60:F2:97:45:3B:92:F7:7C:0A"
			}, {
				"algorithm": "sha-512",
				"value": "21:26:8B:09:5A:F4:AE:E5:A3:10:95:A7:B6:AF:BA:30:06:8C:A9:CF:0F:FA:48:48:80:12:40:45:E7:1B:F7:80:21:4B:77:CE:41:A5:FE:48:60:DA:1E:5D:75:D9:29:CE:49:D7:A0:62:60:AB:95:27:50:B2:A7:49:0A:34:B4:AD"
			}],
			"role": "auto"
		},
		"dtlsState": "new",
		"iceCandidates": [{
			"foundation": "udpcandidate",
			"ip": "127.0.0.1",
			"port": 48559,
			"priority": 1076558079,
			"protocol": "udp",
			"type": "host"
		}, {
			"foundation": "tcpcandidate",
			"ip": "127.0.0.1",
			"port": 41826,
			"priority": 1076302079,
			"protocol": "tcp",
			"tcpType": "passive",
			"type": "host"
		}],
		"iceParameters": {
			"iceLite": true,
			"password": "pnwi0ud2bsfnzesnygxnxf2ropsksbpw",
			"usernameFragment": "si75rpcd2femqh7u"
		},
		"iceRole": "controlled",
		"iceState": "new",
		"id": "30623ea7-4118-4e7a-9bd5-2cd9707a104d",
		"mapSsrcConsumerId": {},
		"producerIds": [],
		"rtpHeaderExtensions": {},
		"rtpListener": {
			"midTable": {},
			"ridTable": {},
			"ssrcTable": {}
		}
	},
	"id": 3
}*/

type msCreateWebrtcTprReq struct {
	ForceTcp  bool `json:"forceTcp"`
	Consuming bool `json:"consuming"`
	Producing bool `json:"producing"`
}

type msJoinReq struct {
	DisplayName string `json:"displayName"`
	Device      struct {
		Flag    string `json:"flag"`
		Version string `json:"version"`
		Name    string `json:"name"`
	} `json:"device"`
	RtpCapabilities RtpCapabilities `json:"rtpCapabilities`
}

type msJoinResp struct {
	Peers []string `json:"peers"`
}

type msConnectRtcTprReq struct {
	TransportId    string           `json:"transportId"`
	DtlsParameters msDtlsParameters `json:"dtlsParameters"`
}

// {"accepted":true,"data":{"dtlsLocalRole":"client"},"id":7}
type msConnectRtcTprResp struct {
	DtlsLocalRole string `json:"client,omitempty"`
}

type RtpParameters struct {
	Mid              string                        `json:"mid"`
	Codecs           []config.MediasoupCodecConfig `json:"codecs"`
	HeaderExtensions []HeaderExtension             `json:"headerExtensions"`
	Encodings        []struct {
		Ssrc uint32 `json:"ssrc"`
	} `json:"encodings"`
	Rtcp struct {
		Cname string `json:"cname"`
	} `json:"rtcp"`
}

type msProduceReq struct {
	TransportId   string                 `json:"transportId"`
	Kind          string                 `json"kind"`
	AppData       map[string]interface{} `json:"appData"` // TODO 还不清楚结构
	RtpParameters RtpParameters          `json:"rtpParameters"`
}

type msProduceResp struct {
	Id string `json:"id"`
}
