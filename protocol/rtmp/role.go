package rtmp

import (
	"context"
	"fmt"
	"log"
)

type PutAvMessage func(m *Message) error

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
}

type Source struct {
	*RtmpHandler
	srcChan chan *PadMessage
	result  chan error
	sinks   map[string]*Sink
}

type Sink struct {
	*RtmpHandler
	result      chan error
	rtmpMsgChan chan *Message
	keyInSrc    string
}

type PadPool struct {
	pipe      chan *PadMessage
	receivers map[string]*Source
	senders   map[string]*Sink
}

var pads *PadPool

func init() {
	pads = &PadPool{
		pipe:      make(chan *PadMessage, 100),
		receivers: make(map[string]*Source),
		senders:   make(map[string]*Sink),
	}
	go pads.pairPad()
}

func RegisterSource(h *RtmpHandler) (output PutAvMessage, err error) {
	src := &Source{
		RtmpHandler: h,
		result:      make(chan error),
		sinks:       make(map[string]*Sink),
		srcChan:     make(chan *PadMessage, 100),
	}
	pads.pipe <- &PadMessage{
		cmd: cmd_register_source,
		msg: src,
	}

	// 同步返回结果
	err = <-src.result
	output = func(m *Message) error {
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
	return
}

func UnregisterSource(h *RtmpHandler) {

	pads.pipe <- &PadMessage{
		cmd: cmd_unregister_source,
		msg: h,
	}
	return
}

func RegisterSink(h *RtmpHandler) (err error) {

	sink := &Sink{
		RtmpHandler: h,
		result:      make(chan error),
		rtmpMsgChan: make(chan *Message, 50),
	}
	pads.pipe <- &PadMessage{
		cmd: cmd_register_sink,
		msg: sink,
	}
	// 同步返回结果
	err = <-sink.result

	return
}

func UnegisterSink(h *RtmpHandler) {
	pads.pipe <- &PadMessage{
		cmd: cmd_unregister_sink,
		msg: h,
	}
	return
}

func (p *PadPool) pairPad() {
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

func (p *PadPool) handleRegisterSource(msg *PadMessage) {

	src, ok := msg.msg.(*Source)
	if !ok {
		panic("handleRegisterSource wrong type")
	}

	key := src.appStreamKey

	_, ok = p.receivers[key]
	log.Println("handleRegisterSource:", key, ok)
	if ok {
		// TODO 存在是否抢流等？
		src.result <- fmt.Errorf("is already exits")
	} else {
		p.receivers[key] = src
		src.result <- nil
		go src.work()
	}
}

func (p *PadPool) handleUnregisterSource(msg *PadMessage) {

	h, ok := msg.msg.(*RtmpHandler)
	if !ok {
		panic("handleUnegisterSource wrong type")
	}

	key := h.appStreamKey

	_, ok = p.receivers[key]
	log.Println("handleUnregisterSource:", key, ok)
	if ok {
		h.Cancel()
		delete(p.receivers, key)
	} else {
		panic(key + " source not registerd")
	}
}

func (p *PadPool) handleRegisterSink(msg *PadMessage) {
	sink, ok := msg.msg.(*Sink)
	if !ok {
		panic("handleRegisterSource wrong type")
	}
	key := sink.appStreamKey

	src, ok := p.receivers[key]
	log.Println("handleRegisterSink:", key, ok)
	if !ok {
		// TODO 推流还不存在，等待？
		sink.result <- fmt.Errorf(key, "%s not exits", key)
	} else {
		sink.result <- nil
		src.srcChan <- msg
		go sink.work()
	}
}

func (p *PadPool) handleUnregisterSink(msg *PadMessage) {
	h, ok := msg.msg.(*RtmpHandler)
	if !ok {
		panic("handleRegisterSource wrong type")
	}
	key := h.appStreamKey

	src, ok := p.receivers[key]
	log.Println("handleUnregisterSink:", key, ok)
	if !ok {
		// 情况1:play时候还没有source, 目前play不存在source返回错误,不可能
		// 情况2:unregsrc和unregsink几乎同时，先unregsrc，的却可能出现这种情况
		log.Println(key + " source not registered(from sink)")
		h.Cancel()
	} else {
		src.srcChan <- msg
	}
}

func (s *Source) work() {
	ctx, _ := context.WithCancel(s.ctx)
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
			rtmpMsg, ok := msg.msg.(*Message)
			if !ok {
				panic("not rtmp message")
			}
			log.Println("srcwork:", rtmpMsg.Timestamp)
			s.writeMessage(rtmpMsg)
		case cmd_unregister_sink:
			s.deleteSink(msg)
		}
	}
}

func (sink *Sink) work() {
	ctx, _ := context.WithCancel(sink.ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-sink.rtmpMsgChan:
			log.Println("=======>receive message")
			sink.writeMessage(m)
		}
	}
}

func (src *Source) connectSink(msg *PadMessage) {
	log.Println("src connectSink")
	sink := msg.msg.(*Sink)
	sink.keyInSrc = sink.appStreamKey + fmt.Sprintf("%p", sink.RtmpHandler)

	if _, ok := src.sinks[sink.keyInSrc]; !ok {
		src.sinks[sink.keyInSrc] = sink
	} else {
		panic("repeater register")
	}

	avMetaData := &Message{
		MessageType: 0x12,
		Timestamp:   0,
		StreamID:    sink.functionalStreamId,
		Payload:     src.avMetaData,
	}

	amsg := &Message{
		MessageType: 8,
		Timestamp:   0,
		StreamID:    sink.functionalStreamId,
		Payload:     src.AACSequenceHeader,
	}

	vmsg := &Message{
		MessageType: 9,
		Timestamp:   0,
		StreamID:    sink.functionalStreamId,
		Payload:     src.VideoDecoderConfigurationRecord,
	}
	log.Println("=======>write metadata", len(vmsg.Payload))
	sink.rtmpMsgChan <- avMetaData
	sink.rtmpMsgChan <- vmsg
	sink.rtmpMsgChan <- amsg
}

func (src *Source) deleteSink(msg *PadMessage) {

	h := msg.msg.(*RtmpHandler)
	keyInSrc := h.appStreamKey + fmt.Sprintf("%p", h)

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

func (src *Source) writeMessage(m *Message) {
	for _, sink := range src.sinks {
		toSinkMsg := new(Message)
		toSinkMsg.MessageType = m.MessageType
		toSinkMsg.Payload = m.Payload
		toSinkMsg.Timestamp = m.Timestamp
		toSinkMsg.StreamID = sink.functionalStreamId
		err := sink.writeMessage(toSinkMsg)
		if err != nil {
			delete(src.sinks, sink.keyInSrc)
		}
	}
}

func (src *Source) cancelSink() {
	for _, sink := range src.sinks {
		sink.Cancel()
	}
}
