package envelope

import (
	"reflect"
)

type (
	Envelope interface {
		Key() string   // Key returns the key of the envelope value
		Payload() any  // Payload returns the value of the envelope
		Bytes() []byte // Bytes returns the serialized envelope containing the key and payload
	}

	Registry interface {
		Register(vs ...any) error
		RegisterFactory(fns ...func() any) error
		Serialize(v any) (Envelope, error)
		Deserialize(data []byte) (Envelope, error)
		IsRegistered(v any) bool
		Build(key string) (any, error)
	}

	Serde interface {
		Serialize(any) ([]byte, error)
		Deserialize([]byte, any) error
	}

	envelope struct {
		key     string
		payload any
		data    []byte
	}

	registry struct {
		serde         Serde
		envelopeSerde Serde
		factories     map[string]func() any
	}
)

// NewRegistry creates a new envelope registry.
//
// The registry is used to register types that can be serialized as concrete types, then
// deserialized back into their original types without knowing ahead of time what those types are.
func NewRegistry(opts ...RegistryOption) Registry {
	r := &registry{
		factories:     make(map[string]func() any),
		serde:         JsonSerde{},
		envelopeSerde: ProtoSerde{},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Register registers one or more types with the registry.
//
// The envelope key is the fully qualified type name of the type being registered,
// or the key will be the result of calling the EnvelopeKey method on the type
// being registered.
func (r *registry) Register(vs ...any) error {
	for _, v := range vs {
		key := getKey(v)
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if err := r.register(key, func() any {
			return reflect.New(t).Interface()
		}); err != nil {
			return err
		}
	}

	return nil
}

// RegisterFactory registers one or more factory functions with the registry.
//
// The factory function should return a pointer to the type being registered.
// The envelope key is the fully qualified type name of the type being registered,
// or the key will be the result of calling the EnvelopeKey method on the type
// being registered.
func (r *registry) RegisterFactory(fns ...func() any) error {
	for _, fn := range fns {
		var v any

		if v = fn(); v == nil {
			return ErrFactoryReturnsNil("")
		}

		key := getKey(v)

		if t := reflect.TypeOf(v); t.Kind() != reflect.Ptr {
			return ErrFactoryDoesNotReturnPointer(key)
		}

		if err := r.register(key, fn); err != nil {
			return err
		}
	}

	return nil
}

// Serialize serializes a value into a byte slice safe for storage.
//
// The value must be registered with the registry before it can be serialized,
// otherwise calls will return an ErrUnregisteredKey error.
func (r *registry) Serialize(v any) (Envelope, error) {
	key := getKey(v)

	if _, exists := r.factories[key]; !exists {
		return nil, ErrUnregisteredKey(key)
	}

	data, err := r.serde.Serialize(v)
	if err != nil {
		return nil, err
	}

	msg := &EnvelopeMsg{
		Key:     &key,
		Payload: data,
	}

	data, err = r.envelopeSerde.Serialize(msg)
	if err != nil {
		return nil, err
	}

	return &envelope{
		key:     key,
		payload: v,
		data:    data,
	}, nil
}

// Deserialize deserializes a byte slice into a value.
//
// The byte slice must have been serialized using the Serialize method of the registry,
// otherwise calls will return an ErrUnregisteredKey error.
func (r *registry) Deserialize(data []byte) (Envelope, error) {
	msg := new(EnvelopeMsg)
	if err := r.envelopeSerde.Deserialize(data, msg); err != nil {
		return nil, err
	}

	key := *msg.Key
	fn, exists := r.factories[key]
	if !exists {
		return nil, ErrUnregisteredKey(key)
	}

	v := fn()
	if err := r.serde.Deserialize(msg.Payload, v); err != nil {
		return nil, err
	}

	return &envelope{
		key:     key,
		payload: v,
		data:    data,
	}, nil
}

// IsRegistered returns true if the type is registered with the registry.
func (r *registry) IsRegistered(v any) bool {
	_, exists := r.factories[getKey(v)]
	return exists
}

// Build creates a new instance of a registered type.
func (r *registry) Build(key string) (any, error) {
	fn, exists := r.factories[key]
	if !exists {
		return nil, ErrUnregisteredKey(key)
	}

	return fn(), nil
}

func (r *registry) register(key string, fn func() any) error {
	if _, exists := r.factories[key]; exists {
		return ErrReregisteredKey(key)
	}

	r.factories[key] = fn
	return nil
}

func (e *envelope) Key() string {
	return e.key
}

func (e *envelope) Payload() any {
	return e.payload
}

func (e *envelope) Bytes() []byte {
	return e.data
}

func getKey(v any) string {
	prefix := ""

	// get an optional prefix for the key
	if prefixer, ok := v.(interface{ EnvelopeKeyPrefix() string }); ok {
		prefix = prefixer.EnvelopeKeyPrefix()
	}
	// get the getKey from the envelope name
	if keyer, ok := v.(interface{ EnvelopeKey() string }); ok {
		return keyer.EnvelopeKey()
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return prefix + t.String()
}
