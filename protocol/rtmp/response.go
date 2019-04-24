package rtmp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/amf"
)

/*
connect response:
+--------------+----------+----------------------------------------+
| Field Name   |     Type |           Description                  |
+--------------+----------+----------------------------------------+
| Command Name |  String  | _result or _error; indicates whether   |
|              |          | the response is result or error.       |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  |     Transaction ID is 1 for connect    |
|      ID      |          |       responses                        |
+--------------+----------+----------------------------------------+
|  Properties  |  Object  |    Name-value pairs that describe the  |
|              |          | properties(fmsver etc.) of the         |
|              |          |      connection.                       |
+--------------+----------+----------------------------------------+
| Information  |  Object  |    Name-value pairs that describe the  |
|              |          | response from|the server. ’code’,      |
|              |          | ’level’, ’description’ are names of few|
|              |          | among such information.                |
+--------------+----------+----------------------------------------+
*/
func handleConnectResponse(m *CommandMessage) (err error) {

	var v interface{}
	r := bytes.NewReader(m.Payload)
	v, err = amf.ReadValue(r)
	if err != nil {
		return
	}
	str, ok := v.(string)
	if !ok || str != "_result" {
		err = fmt.Errorf("connecting wrong response:%s", str)
		return
	}

	v, err = amf.ReadValue(r)
	if err != nil {
		return
	}
	transId, ok := v.(float64)
	if !ok {
		err = fmt.Errorf("connecting wrong transid")
		return
	}
	if int(transId) != 1 {
		err = fmt.Errorf("connecting wrong transid:%d", int(transId))
		return
	}
	connectOk := false
	for {
		v, err = amf.ReadValue(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
		objmap, ok := v.(amf.Object)
		if !ok {
			err = fmt.Errorf("connecting wrong response")
			return
		}
		code, ok := objmap["code"]
		if ok {
			if code.(string) != "NetConnection.Connect.Success" {
				err = fmt.Errorf("connect fail:%s", v.(string))
				return
			} else {
				connectOk = true
				break
			}
		}
	}
	if !connectOk {
		err = fmt.Errorf("connect fail")
		return
	}
	return
}

/*
+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | _result or _error; indicates whether   |
|              |          | the response is result or error.       |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | ID of the command that response belongs|
| ID           |          | to.                                    |
+--------------+----------+----------------------------------------+
| Command      |  Object  | If there exists any command info this  |
| Object       |          | is set, else this is set to null type. |
+--------------+----------+----------------------------------------+
| Stream       |  Number  | The return value is either a stream ID |
| ID           |          | or an error information object.        |
+--------------+----------+----------------------------------------+

@param expectTransId 如果为0,忽略检查
*/
func handleCreateStreamResponse(m *CommandMessage, expectTransId uint32) (functionalStreamId uint32, err error) {

	var v interface{}
	r := bytes.NewReader(m.Payload)
	v, err = amf.ReadValue(r)
	if err != nil {
		return
	}
	str, ok := v.(string)
	if !ok || str != "_result" {
		err = fmt.Errorf("createstream wrong response:%s", str)
		return
	}

	v, err = amf.ReadValue(r)
	if err != nil {
		return
	}
	transId, ok := v.(float64)
	if !ok {
		err = fmt.Errorf("createstream wrong transid")
		return
	}
	if expectTransId != 0 {
		if uint32(transId) != expectTransId {
			err = fmt.Errorf("createstream transid:exp:%d real:%d", uint32(transId), expectTransId)
			return
		}
	}

	v, err = amf.ReadValue(r)
	if err != nil {
		return
	}
	if v != nil {
		err = fmt.Errorf("createstream wrong response")
		return
	}

	v, err = amf.ReadValue(r)
	if err != nil {
		return
	}
	if v != nil {
		// TODO, handle by callback?
	}

	fFunctionalStreamId, ok := v.(float64)
	if !ok {
		err = fmt.Errorf("createstream wrong functionalStreamId")
		return
	}
	functionalStreamId = uint32(fFunctionalStreamId)
	return
}

func handlePublishResponse(m *CommandMessage) error {

	r := bytes.NewReader(m.Payload)
	v, e := amf.ReadValue(r)
	if e != nil {
		return e
	}
	str, ok := v.(string)
	if !ok || str != "onStatus" {
		return fmt.Errorf("publish wrong response:%s", str)
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	transId, ok := v.(float64)
	if !ok {
		return fmt.Errorf("publish wrong transid")
	}
	if uint32(transId) != 0 {
		return fmt.Errorf("publish resp tranid must be zero:%d", uint32(transId))
	}

	v, e = amf.ReadValue(r)
	if e != nil {
		return e
	}
	if v != nil {
		return fmt.Errorf("publish response cmdobj not ni")
	}

	obj, e := amf.ReadValue(r)
	if e != nil {
		return e
	}
	objmap, ok := obj.(amf.Object)
	if !ok {
		return fmt.Errorf("connecting wrong response")
	}
	code, ok := objmap["code"]
	if ok {
		if code.(string) != "NetStream.Publish.Start" {
			return fmt.Errorf("publish fail:%s", code.(string))
		}
	} else {
		return fmt.Errorf("publish fail no code return")
	}

	return nil
}
