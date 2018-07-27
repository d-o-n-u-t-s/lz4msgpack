# Suitabled MessagePack-CSharp serializer for Golang

This serializer created for [MessagePack-CSharp](https://github.com/neuecc/MessagePack-CSharp) + LZ4 encoded data.
It is mutually compatible.

## Installation

```shell
go get -u github.com/d-o-n-u-t-s/lz4msgpack
```

## How to use

```go
func Sample() {
	type Message struct {
		Hello string
	}

	b, err := lz4msgpack.Marshal(&Message{Hello: "World"})
	if err != nil {
		panic(err)
	}

	var message Message
	err = lz4msgpack.Unmarshal(b, &message)
	if err != nil {
		panic(err)
	}
	fmt.Println(message.Hello)
}
```