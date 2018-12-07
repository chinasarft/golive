package amf

import (
	"bytes"
	"fmt"
)

func WriteArrayAsSiblingButElemArrayAsArray(values []interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, c := range values {
		_, err := WriteValue(buf, c)
		if err != nil {
			return nil, fmt.Errorf("WriteValue error: %s", err.Error())
		}
	}
	return buf.Bytes(), nil
}
