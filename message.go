package pram

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stevecallear/pram/proto/prampb"
)

type (
	// Metadata represents message metadata
	Metadata struct {
		ID            string
		Type          string
		CorrelationID string
		Timestamp     time.Time
	}

	// Message represents a message
	Message struct {
		Payload proto.Message
		Metadata
	}
)

// MessageName returns the message name with hyphen separation,
// e.g. my.package.MessageName -> my-package-MessageName
func MessageName(m proto.Message) string {
	return strings.ReplaceAll(string(m.ProtoReflect().Descriptor().FullName()), ".", "-")
}

// Marshal marshals the specified message
func Marshal(m proto.Message, optFns ...func(*Metadata)) ([]byte, error) {
	wm, err := wrap(m, optFns)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(wm)
}

// Unmarshal unmarshals the specified message
func Unmarshal(b []byte, m proto.Message) (Message, error) {
	wm := new(prampb.Message)
	err := proto.Unmarshal(b, wm)
	if err != nil {
		return Message{}, err
	}

	return unwrap(wm, m)
}

// WithCorrelationID sets the message correlation id
func WithCorrelationID(id string) func(*Metadata) {
	return func(md *Metadata) {
		md.CorrelationID = id
	}
}

func wrap(m proto.Message, optFns []func(*Metadata)) (*prampb.Message, error) {
	any, err := anypb.New(m)
	if err != nil {
		return nil, err
	}

	md := Metadata{
		ID:        uuid.NewString(),
		Type:      string(m.ProtoReflect().Descriptor().FullName()),
		Timestamp: time.Now().UTC(),
	}

	for _, opt := range optFns {
		opt(&md)
	}

	return &prampb.Message{
		Id:            md.ID,
		Type:          md.Type,
		CorrelationId: md.CorrelationID,
		Timestamp:     timestamppb.New(md.Timestamp),
		Body:          any,
	}, nil
}

func unwrap(wrapped *prampb.Message, m proto.Message) (Message, error) {
	md := Metadata{
		ID:            wrapped.GetId(),
		Type:          wrapped.GetType(),
		CorrelationID: wrapped.GetCorrelationId(),
		Timestamp:     wrapped.GetTimestamp().AsTime(),
	}

	err := wrapped.Body.UnmarshalTo(m)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Payload:  m,
		Metadata: md,
	}, nil
}
