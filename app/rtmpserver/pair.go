package rtmpserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/chinasarft/golive/protocol/rtmp"
	"github.com/chinasarft/golive/utils/amf"
)

const (
	cmd_register_source = iota
	cmd_register_sink
	cmd_rtmp_message
	cmd_unregister_source
	cmd_unregister_sink
)

type PadMessage struct {
	cmd int
	msg interface{}
	ctx context.Context
}

type Source struct {
	*rtmp.RtmpHandler
	srcChan chan *PadMessage
	result  chan error
	sinks   map[string]*Sink
	err     error

	VideoDecoderConfigurationRecord []byte // avc hevc
	AACSequenceHeader               []byte
	avMetaData                      []byte // 或者至少是在推流h265的时候是不支持的
}

type Sink struct {
	*rtmp.RtmpHandler
	result      chan error
	rtmpMsgChan chan *rtmp.Message
	keyInSrc    string
}

type ConnPool struct {
	pipe      chan *PadMessage
	receivers map[string]*Source
	senders   map[string]*Sink
}

var conns *ConnPool

func init() {
	conns = &ConnPool{
		pipe:      make(chan *PadMessage, 100),
		receivers: make(map[string]*Source),
		senders:   make(map[string]*Sink),
	}
	go conns.pairPad()
}

func (p *ConnPool) pairPad() {
	for {
		var msg *PadMessage
		select {
		case msg = <-p.pipe:
		}
		log.Println("pair")
		switch msg.cmd {
		case cmd_register_source:
			p.handleRegisterSource(msg)
		case cmd_unregister_source:
			p.handleUnregisterSource(msg)
		case cmd_register_sink:
			p.handleRegisterSink(msg)
		case cmd_unregister_sink:
			p.handleUnregisterSink(msg)
		}
	}
}

func (cp *ConnPool) OnSourceDetermined(h *rtmp.RtmpHandler, ctx context.Context) (rtmp.PutAVDMessage, error) {

	src := &Source{
		RtmpHandler: h,
		result:      make(chan error),
		sinks:       make(map[string]*Sink),
		srcChan:     make(chan *PadMessage, 100),
	}
	conns.pipe <- &PadMessage{
		cmd: cmd_register_source,
		msg: src,
		ctx: ctx,
	}

	// 同步返回结果
	err := <-src.result

	output := func(m *rtmp.Message) error {
		if src.err != nil {
			return src.err
		}
		select {
		case src.srcChan <- &PadMessage{
			cmd: cmd_rtmp_message,
			msg: m,
		}:
			log.Println("--------->put message")
		default:
			return fmt.Errorf("chan is full. drop this msg")
		}
		return nil
	}

	return output, err
}

func (p *ConnPool) handleRegisterSource(msg *PadMessage) {

	src, ok := msg.msg.(*Source)
	if !ok {
		panic("handleRegisterSource wrong type")
	}

	key := src.GetAppStreamKey()

	_, ok = p.receivers[key]
	log.Println("handleRegisterSource:", key, ok)
	if ok {
		// TODO 存在是否抢流等？
		src.result <- fmt.Errorf("is already exits")
	} else {
		p.receivers[key] = src
		src.result <- nil
		go src.work(msg.ctx)
	}
}

func (s *Source) work(ctx context.Context) {
	ctx, _ = context.WithCancel(ctx)
	for {
		var msg *PadMessage
		select {
		case <-ctx.Done():
			s.cancelSink()
			return
		case msg = <-s.srcChan:
		}

		switch msg.cmd {
		case cmd_register_sink:
			s.connectSink(msg)
		case cmd_rtmp_message:
			rtmpMsg, ok := msg.msg.(*rtmp.Message)
			if !ok {
				panic("not rtmp message")
			}
			log.Println("srcwork:", rtmpMsg.Timestamp)
			s.handleRtmpMessage(rtmpMsg)
		case cmd_unregister_sink:
			s.deleteSink(msg)
		}
	}
}

func (src *Source) handleRtmpMessage(m *rtmp.Message) {
	switch m.MessageType {
	case 8:
		if m.Payload[1] == 0 {
			src.AACSequenceHeader = m.Payload
		}
	case 9:
		if m.Payload[1] == 0 {
			src.VideoDecoderConfigurationRecord = m.Payload
		}
	case 15:
		fallthrough
	case 18:
		src.handleDataMessaage(m)
	}
	src.writeMessage(m)
}

func (src *Source) writeMessage(m *rtmp.Message) {
	for _, sink := range src.sinks {
		toSinkMsg := new(rtmp.Message)
		toSinkMsg.MessageType = m.MessageType
		toSinkMsg.Payload = m.Payload
		toSinkMsg.Timestamp = m.Timestamp
		toSinkMsg.StreamID = sink.GetFunctionalStreamId()
		err := sink.WriteMessage(toSinkMsg)
		if err != nil {
			delete(src.sinks, sink.keyInSrc)
		}
	}
}

func (src *Source) handleDataMessaage(m *rtmp.Message) {
	r := bytes.NewReader(m.Payload)
	v, e := amf.ReadValue(r)
	if e == nil {
		switch v.(type) {
		case string:
			value := v.(string)
			switch value {
			case "@setDataFrame":
				// @setDataFrame固定长度是16字节
				m.Payload = m.Payload[16:]
				return
			}
		}
	} else {
		if e != io.EOF {
			src.err = e
		}
	}
}

func (p *ConnPool) handleUnregisterSource(msg *PadMessage) {

	h, ok := msg.msg.(*rtmp.RtmpHandler)
	if !ok {
		panic("handleUnegisterSource wrong type")
	}

	key := h.GetAppStreamKey()

	_, ok = p.receivers[key]
	log.Println("handleUnregisterSource:", key, ok)
	if ok {
		delete(p.receivers, key)
	} else {
		panic(key + " source not registerd")
	}
}

func (cp *ConnPool) OnDestroySource(h *rtmp.RtmpHandler) {

	cp.pipe <- &PadMessage{
		cmd: cmd_unregister_source,
		msg: h,
	}

	return
}

func (cp *ConnPool) OnSinkDetermined(h *rtmp.RtmpHandler, ctx context.Context) error {

	sink := &Sink{
		RtmpHandler: h,
		result:      make(chan error),
		rtmpMsgChan: make(chan *rtmp.Message, 50),
	}
	cp.pipe <- &PadMessage{
		cmd: cmd_register_sink,
		msg: sink,
		ctx: ctx,
	}
	// 同步返回结果
	err := <-sink.result

	return err
}

func (p *ConnPool) handleRegisterSink(msg *PadMessage) {
	sink, ok := msg.msg.(*Sink)
	if !ok {
		panic("handleRegisterSource wrong type")
	}
	key := sink.GetAppStreamKey()

	src, ok := p.receivers[key]
	log.Println("handleRegisterSink:", key, ok)
	if !ok {
		// TODO 推流还不存在，等待？
		sink.result <- fmt.Errorf("%s not exits", key)
	} else {
		sink.result <- nil
		src.srcChan <- msg
		go sink.work(msg.ctx)
	}
}

func (p *ConnPool) handleUnregisterSink(msg *PadMessage) {
	h, ok := msg.msg.(*rtmp.RtmpHandler)
	if !ok {
		panic("handleRegisterSource wrong type")
	}
	key := h.GetAppStreamKey()

	src, ok := p.receivers[key]
	log.Println("handleUnregisterSink:", key, ok)
	if !ok {
		// 情况1:play时候还没有source, 目前play不存在source返回错误,不可能
		// 情况2:unregsrc和unregsink几乎同时，先unregsrc，的却可能出现这种情况
		log.Println(key + " source not registered(from sink)")
	} else {
		src.srcChan <- msg
	}
}

func (hs *ConnPool) OnDestroySink(h *rtmp.RtmpHandler) {
	conns.pipe <- &PadMessage{
		cmd: cmd_unregister_sink,
		msg: h,
	}
	return
}

func (sink *Sink) work(ctx context.Context) {
	ctx, _ = context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-sink.rtmpMsgChan:
			log.Println("=======>receive message")
			sink.WriteMessage(m)
		}
	}
}

func (src *Source) connectSink(msg *PadMessage) {
	log.Println("src connectSink")
	sink := msg.msg.(*Sink)
	sink.keyInSrc = sink.GetAppStreamKey() + fmt.Sprintf("%p", sink.RtmpHandler)

	if _, ok := src.sinks[sink.keyInSrc]; !ok {
		src.sinks[sink.keyInSrc] = sink
	} else {
		panic("repeater register")
	}

	if src.avMetaData != nil {
		avMetaData := &rtmp.Message{
			MessageType: 0x12,
			Timestamp:   0,
			StreamID:    sink.GetFunctionalStreamId(),
			Payload:     src.avMetaData,
		}
		sink.rtmpMsgChan <- avMetaData
	}

	if src.AACSequenceHeader != nil {
		amsg := &rtmp.Message{
			MessageType: 8,
			Timestamp:   0,
			StreamID:    sink.GetFunctionalStreamId(),
			Payload:     src.AACSequenceHeader,
		}
		sink.rtmpMsgChan <- amsg
		log.Println("=======>write audio metadata", len(amsg.Payload))
	}

	if src.VideoDecoderConfigurationRecord != nil {
		vmsg := &rtmp.Message{
			MessageType: 9,
			Timestamp:   0,
			StreamID:    sink.GetFunctionalStreamId(),
			Payload:     src.VideoDecoderConfigurationRecord,
		}
		sink.rtmpMsgChan <- vmsg
		log.Println("=======>write video metadata", len(vmsg.Payload))
	}

}

func (src *Source) deleteSink(msg *PadMessage) {

	h := msg.msg.(*rtmp.RtmpHandler)
	keyInSrc := h.GetAppStreamKey() + fmt.Sprintf("%p", h)

	sink, ok := src.sinks[keyInSrc]
	log.Println("src deleteSink", keyInSrc, ok)
	if !ok {
		// 还是有可能的，如果几乎同时关闭，先unregisterSource
		// 在unregisterSink,就会出现
		log.Panicln(keyInSrc + " not conntecd to source")
	} else {
		delete(src.sinks, keyInSrc)
		sink.Cancel()
	}
}

func (src *Source) cancelSink() {
	for _, sink := range src.sinks {
		sink.Cancel()
	}
}
