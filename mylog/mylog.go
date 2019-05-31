package mylog

import (
	"bufio"
	"io"
	"os"
	"time"

	"github.com/chinasarft/golive/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type FileTarget struct {
	file     *os.File
	filePath string
	c        chan struct{}
	w        *bufio.Writer
	l        zerolog.Logger
}

var fileTarget FileTarget

func init() {
	fileTarget.c = make(chan struct{}, 2)
	fileTarget.setOutput(os.Stdout, false)
}

func UpdateConfig(conf *config.LogConfig) error {
	switch conf.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		break
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		break
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		break
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		break
	}

	if conf.Target.Type == "stdout" {
		fileTarget.setOutput(os.Stdout, conf.Position)
	} else if conf.Target.Type == "file" {
		f, err := os.OpenFile(conf.Target.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
		if err != nil {
			return err
		}

		fileTarget.close()

		w := bufio.NewWriter(f)
		fileTarget.setOutput(w, conf.Position)

		fileTarget.start(w)
	}

	return nil
}

func (f *FileTarget) close() {

	if f.file != nil {
		f.c <- struct{}{}
		return
	}
}

func (f *FileTarget) start(w *bufio.Writer) {

	ticker := time.NewTicker(10 * time.Second)
	f.c <- struct{}{}
	go func(w *bufio.Writer) {
		flag := false
		for {
			select {
			case <-ticker.C:
				f.w.Flush()
			case <-f.c:
				if flag == false {
					f.w = w
				} else {
					f.w.Flush()
					f.file.Close()
					f.file = nil
					return
				}
			}
		}
	}(w)

	return
}

func (f *FileTarget) setOutput(w io.Writer, pos bool) {
	fileTarget.l = log.Output(w)
	if pos {
		fileTarget.l = fileTarget.l.With().Caller().Logger()
	}

}

func Debug() *zerolog.Event {
	return fileTarget.l.Debug()
}

func Info() *zerolog.Event {
	return fileTarget.l.Info()
}

func Warn() *zerolog.Event {
	return fileTarget.l.Warn()
}

func Error() *zerolog.Event {
	return fileTarget.l.Error()
}
