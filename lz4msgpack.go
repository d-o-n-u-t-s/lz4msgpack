package lz4msgpack

import (
	"encoding/binary"

	"github.com/pierrec/lz4"
	"github.com/vmihailenco/msgpack"
)

const (
	msgpackCodeExt32 byte = 0xc9
	msgpackCodeInt32 byte = 0xd2
	extCodeLz4       byte = 99

	offsetCodeExt32      = 0
	offsetExtSize        = 1
	offsetCodeInt32      = 5
	offsetCodeLz4        = 6
	offsetUncompressSize = 7
	offsetLength         = 11
)

func Marshal(v ...interface{}) ([]byte, error) {
	data, err := msgpack.Marshal(v...)
	if err != nil {
		return data, err
	}
	buf := make([]byte, offsetLength+lz4.CompressBlockBound(len(data)))
	length, err := lz4.CompressBlockHC(data, buf[offsetLength:], 0)
	if length == 0 || len(data) <= length+offsetLength {
		return data, err
	}

	buf[offsetCodeExt32] = msgpackCodeExt32
	binary.BigEndian.PutUint32(buf[offsetExtSize:offsetCodeInt32], (uint32)(length+offsetCodeInt32))
	buf[offsetCodeInt32] = extCodeLz4
	buf[offsetCodeLz4] = msgpackCodeInt32
	binary.BigEndian.PutUint32(buf[offsetUncompressSize:offsetLength], (uint32)(len(data)))

	return buf[:offsetLength+length], err
}

func Unmarshal(data []byte, v ...interface{}) error {
	if data[offsetCodeExt32] != msgpackCodeExt32 || data[offsetCodeInt32] != extCodeLz4 {
		return msgpack.Unmarshal(data, v...)
	}
	buf := make([]byte, binary.BigEndian.Uint32(data[offsetUncompressSize:offsetLength]))
	_, err := lz4.UncompressBlock(data[offsetLength:], buf)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(buf, v...)
}
