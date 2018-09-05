package byteio

import (
	"encoding/binary"
	"fmt"
	"io"
)

func U8(b []byte) (i uint8) {
	return b[0]
}

func U16BE(b []byte) (i uint16) {
	i = uint16(b[0])
	i <<= 8
	i |= uint16(b[1])
	return
}

func I16BE(b []byte) (i int16) {
	i = int16(b[0])
	i <<= 8
	i |= int16(b[1])
	return
}

func I24BE(b []byte) (i int32) {
	i = int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	return
}

func U24BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	return
}

func I32BE(b []byte) (i int32) {
	i = int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	i <<= 8
	i |= int32(b[3])
	return
}

func U32LE(b []byte) (i uint32) {
	i = uint32(b[3])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[0])
	return
}

func U32BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[3])
	return
}

func U40BE(b []byte) (i uint64) {
	i = uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	return
}

func U64BE(b []byte) (i uint64) {
	i = uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	i <<= 8
	i |= uint64(b[5])
	i <<= 8
	i |= uint64(b[6])
	i <<= 8
	i |= uint64(b[7])
	return
}

func I64BE(b []byte) (i int64) {
	i = int64(int8(b[0]))
	i <<= 8
	i |= int64(b[1])
	i <<= 8
	i |= int64(b[2])
	i <<= 8
	i |= int64(b[3])
	i <<= 8
	i |= int64(b[4])
	i <<= 8
	i |= int64(b[5])
	i <<= 8
	i |= int64(b[6])
	i <<= 8
	i |= int64(b[7])
	return
}

func ReadUint32BE(r io.Reader) (uint32, error) {

	var ret uint32
	err := binary.Read(r, binary.BigEndian, &ret)
	return ret, err
}

func ReadUint32LE(r io.Reader) (uint32, error) {

	var ret uint32
	err := binary.Read(r, binary.LittleEndian, &ret)
	return ret, err
}

func ReadUint24BE(r io.Reader) (uint32, error) {

	b := make([]byte, 3)
	readLen, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	if readLen != 3 {
		return 0, fmt.Errorf("not enough data")
	}

	return uint32(b[0])*65536 + uint32(b[1])*256 + uint32(b[2]), nil
}

func ReadUint24LE(r io.Reader) (uint32, error) {

	b := make([]byte, 3)
	readLen, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	if readLen != 3 {
		return 0, fmt.Errorf("not enough data")
	}

	return uint32(b[2])*65536 + uint32(b[1])*256 + uint32(b[0]), nil
}

func ReadUint16BE(r io.Reader) (uint32, error) {

	b := make([]byte, 2)
	readLen, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	if readLen != 2 {
		return 0, fmt.Errorf("not enough data")
	}

	return uint32(b[0])*256 + uint32(b[1]), nil
}

func ReadUint16LE(r io.Reader) (uint32, error) {

	b := make([]byte, 2)
	readLen, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	if readLen != 2 {
		return 0, fmt.Errorf("not enough data")
	}

	return uint32(b[1])*256 + uint32(b[0]), nil
}

func ReadUint8(r io.Reader) (uint32, error) {
	b := make([]byte, 1)
	readLen, err := r.Read(b)
	if err != nil {
		return 0, err
	}
	if readLen != 1 {
		return 0, fmt.Errorf("not enough data")
	}

	return uint32(b[0]), nil
}
