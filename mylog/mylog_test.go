package mylog

import (
	"os"
	"testing"
	"time"

	"github.com/chinasarft/golive/config"
)

func TestLogStdout(t *testing.T) {
	conf := &config.LogConfig{}
	conf.Level = "info"
	conf.Target.Type = "stdout"
	if err := UpdateConfig(conf); err != nil {
		t.Fatal(err)
	}

	Debug().Msg("this ebug")
	Info().Msg("this info")
}

func TestLogFile(t *testing.T) {
	conf := &config.LogConfig{}
	conf.Level = "info"
	conf.Target.Type = "file"
	conf.Target.Name = "logtest.txt"
	conf.Position = true

	if err := UpdateConfig(conf); err != nil {
		t.Fatal(err)
	}

	Debug().Msg("this ebug")
	Info().Msg("this info")

	if _, err := os.Stat(conf.Target.Name); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 11)
}
