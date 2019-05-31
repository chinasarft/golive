package signaling

import (
	"fmt"
	"testing"

	"github.com/chinasarft/golive/config"
)

// 这个测试还是需要人肉判断
func TestStartMediasoupWorker(t *testing.T) {
	conf := &config.MediasoupWorkerConfig{
		Path:       "/Users/liuye/go/src/github.com/chinasarft/golive/signaling/child",
		MinPort:    49999,
		MaxPort:    59999,
		LogLevel:   "debug",
		LogTag:     []string{"info", "ice", "dtls", "rtp", "srtp", "rtcp"},
		Env:        map[string]string{"version": "3.0.0"},
		ExitSignal: "SIGINT",
	}
	m, err := newMediasoupDirector(conf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(m)

	for i := 0; i < len(m.workers); i++ {
		m.workers[i].ipc.Write([]byte(fmt.Sprintf("from go:%d\n", i)))
	}
	m.wait()
}
