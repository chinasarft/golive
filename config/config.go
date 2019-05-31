package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

/*
{
        "prof": {
                "enable": false
        },
        "log": {
                "level":"debug",
                "position": true,
				"target": {
					"type": "stdout", // stdout, file
					"name": "filename", //only if type is file
					// 用logrotate去做，但是rotate时候可能会丢日志，用APPEND模式打开
				}

        },
        "api": {
                "addr": ":65267"
        },
        "publish": {
                "rtmpflv": {
                	  "addr": ":1935",
                	  "flvmagic": 102
                }
        },
        "sigaling": {
        		"protoo":{
				"addr": ":65273",
			}
        	},
        "relay": {
                "addr": ":61935"
        }
}
*/

type ProfConfig struct {
	Enable bool `json:"enable"`
}

type LogTarget struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
type LogConfig struct {
	Level    string    `json:"level"`
	Target   LogTarget `json:"target"`
	Position bool      `json:"position"`
}

type ApiConfig struct {
	Addr string `json:"addr"`
}

type RtmpFlvConfig struct {
	Addr     string `json:"addr"`
	FlvMagic int    `json:"flvmagic"`
}

type PublishConfig struct {
	IsGopCache bool          `json:"is_gop_cache"`
	RtmpFlv    RtmpFlvConfig `json:"rtmpflv"`
}

type RelayConfig struct {
	Addr string `json:"addr"`
}

type ProtooConfig struct {
	Addr string `json:"addr"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

type MediasoupWorkerConfig struct {
	Path       string            `json:"path"`
	MinPort    int               `json:"rtcMinPort"`
	MaxPort    int               `json:"rtcMaxPort"`
	LogLevel   string            `json:"logLevel"`
	LogTags    []string          `json:"logTags"`
	Env        map[string]string `json:"env"`
	ExitSignal string            `json:"exitSignal"`
}

type RtcpFeedback struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter,omitempty"`
}

// RtpDictionaries.hpp 但是没有kind，是因为可以从MimeType得出?
type MediasoupCodecConfig struct {
	Kind         string                 `json:"kind"`
	MimeType     string                 `json:"mimeType"`
	PayloadType  uint8                  `json:"preferredPayloadType"`
	ClockRate    uint32                 `json:"clockRate,omitempty"`
	Channels     uint8                  `json:"channels,omitempty"`
	Parameters   map[string]interface{} `json:"parameters"`
	RtcpFeedback []RtcpFeedback         `json:"rtcpFeedback"`
}

type MediasoupWebRtcTransportConfig struct {
	ListenIps []struct {
		Ip          string `json:"ip"`
		AnnouncedIp string `json:"announcedIp,omitempty"`
	} `json:"listenIps"`
	MaxIncomingBitrate              int `json:"maxIncomingBitrate"`
	InitialAvailableOutgoingBitrate int `json:"initialAvailableOutgoingBitrate"`
}

type MediasoupRouterConfig struct {
	MediaCodecs []*MediasoupCodecConfig `json:"mediaCodecs"`
}

type MediasoupConfig struct {
	NumWorkers      int                            `json:"numWorkers"`
	Worker          MediasoupWorkerConfig          `json:"worker"`
	Router          MediasoupRouterConfig          `json:"router"`
	WebRtcTransport MediasoupWebRtcTransportConfig `json:"webRtcTransport"`
}

type SigalingConfig struct {
	Protoo    ProtooConfig    `json:"protoo"`
	Mediasoup MediasoupConfig `json:"mediasoup"`
}

type Config struct {
	Prof      ProfConfig     `json:"prof"`
	Log       LogConfig      `json:"log"`
	Api       ApiConfig      `json:"api"`
	Publish   PublishConfig  `json:"publish"`
	Relay     RelayConfig    `json:"relay"`
	Signaling SigalingConfig `json:"signaling"`
}

func checkConfig(conf *Config) error {
	magic := conf.Publish.RtmpFlv.FlvMagic
	if magic > 255 || magic == 3 || magic == 6 || magic < 0 {
		return fmt.Errorf("flvlive magic num:%d", magic)
	}

	switch conf.Log.Level {
	case "debug":
		break
	case "info":
		break
	case "warn":
		break
	case "error":
		break
	default:
		conf.Log.Level = "info"
	}

	return nil
}

func LoadConfig(filename string) (*Config, error) {

	f, err := os.Open(filename)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	if err = json.Unmarshal(content, conf); err != nil {
		return nil, err
	}

	if err = checkConfig(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
