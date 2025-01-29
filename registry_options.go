package envelope

type RegistryOption func(*registry) error

func WithSerde(serde Serde) RegistryOption {
	return func(r *registry) error {
		r.serde = serde
		return nil
	}
}

func WithEnvelopeSerde(serde Serde) RegistryOption {
	return func(r *registry) error {
		r.envelopeSerde = serde
		return nil
	}
}
