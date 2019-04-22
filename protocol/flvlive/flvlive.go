package flvlive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/chinasarft/golive/exchange"
	"github.com/chinasarft/golive/utils/amf"
)

var gRe *regexp.Regexp

func init() {
	gRe = regexp.MustCompile(`\/(\w{1,})\/(\w{1,})`)
}

type Config struct {
	MagicNumber uint8
}

type FlvLiveHandler struct {
	rw     io.ReadWriter
	pad    exchange.Pad
	config Config

	appStreamKey string
	inited       bool

	ctx    context.Context
	cancel context.CancelFunc

	putData exchange.PutData
}

func NewFlvLiveHandler(rw io.ReadWriter, pad exchange.Pad) *FlvLiveHandler {
	return &FlvLiveHandler{
		rw:  rw,
		pad: pad,
		//config: config,
	}
}

func (f *FlvLiveHandler) Start() error {
	var buf [16]byte
	if _, err := f.rw.Read(buf[0:1]); err != nil {
		return err
	}

	// not check magic number
	//if uint8(buf[0]) != f.config.MagicNumber {
	//	return fmt.Errorf("unexpected flv live magic number:%d", uint8(buf[0]))
	//}

	f.ctx, f.cancel = context.WithCancel(context.Background())
	for {
		d, err := GetNextExData(f.rw)
		if err != nil {
			return err
		}
		switch d.DataType {
		case exchange.DataTypeDataAMF0:
			if err = f.handleScriptData(d); err != nil {
				return err
			}
		default:
			if f.putData == nil {
				return fmt.Errorf("source not registered")
			}

			f.putData(d)
		}
	}
}

func (f *FlvLiveHandler) GetAppStreamKey() string {
	return f.appStreamKey
}

func getAppStreamKey(rtmpUrl string) (string, error) {
	if strings.Index(rtmpUrl, "rtmp://") != 0 {
		return "", fmt.Errorf("wrong url:%s", rtmpUrl)
	}

	matches := gRe.FindAllStringSubmatch(rtmpUrl, -1)
	if matches == nil || len(matches) < 0 || len(matches[0]) < 3 {
		return "", fmt.Errorf("wrong url:%s", rtmpUrl)
	}

	return matches[0][1] + "-" + matches[0][2], nil
}

func (f *FlvLiveHandler) handleScriptData(d *exchange.ExData) error {

	objReader := bytes.NewReader(d.Payload)
	objRead, err := amf.ReadObject(objReader)
	if err != nil {
		return err
	}

	if f.inited == false {
		rtmpUrl, ok := objRead["url"].(string)
		if !ok {
			return fmt.Errorf("url not a string")
		}

		f.appStreamKey, err = getAppStreamKey(rtmpUrl)

		isPublish, ok := objRead["publish"].(bool)
		if !ok {
			return fmt.Errorf("publish is not bool")
		}
		if isPublish {
			if f.putData, err = f.pad.OnSourceDetermined(f, f.ctx); err != nil {
				return err
			}
		} else {
			if err = f.pad.OnSinkDetermined(f, f.ctx); err != nil {
				return err
			}
		}
		return nil
	}

	// 暂时没有其它script需要处理
	return nil
}

func (f *FlvLiveHandler) WriteData(m *exchange.ExData) error {
	return fmt.Errorf("flv live must be source")
}

func (f *FlvLiveHandler) Cancel() {
	f.cancel()
}
