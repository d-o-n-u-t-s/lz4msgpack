package lz4msgpack

import (
	"encoding/binary"

	"github.com/pierrec/lz4/v4"
	"github.com/shamaton/msgpack/v2"
)

const (
	msgpackCodeExt8  byte = 0xc7
	msgpackCodeExt16 byte = 0xc8
	msgpackCodeExt32 byte = 0xc9
	msgpackCodeInt32 byte = 0xd2
	extCodeLz4       byte = 99

	offsetMessagePackCode = 0
	offsetExtSize         = 1
	offsetCodeLz4         = 5
	offsetCodeInt32       = 6
	offsetUncompressSize  = 7
	offsetLength          = 11
)

// Marshal returns bytes that is the MessagePack encoded and lz4 compressed.
func Marshal(v interface{}) ([]byte, error) {
	data, err := msgpack.Marshal(v)
	if err != nil {
		return data, err
	}
	return compress(data)
}

// MarshalAsArray returns bytes as array format that is the MessagePack encoded and lz4 compressed.
func MarshalAsArray(v interface{}) ([]byte, error) {
	data, err := msgpack.MarshalAsArray(v)
	if err != nil {
		return data, err
	}
	return compress(data)
}

// compress by lz4
func compress(data []byte) ([]byte, error) {
	buf := make([]byte, offsetLength+lz4.CompressBlockBound(len(data)))
	length, err := lz4.CompressBlockHC(data, buf[offsetLength:], 0, nil, nil)
	if err != nil || length == 0 || len(data) <= length+offsetLength {
		return data, err
	}

	buf[offsetMessagePackCode] = msgpackCodeExt32
	binary.BigEndian.PutUint32(buf[offsetExtSize:offsetCodeLz4], (uint32)(length+offsetCodeLz4))
	buf[offsetCodeLz4] = extCodeLz4
	buf[offsetCodeInt32] = msgpackCodeInt32
	binary.BigEndian.PutUint32(buf[offsetUncompressSize:offsetLength], (uint32)(len(data)))

	return buf[:offsetLength+length], err
}

// Unmarshal decodes the MessagePack-encoded data and stores the result
// in the value pointed to by v.
// In case of data compressed by lz4, it will be uncompressed before decode.
func Unmarshal(data []byte, v interface{}) error {
	return unmarshal(data, v, msgpack.Unmarshal)
}

// UnmarshalAsArray decodes the array format MessagePack-encoded data and stores the result
// in the value pointed to by v.
// In case of data compressed by lz4, it will be uncompressed before decode.
func UnmarshalAsArray(data []byte, v interface{}) error {
	return unmarshal(data, v, msgpack.UnmarshalAsArray)
}

func unmarshal(data []byte, v interface{}, unmarshaler func([]byte, interface{}) error) error {
	// see https://github.com/msgpack/msgpack/blob/master/spec.md#ext-format-family
	var typeCodeOffset byte
	switch data[offsetMessagePackCode] {
	case msgpackCodeExt8:
		typeCodeOffset = 2
	case msgpackCodeExt16:
		typeCodeOffset = 3
	case msgpackCodeExt32:
		typeCodeOffset = 5
	default:
		typeCodeOffset = 1
	}

	if data[typeCodeOffset] != extCodeLz4 {
		return unmarshaler(data, v)
	}

	// typeCode | msgpackCodeInt32 | lz4DataLength(uint32)の順で格納されている
	lz4DataLength := binary.BigEndian.Uint32(data[typeCodeOffset+2 : typeCodeOffset+6])
	buf := make([]byte, lz4DataLength)
	length, err := lz4.UncompressBlock(data[typeCodeOffset+6:], buf)
	if err != nil {
		return err
	}
	return unmarshaler(buf[:length], v)
}
