package envelope_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stackus/envelope"
)

type TestType interface {
	String() string
}

type Test struct {
	Test string
}

type KeyedTest struct {
	Test string
}

func (t Test) String() string {
	return t.Test
}

func (t KeyedTest) EnvelopeKey() string {
	return "test"
}

func (t KeyedTest) String() string {
	return t.Test
}

type brokenSerializer struct{}

func (brokenSerializer) Serialize(any) ([]byte, error) {
	return nil, errors.New("broken")
}

func (brokenSerializer) Deserialize(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type brokenDeserializer struct{}

func (brokenDeserializer) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (brokenDeserializer) Deserialize(data []byte, v any) error {
	return errors.New("broken")
}

func TestRegistry_Register(t *testing.T) {
	type args struct {
		v       []any
		options []envelope.RegistryOption
	}
	tests := map[string]struct {
		options []envelope.RegistryOption
		args    args
		wantErr bool
	}{
		"success": {
			options: []envelope.RegistryOption{},
			args: args{
				v: []any{
					&Test{},
				},
			},
			wantErr: false,
		},
		"keyed success": {
			options: []envelope.RegistryOption{},
			args: args{
				v: []any{
					&KeyedTest{},
				},
			},
			wantErr: false,
		},
		"multiple": {
			options: []envelope.RegistryOption{},
			args: args{
				v: []any{
					&Test{},
					&KeyedTest{},
				},
			},
			wantErr: false,
		},
		"allow no pointer": {
			options: []envelope.RegistryOption{},
			args: args{
				v: []any{
					Test{},
				},
			},
			wantErr: false,
		},
		"multiple same": {
			options: []envelope.RegistryOption{},
			args: args{
				v: []any{
					&Test{},
					Test{},
				},
			},
			wantErr: true,
		},
		"set serdes": {
			options: []envelope.RegistryOption{
				envelope.WithSerde(envelope.JsonSerde{}),
				envelope.WithEnvelopeSerde(envelope.JsonSerde{}), // use the Json one
			},
			args: args{
				v: []any{
					&Test{},
				},
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := envelope.NewRegistry(tt.options...)
			if err := r.Register(tt.args.v...); (err != nil) != tt.wantErr {
				t.Errorf("Registry.Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_RegisterFactory(t *testing.T) {
	type args struct {
		fn func() any
	}
	tests := map[string]struct {
		options []envelope.RegistryOption
		args    []func() any
		wantErr bool
	}{
		"success": {
			options: []envelope.RegistryOption{},
			args: []func() any{
				func() any {
					return &Test{}
				},
			},
			wantErr: false,
		},
		"keyed success": {
			options: []envelope.RegistryOption{},
			args: []func() any{
				func() any {
					return &KeyedTest{}
				},
			},
			wantErr: false,
		},
		"multiple": {
			options: []envelope.RegistryOption{},
			args: []func() any{
				func() any {
					return &Test{}
				},
				func() any {
					return &KeyedTest{}
				},
			},
			wantErr: false,
		},
		"multiple same": {
			options: []envelope.RegistryOption{},
			args: []func() any{
				func() any {
					return &Test{}
				},
				func() any {
					return &Test{}
				},
			},
			wantErr: true,
		},
		"not pointer": {
			options: []envelope.RegistryOption{},
			args: []func() any{
				func() any {
					return Test{}
				},
			},
			wantErr: true,
		},
		"nil": {
			options: []envelope.RegistryOption{},
			args: []func() any{
				func() any {
					return nil
				},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := envelope.NewRegistry(tt.options...)
			if err := r.RegisterFactory(tt.args...); (err != nil) != tt.wantErr {
				t.Errorf("Registry.RegisterFactory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_Serialize(t *testing.T) {
	type args struct {
		v any
	}
	tests := map[string]struct {
		registry envelope.Registry
		args     args
		wantErr  bool
	}{
		"success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				v: &Test{},
			},
			wantErr: false,
		},
		"keyed success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&KeyedTest{})
				return r
			}(),
			args: args{
				v: &KeyedTest{},
			},
			wantErr: false,
		},
		"allow no pointer": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				v: Test{},
			},
			wantErr: false,
		},
		"set serdes": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry(
					envelope.WithSerde(envelope.JsonSerde{}),
					envelope.WithEnvelopeSerde(envelope.JsonSerde{}),
				)
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				v: &Test{},
			},
			wantErr: false,
		},
		"nothing registered": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				return r
			}(),
			args: args{
				v: &Test{},
			},
			wantErr: true,
		},
		"envelope error": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry(
					envelope.WithEnvelopeSerde(brokenSerializer{}),
				)
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				v: &Test{},
			},
			wantErr: true,
		},
		"payload error": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry(
					envelope.WithSerde(brokenSerializer{}),
				)
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				v: &Test{},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := tt.registry.Serialize(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Registry.Serialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_Deserialize(t *testing.T) {
	type args struct {
		data any
	}
	tests := map[string]struct {
		registry envelope.Registry
		args     args
		wantErr  bool
	}{
		"success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				data: &Test{
					Test: "testing",
				},
			},
			wantErr: false,
		},
		"keyed success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&KeyedTest{})
				return r
			}(),
			args: args{
				data: &KeyedTest{
					Test: "testing",
				},
			},
			wantErr: false,
		},
		"allow no pointer": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				data: &Test{
					Test: "testing",
				},
			},
			wantErr: false,
		},
		"set serdes": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry(
					envelope.WithSerde(envelope.JsonSerde{}),
					envelope.WithEnvelopeSerde(envelope.JsonSerde{}),
				)
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				data: &Test{
					Test: "testing",
				},
			},
			wantErr: false,
		},
		"envelope error": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry(
					envelope.WithEnvelopeSerde(brokenDeserializer{}),
				)
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				data: &Test{
					Test: "testing",
				},
			},
			wantErr: true,
		},
		"payload error": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry(
					envelope.WithSerde(brokenDeserializer{}),
				)
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				data: &Test{
					Test: "testing",
				},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			data, _ := tt.registry.Serialize(tt.args.data)
			var dest any
			var err error
			if dest, err = tt.registry.Deserialize(data); (err != nil) != tt.wantErr {
				t.Errorf("Registry.Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if reflect.TypeOf(dest) != reflect.TypeOf(tt.args.data) {
					t.Errorf("Registry.Deserialize() = %v, want %v", reflect.TypeOf(dest), reflect.TypeOf(tt.args.data))
				}
			}
		})
	}
}

func TestRegistry_Build(t *testing.T) {
	type args struct {
		name string
	}
	tests := map[string]struct {
		registry envelope.Registry
		args     args
		wantErr  bool
	}{
		"success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				name: "envelope_test.Test",
			},
			wantErr: false,
		},
		"keyed success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&KeyedTest{})
				return r
			}(),
			args: args{
				name: "test",
			},
			wantErr: false,
		},
		"allow no pointer": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				name: "envelope_test.Test",
			},
			wantErr: false,
		},
		"set serdes": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry(
					envelope.WithSerde(envelope.JsonSerde{}),
					envelope.WithEnvelopeSerde(envelope.JsonSerde{}),
				)
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				name: "envelope_test.Test",
			},
			wantErr: false,
		},
		"not registered": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				return r
			}(),
			args: args{
				name: "envelope_test.Test",
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := tt.registry.Build(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Registry.Build() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_IsRegistered(t *testing.T) {
	type args struct {
		v any
	}
	tests := map[string]struct {
		registry envelope.Registry
		args     args
		want     bool
	}{
		"success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				v: &Test{},
			},
			want: true,
		},
		"keyed success": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&KeyedTest{})
				return r
			}(),
			args: args{
				v: &KeyedTest{},
			},
			want: true,
		},
		"allow no pointer": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				_ = r.Register(&Test{})
				return r
			}(),
			args: args{
				v: Test{},
			},
			want: true,
		},
		"not registered": {
			registry: func() envelope.Registry {
				r := envelope.NewRegistry()
				return r
			}(),
			args: args{
				v: &Test{},
			},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.registry.IsRegistered(tt.args.v); got != tt.want {
				t.Errorf("Registry.IsRegistered() = %v, want %v", got, tt.want)
			}
		})
	}
}
