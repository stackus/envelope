package envelope_test

import (
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

func TestRegistry_Register(t *testing.T) {
	type args struct {
		v       any
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
				v: &Test{},
			},
			wantErr: false,
		},
		"keyed success": {
			options: []envelope.RegistryOption{},
			args: args{
				v: &KeyedTest{},
			},
			wantErr: false,
		},
		"allow no pointer": {
			options: []envelope.RegistryOption{},
			args: args{
				v: Test{},
			},
			wantErr: false,
		},
		"set serdes": {
			options: []envelope.RegistryOption{
				envelope.WithSerde(envelope.JsonSerde{}),
				envelope.WithEnvelopeSerde(envelope.JsonSerde{}), // use the Json one
			},
			args: args{
				v: &Test{},
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r, _ := envelope.NewRegistry(tt.options...)
			if err := r.Register(tt.args.v); (err != nil) != tt.wantErr {
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
		args    args
		wantErr bool
	}{
		"success": {
			options: []envelope.RegistryOption{},
			args: args{
				fn: func() any {
					return &Test{}
				},
			},
			wantErr: false,
		},
		"not pointer": {
			options: []envelope.RegistryOption{},
			args: args{
				fn: func() any {
					return Test{}
				},
			},
			wantErr: true,
		},
		"nil": {
			options: []envelope.RegistryOption{},
			args: args{
				fn: func() any {
					return nil
				},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r, _ := envelope.NewRegistry(tt.options...)
			if err := r.RegisterFactory(tt.args.fn); (err != nil) != tt.wantErr {
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
				r, _ := envelope.NewRegistry()
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
				r, _ := envelope.NewRegistry()
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
				r, _ := envelope.NewRegistry()
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
				r, _ := envelope.NewRegistry(
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
				r, _ := envelope.NewRegistry()
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
				r, _ := envelope.NewRegistry()
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
				r, _ := envelope.NewRegistry()
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
				r, _ := envelope.NewRegistry()
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
				r, _ := envelope.NewRegistry(
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
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			data, _ := tt.registry.Serialize(tt.args.data)
			var dest any
			var err error
			if dest, err = tt.registry.Deserialize(data); (err != nil) != tt.wantErr {
				t.Errorf("Registry.Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if reflect.TypeOf(dest) != reflect.TypeOf(tt.args.data) {
					t.Errorf("Registry.Deserialize() = %v, want %v", reflect.TypeOf(dest), reflect.TypeOf(tt.args.data))
				}
			}
		})
	}
}
