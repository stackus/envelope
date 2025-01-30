package envelope

import (
	"reflect"
)

type (
	registry struct {
		serde         Serde
		envelopeSerde Serde
		registrations map[string]func() any
	}
)

// NewRegistry creates a new envelope registry.
//
// The registry is used to register types that can be serialized as concrete types, then
// deserialized back into their original types without knowing ahead of time what those types are.
func NewRegistry(opts ...RegistryOption) (Registry, error) {
	r := &registry{
		registrations: make(map[string]func() any),
		serde:         JsonSerde{},
		envelopeSerde: ProtoSerde{},
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// Register registers a type with the registry.
//
// The envelope key is the fully qualified type name of the type being registered,
// or the key will be the result of calling the EnvelopeKey method on the type
// being registered.
func (r *registry) Register(v any) error {
	return register(r, v)
}

// RegisterFactory registers a factory function with the registry.
//
// The factory function should return a pointer to the type being registered.
// The envelope key is the fully qualified type name of the type being registered,
// or the key will be the result of calling the EnvelopeKey method on the type
// being registered.
func (r *registry) RegisterFactory(fn func() any) error {
	var v any

	if v = fn(); v == nil {
		return ErrFactoryReturnsNil("")
	}

	key := getKey(v)

	if t := reflect.TypeOf(v); t.Kind() != reflect.Ptr {
		return ErrFactoryDoesNotReturnPointer(key)
	}

	return r.register(key, fn)
}

// Serialize serializes a value into a byte slice safe for storage.
//
// The value must be registered with the registry before it can be serialized,
// otherwise calls will return an ErrUnregisteredKey error.
func (r *registry) Serialize(v any) ([]byte, error) {
	key := getKey(v)

	if _, exists := r.registrations[key]; !exists {
		return nil, ErrUnregisteredKey(key)
	}

	data, err := r.serde.Serialize(v)
	if err != nil {
		return nil, err
	}

	envelope := &Envelope{
		Key:     &key,
		Payload: data,
	}

	return r.envelopeSerde.Serialize(envelope)
}

// Deserialize deserializes a byte slice into a value.
//
// The byte slice must have been serialized using the Serialize method of the registry,
// otherwise calls will return an ErrUnregisteredKey error.
func (r *registry) Deserialize(data []byte) (any, error) {
	envelope := new(Envelope)
	if err := r.envelopeSerde.Deserialize(data, envelope); err != nil {
		return nil, err
	}

	key := *envelope.Key
	fn, exists := r.registrations[key]
	if !exists {
		return nil, ErrUnregisteredKey(key)
	}

	v := fn()
	if err := r.serde.Deserialize(envelope.Payload, v); err != nil {
		return nil, err
	}

	return v, nil
}

// IsRegistered returns true if the type is registered with the registry.
func (r *registry) IsRegistered(v any) bool {
	_, exists := r.registrations[getKey(v)]
	return exists
}

// Build creates a new instance of a registered type.
func (r *registry) Build(key string) (any, error) {
	fn, exists := r.registrations[key]
	if !exists {
		return nil, ErrUnregisteredKey(key)
	}

	return fn(), nil
}

func (r *registry) register(key string, fn func() any) error {
	if _, exists := r.registrations[key]; exists {
		return ErrReregisteredKey(key)
	}

	r.registrations[key] = fn
	return nil
}

func register[T any](reg *registry, v T) error {
	key := getKey(v)
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return reg.register(key, func() any {
			return reflect.New(t.Elem()).Interface()
		})
	}

	return reg.register(key, func() any {
		return new(T)
	})
}

func getKey(v any) string {
	// get the getKey from the envelope name
	if keyer, ok := v.(interface{ EnvelopeKey() string }); ok {
		return keyer.EnvelopeKey()
	}

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}
