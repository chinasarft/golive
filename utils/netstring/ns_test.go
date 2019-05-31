package netstring

import (
	"bytes"
	"io"
	"testing"
)

type testSingleDecodeData struct {
	str string
	err error
	exp []byte
}

var testDecData = []testSingleDecodeData{
	{"0:,", nil, []byte{}},
	{"12:hello world!,", nil, []byte("hello world!")},
	{"15:hello world!123,", TOOLARGE, nil},
}

func TestSingleDecoder(t *testing.T) {
	for idx, tdata := range testDecData {
		r := bytes.NewReader([]byte(tdata.str))
		dec := NewDecoder(r, 12)

		result, err := dec.ReadNetstring()
		if err != tdata.err {
			t.Error(idx, "expect:", tdata.err, " but:", err)
		}
		if bytes.Compare(result, tdata.exp) != 0 {
			t.Error(idx, "expect:", string(tdata.exp), " but:", string(result))
		}
	}
}

type testMultiDecodeData struct {
	str string
	err []error
	exp [][]byte
}

var testMultiDecData = []testMultiDecodeData{
	{
		"0:,12:hello world!,15:hello world!123,1:1,",
		[]error{nil, nil, TOOLARGE, nil},
		[][]byte{[]byte{}, []byte("hello world!"), nil, []byte("1")},
	},
}

func TestMultiDecoder(t *testing.T) {
	for idx, tdata := range testMultiDecData {
		r := bytes.NewReader([]byte(tdata.str))
		dec := NewDecoder(r, 12)

		count := 0
		for {
			result, err := dec.ReadNetstring()
			if err == io.EOF {
				break
			}
			if err != tdata.err[count] {
				t.Error(idx, "expect:", tdata.err[count], " but:", err)
			}
			if bytes.Compare(result, tdata.exp[count]) != 0 {
				t.Error(idx, "expect:", string(tdata.exp[count]), " but:", string(result))
			}
			count++
		}
	}
}
