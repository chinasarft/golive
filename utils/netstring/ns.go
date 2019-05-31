package netstring

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
)

const (
	lengthDelim byte = ':'
	dataDelim   byte = ','
)

var TOOLARGE = errors.New("LARGE")
var NOTEND = errors.New("NOTEND")
var LENERR = errors.New("LENERR")

type Decoder struct {
	r      *bufio.Reader
	maxLen int
	buf    []byte
}

func Encode(data []byte) []byte {

	var buffer bytes.Buffer
	length := strconv.FormatInt(int64(len(data)), 10)
	buffer.WriteString(length)
	buffer.WriteByte(':')
	buffer.Write(data)
	buffer.WriteByte(',')
	return buffer.Bytes()
}

func NewDecoder(r io.Reader, maxLen int) *Decoder {
	return &Decoder{
		r:      bufio.NewReader(r),
		maxLen: maxLen,
		buf:    make([]byte, maxLen),
	}
}

func (d *Decoder) ReadNetstring() ([]byte, error) {

	length, err := d.r.ReadBytes(lengthDelim)
	if err != nil {
		return nil, err
	}

	if len(length) > 6 {
		return nil, LENERR
	}

	l, err := strconv.Atoi(strings.TrimSuffix(string(length), string(lengthDelim)))
	if err != nil {
		return nil, err
	}

	if l > d.maxLen {
		rLen := d.maxLen
		for l > 0 {
			if _, err = io.ReadFull(d.r, d.buf[0:rLen]); err != nil {
				return nil, err
			}
			l -= rLen
			if l > d.maxLen {
				rLen = d.maxLen
			} else {
				rLen = l
			}
		}
		if err = d.readDataDelim(); err != nil {
			return nil, err
		}
		return nil, TOOLARGE
	}

	if l == 0 {
		if err = d.readDataDelim(); err != nil {
			return nil, err
		}
		return d.buf[0:0], nil
	}

	_, err = io.ReadFull(d.r, d.buf[0:l])
	if err != nil {
		return nil, err
	}

	if err = d.readDataDelim(); err != nil {
		return nil, err
	}

	return d.buf[0:l], nil
}

func (d *Decoder) readDataDelim() error {
	next, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	if next != dataDelim {
		d.r.UnreadByte()
		return NOTEND
	}
	return nil
}
