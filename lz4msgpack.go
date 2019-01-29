package lz4msgpack

import (
	"encoding/binary"

	"github.com/pierrec/lz4"
	"github.com/shamaton/msgpack"
)

const (
	msgpackCodeExt32 byte = 0xc9
	msgpackCodeInt32 byte = 0xd2
	extCodeLz4       byte = 99

	offsetCodeExt32      = 0
	offsetExtSize        = 1
	offsetCodeLz4        = 5
	offsetCodeInt32      = 6
	offsetUncompressSize = 7
	offsetLength         = 11
)

// Marshal returns bytes that is the MessagePack encoded and lz4 compressed.
func Marshal(v interface{}) ([]byte, error) {
	return marshal(v, msgpack.Encode)
}

// MarshalAsArray returns bytes as array format that is the MessagePack encoded and lz4 compressed.
func MarshalAsArray(v interface{}) ([]byte, error) {
	return marshal(v, msgpack.EncodeStructAsArray)
}

func marshal(v interface{}, marshaler func(interface{}) ([]byte, error)) ([]byte, error) {
	data, err := marshaler(v)
	if err != nil {
		return data, err
	}
	return compress(data)
}

// compress by lz4
func compress(data []byte) ([]byte, error) {
	buf := make([]byte, offsetLength+lz4.CompressBlockBound(len(data)))
	length, err := lz4.CompressBlockHC(data, buf[offsetLength:], 0)
	if err != nil || length == 0 || len(data) <= length+offsetLength {
		return data, err
	}

	buf[offsetCodeExt32] = msgpackCodeExt32
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
	return unmarshal(data, v, msgpack.Decode)
}

// UnmarshalAsArray decodes the array format MessagePack-encoded data and stores the result
// in the value pointed to by v.
// In case of data compressed by lz4, it will be uncompressed before decode.
func UnmarshalAsArray(data []byte, v interface{}) error {
	return unmarshal(data, v, msgpack.DecodeStructAsArray)
}

func unmarshal(data []byte, v interface{}, unmarshaler func([]byte, interface{}) error) error {
	if data[offsetCodeExt32] != msgpackCodeExt32 || data[offsetCodeLz4] != extCodeLz4 {
		return unmarshaler(data, v)
	}
	buf := make([]byte, binary.BigEndian.Uint32(data[offsetUncompressSize:offsetLength]))
	_, err := lz4.UncompressBlock(data[offsetLength:], buf)
	if err != nil {
		return err
	}
	return unmarshaler(buf, v)
}
