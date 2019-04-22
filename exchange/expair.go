package exchange

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/chinasarft/golive/utils/amf"
)

type PutData func(m *ExData) error
type Pad interface {
	OnSourceDetermined(h StreamHandler, ctx context.Context) (PutData, error)
	OnSinkDetermined(h StreamHandler, ctx context.Context) error
	OnDestroySource(h StreamHandler)
	OnDestroySink(h StreamHandler)
}

type StreamHandler interface {
	GetAppStreamKey() string
	Cancel()
	WriteData(m *ExData) error
}

const (
	cmd_register_source = iota
	cmd_register_sink
	cmd_data_message
	cmd_unregister_source
	cmd_unregister_sink
)

type PadMessage struct {
	cmd int
	msg interface{}
	ctx context.Context
}

type Source struct {
	StreamHandler
	srcChan chan *PadMessage
	result  chan error
	sinks   map[string]*Sink
	err     error

	VideoDecoderConfigurationRecord []byte // avc hevc
	AACSequenceHeader               []byte
	avMetaData                      []byte // 或者至少是在推流h265的时候是不支持的
}

type Sink struct {
	StreamHandler
	result   chan error
	msgChan  chan *ExData
	keyInSrc string
}

type ConnPool struct {
	pipe           chan *PadMessage
	receivers      map[string]*Source
	waitingSenders map[string]map[string]*PadMessage
}

var conns *ConnPool

func init() {
	conns = &ConnPool{
		pipe:           make(chan *PadMessage, 100),
		receivers:      make(map[string]*Source),
		waitingSenders: make(map[string]map[string]*PadMessage),
	}
	go conns.pairPad()
}

func GetExchanger() *ConnPool {
	return conns
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

func (cp *ConnPool) OnSourceDetermined(h StreamHandler, ctx context.Context) (PutData, error) {

	src := &Source{
		StreamHandler: h,
		result:        make(chan error),
		sinks:         make(map[string]*Sink),
		srcChan:       make(chan *PadMessage, 100),
	}
	conns.pipe <- &PadMessage{
		cmd: cmd_register_source,
		msg: src,
		ctx: ctx,
	}

	// 同步返回结果
	err := <-src.result

	output := func(m *ExData) error {
		if src.err != nil {
			return src.err
		}

		select {
		case src.srcChan <- &PadMessage{
			cmd: cmd_data_message,
			msg: m,
		}:
			//log.Println("--------->put message")
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
		p.notify(src)
	}
}

func (p *ConnPool) notify(src *Source) {
	key := src.GetAppStreamKey()
	if waitingSource, srcExits := p.waitingSenders[key]; srcExits {
		for _, regSingMsg := range waitingSource {
			src.srcChan <- regSingMsg
		}
		delete(p.waitingSenders, key)
	}
}

func (s *Source) work(ctx context.Context) {
	log.Println("source work start------->")
	ctx, _ = context.WithCancel(ctx)
	for {
		var msg *PadMessage
		select {
		case <-ctx.Done():
			s.cancelSinks()
			log.Println("source work quit------->")
			return
		case msg = <-s.srcChan:
		}

		switch msg.cmd {
		case cmd_register_sink:
			s.connectSink(msg)
		case cmd_data_message:
			rtmpMsg, ok := msg.msg.(*ExData)
			if !ok {
				panic("not rtmp message")
			}
			//log.Println("srcwork:", rtmpMsg.Timestamp)
			s.handleRtmpMessage(rtmpMsg)
		case cmd_unregister_sink:
			s.deleteSink(msg)
		}
	}
}

func (src *Source) handleRtmpMessage(m *ExData) {
	//fmt.Printf("==========>handle msg:%d %d\n", m.Payload[1], m.DataType)
	switch m.DataType {
	case DataTypeAudioConfig:
		src.AACSequenceHeader = m.Payload
	case DataTypeVideoConfig:
		src.VideoDecoderConfigurationRecord = m.Payload
	case 15:
		fallthrough
	case 18:
		src.handleDataMessaage(m)
	}
	src.writeData(m)
}

func (src *Source) writeData(m *ExData) {
	for _, sink := range src.sinks {
		toSinkMsg := new(ExData)
		toSinkMsg.DataType = m.DataType
		toSinkMsg.Payload = m.Payload
		toSinkMsg.Timestamp = m.Timestamp
		err := sink.writeData(toSinkMsg)
		if err != nil {
			// TODO cancel?
			log.Println(err)
			sink.Cancel()
			delete(src.sinks, sink.keyInSrc)
		}
	}
}

func (sink *Sink) writeData(m *ExData) error {
	select {
	case sink.msgChan <- m:
		return nil
	default:
		return fmt.Errorf("sink rtmpMsgChan full")
	}
}

func (src *Source) handleDataMessaage(m *ExData) {
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

	h, ok := msg.msg.(StreamHandler)
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

func (cp *ConnPool) OnDestroySource(h StreamHandler) {

	cp.pipe <- &PadMessage{
		cmd: cmd_unregister_source,
		msg: h,
	}

	return
}

func (cp *ConnPool) OnSinkDetermined(h StreamHandler, ctx context.Context) error {

	sink := &Sink{
		StreamHandler: h,
		result:        make(chan error),
		msgChan:       make(chan *ExData, 50),
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
		// 推流还不存在，wait
		p.addObserver(sink, msg)

		sink.result <- nil
		//sink.result <- fmt.Errorf("%s not exits", key)
	} else {
		sink.result <- nil
		src.srcChan <- msg
	}
	go sink.work(msg.ctx)
}

func (p *ConnPool) addObserver(sink *Sink, msg *PadMessage) {
	key := sink.GetAppStreamKey()
	sink.keyInSrc = key + fmt.Sprintf("%p", sink.StreamHandler)
	if waitingSource, sourceExits := p.waitingSenders[key]; sourceExits {
		if waitingSource[sink.keyInSrc] != nil {
			panic(sink.keyInSrc + " has waited")
		}
		waitingSource[sink.keyInSrc] = msg
	} else {
		waitingSource = make(map[string]*PadMessage)
		waitingSource[sink.keyInSrc] = msg
		p.waitingSenders[key] = waitingSource
	}
	return
}

func (p *ConnPool) handleUnregisterSink(msg *PadMessage) {
	h, ok := msg.msg.(StreamHandler)
	if !ok {
		panic("handleRegisterSource wrong type")
	}
	key := h.GetAppStreamKey()

	src, ok := p.receivers[key]
	log.Println("handleUnregisterSink:", key, ok)
	if !ok {
		// 情况1:play时候还没有source, 检查是否在wating
		if p.deleteObserver(h) {
			return
		}
		// 情况2:unregsrc和unregsink几乎同时，先unregsrc，的却可能出现这种情况
		log.Println(key + " source not registered(from sink)")
	} else {
		src.srcChan <- msg
	}
}

func (p *ConnPool) deleteObserver(h StreamHandler) bool {
	key := h.GetAppStreamKey()
	if waitingSource, ok := p.waitingSenders[key]; ok {
		keyInSrc := key + fmt.Sprintf("%p", h)
		delete(waitingSource, keyInSrc) // 不会报错，不用检查是否存在
		if len(waitingSource) == 0 {
			delete(p.waitingSenders, key)
		}
		return true
	}
	return false
}

func (hs *ConnPool) OnDestroySink(h StreamHandler) {
	conns.pipe <- &PadMessage{
		cmd: cmd_unregister_sink,
		msg: h,
	}
	return
}

func (sink *Sink) work(ctx context.Context) {
	log.Println("sink work start------->")
	ctx, _ = context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Println("sink work quit------->")
			return
		case m := <-sink.msgChan:
			//log.Println("=======>receive message")
			sink.WriteData(m)
		}
	}
}

func (src *Source) connectSink(msg *PadMessage) {
	log.Println("src connectSink")
	sink := msg.msg.(*Sink)
	sink.keyInSrc = sink.GetAppStreamKey() + fmt.Sprintf("%p", sink.StreamHandler)

	if _, ok := src.sinks[sink.keyInSrc]; !ok {
		src.sinks[sink.keyInSrc] = sink
	} else {
		panic("repeater register")
	}

	if src.avMetaData != nil {
		avMetaData := &ExData{
			DataType:  0x12,
			Timestamp: 0,
			Payload:   src.avMetaData,
		}
		sink.msgChan <- avMetaData
	}

	if src.AACSequenceHeader != nil {
		amsg := &ExData{
			DataType:  DataTypeAudio,
			Timestamp: 0,
			Payload:   src.AACSequenceHeader,
		}
		sink.msgChan <- amsg
		log.Println("=======>write audio metadata", len(amsg.Payload))
	}

	if src.VideoDecoderConfigurationRecord != nil {
		vmsg := &ExData{
			DataType:  DataTypeVideo,
			Timestamp: 0,
			Payload:   src.VideoDecoderConfigurationRecord,
		}
		sink.msgChan <- vmsg
		log.Println("=======>write video metadata", len(vmsg.Payload))
	}

}

func (src *Source) deleteSink(msg *PadMessage) {

	h := msg.msg.(StreamHandler)
	keyInSrc := h.GetAppStreamKey() + fmt.Sprintf("%p", h)

	sink, ok := src.sinks[keyInSrc]
	log.Println("src deleteSink", keyInSrc, ok)
	if !ok {
		// 还是有可能的，如果几乎同时关闭，先unregisterSource
		// 在unregisterSink,就会出现
		log.Println(keyInSrc + " not conntecd to source")
	} else {
		delete(src.sinks, keyInSrc)
		sink.Cancel()
	}
}

func (src *Source) cancelSinks() {
	for _, sink := range src.sinks {
		sink.Cancel()
	}
}
