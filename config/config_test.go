package config

import (
	"encoding/json"
	"testing"
)

var confStr = `
{
	"prof": {
		"enable": false
	},
	"log": {
		"level": "debug",
		"position": true,
		"target": {
			"type": "stdout",
			"name": "filename"
		}
	},
	"api": {
		"addr": ":65267"
	},
	"publish": {
		"is_gop_cache": true,
		"rtmpflv": {
			"addr": ":1935",
			"flvmagic": 102
		}
	},
	"signaling": {
		"protoo": {
			"addr": ":65273"
		},
		"mediasoup": {
			"numWorkers": 2,
			"worker": {
				"path": "toworker",
				"exitSignal": "SIGINT",
				"env": {
					"MEDIASOUP_VERSION": "3.0.0"
				},
				"logLevel": "debug",
				"logTags": [
					"info",
					"ice"
				],
				"rtcMinPort": 40000,
				"rtcMaxPort": 49999
			},
			"router": {
				"mediaCodecs": [{
						"kind": "audio",
						"mimeType": "audio/opus",
						"clockRate": 48000,
						"channels": 2
					},
					{
						"kind": "video",
						"mimeType": "video/h264",
						"clockRate": 90000,
						"parameters": {
							"packetization-mode": 1,
							"profile-level-id": "4d0032"
						}
					}
				]
			},
			"webRtcTransport": {
				"listenIps": [{
					"ip": "127.0.0.1",
					"announcedIp": null
				}],
				"maxIncomingBitrate": 1500000,
				"initialAvailableOutgoingBitrate": 1000000
			}
		}
	},
	"relay": {
		"addr": ":61935"
	}
}
`

func TestLoadConfig(t *testing.T) {
	conf := &Config{}

	if err := json.Unmarshal([]byte(confStr), conf); err != nil {
		t.Fatal(err)
	}

	if conf.Prof.Enable != false {
		t.Fatal("prof expect false")
	}

	if conf.Log.Level != "debug" {
		t.Fatal("log.Level expect debug, but:", conf.Log.Level)
	}
	if conf.Log.Position != true {
		t.Fatal("Log.Position expect true")
	}
	if conf.Log.Target.Type != "stdout" {
		t.Fatal("log.Target.Name expect stdout, but:", conf.Log.Target.Type)
	}
	if conf.Log.Target.Name != "filename" {
		t.Fatal("log.Target.Name expect filename, but:", conf.Log.Target.Name)
	}

	if conf.Api.Addr != ":65267" {
		t.Fatal("conf.Api.Addr expect :65267, but:", conf.Api.Addr)
	}

	if conf.Publish.IsGopCache != true {
		t.Fatal("conf.Publish.IsGopCacher expect true")
	}
	if conf.Publish.RtmpFlv.Addr != ":1935" {
		t.Fatal("conf.Publish.RtmpFlv.Addr expect :1935, but:", conf.Publish.RtmpFlv.Addr)
	}
	if conf.Publish.RtmpFlv.FlvMagic != 102 {
		t.Fatal("conf.Publish.RtmpFlv.flvmagic expect 102, but:", conf.Publish.RtmpFlv.FlvMagic)
	}

	sigConf := &conf.Signaling
	if sigConf.Protoo.Addr != ":65273" {
		t.Fatal("sigConf.Protoo.Addr expect :65273, but:", sigConf.Protoo.Addr)
	}

	m := &conf.Signaling.Mediasoup
	if m.NumWorkers != 2 {
		t.Fatal("m.NumWorkers expect 2, but:", m.NumWorkers)
	}

	w := conf.Signaling.Mediasoup.Worker
	if w.Path != "toworker" {
		t.Fatal("w.Path exepect toworker, but:", w.Path)
	}
	if w.ExitSignal != "SIGINT" {
		t.Fatal("w.ExitSignal exepect SIGINT, but:", w.ExitSignal)
	}
	if len(w.Env) != 1 {
		t.Fatal("expect len(w.Env)==1, but:", len(w.Env))
	} else {
		v, ok := w.Env["MEDIASOUP_VERSION"]
		if !ok || v != "3.0.0" {
			t.Fatal("expect w.Env[MEDIASOUP_VERSION]==3.0.0, but:", v)
		}
	}
	if w.LogLevel != "debug" {
		t.Fatal("w.ExitSignal exepect debug, but:", w.LogLevel)
	}

	if len(w.LogTags) != 2 {
		t.Fatal("expect len(w.LogTags)==2, but:", len(w.LogTags))
	} else {
		tags := []string{"info", "ice"}
		for i, v := range tags {
			if v != w.LogTags[i] {
				t.Fatalf("expect w.LogTags[%d]=%s, but:%s", i, v, w.LogTags[i])
			}
		}
	}
	if w.MaxPort != 49999 {
		t.Fatal("w.MaxPort exepect 49999, but:", w.MaxPort)
	}
	if w.MinPort != 40000 {
		t.Fatal("w.MinPort exepect 40000, but:", w.MinPort)
	}

	if conf.Relay.Addr != ":61935" {
		t.Fatal("conf.Relay expect :61935, but:", conf.Relay.Addr)
	}

	webrtc := &conf.Signaling.Mediasoup.WebRtcTransport
	if webrtc.InitialAvailableOutgoingBitrate != 1000000 {
		t.Fatal("InitialAvailableOutgoingBitrate expect 1000000, but:", webrtc.InitialAvailableOutgoingBitrate)
	}
	if webrtc.MaxIncomingBitrate != 1500000 {
		t.Fatal("MaxIncomingBitrate expect 1500000, but:", webrtc.MaxIncomingBitrate)
	}
	if len(webrtc.ListenIps) != 1 {
		t.Fatal("expect len(webrtc.ListenIps)==2, but:", len(webrtc.ListenIps))
	} else {
		if webrtc.ListenIps[0].Ip != "127.0.0.1" || webrtc.ListenIps[0].AnnouncedIp != "" {
			t.Fatal(`expect 127.0.0.1 and , but:`, webrtc.ListenIps[0].Ip, webrtc.ListenIps[0].AnnouncedIp)
		}
	}

	router := &conf.Signaling.Mediasoup.Router
	if len(router.MediaCodecs) != 2 {
		t.Fatal("expect router.MediaCodecs=2, but", len(router.MediaCodecs))
	} else {
		c := router.MediaCodecs[0]
		if c.Kind != "audio" || c.MimeType != "audio/opus" ||
			c.ClockRate != 48000 || c.Channels != 2 {
			t.Fatalf("expect MediaCodecs[0]: kind:audio mime:audio/opus, clockrate:48000"+
				"channels:2, but %s %s %d %d", c.Kind, c.MimeType, c.ClockRate, c.Channels)
		}

		c = router.MediaCodecs[1]
		if c.Kind != "video" || c.MimeType != "video/h264" ||
			c.ClockRate != 90000 || c.Channels != 0 {
			t.Fatalf("expect MediaCodecs[0]: kind:video mime:video/h264, clockrate:90000"+
				"channels:0, but %s %s %d %d", c.Kind, c.MimeType, c.ClockRate, c.Channels)
		}
		if v, ok := c.Parameters["packetization-mode"]; ok {
			if intv, ok := v.(float64); ok {
				if intv != 1 {
					t.Fatal("Parameters[packetization-mode] expect 1 but:", intv)
				}
			} else {
				t.Fatal("Parameters[packetization-mode] not float64", v)
			}
		} else {
			t.Fatal("Parameters[packetization-mode] not exist")
		}

		if v, ok := c.Parameters["profile-level-id"]; ok {
			if strv, ok := v.(string); ok {
				if strv != "4d0032" {
					t.Fatal("Parameters[profile-level-id] expect 4d0032 but:", strv)
				}
			} else {
				t.Fatal("Parameters[packetization-mode] not string", v)
			}
		} else {
			t.Fatal("Parameters[profile-level-id] not exist")
		}
	}

}
