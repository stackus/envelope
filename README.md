# Envelope

[![Go Report Card](https://goreportcard.com/badge/github.com/stackus/envelope)](https://goreportcard.com/report/github.com/stackus/envelope)
[![](https://godoc.org/github.com/stackus/envelope?status.svg)](https://pkg.go.dev/github.com/stackus/envelope)

This Go library is designed to simplify the serialization and deserialization of concrete types into interfaces.

This library is for all those times you've wanted to unmarshal some data, and all you know ahead of time is that it's an interface type.

Example:
```go
type Event interface {
	EventType() string
}

// dozens of Event implementations

var event Event
err := json.Unmarshal(data, &event)
// 💣 json: cannot unmarshal object into Go value of type Event
```

If you could know ahead of time that the data was a `UserCreated` event, you could unmarshal it into that type, but when you're working with lists of serialized data, you don't know what type you're going to get.

A pattern that provides a solution is to wrap the concrete type in an envelope that includes the type information.

```go
type Envelope struct {
	Type string
	Data json.RawMessage
}
```

This library provides an easy-to-use implementation of this pattern.

Example Usage:
```go
type Event interface {
	EventType() string
}

type UserCreated struct {
	FirstName string
	LastName  string
}

// Register the type with the registry
reg := envelope.NewRegistry()
reg.Register(UserCreated{})

// Serialize the type
userCreated := UserCreated{
	FirstName: "John",
	LastName:  "Doe",
}
data, err := reg.Serialize(userCreated)
if err != nil {
	fmt.Println(err)
	return
}

// Deserialize the type
event, err := reg.Deserialize(data)
if err != nil {
	fmt.Println(err)
	return
}

switch e := event.(type) {
case *UserCreated:
	fmt.Println(e.FirstName, e.LastName)
}
```

## Features
- Serialize and deserialize concrete types into interfaces
- Simple type registration
- Safely serialize and deserialize concrete types into databases, message queues, etc.
- Customizable serialization and deserialization

## Usage

Add the Envelope package to your project:

```bash
go get github.com/stackus/envelope@latest
```

### Create a Registry

Create a new registry and register the types you want to serialize and deserialize.

```go
reg := envelope.NewRegistry()
```

You can provide custom Serde implementations for your types and/or for the Envelope type.

```go
// the envelope.Serde interface
type Serde interface {
	Serialize(any) ([]byte, error)
	Deserialize([]byte, any) error
}

reg := envelope.NewRegistry(
	envelope.WithSerde(envelope.JsonSerde{}),
	envelope.WithEnvelopeSerde(envelope.ProtoSerde{}),
)
```
A `JsonSerde` and `ProtoSerde` are provided out of the box.

By default, the `JsonSerde` is used for the types and the `ProtoSerde` is used for the envelope.

Use your own custom serde that implements the `Serde` interface.

### Type Registration

Register the types you want to serialize and deserialize.

```go
type UserCreated struct {
	FirstName string
	LastName  string
}

// Register the type with the registry using the types reflected name
reg.Register(UserCreated{})

// Complex types that require some initialization can be registered with a factory function
reg.RegisterFactory(func() any {
	return &UserCreated{
		FirstName: "Unknown",
	}
})
```

You may also register types with a custom name by adding the following method on the type:

```go
func (UserCreated) EnvelopeKey() string {
	return "myEntity.userCreated"
}
```

### Serialize & Deserialize
With your types registered, you can now serialize and deserialize them.

```go
// UserCreated implements the Event interface
userCreated := UserCreated{
	FirstName: "John",
	LastName: "Doe",
}

// Serialize the user
data, err := reg.Serialize(userCreated)
if err != nil {
	fmt.Println(err)
	return
}

// store the data in a database, message queue, etc. then read it back in a later process

// Deserialize the event
event, err := reg.Deserialize(data)
if err != nil {
	fmt.Println(err)
	return
}
switch e := event.(type) {
case *UserCreated:
	fmt.Println(e.FirstName, e.LastName)
default:
	fmt.Printf("Unknown event type: %T\n", e)
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
