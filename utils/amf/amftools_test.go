package amf

import (
	"bytes"
	"encoding/hex"
	"testing"
)

type testT1 struct {
	capabilities int    `amf:"capabilities"`
	fmsVer       string `amf:fmsVer`
}

type testT2 struct {
	level          string `amf:"level"`
	code           string `amf:"code"`
	description    string `amf:"description"`
	objectEncoding int    `amd:"objectEncoding"`
}

func TestWriteArrayAsSiblingButElemArrayAsArray(t *testing.T) {

	var values []interface{}
	values = append(values, "_result")
	values = append(values, 1)

	obj1 := &testT1{
		capabilities: 31,
		fmsVer:       "FMS/3,0,1,123",
	}
	values = append(values, obj1)

	obj2 := &testT2{
		level:          "status",
		code:           "NetConnection.Connect.Success",
		description:    "Connection succeeded.",
		objectEncoding: 0,
	}
	values = append(values, obj2)

	data, err := WriteArrayAsSiblingButElemArrayAsArray(values)
	if err != nil {
		t.Fatalf("%s\n", err.Error())
	}

	expectStr := "0200075f726573756c74003ff000000000000003000c6361706162696c697469657300403f00000000" +
		"00000006666d7356657202000d464d532f332c302c312c3132330000090300056c6576656c0200067374617" +
		"475730004636f646502001d4e6574436f6e6e656374696f6e2e436f6e6e6563742e53756363657373000b64" +
		"65736372697074696f6e020015436f6e6e656374696f6e207375636365656465642e000e6f626a656374456" +
		"e636f64696e67000000000000000000000009"

	expect, err := hex.DecodeString(expectStr)
	if err != nil {
		t.Fatalf("hex %s\n", err)
	}

	if !bytes.Equal(data, expect) {
		t.Fatalf("not equal")
	}

}

func TestWriteArrayAsSiblingButElemArrayAsObject(t *testing.T) {
	var values []interface{}
	values = append(values, "_result")
	values = append(values, 1)

	obj1 := []interface{}{
		"fmsVer", "FMS/3,0,1,123",
		"capabilities", 31,
	}
	values = append(values, obj1)

	obj2 := []interface{}{
		"level", "status",
		"code", "NetConnection.Connect.Success",
		"description", "Connection succeeded.",
		"objectEncoding", 0,
	}
	values = append(values, obj2)

	data, err := WriteArrayAsSiblingButElemArrayAsObject(values)
	if err != nil {
		t.Fatalf("WriteArrayAsSiblingButElemArrayAsObject:%s", err.Error())
	}

	connectRespMsg := "0200075f726573756c74003ff0000000000000030006666d7356657202000d464d532f332c302c312c313233000c63617061" +
		"62696c697469657300403f0000000000000000090300056c6576656c0200067374617475730004636f646502001d4e657443" +
		"6f6e6e656374696f6e2e436f6e6e6563742e53756363657373000b6465736372697074696f6e020015436f6e6e656374696f" +
		"6e207375636365656465642e000e6f626a656374456e636f64696e67000000000000000000000009"

	connectRespMsgByte := make([]byte, len(connectRespMsg)/2)
	_, err = hex.Decode(connectRespMsgByte, []byte(connectRespMsg))
	if err != nil {
		t.Errorf("hex decode msg fail:%s", err)
	}

	if !bytes.Equal(data, connectRespMsgByte) {
		t.Fatalf("amftools.go test fail:%d %d", len(data), len(connectRespMsgByte))
	}
}
