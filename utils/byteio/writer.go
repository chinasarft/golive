package byteio

import (
	"encoding/binary"
	"io"
)

func PutU8(b []byte, v uint32) {
	b[0] = byte(v)
}

func PutU16BE(b []byte, v uint16) {
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func PutU16LE(b []byte, v uint32) {
	b[1] = byte(v >> 8)
	b[0] = byte(v)
}

func PutU24BE(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func PutU24LE(b []byte, v uint32) {
	b[2] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[0] = byte(v)
}

func PutU32BE(b []byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func PutU48BE(b []byte, v uint64) {
	b[0] = byte(v >> 40)
	b[1] = byte(v >> 32)
	b[2] = byte(v >> 24)
	b[3] = byte(v >> 16)
	b[4] = byte(v >> 8)
	b[5] = byte(v)
}

func PutU32LE(b []byte, v uint32) {
	b[3] = byte(v >> 24)
	b[2] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[0] = byte(v)
}

func PutU64BE(b []byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func PutU64LE(b []byte, v uint64) {
	b[7] = byte(v >> 56)
	b[6] = byte(v >> 48)
	b[5] = byte(v >> 40)
	b[4] = byte(v >> 32)
	b[3] = byte(v >> 24)
	b[2] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[0] = byte(v)
}

func WriteU8(r io.Writer, v uint32) (int, error) {
	b := make([]byte, 1)
	PutU8(b, v)
	return r.Write(b)
}

func WriteU16BE(r io.Writer, v uint16) (int, error) {
	b := make([]byte, 2)
	PutU16BE(b, v)
	return r.Write(b)
}

func WriteU16LE(r io.Writer, v uint32) (int, error) {
	b := make([]byte, 2)
	PutU16LE(b, v)
	return r.Write(b)
}

func WriteU24BE(r io.Writer, v uint32) (int, error) {
	b := make([]byte, 3)
	PutU24BE(b, v)
	return r.Write(b)
}

func WriteU24LE(r io.Writer, v uint32) (int, error) {
	b := make([]byte, 3)
	PutU24LE(b, v)
	return r.Write(b)
}

func WriteU32BE(r io.Writer, v uint32) (int, error) {
	b := make([]byte, 4)
	PutU32BE(b, v)
	return r.Write(b)
}

func WriteU32LE(r io.Writer, v uint32) (int, error) {
	b := make([]byte, 4)
	PutU32LE(b, v)
	return r.Write(b)
}

func WriteI32BE(r io.Writer, v int32) (int, error) {
	b := []byte{0, 0, 0, 0}
	PutU32BE(b, uint32(v))
	return r.Write(b)
}

func WriteU64BE(r io.Writer, v uint64) (int, error) {
	b := make([]byte, 8)
	PutU64BE(b, v)
	return r.Write(b)
}

func WriteU64LE(r io.Writer, v uint64) (int, error) {
	b := make([]byte, 8)
	PutU64LE(b, v)
	return r.Write(b)
}

func WriteFloat32BE(w io.Writer, f float32) error {
	return binary.Write(w, binary.BigEndian, f)
}

func WriteFloat32LE(w io.Writer, f float32) error {
	return binary.Write(w, binary.LittleEndian, f)
}

func WriteFloat64BE(w io.Writer, f float64) error {
	return binary.Write(w, binary.BigEndian, f)
}

func WriteFloat64LE(w io.Writer, f float64) error {
	return binary.Write(w, binary.LittleEndian, f)
}
