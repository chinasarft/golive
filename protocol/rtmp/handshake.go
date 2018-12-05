package rtmp

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/chinasarft/golive/utils/byteio"
)

var (
	hsClientFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ', 'F', 'l',
		'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ', '0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
		0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB, 0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsServerFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ', 'F', 'l',
		'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ', 'S', 'e', 'r', 'v', 'e', 'r',
		' ', '0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
		0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB, 0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsClientPartialKey = hsClientFullKey[:30]
	hsServerPartialKey = hsServerFullKey[:36]
)

var (
	schema0 = 0
	schema1 = 1
)

type keyBlock struct {
	random1   []byte
	key       []byte // 128位固定长度
	random2   []byte
	offset    []byte
	keyOffset int
}

// 包含了全部的C1或者S1的数据
// digestblock格式如意，以schema0为例子
// |time|version|keyblock|(digestblock|offset|random1|digest|random2)|
// digest的是计算digest前的数据即:|time|version|keyblock|(digestblock|offset|random1|
//             digest后的数据即:|random2)|
// 这里其实random1是 digest前的数据
//        random2是 digest后的数据
type digestBlock struct {
	random1      []byte
	digest       []byte // 32位固定长度
	random2      []byte
	digestOffset int
}

/*
func parseHandshakeKey(data []byte) (key *keyBlock, err error) {
	key = &keyBlock{}
	key.offset = data[760:764]

	pos := 0
	for i := 0; i < 4; i++ {
		pos += int(key.key[i])
	}
	// 764 - 128 - 4 //按照digest算法，并没有看到解析key的，key不用解析？
	pos = (pos % 632) + 4

	key.keyOffset = pos
	key.random1 = data[0:pos]
	key.key = data[pos : pos+128]
	key.random2 = data[pos+128 : 760]
	return
}
*/

func parseHandshakeDigest(c1 []byte, schema int) (block *digestBlock) {
	base := 8
	blockData := c1[base : 764+base]
	if schema == schema1 {
		base = 764 + 8
		blockData = c1[base:1536]
	}
	block = &digestBlock{}
	offset := blockData[0:4]

	pos := 0
	for i := 0; i < 4; i++ {
		pos += int(offset[i])
	}
	// 764 - 32 - 4
	pos = (pos % 728) + 4 + base

	block.digestOffset = pos
	block.random1 = c1[:pos]
	block.digest = c1[pos : pos+32]
	block.random2 = c1[pos+32:]
	return
}

func (digest *digestBlock) makeC1Digest() (s256 []byte) {
	h := hmac.New(sha256.New, hsClientPartialKey)

	h.Write(digest.random1)
	h.Write(digest.random2)

	return h.Sum(nil)
}

func (digest *digestBlock) makeS1Digest() {
	h := hmac.New(sha256.New, hsServerPartialKey)

	h.Write(digest.random1)
	h.Write(digest.random2)

	s256 := h.Sum(nil)
	copy(digest.digest, s256)
}

// 从代码来看，s2 digest是对C1的digest在做一个sha265作为s2 digest的key
func (digest *digestBlock) makeS2DigestKey() (s256 []byte) {
	h := hmac.New(sha256.New, hsServerFullKey)

	h.Write(digest.digest)

	return h.Sum(nil)
}

func handshake(rw io.ReadWriter) (err error) {
	var header [9]byte

	C0 := header[:1] // C0就一个字节
	C1TimeAndVersion := header[1:9]

	// 先读9字节(C0 和 C1的前8字节)
	if _, err = io.ReadFull(rw, header[0:9]); err != nil {
		return
	}

	// C1的前四个4字节是time， 后四个字节如果是复杂握手就是非0表示version，否则必须是0
	//clitime := byteio.U32BE(C1TiimeAndVersion[0:4])
	cliver := byteio.U32BE(C1TimeAndVersion[4:8])

	if cliver != 0 {
		// 复杂握手3 表示明文， 6表示密文，不支持密文
		if C0[0] == 6 {
			err = fmt.Errorf("rtmp: complex handshake not support crypto")
			return
		}
		if C0[0] != 3 {
			err = fmt.Errorf("rtmp: simple handshake version=%d invalid", C0[0])
			return
		}
		return complexHandshake(rw, C1TimeAndVersion)
	} else {
		// 简单握手3 表示版本，目前必须为3
		if C0[0] != 3 {
			err = fmt.Errorf("rtmp: simple handshake version=%d invalid", C0[0])
			return
		}
		return simpleHandshake(rw, C1TimeAndVersion)
	}
}

func complexHandshake(rw io.ReadWriter, C1TimeAndVersion []byte) (err error) {
	var buf [(1536*2)*2 + 1]byte
	C1C2 := buf[:1536*2]
	C1 := C1C2[0:1536]
	C2 := C1C2[1536 : 1536*2]

	copy(C1, C1TimeAndVersion)
	// 读取剩下的C1
	if _, err = io.ReadFull(rw, C1[8:]); err != nil {
		return
	}

	ok, block := complexHandshakeC1CheckAndDigest(C1)
	if !ok {
		err = fmt.Errorf("c1 digest check fail")
		return
	}
	fmt.Println(ok, block)
	S0S1S2 := buf[1536*2:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S2 := S0S1S2[1536+1:]

	S0[0] = 3

	createComplexS1(S1)

	s2DigestKey := block.makeS2DigestKey()
	createComplexS2(S2, s2DigestKey)

	// 发送 S0S1S2
	if _, err = rw.Write(S0S1S2); err != nil {
		return
	}

	// 读取C2
	if _, err = io.ReadFull(rw, C2); err != nil {
		return
	}
	return
}

func complexHandshakeC1CheckAndDigest(C1 []byte) (ok bool, block *digestBlock) {
	if ok, block = complexHandshakeC1CheckAndDigestSchema(C1, schema0); !ok {
		fmt.Println("try schema1")
		ok, block = complexHandshakeC1CheckAndDigestSchema(C1, schema1)
	}
	return
}

// schema0 time: 4bytes version:4bytes key:764bytes digest:764bytes
// schema1 time: 4bytes version:4bytes digest:764bytes key:764bytes
func complexHandshakeC1CheckAndDigestSchema(C1 []byte, shcema int) (ok bool, block *digestBlock) {
	block = parseHandshakeDigest(C1, shcema)
	s256 := block.makeC1Digest()

	if bytes.Compare(block.digest, s256) != 0 {
		block = nil
		return
	}
	ok = true
	return
}

func createComplexS1(s1 []byte) {

	rand.Read(s1[8:])

	byteio.PutU32BE(s1[0:4], 0)
	// 这个version并没有特殊说法，参考的livego使用的这个值
	s1ver := uint32(0x0d0e0a0d)
	byteio.PutU32BE(s1[4:8], s1ver)

	block := parseHandshakeDigest(s1, schema0)
	block.makeS1Digest()
}

func createComplexS2(s2 []byte, s2DigesKeyt []byte) {
	rand.Read(s2)
	gap := len(s2) - 32

	h := hmac.New(sha256.New, s2DigesKeyt)
	h.Write(s2[:gap])
	digest := h.Sum(nil)

	copy(s2[gap:], digest)
}

func simpleHandshake(rw io.ReadWriter, C1TimeAndVersion []byte) (err error) {
	var buf [(1536*2)*2 + 1]byte
	C1C2 := buf[:1536*2]
	C1 := C1C2[0:1536]
	C2 := C1C2[1536 : 1536*2]

	copy(C1, C1TimeAndVersion)
	// 读取剩下的C1
	if _, err = io.ReadFull(rw, C1[8:]); err != nil {
		return
	}

	S0S1S2 := buf[1536*2:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S2 := S0S1S2[1536+1:]

	S0[0] = 3
	byteio.PutU32BE(S1[0:4], 0) //S1 time设置为0
	rand.Read(S1[4:])
	copy(S2, C1)

	// 发送 S0S1S2
	if _, err = rw.Write(S0S1S2); err != nil {
		return
	}

	// 读取C2
	if _, err = io.ReadFull(rw, C2); err != nil {
		return
	}
	return

}
