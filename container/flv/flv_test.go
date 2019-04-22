package flv

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/chinasarft/golive/utils/amf"
	"github.com/chinasarft/golive/utils/byteio"
)

func TestAmfParse(t *testing.T) {
	buf := new(bytes.Buffer)

	buf.Write([]byte{0x12})

	bufObj := new(bytes.Buffer)
	obj := amf.Object{"url": "rtmp://aa.bb.cc/live/t1", "publish": true}
	_, err := amf.WriteObject(bufObj, obj)
	if err != nil {
		t.Fatalf("Object %s", err)
	}

	objData := bufObj.Bytes()
	byteio.WriteU24BE(buf, uint32(len(objData)))
	buf.Write([]byte{0, 0, 0, 0, 0, 0, 0})

	buf.Write(objData)

	byteio.WriteU32BE(buf, uint32(len(buf.Bytes())))

	scriptData := buf.Bytes()
	r := bytes.NewReader(scriptData)

	tag, err := ParseTag(r)
	if err != nil {
		t.Fatal("TestAmfParse GetNextExData fail")
	}

	fmt.Print(tag)

	if tag.TagType != FlvTagAMF0 {
		t.Fatalf("expected tagType:%d, but is:%d", FlvTagAMF0, tag.TagType)
	}

	if tag.Timestamp != 0 {
		t.Fatalf("expected timestamp:0, but is:%d", tag.Timestamp)
	}

	objReader := bytes.NewReader(tag.Data)
	objRead, err := amf.ReadObject(objReader)
	if err != nil {
		t.Fatalf("payload not AMF object:%s", err.Error())
	}

	fmt.Println(objRead["url"])
	fmt.Println(objRead["publish"])

	if v, ok := objRead["url"].(string); ok != true || v != "rtmp://aa.bb.cc/live/t1" {
		t.Errorf("expect ")
	}

	if v, ok := objRead["publish"].(bool); ok != true || v != true {
		t.Error("expect publish is true:", ok, v)
	}
}
