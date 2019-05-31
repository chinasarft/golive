package signaling

import (
	"errors"
	"io"
	"sync"
	"time"

	log "github.com/chinasarft/golive/mylog"
	"github.com/chinasarft/golive/utils/netstring"
)

type socketSync struct {
	mutex     sync.Mutex
	respMap   map[uint64]chan *msWorkerResp
	chanPool  sync.Pool
	timerPool sync.Pool
}

func newSocketSync() *socketSync {
	return &socketSync{
		chanPool: sync.Pool{
			New: func() interface{} {
				return make(chan *msWorkerResp)
			},
		},
		timerPool: sync.Pool{
			New: func() interface{} {
				return time.NewTimer(time.Second * 10)
			},
		},
		respMap: make(map[uint64]chan *msWorkerResp),
	}
}

func (s *socketSync) doRequest(w io.Writer, id uint64, str string) (*msWorkerResp, error) {
	s.mutex.Lock()
	c := s.chanPool.Get().(chan *msWorkerResp)
	s.respMap[id] = c
	s.mutex.Unlock()

	reqByte := netstring.Encode([]byte(str))
	if _, err := w.Write(reqByte); err != nil {
		s.mutex.Lock()
		if _, ok := s.respMap[id]; ok {
			delete(s.respMap, id)
		}
		s.mutex.Unlock()
		return nil, err
	}

	t := s.timerPool.Get().(*time.Timer)
	t.Reset(time.Second * 10)
	// 防止之前的数据，因为为了重用并没有stop timer
OLDDATA:
	for {
		select {
		case <-t.C:
		default:
			break OLDDATA
		}
	}

	var msg *msWorkerResp = nil
	select {
	case <-t.C:
		s.mutex.Lock()
		select {
		case msg = <-c:
		default:
			msg = &msWorkerResp{
				err: errors.New("reqtimeout"),
			}
		}
		if _, ok := s.respMap[id]; ok {
			delete(s.respMap, id)
		}
		s.mutex.Unlock()
	case msg = <-c:
	}

	s.chanPool.Put(c)
	s.timerPool.Put(t)
	return msg, nil
}

// TODO 除了accept:true之外是否还有其它错误返回，err可能不是nil
func (s *socketSync) doResponse(id uint64, msg *msWorkerResp) {
	s.mutex.Lock()
	defer func() {
		s.mutex.Unlock()
	}()
	if c, ok := s.respMap[id]; ok {
		c <- msg
		delete(s.respMap, id)
	} else {
		log.Warn().Uint64("id", id).Msg("resptimeout")
	}
}
