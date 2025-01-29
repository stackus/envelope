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

func (r *registry) Register(v any) error {
	return register(r, v)
}

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
