package envelope

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"
)

type JsonSerde struct{}

func (s JsonSerde) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (s JsonSerde) Deserialize(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type ProtoSerde struct{}

func (s ProtoSerde) Serialize(v any) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (s ProtoSerde) Deserialize(data []byte, v any) error {
	return proto.Unmarshal(data, v.(proto.Message))
}
