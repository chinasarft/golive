package amf

import (
	"bytes"
	"fmt"
	"reflect"
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

func WriteArrayAsSiblingButElemArrayAsObject(values []interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	for _, c := range values {
		v := reflect.ValueOf(c)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			_, err := WriteObjectMarker(buf)
			if err != nil {
				return nil, err
			}
			for i := 0; i < v.Len(); i += 2 {
				_, err = WriteObjectName(buf, v.Index(i).Interface().(string))
				if err != nil {
					return nil, err
				}

				_, err = WriteValue(buf, v.Index(i+1).Interface())
				if err != nil {
					return nil, err
				}
			}
			_, err = WriteObjectEndMarker(buf)
			if err != nil {
				return nil, err
			}

		} else {

			_, err := WriteValue(buf, c)
			if err != nil {
				return nil, fmt.Errorf("WriteValue error: %s", err.Error())
			}
		}
	}
	return buf.Bytes(), nil
}
