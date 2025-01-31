package envelope

type RegistryOption func(*registry)

func WithSerde(serde Serde) RegistryOption {
	return func(r *registry) {
		r.serde = serde
	}
}

func WithEnvelopeSerde(serde Serde) RegistryOption {
	return func(r *registry) {
		r.envelopeSerde = serde
	}
}
