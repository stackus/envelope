package envelope

type (
	Registry interface {
		Register(v any) error
		RegisterFactory(fn func() any) error
		Serialize(v any) ([]byte, error)
		Deserialize(data []byte) (any, error)
	}

	Serde interface {
		Serialize(any) ([]byte, error)
		Deserialize([]byte, any) error
	}
)
