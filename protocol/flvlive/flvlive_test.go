package flvlive

import (
	"testing"
)

func Test_getAppStreamKey(t *testing.T) {
	key, err := getAppStreamKey("rtmp://a.b.c/app/name")
	if err != nil {
		t.Fatalf("fail:%s\n", err.Error())
	}
	if key != "app-name" {
		t.Fatalf("unpected key:%s\n", key)
	}

	key, err = getAppStreamKey("rtmp://a.b.c/app/name?token=abc")
	if err != nil {
		t.Fatalf("fail:%s\n", err.Error())
	}
	if key != "app-name" {
		t.Fatalf("unpected key:%s\n", key)
	}

	key, err = getAppStreamKey("rtmp://a.b.c/app/name?token=abc&/app2/name2")
	if err != nil {
		t.Fatalf("fail:%s\n", err.Error())
	}
	if key != "app-name" {
		t.Fatalf("unpected key:%s\n", key)
	}

	key, err = getAppStreamKey("rtmp://a.b.c/app/name?token=abc&/app2/name2&debug=true")
	if err != nil {
		t.Fatalf("fail:%s\n", err.Error())
	}
	if key != "app-name" {
		t.Fatalf("unpected key:%s\n", key)
	}
}
