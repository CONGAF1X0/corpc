package serializer

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

var NotImplementProtoMessageError = errors.New("param does not implement proto.Message")

var Proto = ProtoSerializer{}

type ProtoSerializer struct {
}

func (_ ProtoSerializer) Marshal(v interface{}) ([]byte, error) {
	var body proto.Message
	if v == nil {
		return nil, NotImplementProtoMessageError
	}
	var ok bool
	if body, ok = v.(proto.Message); !ok {
		return nil, NotImplementProtoMessageError
	}
	return proto.Marshal(body)
}

func (_ ProtoSerializer) Unmarshal(data []byte, v interface{}) error {
	var body proto.Message
	if v == nil {
		return NotImplementProtoMessageError
	}
	var ok bool
	if body, ok = v.(proto.Message); !ok {
		return NotImplementProtoMessageError
	}
	return proto.Unmarshal(data, body)
}
