package lz4msgpack_test

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/d-o-n-u-t-s/lz4msgpack"
	"github.com/pierrec/lz4/v4"
	"github.com/shamaton/msgpack/v2"
)

func Test(t *testing.T) {
	type Data struct {
		A int
		B int8
		C int16
		D int32
		E int64
		F uint
		G uint8
		H uint16
		I uint32
		J uint64
		// K uintptr // unsupported
		L float32
		M float64
		N []string
		O time.Time
		P []rune
		Q []byte
	}

	data := Data{
		A: 4578234323,
		B: math.MaxInt8,
		C: math.MaxInt16,
		D: math.MaxInt32,
		E: math.MaxInt64,
		F: ^uint(0),
		G: ^uint8(0),
		H: ^uint16(0),
		I: ^uint32(0),
		J: ^uint64(0),
		// K: ^uintptr(0),
		L: math.MaxFloat32,
		M: math.MaxFloat64,
		N: []string{"Hello World", "Hello World", "Hello World", "Hello World", "Hello World"},
		O: time.Date(1999, 12, 31, 7, 7, 7, 77777, time.Local),
		P: []rune("Hello World"),
		Q: []byte("Hello World"),
	}
	t.Log(data)

	tester := func(name string, marshaler func(v interface{}) ([]byte, error), unmarshaler func(data []byte, v interface{}) error) {
		b, err := marshaler(&data)
		if err != nil {
			t.Fatal("marshal", err)
		}
		t.Logf("%s: %d", name, len(b))
		var data1 Data
		if err = unmarshaler(b, &data1); err != nil {
			t.Fatal("unmarshal", err)
		}
		if !reflect.DeepEqual(data, data1) {
			t.Fatal("error", name)
		}
	}

	tester("          msgpack.Marshal", msgpack.Marshal, msgpack.Unmarshal)
	tester("   msgpack.MarshalAsArray", msgpack.MarshalAsArray, msgpack.UnmarshalAsArray)
	tester("       lz4msgpack.Marshal", lz4msgpack.Marshal, lz4msgpack.Unmarshal)
	tester("lz4msgpack.MarshalAsArray", lz4msgpack.MarshalAsArray, lz4msgpack.UnmarshalAsArray)
}

func TestExtUnmarshal(t *testing.T) {
	data := "qwertyuioppasdfghjkl;'zxcvbnm,../"
	msgpackData, err := msgpack.Marshal(data)
	if err != nil {
		t.Fatal("msgpack", err)
	}

	lz4MaxLength := lz4.CompressBlockBound(len(msgpackData))
	lz4Data := make([]byte, lz4MaxLength)
	lz4Length, _ := lz4.CompressBlockHC(msgpackData, lz4Data, 0, nil, nil)
	if err != nil {
		t.Fatal("lz4", err)
	}

	// ext8
	ext8 := []byte{0xc7, 0, 99, 0xd2}
	ext8 = binary.BigEndian.AppendUint32(ext8, uint32(lz4MaxLength))
	ext8 = append(ext8, lz4Data...)
	var ext8umarshaled string
	if err = lz4msgpack.Unmarshal(ext8[:8+lz4Length], &ext8umarshaled); err != nil {
		t.Fatal("unmarshal ext8", err)
	}
	if !reflect.DeepEqual(data, ext8umarshaled) {
		t.Fatal("error ext8")
	}

	// ext16
	ext16 := []byte{0xc8, 0, 0, 99, 0xd2}
	ext16 = binary.BigEndian.AppendUint32(ext16, uint32(lz4MaxLength))
	ext16 = append(ext16, lz4Data...)
	var ext16umarshaled string
	if err = lz4msgpack.Unmarshal(ext16[:9+lz4Length], &ext16umarshaled); err != nil {
		t.Fatal("unmarshal ext16", err)
	}
	if !reflect.DeepEqual(data, ext16umarshaled) {
		t.Fatal("error ext16")
	}

	// ext32
	ext32 := []byte{0xc9, 0, 0, 0, 0, 99, 0xd2}
	ext32 = binary.BigEndian.AppendUint32(ext32, uint32(lz4MaxLength))
	ext32 = append(ext32, lz4Data...)
	var ext32umarshaled string
	if err = lz4msgpack.Unmarshal(ext32[:11+lz4Length], &ext32umarshaled); err != nil {
		t.Fatal("unmarshal ext32", err)
	}
	if !reflect.DeepEqual(data, ext32umarshaled) {
		t.Fatal("error ext32")
	}
}
