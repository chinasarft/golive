package signaling

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"syscall"

	"github.com/chinasarft/golive/config"
	log "github.com/chinasarft/golive/mylog"
	"github.com/chinasarft/golive/utils/netstring"
)

type mediasoupWorker struct {
	stdout io.ReadCloser
	stderr io.ReadCloser
	cmd    *exec.Cmd
	ipc    *os.File
	p      *mediasoupDirector
	ss     *socketSync
}

type mediasoupDirector struct {
	Args       []string
	Path       string
	workers    []*mediasoupWorker
	conf       config.MediasoupConfig
	waitChan   chan int
	roundRobin int
}

type msWorkerResp struct {
	Accept   bool            `json:"accept"`
	Id       uint64          `json:"id"`
	Data     json.RawMessage `json:"data"`
	Event    string          `json:"event"`
	TargetId string          `json:"targetId"`
	err      error
}

var reqCounter uint64

func getGlobalReqCounter() uint64 {
	v := atomic.AddUint64(&reqCounter, 1)
	return v
}

func socketpair() (*os.File, *os.File, error) {
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, err
	}
	// 只有在程序启动后在手动关闭不需要的一个了
	// 没有办法操作cmd的closeAfterWait closeAfterStart这两个成员
	syscall.CloseOnExec(fds[0])
	syscall.CloseOnExec(fds[1])

	return os.NewFile(uintptr(fds[0]), ""), os.NewFile(uintptr(fds[1]), ""), nil
}

// 因为cmd的资源清理问题(调用wait后才清理)，所以处理的比较奇怪
func newMediasoupDirector(conf *config.MediasoupConfig) (*mediasoupDirector, error) {

	numWorker := conf.NumWorkers
	if numWorker < 1 {
		numWorker = runtime.NumCPU()
	}

	idx := 0
	pairs := make([][2]*os.File, numWorker)

	m := &mediasoupDirector{
		Path:     conf.Worker.Path,
		conf:     *conf,
		waitChan: make(chan int, numWorker),
	}

	if conf.Worker.LogLevel != "" {
		m.Args = append(m.Args, "--logLevel="+conf.Worker.LogLevel)
	}
	for i := 0; i < len(conf.Worker.LogTags); i++ {
		m.Args = append(m.Args, "--logTag="+conf.Worker.LogTags[i])
	}
	m.Args = append(m.Args, fmt.Sprint("--rtcMinPort=", conf.Worker.MinPort))
	m.Args = append(m.Args, fmt.Sprint("--rtcMaxPort=", conf.Worker.MaxPort))

	var err error

	defer func() {
		if err != nil {
			for i := idx; i < numWorker; i++ {
				pairs[i][0].Close()
				pairs[i][1].Close()
			}
			m.stop()
		}
	}()

	for idx = 0; idx < numWorker; idx++ {
		pairs[idx][0], pairs[idx][1], err = socketpair()
		if err != nil {
			return m, err
		}
	}

	for i := 0; i < numWorker; i++ {

		cmd := exec.Command(m.Path, m.Args...)
		for k, v := range conf.Worker.Env {
			log.Debug().Msg("env:" + k + "=" + v)
			cmd.Env = append(cmd.Env, k+"="+v)
		}

		woker := &mediasoupWorker{
			cmd: cmd,
			ipc: pairs[i][0],
			p:   m,
			ss:  newSocketSync(),
		}

		woker.cmd = cmd
		if woker.stdout, err = cmd.StdoutPipe(); err == nil {
			woker.stderr, err = cmd.StderrPipe()
		}

		cmd.ExtraFiles = append(cmd.ExtraFiles, pairs[i][1])

		m.workers = append(m.workers, woker)

		if startErr := woker.cmd.Start(); startErr != nil {
			log.Error().Err(startErr).Msg("cmd start fail")
		} else {
			go woker.start()
			pairs[i][1].Close()
		}

		if err != nil {
			break
		}

	}

	return m, nil
}

func (m *mediasoupDirector) stop() {
	for i := 0; i < len(m.workers); i++ {
		switch m.conf.Worker.ExitSignal {
		case "SIGINT":
			m.workers[i].cmd.Process.Signal(os.Interrupt)
		default:
			m.workers[i].cmd.Process.Kill()
		}
	}
}

func (d *mediasoupDirector) getWorker() *mediasoupWorker {
	w := d.workers[d.roundRobin]
	d.roundRobin++
	if d.roundRobin >= len(d.workers) {
		d.roundRobin = 0
	}

	return w
}

func (s *signalingServer) startMediasoupWorker(conf *config.SigalingConfig) (*mediasoupDirector, error) {
	return nil, nil
}

func (w *mediasoupWorker) start() {
	go w.readStdout()
	go w.readStderr()
	go w.handleSocketPair()
	w.cmd.Wait()
	var status syscall.WaitStatus
	var ok bool
	if status, ok = w.cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
		log.Info().Int("exitCode", status.ExitStatus()).Msg("")
	}
	w.p.waitChan <- int(status)
}

func (w *mediasoupDirector) wait() {
	num := len(w.workers)
	for num > 0 {
		<-w.waitChan
		num--
	}
}

func (w *mediasoupWorker) readStdout() {
	if w.stdout != nil {
		w.readoutput(w.stdout, "stdout")
	}
}

func (w *mediasoupWorker) readStderr() {
	if w.stderr != nil {
		w.readoutput(w.stderr, "stderr")
	}
}

func (w *mediasoupWorker) readoutput(r io.ReadCloser, name string) {
	var err error
	var line []byte
	rc := w.stdout
	if name == "stderr" {
		rc = w.stderr
	}
	lineReader := bufio.NewReader(rc)
	for {
		if line, _, err = lineReader.ReadLine(); err != nil {
			log.Error().Err(err).Str("name", name).Msg("read")
			break
		} else {
			log.Info().Str("name", name).Msg(string(line))
		}
	}
}

func (w *mediasoupWorker) handleSocketPair() {

	d := netstring.NewDecoder(w.ipc, 65543)

	for {
		cnt, err := d.ReadNetstring()
		if err != nil {
			log.Error().Err(err).Msg("handle socketpair")
			break
		}
		switch cnt[0] {
		case '{':
			fmt.Printf("=-=:[%s]\n", string(cnt))
			w.handleSocketPairMessage(cnt)
		case 'D':
			log.Debug().Msg(string(cnt[1:]))
		case 'W':
			log.Warn().Msg(string(cnt[1:]))
		case 'E':
			log.Error().Msg(string(cnt[1:]))
		default:
			log.Warn().Str("type", "unknown").Msg(string(cnt))
		}
	}
}

func (w *mediasoupWorker) handleSocketPairMessage(msgByte []byte) {
	msg := &msWorkerResp{}
	if err := json.Unmarshal(msgByte, msg); err != nil {
		log.Error().Err(err).Msg("unmarshal msg")
		return
	}
	fmt.Println("------------>", msg.Id, string(msgByte))
	if msg.Id != 0 {
		w.ss.doResponse(msg.Id, msg)
	} else if msg.Event != "" {
		switch msg.Event {
		case "running":
			log.Info().Msg("running")
		default:
			log.Warn().Str("event", msg.Event).Msg("unknown event")
		}
	}
}
