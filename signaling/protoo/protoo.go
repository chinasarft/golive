package protoo

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Request struct {
	Request bool            `json:"request"`
	Id      int             `json:"id"`
	Method  string          `json:"method"`
	Data    json.RawMessage `json:"data"`
}

type Response struct {
	Response    bool        `json:"response"`
	Id          int         `json:"id"`
	Ok          bool        `json:"ok"`
	Data        interface{} `json:"data"`
	ErrorCode   int         `json:"errorCode,omitempty"`
	ErrorReason string      `json:"errorReason,omitempty"`
}

type Notification struct {
	Notification bool        `json:"notification"`
	Method       string      `json:"method"`
	Data         interface{} `json:"data"`
}

func (r *Request) ErrorResponse(w *websocket.Conn, reason string) error {
	resp := Response{
		Response:    true,
		Ok:          false,
		ErrorReason: reason,
	}

	b, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	return w.WriteMessage(websocket.TextMessage, b)
}

func (r *Request) GetResponseData(d interface{}) []byte {
	resp := Response{
		Response: true,
		Id:       r.Id,
		Ok:       true,
		Data:     d,
	}
	b, _ := json.Marshal(resp)

	return b
}
