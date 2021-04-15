package lz4msgpack_test

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/d-o-n-u-t-s/lz4msgpack"
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
