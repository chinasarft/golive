package rtmp

import (
	"fmt"
	"log"
)

type PutAvMessage func(m *Message) error

const (
	cmd_register_source = iota
	cmd_register_sink
	cmd_rtmp_message
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
		case cmd_register_sink:
			p.handleRegisterSink(msg)
		}
	}
}

func (p *PadPool) handleRegisterSource(msg *PadMessage) {

	src, ok := msg.msg.(*Source)
	if !ok {
		panic("handleRegisterSource wrong type")
	}

	key := src.appStreamKey
	log.Println("handleRegisterSource:", key)

	_, ok = p.receivers[key]
	if ok {
		// TODO 存在是否抢流等？
		src.result <- fmt.Errorf("is already exits")
	} else {
		p.receivers[key] = src
		src.result <- nil
		go src.work()
	}
}

func (p *PadPool) handleRegisterSink(msg *PadMessage) {
	sink, ok := msg.msg.(*Sink)
	if !ok {
		panic("handleRegisterSource wrong type")
	}
	key := sink.appStreamKey
	log.Println("handleRegisterSink:", key)

	src, ok := p.receivers[key]
	if !ok {
		// TODO 推流还不存在，等待？
		sink.result <- fmt.Errorf(key, "%s not exits", key)
	} else {
		sink.result <- nil
		src.srcChan <- msg
		go sink.work()
	}
}

func (s *Source) work() {
	for {
		var msg *PadMessage
		select {
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
		}
	}
}

func (sink *Sink) work() {
	for {
		select {
		case m := <-sink.rtmpMsgChan:
			log.Println("=======>receive message")
			sink.writeMessage(m)
		}
	}
}

func (src *Source) connectSink(msg *PadMessage) {
	log.Println("src connectSink")
	sink := msg.msg.(*Sink)
	sink.keyInSrc = sink.appStreamKey + fmt.Sprintf("%p", sink)

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
		Payload:     src.AVCDecoderConfigurationRecord,
	}
	log.Println("=======>write metadata", len(vmsg.Payload))
	sink.rtmpMsgChan <- avMetaData
	sink.rtmpMsgChan <- vmsg
	sink.rtmpMsgChan <- amsg
}

func (src *Source) writeMessage(m *Message) {
	for _, sink := range src.sinks {
		m.StreamID = sink.functionalStreamId
		err := sink.writeMessage(m)
		if err != nil {
			delete(src.sinks, sink.keyInSrc)
		}
	}
}
