package pram_test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stevecallear/pram"
	"github.com/stevecallear/pram/internal/assert"
	"github.com/stevecallear/pram/proto/testpb"
)

func TestMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input proto.Message
		mdFn  func(*pram.Metadata)
		exp   proto.Message
	}{
		{
			name:  "should marshal and unmarshal the message",
			input: &testpb.Message{Value: "value"},
			exp:   &testpb.Message{Value: "value"},
		},
		{
			name:  "should apply metadata options",
			input: &testpb.Message{Value: "value"},
			mdFn: func(md *pram.Metadata) {
				md.CorrelationID = "correlation-id"
			},
			exp: &testpb.Message{Value: "value"},
		},
	}

	for _, tt := range tests {
		exp := pram.Message{
			Payload: tt.exp,
		}

		enc, err := pram.Marshal(tt.input, func(md *pram.Metadata) {
			if tt.mdFn != nil {
				tt.mdFn(md)
			}
			exp.Metadata = *md
		})

		assert.ErrorExists(t, err, false)

		act, err := pram.Unmarshal(enc, tt.exp)

		assert.ErrorExists(t, err, false)
		assert.DeepEqual(t, act, exp)
	}
}

func TestWithCorrelationID(t *testing.T) {
	t.Run("should set the correlation id", func(t *testing.T) {
		const exp = "expected"

		md := pram.Metadata{}
		pram.WithCorrelationID(exp)(&md)

		if md.CorrelationID != exp {
			t.Errorf("got %s, expected %s", md.CorrelationID, exp)
		}
	})
}
